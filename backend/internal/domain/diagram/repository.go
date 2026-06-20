package diagram

import "context"

type Repository interface {
	FindAll(ctx context.Context) ([]Diagram, error)
	FindByID(ctx context.Context, id string) (*Diagram, error)
	Create(ctx context.Context, d *Diagram) error
	Update(ctx context.Context, d *Diagram) error
	Delete(ctx context.Context, id string) error
}
