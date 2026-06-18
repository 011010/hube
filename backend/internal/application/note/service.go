package note

import (
	"context"

	"github.com/husari/hube/internal/domain/note"
)

type Service struct{ repo note.Repository }

func NewService(repo note.Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context, folderID *string) ([]note.Note, error) {
	return s.repo.FindAll(ctx, folderID)
}

func (s *Service) Get(ctx context.Context, id string) (*note.Note, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Search(ctx context.Context, query string) ([]note.Note, error) {
	return s.repo.Search(ctx, query)
}

func (s *Service) Create(ctx context.Context, n *note.Note) error {
	if n.Tags == nil {
		n.Tags = []string{}
	}
	return s.repo.Create(ctx, n)
}

func (s *Service) Update(ctx context.Context, n *note.Note) error {
	if n.Tags == nil {
		n.Tags = []string{}
	}
	return s.repo.Update(ctx, n)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
