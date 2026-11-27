package note

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
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
	// Properties are persisted only by SetProperties; accepting them here would
	// echo values the create path never stores.
	if len(n.Properties) > 0 {
		return &note.ValidationError{
			Field:   "properties",
			Message: "properties must be set via PUT /notes/{id}/properties",
		}
	}
	n.Normalize()
	if err := n.Validate(); err != nil {
		return err
	}
	if n.Tags == nil {
		n.Tags = []string{}
	}
	if n.Properties == nil {
		n.Properties = []note.Property{}
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
	if n.Properties == nil {
		n.Properties = []note.Property{}
	}
	return s.repo.Update(ctx, n)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// SetProperties replaces the note's property set. A missing note returns
// (nil, nil) so the HTTP layer can answer 404 without inventing an error type.
func (s *Service) SetProperties(ctx context.Context, id string, props []note.Property) ([]note.Property, error) {
	n, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, nil
	}
	if err := note.ValidateProperties(props); err != nil {
		return nil, err
	}
	if props == nil {
		props = []note.Property{}
	}
	if err := s.repo.SetProperties(ctx, id, props); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return props, nil
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
	if len(n.Tags) > 0 {
		sb.WriteString("tags:\n")
		for _, tag := range n.Tags {
			fmt.Fprintf(&sb, "  - %s\n", yamlQuote(tag))
		}
	} else {
		sb.WriteString("tags: []\n")
	}
	for _, p := range n.Properties {
		if note.IsReservedPropertyKey(p.Key) {
			return "", fmt.Errorf("export property %q: reserved key", p.Key)
		}
		switch p.Type {
		case note.PropertyNumber, note.PropertyBoolean:
			fmt.Fprintf(&sb, "%s: %v\n", yamlQuote(p.Key), p.Value)
		case note.PropertyList:
			items, ok := note.StringList(p.Value)
			if !ok {
				return "", fmt.Errorf("export property %q: value is not a string list (%T)", p.Key, p.Value)
			}
			if len(items) == 0 {
				fmt.Fprintf(&sb, "%s: []\n", yamlQuote(p.Key))
				continue
			}
			fmt.Fprintf(&sb, "%s:\n", yamlQuote(p.Key))
			for _, item := range items {
				fmt.Fprintf(&sb, "  - %s\n", yamlQuote(item))
			}
		default:
			fmt.Fprintf(&sb, "%s: %s\n", yamlQuote(p.Key), yamlQuote(fmt.Sprint(p.Value)))
		}
	}
	fmt.Fprintf(&sb, "created_at: %s\n", n.CreatedAt.UTC().Format(time.RFC3339))
	fmt.Fprintf(&sb, "updated_at: %s\n", n.UpdatedAt.UTC().Format(time.RFC3339))
	sb.WriteString("---\n\n")
	sb.WriteString(n.Content)
	sb.WriteString("\n")

	return sb.String(), nil
}

// yamlQuote renders s as a YAML double-quoted scalar. strconv.Quote escapes
// quotes, backslashes, newlines, and control characters using the same escape
// sequences YAML double-quoted scalars accept, so the value round-trips exactly.
func yamlQuote(s string) string {
	return strconv.Quote(s)
}
