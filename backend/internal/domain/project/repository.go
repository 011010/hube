package project

import "context"

type Repository interface {
	FindAll(ctx context.Context) ([]Project, error)
	FindByID(ctx context.Context, id string) (*Project, error)
	Create(ctx context.Context, p *Project) error
	Update(ctx context.Context, p *Project) error
	Delete(ctx context.Context, id string) error
}
