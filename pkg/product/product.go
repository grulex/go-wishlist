package product

import (
	"errors"
	"github.com/bojanz/currency"
	"github.com/grulex/go-wishlist/pkg/image"
	"gopkg.in/guregu/null.v4"
)

var ErrNotFound = errors.New("product not found")

type ID string

type Product struct {
	ID          ID
	Title       string
	ImageID     image.ID
	PriceFrom   *currency.Amount
	PriceTo     *currency.Amount
	Description null.String
	Url         null.String
}
