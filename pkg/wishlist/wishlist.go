package wishlist

import (
	"github.com/grulex/go-wishlist/pkg/image"
	"github.com/grulex/go-wishlist/pkg/product"
	"github.com/grulex/go-wishlist/pkg/user"
	"time"
)

type ID string

type Wishlist struct {
	ID          ID
	UserID      user.ID
	IsDefault   bool
	Title       string
	Avatar      *image.ID
	Description string
	IsArchived  bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Item struct {
	ID                 ItemID
	IsBookingAvailable bool
	IsBookedBy         *user.ID
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type ItemID struct {
	WishlistID ID         `json:"wishlist_id"`
	ProductID  product.ID `json:"product_id"`
}
