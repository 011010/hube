package sqlite

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/husari/hube/internal/domain/folder"
	"github.com/jmoiron/sqlx"
)

type FolderRepo struct{ db *sqlx.DB }

func NewFolderRepo(db *sqlx.DB) *FolderRepo { return &FolderRepo{db: db} }

func (r *FolderRepo) FindAll(ctx context.Context) ([]folder.Folder, error) {
	folders := make([]folder.Folder, 0)
	return folders, r.db.SelectContext(ctx, &folders, `SELECT id,name,parent_id,created_at,updated_at FROM folders ORDER BY name`)
}

func (r *FolderRepo) Create(ctx context.Context, f *folder.Folder) error {
	f.ID = uuid.NewString()
	now := time.Now()
	f.CreatedAt = now
	f.UpdatedAt = now
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO folders(id,name,parent_id,created_at,updated_at) VALUES(?,?,?,?,?)`,
		f.ID, f.Name, f.ParentID, f.CreatedAt, f.UpdatedAt,
	)
	return err
}

func (r *FolderRepo) Update(ctx context.Context, f *folder.Folder) error {
	f.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE folders SET name=?,parent_id=?,updated_at=? WHERE id=?`,
		f.Name, f.ParentID, f.UpdatedAt, f.ID,
	)
	return err
}

func (r *FolderRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM folders WHERE id=?`, id)
	return err
}
