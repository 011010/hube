package task

import (
	"fmt"
	"strings"
	"time"
)

type Priority string
type Status string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"

	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"

	// Recurrence values the scheduler knows how to advance. Any other
	// value makes isDue return false, so the task would silently never
	// repeat; Validate rejects them instead.
	RecurrenceNone    = ""
	RecurrenceDaily   = "daily"
	RecurrenceWeekly  = "weekly"
	RecurrenceMonthly = "monthly"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type Task struct {
	ID             string     `db:"id"               json:"id"`
	Title          string     `db:"title"            json:"title"`
	Description    string     `db:"description"      json:"description"`
	Priority       Priority   `db:"priority"         json:"priority"`
	Status         Status     `db:"status"           json:"status"`
	DueDate        *time.Time `db:"due_date"         json:"due_date"`
	ProjectID      *string    `db:"project_id"       json:"project_id"`
	NoteID         *string    `db:"note_id"          json:"note_id"`
	Recurrence     string     `db:"recurrence"       json:"recurrence"`
	LastRecurredAt *time.Time `db:"last_recurred_at" json:"last_recurred_at,omitempty"`
	CreatedAt      time.Time  `db:"created_at"       json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"       json:"updated_at"`
}

// Normalize fills in the defaults a partially specified task is allowed to
// omit and trims the free-text fields, so Validate compares against clean
// values.
func (t *Task) Normalize() {
	t.Title = strings.TrimSpace(t.Title)
	t.Recurrence = strings.TrimSpace(t.Recurrence)
	if t.Status == "" {
		t.Status = StatusTodo
	}
	if t.Priority == "" {
		t.Priority = PriorityMedium
	}
}

// Validate reports the first field that would make the task unusable. Call
// Normalize first: Validate does not trim.
func (t *Task) Validate() error {
	if t.Title == "" {
		return &ValidationError{Field: "title", Message: "title is required"}
	}
	switch t.Status {
	case StatusTodo, StatusInProgress, StatusDone:
	default:
		return &ValidationError{Field: "status", Message: fmt.Sprintf("invalid status: %s", t.Status)}
	}
	switch t.Priority {
	case PriorityLow, PriorityMedium, PriorityHigh:
	default:
		return &ValidationError{Field: "priority", Message: fmt.Sprintf("invalid priority: %s", t.Priority)}
	}
	switch t.Recurrence {
	case RecurrenceNone, RecurrenceDaily, RecurrenceWeekly, RecurrenceMonthly:
	default:
		return &ValidationError{Field: "recurrence", Message: fmt.Sprintf("invalid recurrence: %s", t.Recurrence)}
	}
	return nil
}
