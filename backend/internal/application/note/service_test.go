package note

import (
	"context"
	"errors"
	"testing"

	"github.com/husari/hube/internal/domain/note"
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
