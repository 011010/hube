package task

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/husari/hube/internal/domain/task"
)

type Service struct {
	repo task.Repository
}

func NewService(repo task.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]task.Task, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) ListByProject(ctx context.Context, projectID string) ([]task.Task, error) {
	return s.repo.FindByProject(ctx, projectID)
}

func (s *Service) Get(ctx context.Context, id string) (*task.Task, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, t *task.Task) error {
	t.Normalize()
	// A new task always starts in todo, whatever the caller asked for, so
	// the status is forced before validation rather than rejected.
	t.Status = task.StatusTodo
	if err := t.Validate(); err != nil {
		return err
	}
	t.ID = uuid.New().String()
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	return s.repo.Create(ctx, t)
}

func (s *Service) Update(ctx context.Context, t *task.Task) error {
	t.Normalize()
	if err := t.Validate(); err != nil {
		return err
	}
	t.UpdatedAt = time.Now()
	return s.repo.Update(ctx, t)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
