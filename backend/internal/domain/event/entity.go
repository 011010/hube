package event

import (
	"fmt"
	"strings"
	"time"
)

// defaultColor is the swatch an event falls back to when the client does not
// pick one. The events table declares the same default, but the repository
// always writes the column explicitly, so the SQL default never applies.
const defaultColor = "#6366f1"

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

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

// Normalize fills in the defaults a partially specified event is allowed to
// omit and trims the free-text fields, so Validate compares against clean
// values.
func (e *Event) Normalize() {
	e.Title = strings.TrimSpace(e.Title)
	if e.Color == "" {
		e.Color = defaultColor
	}
}

// Validate reports the first field that would make the event unusable. Call
// Normalize first: Validate does not trim.
//
// The timestamps are checked because the calendar range query filters on
// start_at alone. An event left at the zero time sits in year 1, outside
// every range a user will ever browse, so it can be created but never seen
// or deleted from the calendar.
func (e *Event) Validate() error {
	if e.Title == "" {
		return &ValidationError{Field: "title", Message: "title is required"}
	}
	if e.StartAt.IsZero() {
		return &ValidationError{Field: "start_at", Message: "start_at is required"}
	}
	if e.EndAt.IsZero() {
		return &ValidationError{Field: "end_at", Message: "end_at is required"}
	}
	// Equal timestamps are allowed: a reminder marks an instant, not a span.
	if e.EndAt.Before(e.StartAt) {
		return &ValidationError{Field: "end_at", Message: "end_at must not be before start_at"}
	}
	return nil
}
