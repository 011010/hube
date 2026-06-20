package diagram

import "time"

type Diagram struct {
	ID        string    `db:"id"         json:"id"`
	Name      string    `db:"name"       json:"name"`
	Nodes     string    `db:"nodes"      json:"nodes"`
	Edges     string    `db:"edges"      json:"edges"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
