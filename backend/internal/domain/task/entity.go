package task

import "time"

type Priority string
type Status string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"

	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

// Recurrence values: "", "daily", "weekly", "monthly"
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
