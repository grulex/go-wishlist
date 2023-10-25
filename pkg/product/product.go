package product

import (
	"errors"
	"github.com/bojanz/currency"
	"github.com/grulex/go-wishlist/pkg/image"
	"gopkg.in/guregu/null.v4"
	"time"
)

var ErrNotFound = errors.New("product not found")

const MaxTitleLength = 40

type ID string

type Product struct {
	ID          ID
	Title       string
	ImageID     *image.ID
	Price       *currency.Amount
	Description null.String
	Url         null.String
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
