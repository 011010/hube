package sqlite

import (
	"testing"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func noteColumnCount(t *testing.T, db *sqlx.DB, col string) int {
	t.Helper()
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM pragma_table_info('notes') WHERE name = ?`, col); err != nil {
		t.Fatalf("check column %q: %v", col, err)
	}
	return count
}

func TestMigrations_NoteProperties(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	for _, col := range []string{"status", "blocks"} {
		if got := noteColumnCount(t, db, col); got != 1 {
			t.Fatalf("expected column %q to exist, got count %d", col, got)
		}
	}

	// priority and due_date are task semantics; notes must not carry them.
	for _, col := range []string{"priority", "due_date"} {
		if got := noteColumnCount(t, db, col); got != 0 {
			t.Fatalf("expected column %q to be dropped, got count %d", col, got)
		}
	}
}

func TestMigrations_NotePropertiesUpgrade(t *testing.T) {
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE notes (
		id         TEXT PRIMARY KEY,
		title      TEXT NOT NULL DEFAULT '',
		content    TEXT NOT NULL DEFAULT '',
		folder_id  TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("create pre-migration notes table: %v", err)
	}

	_, err = db.Exec(`INSERT INTO notes (id, title) VALUES ('1', 'Test')`)
	if err != nil {
		t.Fatalf("insert note: %v", err)
	}

	if err := RunMigrations(db); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	var status, blocks string
	err = db.QueryRow(`SELECT status, blocks FROM notes WHERE id = '1'`).Scan(&status, &blocks)
	if err != nil {
		t.Fatalf("select migrated note: %v", err)
	}

	if status != "draft" {
		t.Errorf("status default = %q, want %q", status, "draft")
	}
	if blocks != "" {
		t.Errorf("blocks default = %q, want empty string", blocks)
	}
}

// A database already carrying the task-shaped note columns must have them
// removed, without losing the rows themselves.
func TestMigrations_DropsNoteTaskFieldsFromExistingDB(t *testing.T) {
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE notes (
		id         TEXT PRIMARY KEY,
		title      TEXT NOT NULL DEFAULT '',
		content    TEXT NOT NULL DEFAULT '',
		folder_id  TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		status     TEXT NOT NULL DEFAULT 'draft',
		priority   TEXT NOT NULL DEFAULT 'medium',
		due_date   TEXT,
		blocks     TEXT NOT NULL DEFAULT ''
	)`)
	if err != nil {
		t.Fatalf("create post-009 notes table: %v", err)
	}

	// Mark 009 as already applied so the runner reaches 012 on this schema.
	if _, err = db.Exec(`CREATE TABLE schema_migrations (
		name       TEXT PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		t.Fatalf("create schema_migrations: %v", err)
	}
	if _, err = db.Exec(`INSERT INTO schema_migrations (name) VALUES ('009_note_properties')`); err != nil {
		t.Fatalf("mark 009 applied: %v", err)
	}

	if _, err = db.Exec(`INSERT INTO notes (id, title, priority, due_date) VALUES ('1', 'Keep Me', 'high', '2026-08-01')`); err != nil {
		t.Fatalf("insert legacy note: %v", err)
	}

	if err := RunMigrations(db); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	for _, col := range []string{"priority", "due_date"} {
		if got := noteColumnCount(t, db, col); got != 0 {
			t.Errorf("expected column %q to be dropped, got count %d", col, got)
		}
	}

	var title string
	if err := db.QueryRow(`SELECT title FROM notes WHERE id = '1'`).Scan(&title); err != nil {
		t.Fatalf("select surviving note: %v", err)
	}
	if title != "Keep Me" {
		t.Errorf("title = %q, want %q", title, "Keep Me")
	}
}
