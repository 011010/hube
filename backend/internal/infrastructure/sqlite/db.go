package sqlite

import (
	_ "embed"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

//go:embed migrations/001_init.sql
var initSQL string

func Open(path string) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1)

	if _, err = db.Exec("PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;"); err != nil {
		return nil, fmt.Errorf("sqlite pragmas: %w", err)
	}

	if _, err = db.Exec(initSQL); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return db, nil
}
