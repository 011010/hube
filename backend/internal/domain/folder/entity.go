package folder

import "time"

type Folder struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	ParentID  *string   `json:"parent_id" db:"parent_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
