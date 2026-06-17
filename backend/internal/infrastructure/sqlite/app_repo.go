package sqlite

import (
	"context"

	"github.com/husari/hube/internal/domain/app"
	"github.com/jmoiron/sqlx"
)

type AppRepo struct{ db *sqlx.DB }

func NewAppRepo(db *sqlx.DB) *AppRepo { return &AppRepo{db: db} }

func (r *AppRepo) FindAll(ctx context.Context) ([]app.App, error) {
	apps := make([]app.App, 0)
	err := r.db.SelectContext(ctx, &apps, `SELECT * FROM apps WHERE active=1 ORDER BY sort_order ASC`)
	return apps, err
}

func (r *AppRepo) FindByID(ctx context.Context, id string) (*app.App, error) {
	var a app.App
	err := r.db.GetContext(ctx, &a, `SELECT * FROM apps WHERE id=?`, id)
	return &a, err
}

func (r *AppRepo) Create(ctx context.Context, a *app.App) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO apps (id, name, description, url, icon, color, sort_order, active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.Name, a.Description, a.URL, a.Icon, a.Color, a.SortOrder, a.Active,
	)
	return err
}

func (r *AppRepo) Update(ctx context.Context, a *app.App) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE apps SET name=?, description=?, url=?, icon=?, color=?, sort_order=?, active=?
		WHERE id=?`,
		a.Name, a.Description, a.URL, a.Icon, a.Color, a.SortOrder, a.Active, a.ID,
	)
	return err
}

func (r *AppRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM apps WHERE id=?`, id)
	return err
}
