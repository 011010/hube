package project

import "time"

type Status string

const (
	StatusPlanning  Status = "planning"
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
	StatusOnHold    Status = "on_hold"
)

type Project struct {
	ID             string    `json:"id"              db:"id"`
	Name           string    `json:"name"            db:"name"`
	Description    string    `json:"description"     db:"description"`
	Status         Status    `json:"status"          db:"status"`
	Color          string    `json:"color"           db:"color"`
	DueDate        *string   `json:"due_date"        db:"due_date"`
	NoteID         *string   `json:"note_id"         db:"note_id"`
	TaskCount      int       `json:"task_count"      db:"task_count"`
	CompletedCount int       `json:"completed_count" db:"completed_count"`
	CreatedAt      time.Time `json:"created_at"      db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"      db:"updated_at"`
}

func (p *Project) Progress() int {
	if p.TaskCount == 0 {
		return 0
	}
	return (p.CompletedCount * 100) / p.TaskCount
}
