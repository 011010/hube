package sqlite

import (
	_ "embed"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

//go:embed migrations/001_init.sql
var initSQL string

//go:embed migrations/002_notes.sql
var notesSQL string

//go:embed migrations/003_projects.sql
var projectsSQL string

//go:embed migrations/004_settings.sql
var settingsSQL string

//go:embed migrations/005_wishlist.sql
var wishlistSQL string

//go:embed migrations/006_task_recurrence.sql
var taskRecurrenceSQL string

//go:embed migrations/007_diagrams.sql
var diagramsSQL string

//go:embed migrations/008_note_embeddings.sql
var noteEmbeddingsSQL string

//go:embed migrations/009_note_properties.sql
var notePropertiesSQL string

//go:embed migrations/010_note_links.sql
var noteLinksSQL string

//go:embed migrations/011_task_project_note_id.sql
var taskProjectNoteIDSQL string

type migration struct {
	name string
	sql  string
}

var migrations = []migration{
	{"001_init", initSQL},
	{"002_notes", notesSQL},
	{"003_projects", projectsSQL},
	{"004_settings", settingsSQL},
	{"005_wishlist", wishlistSQL},
	{"006_task_recurrence", taskRecurrenceSQL},
	{"007_diagrams", diagramsSQL},
	{"008_note_embeddings", noteEmbeddingsSQL},
	{"009_note_properties", notePropertiesSQL},
	{"010_note_links", noteLinksSQL},
	{"011_task_project_note_id", taskProjectNoteIDSQL},
}

// RunMigrations applies all pending migrations to the provided database.
// It is exported for tests that need to exercise migration behavior on a
// pre-seeded schema.
func RunMigrations(db *sqlx.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		name       TEXT PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	for _, m := range migrations {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE name = ?`, m.name).Scan(&count); err != nil {
			return fmt.Errorf("check migration %s: %w", m.name, err)
		}
		if count > 0 {
			continue
		}
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", m.name, err)
		}
		if _, err = tx.Exec(m.sql); err != nil {
			tx.Rollback()
			return fmt.Errorf("run migration %s: %w", m.name, err)
		}
		if _, err = tx.Exec(`INSERT INTO schema_migrations (name) VALUES (?)`, m.name); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", m.name, err)
		}
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", m.name, err)
		}
	}

	return nil
}

func Open(path string) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1)

	if _, err = db.Exec("PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;"); err != nil {
		return nil, fmt.Errorf("sqlite pragmas: %w", err)
	}

	if err := RunMigrations(db); err != nil {
		return nil, err
	}

	return db, nil
}
