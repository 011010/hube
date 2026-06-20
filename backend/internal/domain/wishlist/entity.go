package wishlist

import "time"

type Status string
type Priority string

const (
	StatusPending   Status = "pending"
	StatusPurchased Status = "purchased"

	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type Item struct {
	ID           string    `db:"id"            json:"id"`
	Name         string    `db:"name"          json:"name"`
	Description  string    `db:"description"   json:"description"`
	URL          string    `db:"url"           json:"url"`
	Store        string    `db:"store"         json:"store"`
	TargetPrice  float64   `db:"target_price"  json:"target_price"`
	CurrentPrice float64   `db:"current_price" json:"current_price"`
	Currency     string    `db:"currency"      json:"currency"`
	Priority     Priority  `db:"priority"      json:"priority"`
	Status       Status    `db:"status"        json:"status"`
	Notes        string    `db:"notes"         json:"notes"`
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"    json:"updated_at"`
}
