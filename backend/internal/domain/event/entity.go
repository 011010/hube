package event

import "time"

type Event struct {
	ID          string    `db:"id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	StartAt     time.Time `db:"start_at"`
	EndAt       time.Time `db:"end_at"`
	AllDay      bool      `db:"all_day"`
	Color       string    `db:"color"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
