package note

import "context"

type Repository interface {
	FindAll(ctx context.Context, folderID *string) ([]Note, error)
	FindByID(ctx context.Context, id string) (*Note, error)
	Search(ctx context.Context, query string) ([]Note, error)
	Create(ctx context.Context, n *Note) error
	Update(ctx context.Context, n *Note) error
	Delete(ctx context.Context, id string) error
	SetTags(ctx context.Context, noteID string, tags []string) error
	FindAllLinks(ctx context.Context) ([]Link, error)
	Graph(ctx context.Context) (*Graph, error)
}

// Link represents a `[[Title]]` reference from one note's blocks to another
// note, resolved to note IDs.
type Link struct {
	NoteID       string `json:"note_id" db:"note_id"`
	TargetNoteID string `json:"target_note_id" db:"target_note_id"`
}

// GraphNode is a single vertex in the notes graph: a note, a task, or a
// project. IDs are namespaced by type (e.g. "note:<id>") so they stay unique
// across entity kinds.
type GraphNode struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

// GraphEdge connects two graph nodes. Type describes the relationship:
// "link" for a [[Title]] note-to-note reference, "task" for a task linked to
// a note, "project" for a project linked to a note.
type GraphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
}

// Graph is the full notes graph: notes, tasks, and projects as nodes, with
// note_links and note_id foreign keys as edges.
type Graph struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

type EmbeddingRecord struct {
	ID        string
	Embedding []byte
}

type EmbeddingRepository interface {
	Repository
	StoreEmbedding(ctx context.Context, noteID string, embedding []byte) error
	FindAllEmbeddings(ctx context.Context) ([]EmbeddingRecord, error)
}
