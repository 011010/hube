package folder

import (
	"context"

	"github.com/husari/hube/internal/domain/folder"
)

type Service struct{ repo folder.Repository }

func NewService(repo folder.Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context) ([]folder.Folder, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) Create(ctx context.Context, f *folder.Folder) error {
	return s.repo.Create(ctx, f)
}

func (s *Service) Update(ctx context.Context, f *folder.Folder) error {
	return s.repo.Update(ctx, f)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
