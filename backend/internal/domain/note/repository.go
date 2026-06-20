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
