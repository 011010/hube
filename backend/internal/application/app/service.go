package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/husari/hube/internal/domain/app"
)

type Service struct {
	repo app.Repository
}

func NewService(repo app.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]app.App, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) Get(ctx context.Context, id string) (*app.App, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, a *app.App) error {
	a.ID = uuid.New().String()
	a.Active = true
	return s.repo.Create(ctx, a)
}

func (s *Service) Update(ctx context.Context, a *app.App) error {
	return s.repo.Update(ctx, a)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
