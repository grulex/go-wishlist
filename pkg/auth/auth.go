package auth

import (
	"errors"
	"github.com/grulex/go-wishlist/pkg/user"
	"gopkg.in/guregu/null.v4"
)

var ErrNotFound = errors.New("auth not found")

type Method string

const (
	MethodTelegram Method = "telegram"
)

type SocialID null.String

type Auth struct {
	UserID   user.ID
	Method   Method
	SocialID SocialID
}
