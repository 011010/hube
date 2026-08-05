package note

import (
	"fmt"
	"strings"
	"time"
)

const (
	StatusDraft      = "draft"
	StatusInProgress = "in_progress"
	StatusPublished  = "published"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type Note struct {
	ID        string    `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	Blocks    string    `json:"blocks" db:"blocks"`
	Status    string    `json:"status" db:"status"`
	FolderID  *string   `json:"folder_id" db:"folder_id"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (n *Note) Normalize() {
	n.Title = strings.TrimSpace(n.Title)
	if n.Status == "" {
		n.Status = StatusDraft
	}
}

func (n *Note) Validate() error {
	if n.Title == "" {
		return &ValidationError{Field: "title", Message: "title is required"}
	}
	switch n.Status {
	case StatusDraft, StatusInProgress, StatusPublished:
	default:
		return &ValidationError{Field: "status", Message: fmt.Sprintf("invalid status: %s", n.Status)}
	}
	return nil
}
