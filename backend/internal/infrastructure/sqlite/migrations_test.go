package sqlite

import (
	"testing"
)

func TestMigrations_NoteProperties(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	var count int
	err = db.Get(&count, `SELECT COUNT(*) FROM pragma_table_info('notes') WHERE name IN ('status', 'priority', 'due_date', 'blocks')`)
	if err != nil {
		t.Fatalf("query columns: %v", err)
	}
	if count != 4 {
		t.Fatalf("expected 4 new columns, got %d", count)
	}
}
