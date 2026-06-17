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

func (s *Service) Get(ctx context.Context, id string) (*task.Task, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, t *task.Task) error {
	t.ID = uuid.New().String()
	t.Status = task.StatusTodo
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	return s.repo.Create(ctx, t)
}

func (s *Service) Update(ctx context.Context, t *task.Task) error {
	t.UpdatedAt = time.Now()
	return s.repo.Update(ctx, t)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
