package project

import (
	"fmt"
	"strings"
	"time"
)

type Status string

const (
	StatusPlanning  Status = "planning"
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
	StatusOnHold    Status = "on_hold"

	// defaultColor is the swatch a project falls back to when the client
	// does not pick one.
	defaultColor = "#6366f1"

	// dateLayout is the wire format for DueDate, which is carried as a
	// string rather than a time.Time.
	dateLayout = "2006-01-02"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

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

// Normalize fills in the defaults a partially specified project is allowed to
// omit and trims the free-text fields, so Validate compares against clean
// values.
func (p *Project) Normalize() {
	p.Name = strings.TrimSpace(p.Name)
	if p.Status == "" {
		p.Status = StatusPlanning
	}
	if p.Color == "" {
		p.Color = defaultColor
	}
	if p.DueDate != nil && strings.TrimSpace(*p.DueDate) == "" {
		p.DueDate = nil
	}
}

// Validate reports the first field that would make the project unusable. Call
// Normalize first: Validate does not trim.
//
// Color is deliberately not format-checked. Existing rows predate this
// validation and rejecting them would break editing a project that saves
// fine today.
func (p *Project) Validate() error {
	if p.Name == "" {
		return &ValidationError{Field: "name", Message: "name is required"}
	}
	switch p.Status {
	case StatusPlanning, StatusActive, StatusCompleted, StatusOnHold:
	default:
		return &ValidationError{Field: "status", Message: fmt.Sprintf("invalid status: %s", p.Status)}
	}
	if p.DueDate != nil {
		if _, err := time.Parse(dateLayout, *p.DueDate); err != nil {
			return &ValidationError{Field: "due_date", Message: fmt.Sprintf("invalid due_date: %v", err)}
		}
	}
	return nil
}

// Progress reports completion as a whole percentage, truncated. TaskCount and
// CompletedCount are filled by a SQL aggregate over the joined tasks, so
// CompletedCount never exceeds TaskCount.
func (p *Project) Progress() int {
	if p.TaskCount == 0 {
		return 0
	}
	return (p.CompletedCount * 100) / p.TaskCount
}
