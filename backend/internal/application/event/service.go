package event

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/husari/hube/internal/domain/event"
)

type Service struct {
	repo event.Repository
}

func NewService(repo event.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]event.Event, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) ListByRange(ctx context.Context, from, to time.Time) ([]event.Event, error) {
	return s.repo.FindByRange(ctx, from, to)
}

func (s *Service) Get(ctx context.Context, id string) (*event.Event, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, e *event.Event) error {
	e.Normalize()
	if err := e.Validate(); err != nil {
		return err
	}
	e.ID = uuid.New().String()
	e.CreatedAt = time.Now()
	e.UpdatedAt = time.Now()
	return s.repo.Create(ctx, e)
}

func (s *Service) Update(ctx context.Context, e *event.Event) error {
	e.Normalize()
	if err := e.Validate(); err != nil {
		return err
	}
	e.UpdatedAt = time.Now()
	return s.repo.Update(ctx, e)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
