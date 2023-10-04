package subscribe

import (
	"errors"
	"github.com/grulex/go-wishlist/pkg/user"
	"github.com/grulex/go-wishlist/pkg/wishlist"
	"time"
)

var ErrNotFound = errors.New("subscribe not found")

type Subscribe struct {
	UserID     user.ID
	WishlistID wishlist.ID
	CreatedAt  time.Time
}
