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

func tableColumnCount(t *testing.T, db *sqlx.DB, table, col string) int {
	t.Helper()
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?`, table, col); err != nil {
		t.Fatalf("check column %s.%s: %v", table, col, err)
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

func TestMigrations_NotePropertiesTable(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM pragma_table_info('note_properties')`); err != nil {
		t.Fatalf("inspect note_properties: %v", err)
	}
	if count != 4 {
		t.Fatalf("expected 4 note_properties columns, got %d", count)
	}

	for _, col := range []string{"note_id", "key", "type", "value"} {
		if got := tableColumnCount(t, db, "note_properties", col); got != 1 {
			t.Fatalf("expected column note_properties.%s, got count %d", col, got)
		}
	}

	var index int
	if err := db.Get(&index, `SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='idx_note_properties_key'`); err != nil {
		t.Fatalf("inspect index: %v", err)
	}
	if index != 1 {
		t.Fatalf("expected idx_note_properties_key, got %d", index)
	}

	var pk int
	if err := db.Get(&pk, `SELECT pk FROM pragma_table_info('note_properties') WHERE name = 'key'`); err != nil {
		t.Fatalf("inspect primary key: %v", err)
	}
	if pk != 2 {
		t.Fatalf("expected key to be second PK column, got ordinal %d", pk)
	}

	var onDelete string
	if err := db.Get(&onDelete, `SELECT "on_delete" FROM pragma_foreign_key_list('note_properties') WHERE "from" = 'note_id'`); err != nil {
		t.Fatalf("inspect foreign key: %v", err)
	}
	if onDelete != "CASCADE" {
		t.Fatalf("note_id on_delete = %q, want CASCADE", onDelete)
	}

	if _, err := db.Exec(`INSERT INTO notes (id) VALUES ('np-test')`); err != nil {
		t.Fatalf("insert note: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO note_properties (note_id, key, type, value) VALUES ('np-test', 'x', 'widget', '1')`); err == nil {
		t.Fatal("expected CHECK constraint to reject unknown property type")
	}
}

// An existing database at migration 012 must upgrade cleanly, keep its notes,
// and expose an empty property list for them.
func TestMigrations_NotePropertiesTableUpgradeFrom012(t *testing.T) {
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE schema_migrations (
		name       TEXT PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		t.Fatalf("create schema_migrations: %v", err)
	}

	const target = "013_note_properties_table"
	found := false
	for _, m := range migrations {
		if m.name == target {
			found = true
			break
		}
		if _, err := db.Exec(m.sql); err != nil {
			t.Fatalf("apply migration %s: %v", m.name, err)
		}
		if _, err := db.Exec(`INSERT INTO schema_migrations (name) VALUES (?)`, m.name); err != nil {
			t.Fatalf("record migration %s: %v", m.name, err)
		}
	}
	if !found {
		t.Fatalf("migration %s not found in migrations slice", target)
	}

	if _, err := db.Exec(
		`INSERT INTO notes (id, title, status, blocks) VALUES ('upgrade-1', 'Upgrade Note', 'draft', '')`); err != nil {
		t.Fatalf("insert note: %v", err)
	}

	if err := RunMigrations(db); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	assertAppliedOnce := func() {
		t.Helper()
		var count int
		if err := db.Get(&count, `SELECT COUNT(*) FROM schema_migrations WHERE name = ?`, target); err != nil {
			t.Fatalf("count migration %s: %v", target, err)
		}
		if count != 1 {
			t.Fatalf("expected %s to be recorded once, got %d", target, count)
		}
	}
	assertAppliedOnce()

	var title string
	if err := db.QueryRow(`SELECT title FROM notes WHERE id = 'upgrade-1'`).Scan(&title); err != nil {
		t.Fatalf("select surviving note: %v", err)
	}
	if title != "Upgrade Note" {
		t.Errorf("title = %q, want %q", title, "Upgrade Note")
	}

	var tableCount int
	if err := db.Get(&tableCount,
		`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'note_properties'`); err != nil {
		t.Fatalf("inspect note_properties table: %v", err)
	}
	if tableCount != 1 {
		t.Fatalf("expected note_properties table to exist, got %d", tableCount)
	}

	var properties int
	if err := db.Get(&properties, `SELECT COUNT(*) FROM note_properties`); err != nil {
		t.Fatalf("count note_properties rows: %v", err)
	}
	if properties != 0 {
		t.Fatalf("expected 0 note properties after upgrade, got %d", properties)
	}

	if err := RunMigrations(db); err != nil {
		t.Fatalf("run migrations again: %v", err)
	}
	assertAppliedOnce()
}
