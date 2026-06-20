package backup

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/jmoiron/sqlx"
)

// Service handles periodic SQLite backups using VACUUM INTO.
type Service struct {
	db     *sqlx.DB
	dir    string
	retain int
	dbPath string // original DB file path (for display only)
}

// NewService creates a new backup Service.
// dir defaults to "./backups" when empty.
// retain defaults to 7 when <= 0.
func NewService(db *sqlx.DB, dbPath, dir string, retain int) *Service {
	if dir == "" {
		if v := os.Getenv("BACKUP_DIR"); v != "" {
			dir = v
		} else {
			dir = "./backups"
		}
	}
	if retain <= 0 {
		retain = 7
		if v := os.Getenv("BACKUP_RETAIN"); v != "" {
			fmt.Sscanf(v, "%d", &retain)
		}
	}
	return &Service{
		db:     db,
		dir:    dir,
		retain: retain,
		dbPath: dbPath,
	}
}

// Run ticks every 12 hours and calls CreateBackup until ctx is cancelled.
func (s *Service) Run(ctx context.Context) {
	ticker := time.NewTicker(12 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if path, err := s.CreateBackup(ctx); err != nil {
				log.Printf("backup: error creating backup: %v", err)
			} else {
				log.Printf("backup: created %s", path)
			}
		}
	}
}

// CreateBackup copies the live SQLite database to a timestamped file using
// VACUUM INTO, which is the only concurrent-safe approach for SQLite.
// It returns the absolute path of the new backup file.
func (s *Service) CreateBackup(ctx context.Context) (string, error) {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return "", fmt.Errorf("backup: create dir %s: %w", s.dir, err)
	}

	stamp := time.Now().UTC().Format("20060102_150405")
	name := fmt.Sprintf("hube_backup_%s.db", stamp)
	dest := filepath.Join(s.dir, name)

	if _, err := s.db.ExecContext(ctx, fmt.Sprintf("VACUUM INTO '%s'", dest)); err != nil {
		return "", fmt.Errorf("backup: VACUUM INTO %s: %w", dest, err)
	}

	if err := s.pruneOld(); err != nil {
		log.Printf("backup: prune old backups: %v", err)
	}

	abs, err := filepath.Abs(dest)
	if err != nil {
		abs = dest
	}
	return abs, nil
}

// pruneOld removes the oldest backup files when the count exceeds s.retain.
func (s *Service) pruneOld() error {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return fmt.Errorf("read backup dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if matched, _ := filepath.Match("hube_backup_*.db", e.Name()); matched {
			files = append(files, filepath.Join(s.dir, e.Name()))
		}
	}

	sort.Strings(files) // ascending by name = ascending by timestamp

	for len(files) > s.retain {
		if err := os.Remove(files[0]); err != nil {
			return fmt.Errorf("remove %s: %w", files[0], err)
		}
		files = files[1:]
	}
	return nil
}
