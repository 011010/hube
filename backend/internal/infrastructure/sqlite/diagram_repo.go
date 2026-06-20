package sqlite

import (
	"context"
	"time"

	"github.com/husari/hube/internal/domain/diagram"
	"github.com/jmoiron/sqlx"
)

type DiagramRepo struct{ db *sqlx.DB }

func NewDiagramRepo(db *sqlx.DB) *DiagramRepo { return &DiagramRepo{db: db} }

func (r *DiagramRepo) FindAll(ctx context.Context) ([]diagram.Diagram, error) {
	items := make([]diagram.Diagram, 0)
	err := r.db.SelectContext(ctx, &items, `SELECT * FROM diagrams ORDER BY updated_at DESC`)
	return items, err
}

func (r *DiagramRepo) FindByID(ctx context.Context, id string) (*diagram.Diagram, error) {
	var d diagram.Diagram
	err := r.db.GetContext(ctx, &d, `SELECT * FROM diagrams WHERE id = ?`, id)
	return &d, err
}

func (r *DiagramRepo) Create(ctx context.Context, d *diagram.Diagram) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO diagrams (id, name, nodes, edges, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		d.ID, d.Name, d.Nodes, d.Edges, d.CreatedAt, d.UpdatedAt,
	)
	return err
}

func (r *DiagramRepo) Update(ctx context.Context, d *diagram.Diagram) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE diagrams SET name=?, nodes=?, edges=?, updated_at=? WHERE id=?`,
		d.Name, d.Nodes, d.Edges, time.Now(), d.ID,
	)
	return err
}

func (r *DiagramRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM diagrams WHERE id=?`, id)
	return err
}
