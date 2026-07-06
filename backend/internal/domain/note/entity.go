package note

import "time"

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
