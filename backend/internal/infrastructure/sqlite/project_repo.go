package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/husari/hube/internal/domain/project"
	"github.com/jmoiron/sqlx"
)

type ProjectRepo struct{ db *sqlx.DB }

func NewProjectRepo(db *sqlx.DB) *ProjectRepo { return &ProjectRepo{db: db} }

const projectSelect = `
	SELECT
		p.id, p.name, p.description, p.status, p.color, p.due_date, p.note_id,
		p.created_at, p.updated_at,
		COUNT(t.id) AS task_count,
		SUM(CASE WHEN t.status = 'done' THEN 1 ELSE 0 END) AS completed_count
	FROM projects p
	LEFT JOIN tasks t ON t.project_id = p.id
`

func (r *ProjectRepo) FindAll(ctx context.Context) ([]project.Project, error) {
	projects := make([]project.Project, 0)
	err := r.db.SelectContext(ctx, &projects,
		projectSelect+` GROUP BY p.id ORDER BY p.updated_at DESC`,
	)
	return projects, err
}

func (r *ProjectRepo) FindByID(ctx context.Context, id string) (*project.Project, error) {
	var p project.Project
	err := r.db.GetContext(ctx, &p,
		projectSelect+` WHERE p.id = ? GROUP BY p.id`, id,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func (r *ProjectRepo) Create(ctx context.Context, p *project.Project) error {
	p.ID = uuid.NewString()
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO projects (id,name,description,status,color,due_date,note_id,created_at,updated_at)
		 VALUES (?,?,?,?,?,?,?,?,?)`,
		p.ID, p.Name, p.Description, p.Status, p.Color, p.DueDate, p.NoteID, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *ProjectRepo) Update(ctx context.Context, p *project.Project) error {
	p.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE projects SET name=?,description=?,status=?,color=?,due_date=?,note_id=?,updated_at=? WHERE id=?`,
		p.Name, p.Description, p.Status, p.Color, p.DueDate, p.NoteID, p.UpdatedAt, p.ID,
	)
	return err
}

func (r *ProjectRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM projects WHERE id=?`, id)
	return err
}
