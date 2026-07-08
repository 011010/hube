package sqlite

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/husari/hube/internal/domain/note"
	"github.com/husari/hube/internal/domain/task"
	"github.com/jmoiron/sqlx"
)

func createTestNote(t *testing.T, db *sqlx.DB, title string) string {
	t.Helper()
	noteRepo := NewNoteRepo(db)
	n := &note.Note{Title: title, Status: note.StatusDraft, Priority: note.PriorityMedium}
	if err := noteRepo.Create(context.Background(), n); err != nil {
		t.Fatalf("create test note: %v", err)
	}
	return n.ID
}

func TestTaskRepo_NoteIDRoundTrip(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := NewTaskRepo(db)
	ctx := context.Background()

	noteID := createTestNote(t, db, "Note A")
	tsk := &task.Task{
		ID:       uuid.NewString(),
		Title:    "Linked task",
		Priority: task.PriorityMedium,
		Status:   task.StatusTodo,
		NoteID:   &noteID,
	}
	if err := repo.Create(ctx, tsk); err != nil {
		t.Fatalf("create task: %v", err)
	}

	fetched, err := repo.FindByID(ctx, tsk.ID)
	if err != nil {
		t.Fatalf("find task: %v", err)
	}
	if fetched.NoteID == nil || *fetched.NoteID != noteID {
		t.Fatalf("expected note_id %q, got %v", noteID, fetched.NoteID)
	}

	otherNoteID := createTestNote(t, db, "Note B")
	fetched.NoteID = &otherNoteID
	if err := repo.Update(ctx, fetched); err != nil {
		t.Fatalf("update task: %v", err)
	}

	updated, err := repo.FindByID(ctx, tsk.ID)
	if err != nil {
		t.Fatalf("find updated task: %v", err)
	}
	if updated.NoteID == nil || *updated.NoteID != otherNoteID {
		t.Fatalf("expected note_id %q after update, got %v", otherNoteID, updated.NoteID)
	}
}

func TestTaskRepo_NoteIDNilByDefault(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := NewTaskRepo(db)
	ctx := context.Background()

	tsk := &task.Task{
		ID:       uuid.NewString(),
		Title:    "No note",
		Priority: task.PriorityMedium,
		Status:   task.StatusTodo,
	}
	if err := repo.Create(ctx, tsk); err != nil {
		t.Fatalf("create task: %v", err)
	}

	fetched, err := repo.FindByID(ctx, tsk.ID)
	if err != nil {
		t.Fatalf("find task: %v", err)
	}
	if fetched.NoteID != nil {
		t.Fatalf("expected nil note_id, got %v", *fetched.NoteID)
	}
}
