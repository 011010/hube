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

type migration struct {
	name string
	sql  string
}

var migrations = []migration{
	{"001_init", initSQL},
	{"002_notes", notesSQL},
	{"003_projects", projectsSQL},
	{"004_settings", settingsSQL},
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

	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		name       TEXT PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return nil, fmt.Errorf("create schema_migrations: %w", err)
	}

	for _, m := range migrations {
		var count int
		db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE name = ?`, m.name).Scan(&count)
		if count > 0 {
			continue
		}
		if _, err = db.Exec(m.sql); err != nil {
			return nil, fmt.Errorf("run migration %s: %w", m.name, err)
		}
		if _, err = db.Exec(`INSERT INTO schema_migrations (name) VALUES (?)`, m.name); err != nil {
			return nil, fmt.Errorf("record migration %s: %w", m.name, err)
		}
	}

	return db, nil
}
