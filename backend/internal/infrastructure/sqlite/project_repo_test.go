package sqlite

import (
	"context"
	"testing"

	"github.com/husari/hube/internal/domain/project"
)

func TestProjectRepo_NoteIDRoundTrip(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := NewProjectRepo(db)
	ctx := context.Background()

	noteID := createTestNote(t, db, "Note A")
	p := &project.Project{
		Name:   "Linked project",
		Status: project.StatusPlanning,
		Color:  "#6366f1",
		NoteID: &noteID,
	}
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("create project: %v", err)
	}

	fetched, err := repo.FindByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("find project: %v", err)
	}
	if fetched.NoteID == nil || *fetched.NoteID != noteID {
		t.Fatalf("expected note_id %q, got %v", noteID, fetched.NoteID)
	}

	otherNoteID := createTestNote(t, db, "Note B")
	fetched.NoteID = &otherNoteID
	if err := repo.Update(ctx, fetched); err != nil {
		t.Fatalf("update project: %v", err)
	}

	updated, err := repo.FindByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("find updated project: %v", err)
	}
	if updated.NoteID == nil || *updated.NoteID != otherNoteID {
		t.Fatalf("expected note_id %q after update, got %v", otherNoteID, updated.NoteID)
	}
}

func TestProjectRepo_NoteIDNilByDefault(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := NewProjectRepo(db)
	ctx := context.Background()

	p := &project.Project{
		Name:   "No note",
		Status: project.StatusPlanning,
		Color:  "#6366f1",
	}
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("create project: %v", err)
	}

	fetched, err := repo.FindByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("find project: %v", err)
	}
	if fetched.NoteID != nil {
		t.Fatalf("expected nil note_id, got %v", *fetched.NoteID)
	}
}
