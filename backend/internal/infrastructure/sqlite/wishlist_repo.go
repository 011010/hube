package sqlite

import (
	"context"
	"time"

	"github.com/husari/hube/internal/domain/wishlist"
	"github.com/jmoiron/sqlx"
)

type WishlistRepo struct{ db *sqlx.DB }

func NewWishlistRepo(db *sqlx.DB) *WishlistRepo { return &WishlistRepo{db: db} }

func (r *WishlistRepo) FindAll(ctx context.Context) ([]wishlist.Item, error) {
	items := make([]wishlist.Item, 0)
	err := r.db.SelectContext(ctx, &items, `SELECT * FROM wishlist_items ORDER BY created_at DESC`)
	return items, err
}

func (r *WishlistRepo) FindByID(ctx context.Context, id string) (*wishlist.Item, error) {
	var item wishlist.Item
	err := r.db.GetContext(ctx, &item, `SELECT * FROM wishlist_items WHERE id = ?`, id)
	return &item, err
}

func (r *WishlistRepo) Create(ctx context.Context, item *wishlist.Item) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO wishlist_items
			(id, name, description, url, store, target_price, current_price, currency, priority, status, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.Name, item.Description, item.URL, item.Store,
		item.TargetPrice, item.CurrentPrice, item.Currency,
		item.Priority, item.Status, item.Notes,
		item.CreatedAt, item.UpdatedAt,
	)
	return err
}

func (r *WishlistRepo) Update(ctx context.Context, item *wishlist.Item) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE wishlist_items SET
			name=?, description=?, url=?, store=?, target_price=?, current_price=?,
			currency=?, priority=?, status=?, notes=?, updated_at=?
		WHERE id=?`,
		item.Name, item.Description, item.URL, item.Store,
		item.TargetPrice, item.CurrentPrice, item.Currency,
		item.Priority, item.Status, item.Notes,
		time.Now(), item.ID,
	)
	return err
}

func (r *WishlistRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM wishlist_items WHERE id=?`, id)
	return err
}
