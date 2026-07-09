package note

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	n.Normalize()
	if err := n.Validate(); err != nil {
		return err
	}
	if n.Tags == nil {
		n.Tags = []string{}
	}
	return s.repo.Create(ctx, n)
}

func (s *Service) Update(ctx context.Context, n *note.Note) error {
	n.Normalize()
	if err := n.Validate(); err != nil {
		return err
	}
	if n.Tags == nil {
		n.Tags = []string{}
	}
	return s.repo.Update(ctx, n)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) Graph(ctx context.Context) (*note.Graph, error) {
	return s.repo.Graph(ctx)
}

// ExportAgentMarkdown renders a note as a self-contained Markdown document:
// YAML frontmatter with the note's structured properties, followed by the
// note's plain-text body. Intended for consumption by agents/tools that
// prefer flat Markdown over JSON.
func (s *Service) ExportAgentMarkdown(_ context.Context, n *note.Note) (string, error) {
	var sb strings.Builder

	sb.WriteString("---\n")
	fmt.Fprintf(&sb, "title: %s\n", yamlQuote(n.Title))
	fmt.Fprintf(&sb, "status: %s\n", n.Status)
	fmt.Fprintf(&sb, "priority: %s\n", n.Priority)
	if n.DueDate != nil {
		fmt.Fprintf(&sb, "due_date: %s\n", *n.DueDate)
	} else {
		sb.WriteString("due_date: null\n")
	}
	if len(n.Tags) > 0 {
		sb.WriteString("tags:\n")
		for _, tag := range n.Tags {
			fmt.Fprintf(&sb, "  - %s\n", yamlQuote(tag))
		}
	} else {
		sb.WriteString("tags: []\n")
	}
	fmt.Fprintf(&sb, "created_at: %s\n", n.CreatedAt.UTC().Format(time.RFC3339))
	fmt.Fprintf(&sb, "updated_at: %s\n", n.UpdatedAt.UTC().Format(time.RFC3339))
	sb.WriteString("---\n\n")
	sb.WriteString(n.Content)
	sb.WriteString("\n")

	return sb.String(), nil
}

// yamlQuote wraps a scalar in double quotes and escapes embedded quotes/backslashes
// so it is safe to emit as a YAML string value regardless of its contents.
func yamlQuote(s string) string {
	escaped := strings.ReplaceAll(s, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}
