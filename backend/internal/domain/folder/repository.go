package folder

import "context"

type Repository interface {
	FindAll(ctx context.Context) ([]Folder, error)
	Create(ctx context.Context, f *Folder) error
	Update(ctx context.Context, f *Folder) error
	Delete(ctx context.Context, id string) error
}
