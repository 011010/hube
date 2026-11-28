package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/husari/hube/internal/domain/note"
	"github.com/jmoiron/sqlx"
)

type NoteRepo struct{ db *sqlx.DB }

func NewNoteRepo(db *sqlx.DB) *NoteRepo { return &NoteRepo{db: db} }

type noteRow struct {
	ID        string    `db:"id"`
	Title     string    `db:"title"`
	Content   string    `db:"content"`
	Blocks    string    `db:"blocks"`
	Status    string    `db:"status"`
	FolderID  *string   `db:"folder_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
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

var linkPattern = regexp.MustCompile(`\[\[(.+?)\]\]`)

// extractLinks parses `[[Title]]` references out of a note's raw blocks
// string, returning the referenced titles in first-seen order with
// duplicates removed.
func extractLinks(blocks string) []string {
	if blocks == "" {
		return nil
	}
	matches := linkPattern.FindAllStringSubmatch(blocks, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(matches))
	titles := make([]string, 0, len(matches))
	for _, m := range matches {
		title := strings.TrimSpace(m[1])
		if title == "" {
			continue
		}
		if _, ok := seen[title]; ok {
			continue
		}
		seen[title] = struct{}{}
		titles = append(titles, title)
	}
	return titles
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

type propertyRow struct {
	NoteID string `db:"note_id"`
	Key    string `db:"key"`
	Type   string `db:"type"`
	Value  string `db:"value"`
}

func (r *NoteRepo) loadProperties(ctx context.Context, noteID string) ([]note.Property, error) {
	rows := make([]propertyRow, 0)
	err := r.db.SelectContext(ctx, &rows,
		`SELECT note_id,key,type,value FROM note_properties WHERE note_id = ? ORDER BY key`, noteID)
	if err != nil {
		return nil, err
	}
	return decodeProperties(rows)
}

func decodeProperties(rows []propertyRow) ([]note.Property, error) {
	props := make([]note.Property, 0, len(rows))
	for _, row := range rows {
		var value any
		if err := json.Unmarshal([]byte(row.Value), &value); err != nil {
			return nil, fmt.Errorf("decode property %s: %w", row.Key, err)
		}
		props = append(props, note.Property{Key: row.Key, Type: row.Type, Value: value})
	}
	return props, nil
}

const propertiesBatchSize = 500

func (r *NoteRepo) loadPropertiesBatch(ctx context.Context, ids []string) (map[string][]note.Property, error) {
	out := make(map[string][]note.Property, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	for start := 0; start < len(ids); start += propertiesBatchSize {
		end := start + propertiesBatchSize
		if end > len(ids) {
			end = len(ids)
		}
		query, args, err := sqlx.In(
			`SELECT note_id,key,type,value FROM note_properties WHERE note_id IN (?) ORDER BY note_id, key`, ids[start:end])
		if err != nil {
			return nil, err
		}
		rows := make([]propertyRow, 0)
		if err := r.db.SelectContext(ctx, &rows, r.db.Rebind(query), args...); err != nil {
			return nil, err
		}
		grouped := make(map[string][]propertyRow)
		for _, row := range rows {
			grouped[row.NoteID] = append(grouped[row.NoteID], row)
		}
		for id, group := range grouped {
			props, err := decodeProperties(group)
			if err != nil {
				return nil, err
			}
			out[id] = props
		}
	}
	return out, nil
}

// withProperties assigns props to n, guaranteeing a non-nil slice so JSON
// responses marshal an empty array instead of null.
func withProperties(n *note.Note, props []note.Property) {
	if props == nil {
		props = []note.Property{}
	}
	n.Properties = props
}

// SetProperties replaces a note's properties and bumps the note's updated_at
// in the same transaction. A missing note yields an error wrapping
// sql.ErrNoRows, so a caller that raced with a delete can answer 404.
func (r *NoteRepo) SetProperties(ctx context.Context, noteID string, props []note.Property) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin properties tx: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `UPDATE notes SET updated_at = ? WHERE id = ?`, time.Now(), noteID)
	if err != nil {
		return fmt.Errorf("touch note: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("touch note rows: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("set properties: %w", sql.ErrNoRows)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM note_properties WHERE note_id = ?`, noteID); err != nil {
		return fmt.Errorf("clear properties: %w", err)
	}
	for _, p := range props {
		encoded, err := json.Marshal(p.Value)
		if err != nil {
			return fmt.Errorf("encode property %s: %w", p.Key, err)
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO note_properties (note_id,key,type,value) VALUES (?,?,?,?)`,
			noteID, p.Key, p.Type, string(encoded)); err != nil {
			return fmt.Errorf("insert property %q: %w", p.Key, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit properties: %w", err)
	}
	return nil
}

// FindNoteIDByProperty returns the id of the first note holding the exact
// property value, or "" when none does. value is the JSON-encoded stored form.
func (r *NoteRepo) FindNoteIDByProperty(ctx context.Context, key, value string) (string, error) {
	var noteID string
	err := r.db.GetContext(ctx, &noteID,
		`SELECT note_id FROM note_properties WHERE key = ? AND value = ? ORDER BY note_id LIMIT 1`,
		key, value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("find note by property %q: %w", key, err)
	}
	return noteID, nil
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
		FolderID:  row.FolderID,
		Tags:      tags,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *NoteRepo) attachProperties(ctx context.Context, notes []note.Note) error {
	if len(notes) == 0 {
		return nil
	}
	ids := make([]string, len(notes))
	for i := range notes {
		ids[i] = notes[i].ID
	}
	propsByNote, err := r.loadPropertiesBatch(ctx, ids)
	if err != nil {
		return err
	}
	for i := range notes {
		withProperties(&notes[i], propsByNote[notes[i].ID])
	}
	return nil
}

func (r *NoteRepo) FindAll(ctx context.Context, folderID *string) ([]note.Note, error) {
	rows := make([]noteRow, 0)
	var err error
	if folderID == nil {
		err = r.db.SelectContext(ctx, &rows, `SELECT id,title,content,blocks,status,folder_id,created_at,updated_at FROM notes ORDER BY updated_at DESC`)
	} else {
		err = r.db.SelectContext(ctx, &rows, `SELECT id,title,content,blocks,status,folder_id,created_at,updated_at FROM notes WHERE folder_id = ? ORDER BY updated_at DESC`, *folderID)
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
	if err := r.attachProperties(ctx, notes); err != nil {
		return nil, err
	}
	return notes, nil
}

func (r *NoteRepo) FindByID(ctx context.Context, id string) (*note.Note, error) {
	var row noteRow
	err := r.db.GetContext(ctx, &row, `SELECT id,title,content,blocks,status,folder_id,created_at,updated_at FROM notes WHERE id = ?`, id)
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
	props, err := r.loadProperties(ctx, id)
	if err != nil {
		return nil, err
	}
	withProperties(&n, props)
	return &n, nil
}

func (r *NoteRepo) Search(ctx context.Context, query string) ([]note.Note, error) {
	if len(query) > 500 {
		query = query[:500]
	}
	rows := make([]noteRow, 0)
	err := r.db.SelectContext(ctx, &rows, `
		SELECT n.id, n.title, n.content, n.blocks, n.status, n.folder_id, n.created_at, n.updated_at
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
	if err := r.attachProperties(ctx, notes); err != nil {
		return nil, err
	}
	return notes, nil
}

func (r *NoteRepo) Create(ctx context.Context, n *note.Note) error {
	n.ID = uuid.NewString()
	if n.Blocks != "" {
		content, err := blocksToText(n.Blocks)
		if err != nil {
			return &note.ValidationError{Field: "blocks", Message: err.Error()}
		}
		n.Content = content
	}
	now := time.Now()
	n.CreatedAt = now
	n.UpdatedAt = now
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO notes (id,title,content,blocks,status,folder_id,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?)`,
		n.ID, n.Title, n.Content, n.Blocks, n.Status, n.FolderID, n.CreatedAt, n.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if err := r.storeLinks(ctx, n.ID, n.Blocks); err != nil {
		return fmt.Errorf("store links: %w", err)
	}
	return r.SetTags(ctx, n.ID, n.Tags)
}

func (r *NoteRepo) Update(ctx context.Context, n *note.Note) error {
	if n.Blocks != "" {
		content, err := blocksToText(n.Blocks)
		if err != nil {
			return &note.ValidationError{Field: "blocks", Message: err.Error()}
		}
		n.Content = content
	}
	n.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE notes SET title=?,content=?,blocks=?,status=?,folder_id=?,updated_at=? WHERE id=?`,
		n.Title, n.Content, n.Blocks, n.Status, n.FolderID, n.UpdatedAt, n.ID,
	)
	if err != nil {
		return err
	}
	if err := r.storeLinks(ctx, n.ID, n.Blocks); err != nil {
		return fmt.Errorf("store links: %w", err)
	}
	return r.SetTags(ctx, n.ID, n.Tags)
}

// storeLinks replaces the stored `[[Title]]` links for a note based on its
// current blocks content. Titles that do not resolve to an existing note, or
// that resolve back to the note itself, are silently skipped.
func (r *NoteRepo) storeLinks(ctx context.Context, noteID string, blocks string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM note_links WHERE note_id = ?`, noteID); err != nil {
		return fmt.Errorf("clear links: %w", err)
	}
	titles := extractLinks(blocks)
	for _, title := range titles {
		var targetID string
		err := r.db.GetContext(ctx, &targetID, `SELECT id FROM notes WHERE title = ? LIMIT 1`, title)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return fmt.Errorf("resolve link title %q: %w", title, err)
		}
		if targetID == noteID {
			continue
		}
		if _, err := r.db.ExecContext(ctx, `INSERT OR IGNORE INTO note_links(note_id,target_note_id) VALUES(?,?)`, noteID, targetID); err != nil {
			return fmt.Errorf("insert link: %w", err)
		}
	}
	return nil
}

// FindAllLinks returns every stored note-to-note link, used to build the
// notes graph.
func (r *NoteRepo) FindAllLinks(ctx context.Context) ([]note.Link, error) {
	links := make([]note.Link, 0)
	err := r.db.SelectContext(ctx, &links, `SELECT note_id, target_note_id FROM note_links`)
	return links, err
}

type graphEntityRow struct {
	ID     string `db:"id"`
	Label  string `db:"label"`
	NoteID string `db:"note_id"`
}

// Graph builds the full notes graph: every note as a node, plus tasks and
// projects that are linked to a note, with edges for [[Title]] note links
// and note_id foreign keys.
func (r *NoteRepo) Graph(ctx context.Context) (*note.Graph, error) {
	graph := &note.Graph{
		Nodes: make([]note.GraphNode, 0),
		Edges: make([]note.GraphEdge, 0),
	}

	noteRows := make([]struct {
		ID    string `db:"id"`
		Title string `db:"title"`
	}, 0)
	if err := r.db.SelectContext(ctx, &noteRows, `SELECT id, title FROM notes`); err != nil {
		return nil, fmt.Errorf("load notes for graph: %w", err)
	}
	for _, row := range noteRows {
		graph.Nodes = append(graph.Nodes, note.GraphNode{
			ID:    "note:" + row.ID,
			Label: row.Title,
			Type:  "note",
		})
	}

	links, err := r.FindAllLinks(ctx)
	if err != nil {
		return nil, fmt.Errorf("load links for graph: %w", err)
	}
	for _, l := range links {
		graph.Edges = append(graph.Edges, note.GraphEdge{
			Source: "note:" + l.NoteID,
			Target: "note:" + l.TargetNoteID,
			Type:   "link",
		})
	}

	taskRows := make([]graphEntityRow, 0)
	if err := r.db.SelectContext(ctx, &taskRows,
		`SELECT id, title AS label, note_id FROM tasks WHERE note_id IS NOT NULL`,
	); err != nil {
		return nil, fmt.Errorf("load tasks for graph: %w", err)
	}
	for _, row := range taskRows {
		graph.Nodes = append(graph.Nodes, note.GraphNode{
			ID:    "task:" + row.ID,
			Label: row.Label,
			Type:  "task",
		})
		graph.Edges = append(graph.Edges, note.GraphEdge{
			Source: "task:" + row.ID,
			Target: "note:" + row.NoteID,
			Type:   "task",
		})
	}

	projectRows := make([]graphEntityRow, 0)
	if err := r.db.SelectContext(ctx, &projectRows,
		`SELECT id, name AS label, note_id FROM projects WHERE note_id IS NOT NULL`,
	); err != nil {
		return nil, fmt.Errorf("load projects for graph: %w", err)
	}
	for _, row := range projectRows {
		graph.Nodes = append(graph.Nodes, note.GraphNode{
			ID:    "project:" + row.ID,
			Label: row.Label,
			Type:  "project",
		})
		graph.Edges = append(graph.Edges, note.GraphEdge{
			Source: "project:" + row.ID,
			Target: "note:" + row.NoteID,
			Type:   "project",
		})
	}

	return graph, nil
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
