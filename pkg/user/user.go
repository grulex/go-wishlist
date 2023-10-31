package user

import (
	"errors"
	"github.com/grulex/go-wishlist/pkg/notify"
	"time"
)

var ErrNotFound = errors.New("user not found")

type ID string
type Language string

type User struct {
	ID              ID
	FullName        string
	Language        Language
	NotifyType      *notify.Type
	NotifyChannelID *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
