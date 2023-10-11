package postgres

import (
	"context"
	"database/sql"
	"errors"
	userPkg "github.com/grulex/go-wishlist/pkg/user"
	wishlistPkg "github.com/grulex/go-wishlist/pkg/wishlist"
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	db *sqlx.DB
}

func NewImageStorage(db *sqlx.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) Upsert(ctx context.Context, w *wishlistPkg.Wishlist) error {
	query := `INSERT INTO wishlist (
		id,
		user_id,
		is_default,
		title,
		image_id,
		description,
		is_archived,
		created_at,
		updated_at
	) VALUES (
		:id,
		:user_id,
		:is_default,
		:title,
		:image_id,
		:description,
		:is_archived,
		:created_at,
		:updated_at
	) ON CONFLICT (id) DO UPDATE SET
		is_default = :is_default,
		title = :title,
		image_id = :image_id,
		description = :description,
		is_archived = :is_archived,
		updated_at = :updated_at`
	wishlistPersistent := wishlistPersistent{}.FromWishlist(w)
	_, err := s.db.NamedExecContext(ctx, query, wishlistPersistent)
	return err
}

func (s *Storage) Get(ctx context.Context, id wishlistPkg.ID) (*wishlistPkg.Wishlist, error) {
	query := `SELECT * FROM wishlist WHERE id = $1`
	w := &wishlistPersistent{}
	err := s.db.GetContext(ctx, w, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, wishlistPkg.ErrNotFound
		}
		return nil, err
	}

	return w.ToWishlist(), nil
}

func (s *Storage) GetByUserID(ctx context.Context, userID userPkg.ID) ([]*wishlistPkg.Wishlist, error) {
	query := `SELECT * FROM wishlist WHERE user_id = $1`
	wishlistsPersistent := make([]*wishlistPersistent, 0)
	err := s.db.SelectContext(ctx, &wishlistsPersistent, query, userID)
	if err != nil {
		return nil, err
	}
	wishlists := make([]*wishlistPkg.Wishlist, 0)
	for _, w := range wishlistsPersistent {
		wishlists = append(wishlists, w.ToWishlist())
	}

	return wishlists, nil
}

func (s *Storage) GetWishlistItems(ctx context.Context, wishlistID wishlistPkg.ID, limit, offset uint) (items []*wishlistPkg.Item, haveMore bool, err error) {
	itemsPersistent := make([]*itemPersistent, 0)
	query := `SELECT * FROM wishlist_item WHERE wishlist_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	err = s.db.SelectContext(ctx, &itemsPersistent, query, wishlistID, limit+1, offset)
	if err != nil {
		return nil, false, err
	}
	items = make([]*wishlistPkg.Item, 0, limit)
	hasMore := false
	for i, itemPersistent := range itemsPersistent {
		if i == int(limit) {
			hasMore = true
			break
		}
		items = append(items, itemPersistent.ToItem())
	}

	return items, hasMore, nil
}

func (s *Storage) UpsertWishlistItem(ctx context.Context, item *wishlistPkg.Item) error {
	query := `INSERT INTO wishlist_item (
		wishlist_id,
		product_id,
		is_booking_available,
		is_booked_by,
		created_at,
		updated_at
	) VALUES (
		:wishlist_id,
		:product_id,
		:is_booking_available,
		:is_booked_by,
		:created_at,
		:updated_at
	) ON CONFLICT (wishlist_id, product_id) DO UPDATE SET
		is_booking_available = :is_booking_available,
		is_booked_by = :is_booked_by,
		updated_at = :updated_at`

	itemPersistent := itemPersistent{}.FromItem(item)
	_, err := s.db.NamedExecContext(ctx, query, itemPersistent)
	return err
}

func (s *Storage) DeleteWishlistItem(ctx context.Context, itemID wishlistPkg.ItemID) error {
	query := `DELETE FROM wishlist_item WHERE wishlist_id = $1 AND product_id = $2`
	_, err := s.db.ExecContext(ctx, query, itemID.WishlistID, itemID.ProductID)
	return err
}

func (s *Storage) GetWishlistItemByID(ctx context.Context, itemID wishlistPkg.ItemID) (*wishlistPkg.Item, error) {
	query := `SELECT * FROM wishlist_item WHERE wishlist_id = $1 AND product_id = $2`
	itemPersistent := &itemPersistent{}
	err := s.db.GetContext(ctx, itemPersistent, query, itemID.WishlistID, itemID.ProductID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, wishlistPkg.ErrItemNotFound
		}
		return nil, err
	}

	return itemPersistent.ToItem(), nil
}
