package wishlist

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/husari/hube/internal/domain/wishlist"
)

type Service struct {
	repo wishlist.Repository
}

func NewService(repo wishlist.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]wishlist.Item, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) Get(ctx context.Context, id string) (*wishlist.Item, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, item *wishlist.Item) error {
	item.ID = uuid.New().String()
	if item.Status == "" {
		item.Status = wishlist.StatusPending
	}
	if item.Priority == "" {
		item.Priority = wishlist.PriorityMedium
	}
	if item.Currency == "" {
		item.Currency = "USD"
	}
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	return s.repo.Create(ctx, item)
}

func (s *Service) Update(ctx context.Context, item *wishlist.Item) error {
	item.UpdatedAt = time.Now()
	return s.repo.Update(ctx, item)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
