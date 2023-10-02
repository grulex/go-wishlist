package wishlist

import "github.com/grulex/go-wishlist/pkg/user"

type Type string
type ID string

const (
	TypePersonal Type = "personal"
	TypePublic   Type = "public"
)

type Profile struct {
	ID          ID
	UserID      user.ID
	Type        Type
	Title       string
	AvatarURL   string
	Description string
}
