package diagram

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/husari/hube/internal/domain/diagram"
)

type Service struct {
	repo diagram.Repository
}

func NewService(repo diagram.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]diagram.Diagram, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) Get(ctx context.Context, id string) (*diagram.Diagram, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, d *diagram.Diagram) error {
	d.ID = uuid.New().String()
	if d.Nodes == "" {
		d.Nodes = "[]"
	}
	if d.Edges == "" {
		d.Edges = "[]"
	}
	d.CreatedAt = time.Now()
	d.UpdatedAt = time.Now()
	return s.repo.Create(ctx, d)
}

func (s *Service) Update(ctx context.Context, d *diagram.Diagram) error {
	d.UpdatedAt = time.Now()
	return s.repo.Update(ctx, d)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
