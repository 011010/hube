package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/husari/hube/internal/domain/note"
	"github.com/jmoiron/sqlx"
)

type NoteRepo struct{ db *sqlx.DB }

func NewNoteRepo(db *sqlx.DB) *NoteRepo { return &NoteRepo{db: db} }

type noteRow struct {
	ID        string       `db:"id"`
	Title     string       `db:"title"`
	Content   string       `db:"content"`
	Blocks    string       `db:"blocks"`
	Status    string       `db:"status"`
	Priority  string       `db:"priority"`
	DueDate   *string      `db:"due_date"`
	FolderID  *string      `db:"folder_id"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
}

func walkNode(node map[string]any, out *[]string) {
	if text, ok := node["text"].(string); ok {
		*out = append(*out, text)
	}
	if children, ok := node["content"].([]any); ok {
		for _, child := range children {
			if childMap, ok := child.(map[string]any); ok {
				walkNode(childMap, out)
			}
		}
	}
}

func blocksToText(blocks string) (string, error) {
	if blocks == "" {
		return "", nil
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(blocks), &doc); err != nil {
		return "", err
	}
	var parts []string
	walkNode(doc, &parts)
	return strings.Join(parts, " "), nil
}

func (r *NoteRepo) loadTags(ctx context.Context, noteID string) ([]string, error) {
	tags := make([]string, 0)
	return tags, r.db.SelectContext(ctx, &tags, `SELECT tag FROM note_tags WHERE note_id = ? ORDER BY tag`, noteID)
}

func (r *NoteRepo) toNote(ctx context.Context, row noteRow) (note.Note, error) {
	tags, err := r.loadTags(ctx, row.ID)
	if err != nil {
		return note.Note{}, err
	}
	return note.Note{
		ID:        row.ID,
		Title:     row.Title,
		Content:   row.Content,
		Blocks:    row.Blocks,
		Status:    row.Status,
		Priority:  row.Priority,
		DueDate:   row.DueDate,
		FolderID:  row.FolderID,
		Tags:      tags,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *NoteRepo) FindAll(ctx context.Context, folderID *string) ([]note.Note, error) {
	rows := make([]noteRow, 0)
	var err error
	if folderID == nil {
		err = r.db.SelectContext(ctx, &rows, `SELECT id,title,content,blocks,status,priority,due_date,folder_id,created_at,updated_at FROM notes ORDER BY updated_at DESC`)
	} else {
		err = r.db.SelectContext(ctx, &rows, `SELECT id,title,content,blocks,status,priority,due_date,folder_id,created_at,updated_at FROM notes WHERE folder_id = ? ORDER BY updated_at DESC`, *folderID)
	}
	if err != nil {
		return nil, err
	}
	notes := make([]note.Note, 0, len(rows))
	for _, row := range rows {
		n, err := r.toNote(ctx, row)
		if err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, nil
}

func (r *NoteRepo) FindByID(ctx context.Context, id string) (*note.Note, error) {
	var row noteRow
	err := r.db.GetContext(ctx, &row, `SELECT id,title,content,blocks,status,priority,due_date,folder_id,created_at,updated_at FROM notes WHERE id = ?`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	n, err := r.toNote(ctx, row)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *NoteRepo) Search(ctx context.Context, query string) ([]note.Note, error) {
	if len(query) > 500 {
		query = query[:500]
	}
	rows := make([]noteRow, 0)
	err := r.db.SelectContext(ctx, &rows, `
		SELECT n.id, n.title, n.content, n.blocks, n.status, n.priority, n.due_date, n.folder_id, n.created_at, n.updated_at
		FROM notes n
		JOIN notes_fts ON notes_fts.rowid = n.rowid
		WHERE notes_fts MATCH ?
		ORDER BY rank
	`, query+"*")
	if err != nil {
		// FTS syntax errors are user input errors, not server faults.
		return []note.Note{}, nil
	}
	notes := make([]note.Note, 0, len(rows))
	for _, row := range rows {
		n, err := r.toNote(ctx, row)
		if err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, nil
}

func (r *NoteRepo) Create(ctx context.Context, n *note.Note) error {
	n.ID = uuid.NewString()
	if n.Blocks != "" {
		content, err := blocksToText(n.Blocks)
		if err != nil {
			return fmt.Errorf("blocksToText: %w", err)
		}
		n.Content = content
	}
	now := time.Now()
	n.CreatedAt = now
	n.UpdatedAt = now
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO notes (id,title,content,blocks,status,priority,due_date,folder_id,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		n.ID, n.Title, n.Content, n.Blocks, n.Status, n.Priority, n.DueDate, n.FolderID, n.CreatedAt, n.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return r.SetTags(ctx, n.ID, n.Tags)
}

func (r *NoteRepo) Update(ctx context.Context, n *note.Note) error {
	if n.Blocks != "" {
		content, err := blocksToText(n.Blocks)
		if err != nil {
			return fmt.Errorf("blocksToText: %w", err)
		}
		n.Content = content
	}
	n.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE notes SET title=?,content=?,blocks=?,status=?,priority=?,due_date=?,folder_id=?,updated_at=? WHERE id=?`,
		n.Title, n.Content, n.Blocks, n.Status, n.Priority, n.DueDate, n.FolderID, n.UpdatedAt, n.ID,
	)
	if err != nil {
		return err
	}
	return r.SetTags(ctx, n.ID, n.Tags)
}

func (r *NoteRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM notes WHERE id = ?`, id)
	return err
}

func (r *NoteRepo) StoreEmbedding(ctx context.Context, noteID string, embedding []byte) error {
	_, err := r.db.ExecContext(ctx, `UPDATE notes SET embedding=? WHERE id=?`, string(embedding), noteID)
	return err
}

type noteEmbeddingRow struct {
	ID        string `db:"id"`
	Embedding string `db:"embedding"`
}

func (r *NoteRepo) FindAllEmbeddings(ctx context.Context) ([]note.EmbeddingRecord, error) {
	rows := make([]noteEmbeddingRow, 0)
	if err := r.db.SelectContext(ctx, &rows, `SELECT id, embedding FROM notes WHERE embedding != ''`); err != nil {
		return nil, err
	}
	result := make([]note.EmbeddingRecord, len(rows))
	for i, row := range rows {
		result[i] = note.EmbeddingRecord{ID: row.ID, Embedding: []byte(row.Embedding)}
	}
	return result, nil
}

func (r *NoteRepo) SetTags(ctx context.Context, noteID string, tags []string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM note_tags WHERE note_id = ?`, noteID); err != nil {
		return fmt.Errorf("clear tags: %w", err)
	}
	for _, tag := range tags {
		if _, err := r.db.ExecContext(ctx, `INSERT OR IGNORE INTO note_tags(note_id,tag) VALUES(?,?)`, noteID, tag); err != nil {
			return fmt.Errorf("insert tag: %w", err)
		}
	}
	return nil
}
