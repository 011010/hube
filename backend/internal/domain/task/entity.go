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

type Task struct {
	ID          string     `db:"id"`
	Title       string     `db:"title"`
	Description string     `db:"description"`
	Priority    Priority   `db:"priority"`
	Status      Status     `db:"status"`
	DueDate     *time.Time `db:"due_date"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}
