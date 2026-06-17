package event

import (
	"context"
	"time"
)

type Repository interface {
	FindAll(ctx context.Context) ([]Event, error)
	FindByRange(ctx context.Context, from, to time.Time) ([]Event, error)
	FindByID(ctx context.Context, id string) (*Event, error)
	Create(ctx context.Context, e *Event) error
	Update(ctx context.Context, e *Event) error
	Delete(ctx context.Context, id string) error
}
