package project

import (
	"context"

	"github.com/husari/hube/internal/domain/project"
)

type Service struct{ repo project.Repository }

func NewService(repo project.Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context) ([]project.Project, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) Get(ctx context.Context, id string) (*project.Project, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, p *project.Project) error {
	if p.Status == "" {
		p.Status = project.StatusPlanning
	}
	if p.Color == "" {
		p.Color = "#6366f1"
	}
	return s.repo.Create(ctx, p)
}

func (s *Service) Update(ctx context.Context, p *project.Project) error {
	return s.repo.Update(ctx, p)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
