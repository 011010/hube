package note

import (
	"errors"
	"fmt"
	"time"
)

const (
	StatusDraft      = "draft"
	StatusInProgress = "in_progress"
	StatusPublished  = "published"

	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"
)

type Note struct {
	ID        string    `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	Blocks    string    `json:"blocks" db:"blocks"`
	Status    string    `json:"status" db:"status"`
	Priority  string    `json:"priority" db:"priority"`
	DueDate   *string   `json:"due_date" db:"due_date"`
	FolderID  *string   `json:"folder_id" db:"folder_id"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (n *Note) Normalize() {
	if n.Status == "" {
		n.Status = StatusDraft
	}
	if n.Priority == "" {
		n.Priority = PriorityMedium
	}
}

func (n *Note) Validate() error {
	if n.Title == "" {
		return errors.New("title is required")
	}
	switch n.Status {
	case StatusDraft, StatusInProgress, StatusPublished:
	default:
		return fmt.Errorf("invalid status: %s", n.Status)
	}
	switch n.Priority {
	case PriorityLow, PriorityMedium, PriorityHigh:
	default:
		return fmt.Errorf("invalid priority: %s", n.Priority)
	}
	if n.DueDate != nil {
		if _, err := time.Parse("2006-01-02", *n.DueDate); err != nil {
			return fmt.Errorf("invalid due_date: %w", err)
		}
	}
	return nil
}
