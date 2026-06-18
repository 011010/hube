package task

import "context"

type Repository interface {
	FindAll(ctx context.Context) ([]Task, error)
	FindByProject(ctx context.Context, projectID string) ([]Task, error)
	FindByID(ctx context.Context, id string) (*Task, error)
	Create(ctx context.Context, t *Task) error
	Update(ctx context.Context, t *Task) error
	Delete(ctx context.Context, id string) error
}
