package subscribe

import (
	"github.com/grulex/go-wishlist/pkg/user"
	"github.com/grulex/go-wishlist/pkg/wishlist"
)

type Subscribe struct {
	UserID    user.ID
	ProfileID wishlist.ID
}
