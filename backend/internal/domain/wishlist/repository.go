package wishlist

import "context"

type Repository interface {
	FindAll(ctx context.Context) ([]Item, error)
	FindByID(ctx context.Context, id string) (*Item, error)
	Create(ctx context.Context, item *Item) error
	Update(ctx context.Context, item *Item) error
	Delete(ctx context.Context, id string) error
}
