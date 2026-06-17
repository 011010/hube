package app

import "context"

type Repository interface {
	FindAll(ctx context.Context) ([]App, error)
	FindByID(ctx context.Context, id string) (*App, error)
	Create(ctx context.Context, a *App) error
	Update(ctx context.Context, a *App) error
	Delete(ctx context.Context, id string) error
}
