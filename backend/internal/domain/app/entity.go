package app

type App struct {
	ID          string `db:"id"          json:"id"`
	Name        string `db:"name"         json:"name"`
	Description string `db:"description"  json:"description"`
	URL         string `db:"url"          json:"url"`
	Icon        string `db:"icon"         json:"icon"`
	Color       string `db:"color"        json:"color"`
	SortOrder   int    `db:"sort_order"   json:"sort_order"`
	Active      bool   `db:"active"       json:"active"`
}
