package app

type App struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	URL         string `db:"url"`
	Icon        string `db:"icon"`
	Color       string `db:"color"`
	SortOrder   int    `db:"sort_order"`
	Active      bool   `db:"active"`
}
