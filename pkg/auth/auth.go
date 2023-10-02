package auth

import (
	"github.com/grulex/go-wishlist/pkg/user"
	"gopkg.in/guregu/null.v4"
)

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
