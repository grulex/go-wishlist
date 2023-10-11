package postgres

import (
	imagePkg "github.com/grulex/go-wishlist/pkg/image"
	productPkg "github.com/grulex/go-wishlist/pkg/product"
	userPkg "github.com/grulex/go-wishlist/pkg/user"
	wishlistPkg "github.com/grulex/go-wishlist/pkg/wishlist"
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

func (w wishlistPersistent) ToWishlist() *wishlistPkg.Wishlist {
	var avatar *imagePkg.ID
	if w.ImageId != nil {
		avatarID := imagePkg.ID(*w.ImageId)
		avatar = &avatarID
	}
	return &wishlistPkg.Wishlist{
		ID:          wishlistPkg.ID(w.ID),
		UserID:      userPkg.ID(w.UserID),
		IsDefault:   w.IsDefault,
		Title:       w.Title,
		Avatar:      avatar,
		Description: w.Description,
		IsArchived:  w.IsArchived,
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
	}
}

func (w wishlistPersistent) FromWishlist(wishlist *wishlistPkg.Wishlist) *wishlistPersistent {
	var avatar *string
	if wishlist.Avatar != nil {
		stringAvatar := string(*wishlist.Avatar)
		avatar = &stringAvatar
	}
	return &wishlistPersistent{
		ID:          string(wishlist.ID),
		UserID:      string(wishlist.UserID),
		IsDefault:   wishlist.IsDefault,
		Title:       wishlist.Title,
		ImageId:     avatar,
		Description: wishlist.Description,
		IsArchived:  wishlist.IsArchived,
		CreatedAt:   wishlist.CreatedAt,
		UpdatedAt:   wishlist.UpdatedAt,
	}
}

type itemPersistent struct {
	WishlistID         string    `db:"wishlist_id"`
	ProductID          string    `db:"product_id"`
	IsBookingAvailable bool      `db:"is_booking_available"`
	IsBookedBy         *string   `db:"is_booked_by"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

func (i itemPersistent) ToItem() *wishlistPkg.Item {
	var isBookedBy *userPkg.ID
	if i.IsBookedBy != nil {
		userID := userPkg.ID(*i.IsBookedBy)
		isBookedBy = &userID
	}
	return &wishlistPkg.Item{
		ID:                 wishlistPkg.ItemID{WishlistID: wishlistPkg.ID(i.WishlistID), ProductID: productPkg.ID(i.ProductID)},
		IsBookingAvailable: i.IsBookingAvailable,
		IsBookedBy:         isBookedBy,
		CreatedAt:          i.CreatedAt,
		UpdatedAt:          i.UpdatedAt,
	}
}

func (i itemPersistent) FromItem(item *wishlistPkg.Item) *itemPersistent {
	var isBookedBy *string
	if item.IsBookedBy != nil {
		userIDString := string(*item.IsBookedBy)
		isBookedBy = &userIDString
	}
	return &itemPersistent{
		WishlistID:         string(item.ID.WishlistID),
		ProductID:          string(item.ID.ProductID),
		IsBookingAvailable: item.IsBookingAvailable,
		IsBookedBy:         isBookedBy,
		CreatedAt:          item.CreatedAt,
		UpdatedAt:          item.UpdatedAt,
	}
}
