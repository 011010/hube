package sqlite

import (
	"context"
	"time"

	"github.com/husari/hube/internal/domain/event"
	"github.com/jmoiron/sqlx"
)

type EventRepo struct{ db *sqlx.DB }

func NewEventRepo(db *sqlx.DB) *EventRepo { return &EventRepo{db: db} }

func (r *EventRepo) FindAll(ctx context.Context) ([]event.Event, error) {
	var events []event.Event
	err := r.db.SelectContext(ctx, &events, `SELECT * FROM events ORDER BY start_at ASC`)
	return events, err
}

func (r *EventRepo) FindByRange(ctx context.Context, from, to time.Time) ([]event.Event, error) {
	var events []event.Event
	err := r.db.SelectContext(ctx, &events,
		`SELECT * FROM events WHERE start_at >= ? AND start_at <= ? ORDER BY start_at ASC`,
		from, to,
	)
	return events, err
}

func (r *EventRepo) FindByID(ctx context.Context, id string) (*event.Event, error) {
	var e event.Event
	err := r.db.GetContext(ctx, &e, `SELECT * FROM events WHERE id = ?`, id)
	return &e, err
}

func (r *EventRepo) Create(ctx context.Context, e *event.Event) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO events (id, title, description, start_at, end_at, all_day, color, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.Title, e.Description, e.StartAt, e.EndAt, e.AllDay, e.Color, e.CreatedAt, e.UpdatedAt,
	)
	return err
}

func (r *EventRepo) Update(ctx context.Context, e *event.Event) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE events SET title=?, description=?, start_at=?, end_at=?, all_day=?, color=?, updated_at=?
		WHERE id=?`,
		e.Title, e.Description, e.StartAt, e.EndAt, e.AllDay, e.Color, time.Now(), e.ID,
	)
	return err
}

func (r *EventRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM events WHERE id=?`, id)
	return err
}
