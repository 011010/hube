package sqlite

import (
	"database/sql"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestMigrations_NoteProperties(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	for _, col := range []string{"status", "priority", "due_date", "blocks"} {
		var exists int
		err := db.Get(&exists, `SELECT COUNT(*) FROM pragma_table_info('notes') WHERE name = ?`, col)
		if err != nil {
			t.Fatalf("check column %q: %v", col, err)
		}
		if exists != 1 {
			t.Fatalf("expected column %q to exist, got count %d", col, exists)
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

	var status, priority, blocks string
	var dueDate sql.NullString
	err = db.QueryRow(`SELECT status, priority, due_date, blocks FROM notes WHERE id = '1'`).
		Scan(&status, &priority, &dueDate, &blocks)
	if err != nil {
		t.Fatalf("select migrated note: %v", err)
	}

	if status != "draft" {
		t.Errorf("status default = %q, want %q", status, "draft")
	}
	if priority != "medium" {
		t.Errorf("priority default = %q, want %q", priority, "medium")
	}
	if blocks != "" {
		t.Errorf("blocks default = %q, want empty string", blocks)
	}
	if dueDate.Valid {
		t.Errorf("due_date default = %q, want NULL", dueDate.String)
	}
}
