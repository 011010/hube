package sqlite

import (
	"context"
	"time"

	"github.com/husari/hube/internal/domain/task"
	"github.com/jmoiron/sqlx"
)

type TaskRepo struct{ db *sqlx.DB }

func NewTaskRepo(db *sqlx.DB) *TaskRepo { return &TaskRepo{db: db} }

func (r *TaskRepo) FindAll(ctx context.Context) ([]task.Task, error) {
	var tasks []task.Task
	err := r.db.SelectContext(ctx, &tasks, `SELECT * FROM tasks ORDER BY created_at DESC`)
	return tasks, err
}

func (r *TaskRepo) FindByID(ctx context.Context, id string) (*task.Task, error) {
	var t task.Task
	err := r.db.GetContext(ctx, &t, `SELECT * FROM tasks WHERE id = ?`, id)
	return &t, err
}

func (r *TaskRepo) Create(ctx context.Context, t *task.Task) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tasks (id, title, description, priority, status, due_date, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.Title, t.Description, t.Priority, t.Status, t.DueDate, t.CreatedAt, t.UpdatedAt,
	)
	return err
}

func (r *TaskRepo) Update(ctx context.Context, t *task.Task) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE tasks SET title=?, description=?, priority=?, status=?, due_date=?, updated_at=?
		WHERE id=?`,
		t.Title, t.Description, t.Priority, t.Status, t.DueDate, time.Now(), t.ID,
	)
	return err
}

func (r *TaskRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM tasks WHERE id=?`, id)
	return err
}
