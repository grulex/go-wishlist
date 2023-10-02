package product

import (
	"github.com/bojanz/currency"
	"gopkg.in/guregu/null.v4"
)

type ID string

type ImageID null.String

type Product struct {
	ID          ID
	Title       string
	ImageID     ImageID
	PriceFrom   *currency.Amount
	PriceTo     *currency.Amount
	Description null.String
	Url         null.String
}
