package event

import "time"

type Event struct {
	ID          string    `db:"id"          json:"id"`
	Title       string    `db:"title"        json:"title"`
	Description string    `db:"description"  json:"description"`
	StartAt     time.Time `db:"start_at"     json:"start_at"`
	EndAt       time.Time `db:"end_at"       json:"end_at"`
	AllDay      bool      `db:"all_day"      json:"all_day"`
	Color       string    `db:"color"        json:"color"`
	CreatedAt   time.Time `db:"created_at"   json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"   json:"updated_at"`
}
