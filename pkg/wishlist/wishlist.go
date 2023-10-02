package wishlist

import (
	"github.com/grulex/go-wishlist/pkg/image"
	"github.com/grulex/go-wishlist/pkg/product"
	"github.com/grulex/go-wishlist/pkg/user"
)

type ID string

type Wishlist struct {
	ID          ID
	UserID      user.ID
	IsDefault   bool
	Title       string
	Avatar      image.Image
	Description string
	IsArchived  bool
}

type Item struct {
	ID                 ItemID
	IsBookingAvailable bool
	IsBookedBy         *user.ID
}

type ItemID struct {
	WishlistID ID
	ProductID  product.ID
}
