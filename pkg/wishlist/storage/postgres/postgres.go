package postgres

import (
	"context"
	"database/sql"
	"errors"
	productPkg "github.com/grulex/go-wishlist/pkg/product"
	userPkg "github.com/grulex/go-wishlist/pkg/user"
	wishlistPkg "github.com/grulex/go-wishlist/pkg/wishlist"
	"github.com/jmoiron/sqlx"
	"time"
)

type wishlistPersistent struct {
	ID          string    `db:"id"`
	UserID      string    `db:"user_id"`
	IsDefault   bool      `db:"is_default"`
	Title       string    `db:"title"`
	ImageId     *string   `db:"image_id"`
	Description string    `db:"description"`
	IsArchived  bool      `db:"is_archived"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type itemPersistent struct {
	WishlistID         string    `db:"wishlist_id"`
	ProductID          string    `db:"product_id"`
	IsBookingAvailable bool      `db:"is_booking_available"`
	IsBookedBy         *string   `db:"is_booked_by"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

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
	wishlistPersistent := wishlistPersistent{
		ID:          string(w.ID),
		UserID:      string(w.UserID),
		IsDefault:   w.IsDefault,
		Title:       w.Title,
		ImageId:     nil,
		Description: w.Description,
		IsArchived:  w.IsArchived,
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
	}
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
	wishlist := &wishlistPkg.Wishlist{
		ID:          wishlistPkg.ID(w.ID),
		UserID:      userPkg.ID(w.UserID),
		IsDefault:   w.IsDefault,
		Title:       w.Title,
		Avatar:      nil,
		Description: w.Description,
		IsArchived:  w.IsArchived,
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
	}

	return wishlist, nil
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
		wishlists = append(wishlists, &wishlistPkg.Wishlist{
			ID:          wishlistPkg.ID(w.ID),
			UserID:      userPkg.ID(w.UserID),
			IsDefault:   w.IsDefault,
			Title:       w.Title,
			Avatar:      nil,
			Description: w.Description,
			IsArchived:  w.IsArchived,
			CreatedAt:   w.CreatedAt,
			UpdatedAt:   w.UpdatedAt,
		})
	}

	return wishlists, nil
}

func (s *Storage) GetWishlistItems(ctx context.Context, wishlistID wishlistPkg.ID, limit, offset uint) (items []*wishlistPkg.Item, haveMore bool, err error) {
	itemsPersistent := make([]*itemPersistent, 0)
	query := `SELECT * FROM wishlist_item WHERE wishlist_id = $1 LIMIT $2 OFFSET $3`
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
		var isBookedBy *userPkg.ID
		if itemPersistent.IsBookedBy != nil {
			isBookedBy = new(userPkg.ID)
			*isBookedBy = userPkg.ID(*itemPersistent.IsBookedBy)
		}
		items = append(items, &wishlistPkg.Item{
			ID:                 wishlistPkg.ItemID{WishlistID: wishlistID, ProductID: productPkg.ID(itemPersistent.ProductID)},
			IsBookingAvailable: itemPersistent.IsBookingAvailable,
			IsBookedBy:         isBookedBy,
			CreatedAt:          itemPersistent.CreatedAt,
			UpdatedAt:          itemPersistent.UpdatedAt,
		})
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

	isBookedBy := (*string)(nil)
	if item.IsBookedBy != nil {
		isBookedBy = new(string)
		*isBookedBy = string(*item.IsBookedBy)
	}
	itemPersistent := itemPersistent{
		WishlistID:         string(item.ID.WishlistID),
		ProductID:          string(item.ID.ProductID),
		IsBookingAvailable: item.IsBookingAvailable,
		IsBookedBy:         isBookedBy,
		CreatedAt:          item.CreatedAt,
		UpdatedAt:          item.UpdatedAt,
	}
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
	var isBookedBy *userPkg.ID
	if itemPersistent.IsBookedBy != nil {
		isBookedBy = new(userPkg.ID)
		*isBookedBy = userPkg.ID(*itemPersistent.IsBookedBy)
	}
	item := &wishlistPkg.Item{
		ID:                 itemID,
		IsBookingAvailable: itemPersistent.IsBookingAvailable,
		IsBookedBy:         isBookedBy,
		CreatedAt:          itemPersistent.CreatedAt,
		UpdatedAt:          itemPersistent.UpdatedAt,
	}
	return item, nil
}
