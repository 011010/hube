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
