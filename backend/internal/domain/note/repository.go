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
}

// Link represents a `[[Title]]` reference from one note's blocks to another
// note, resolved to note IDs.
type Link struct {
	NoteID       string `json:"note_id" db:"note_id"`
	TargetNoteID string `json:"target_note_id" db:"target_note_id"`
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
