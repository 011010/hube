package note

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/husari/hube/internal/domain/note"
	"gopkg.in/yaml.v3"
)

type mockRepo struct {
	created *note.Note
	updated *note.Note
}

func (m *mockRepo) FindAll(ctx context.Context, folderID *string) ([]note.Note, error) {
	return nil, nil
}

func (m *mockRepo) FindByID(ctx context.Context, id string) (*note.Note, error) {
	return nil, nil
}

func (m *mockRepo) Search(ctx context.Context, query string) ([]note.Note, error) {
	return nil, nil
}

func (m *mockRepo) Create(ctx context.Context, n *note.Note) error {
	m.created = n
	return nil
}

func (m *mockRepo) Update(ctx context.Context, n *note.Note) error {
	m.updated = n
	return nil
}

func (m *mockRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockRepo) SetTags(ctx context.Context, noteID string, tags []string) error {
	return nil
}

func (m *mockRepo) FindAllLinks(ctx context.Context) ([]note.Link, error) {
	return nil, nil
}

func (m *mockRepo) Graph(ctx context.Context) (*note.Graph, error) {
	return &note.Graph{}, nil
}

func TestService_Create_NormalizesAndValidates(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	n := &note.Note{Title: "Test"}
	if err := svc.Create(context.Background(), n); err != nil {
		t.Fatalf("Create returned unexpected error: %v", err)
	}

	if n.Status != note.StatusDraft {
		t.Errorf("Status = %q, want %q", n.Status, note.StatusDraft)
	}
	if n.Priority != note.PriorityMedium {
		t.Errorf("Priority = %q, want %q", n.Priority, note.PriorityMedium)
	}
	if repo.created != n {
		t.Error("Create did not persist the normalized note")
	}
}

func TestService_Create_InvalidStatus(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	n := &note.Note{Title: "Test", Status: "invalid", Priority: note.PriorityMedium}
	err := svc.Create(context.Background(), n)
	if err == nil {
		t.Fatal("Create expected error for invalid status")
	}
	var valErr *note.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("error type = %T, want *note.ValidationError", err)
	}
	if valErr.Field != "status" {
		t.Errorf("ValidationError.Field = %q, want %q", valErr.Field, "status")
	}
	if repo.created != nil {
		t.Error("Create should not persist an invalid note")
	}
}

func TestService_Create_EmptyTitle(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	n := &note.Note{Status: note.StatusDraft, Priority: note.PriorityMedium}
	err := svc.Create(context.Background(), n)
	if err == nil {
		t.Fatal("Create expected error for empty title")
	}
	var valErr *note.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("error type = %T, want *note.ValidationError", err)
	}
	if valErr.Field != "title" {
		t.Errorf("ValidationError.Field = %q, want %q", valErr.Field, "title")
	}
	if repo.created != nil {
		t.Error("Create should not persist an invalid note")
	}
}

func TestService_Update_NormalizesEmptyStatusAndPriority(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	n := &note.Note{ID: "1", Title: "Test"}
	if err := svc.Update(context.Background(), n); err != nil {
		t.Fatalf("Update returned unexpected error: %v", err)
	}

	if n.Status != note.StatusDraft {
		t.Errorf("Status = %q, want %q", n.Status, note.StatusDraft)
	}
	if n.Priority != note.PriorityMedium {
		t.Errorf("Priority = %q, want %q", n.Priority, note.PriorityMedium)
	}
	if repo.updated != n {
		t.Error("Update did not persist the normalized note")
	}
}

func TestService_Update_InvalidDueDate(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	invalidDate := "not-a-date"
	n := &note.Note{ID: "1", Title: "Test", Status: note.StatusDraft, Priority: note.PriorityMedium, DueDate: &invalidDate}
	err := svc.Update(context.Background(), n)
	if err == nil {
		t.Fatal("Update expected error for invalid due_date")
	}
	var valErr *note.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("error type = %T, want *note.ValidationError", err)
	}
	if valErr.Field != "due_date" {
		t.Errorf("ValidationError.Field = %q, want %q", valErr.Field, "due_date")
	}
	if repo.updated != nil {
		t.Error("Update should not persist an invalid note")
	}
}

func TestService_ExportAgentMarkdown(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	due := "2026-08-01"
	created := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	updated := time.Date(2026, 7, 2, 9, 30, 0, 0, time.UTC)
	n := &note.Note{
		ID:        "1",
		Title:     "My Note",
		Content:   "Hello world",
		Status:    note.StatusDraft,
		Priority:  note.PriorityHigh,
		DueDate:   &due,
		Tags:      []string{"go", "test"},
		CreatedAt: created,
		UpdatedAt: updated,
	}

	md, err := svc.ExportAgentMarkdown(context.Background(), n)
	if err != nil {
		t.Fatalf("ExportAgentMarkdown returned unexpected error: %v", err)
	}

	if !strings.HasPrefix(md, "---\n") {
		t.Fatalf("expected markdown to start with YAML frontmatter delimiter, got: %q", md)
	}
	wantFields := []string{
		`title: "My Note"`,
		"status: draft",
		"priority: high",
		"due_date: 2026-08-01",
		"tags:",
		`  - "go"`,
		`  - "test"`,
		"created_at: 2026-07-01T12:00:00Z",
		"updated_at: 2026-07-02T09:30:00Z",
	}
	for _, want := range wantFields {
		if !strings.Contains(md, want) {
			t.Errorf("expected markdown to contain %q, got:\n%s", want, md)
		}
	}
	if !strings.HasSuffix(strings.TrimRight(md, "\n"), "Hello world") {
		t.Errorf("expected markdown body to end with note content, got:\n%s", md)
	}
	// Frontmatter must be closed before the body.
	parts := strings.SplitN(md, "---\n", 3)
	if len(parts) != 3 {
		t.Fatalf("expected exactly two --- delimiters, got markdown:\n%s", md)
	}
	if !strings.Contains(parts[2], "Hello world") {
		t.Errorf("expected body section to contain note content, got: %q", parts[2])
	}
}

func TestService_ExportAgentMarkdown_NoDueDateOrTags(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	n := &note.Note{
		ID:       "1",
		Title:    "Bare Note",
		Content:  "",
		Status:   note.StatusDraft,
		Priority: note.PriorityMedium,
	}

	md, err := svc.ExportAgentMarkdown(context.Background(), n)
	if err != nil {
		t.Fatalf("ExportAgentMarkdown returned unexpected error: %v", err)
	}
	if !strings.Contains(md, "due_date: null") {
		t.Errorf("expected due_date: null for nil DueDate, got:\n%s", md)
	}
	if !strings.Contains(md, "tags: []") {
		t.Errorf("expected tags: [] for empty tags, got:\n%s", md)
	}
}

func TestService_ExportAgentMarkdown_EscapesTagsAgainstYAMLInjection(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	n := &note.Note{
		ID:       "1",
		Title:    "Injection Note",
		Content:  "body",
		Status:   note.StatusDraft,
		Priority: note.PriorityMedium,
		// A tag containing a raw newline followed by YAML syntax used to be
		// written unescaped, letting it break out of the tags list and inject
		// a sibling "injected" key into the frontmatter.
		Tags: []string{"safe", "evil\ninjected: true"},
	}

	md, err := svc.ExportAgentMarkdown(context.Background(), n)
	if err != nil {
		t.Fatalf("ExportAgentMarkdown returned unexpected error: %v", err)
	}

	parts := strings.SplitN(md, "---\n", 3)
	if len(parts) != 3 {
		t.Fatalf("expected exactly two --- delimiters, got markdown:\n%s", md)
	}
	frontmatter := parts[1]

	var doc map[string]any
	if err := yaml.Unmarshal([]byte(frontmatter), &doc); err != nil {
		t.Fatalf("frontmatter is not valid YAML (tag injection broke the document): %v\nfrontmatter:\n%s", err, frontmatter)
	}

	if _, injected := doc["injected"]; injected {
		t.Errorf("tag value injected a sibling top-level YAML key, got parsed doc: %#v", doc)
	}

	tags, ok := doc["tags"].([]any)
	if !ok || len(tags) != 2 {
		t.Fatalf("expected tags to parse as a 2-element list, got: %#v", doc["tags"])
	}
	// YAML double-quoted scalars fold an embedded raw line break into a single
	// space (per the YAML spec's line-folding rules), so the round-tripped tag
	// is not byte-identical to the original — that's expected. What matters is
	// that it stays a single, safely-parsed scalar with no injected key (checked
	// above), and no longer breaks into two lines.
	if tags[0] != "safe" || tags[1] != "evil injected: true" {
		t.Errorf("expected tags to parse as safe folded scalars, got: %#v", tags)
	}
}
