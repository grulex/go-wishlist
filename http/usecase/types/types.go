package types

import (
	"github.com/bojanz/currency"
	"github.com/grulex/go-wishlist/pkg/image"
	productPkg "github.com/grulex/go-wishlist/pkg/product"
	"github.com/grulex/go-wishlist/pkg/user"
	"github.com/grulex/go-wishlist/pkg/wishlist"
	"gopkg.in/guregu/null.v4"
)

type User struct {
	ID user.ID `json:"id"`
}

type Wishlist struct {
	ID          wishlist.ID `json:"id"`
	IsDefault   bool        `json:"is_default"`
	Title       string      `json:"title"`
	Avatar      *Image      `json:"avatar,omitempty"`
	Description string      `json:"description"`
}

type Item struct {
	ID                    wishlist.ItemID `json:"id"`
	IsBookingAvailable    bool            `json:"is_booking_available"`
	IsBookedByCurrentUser bool            `json:"is_booked_by_current_user"`
	Product               Product         `json:"product"`
}

type Product struct {
	ID          *productPkg.ID   `json:"id,omitempty"`
	Title       string           `json:"title"`
	PriceFrom   *currency.Amount `json:"price_from,omitempty"`
	PriceTo     *currency.Amount `json:"price_to,omitempty"`
	Description null.String      `json:"description,omitempty"`
	Url         null.String      `json:"url,omitempty"`
	Image       *Image           `json:"image,omitempty"`
}

type Image struct {
	ID   image.ID `json:"id"`
	Link string   `json:"link,omitempty"`
}

type Subscribe struct {
	ID wishlist.ID `json:"id"`
}
