package wish

import (
	"github.com/grulex/go-wishlist/pkg/eventmanager"
	"github.com/grulex/go-wishlist/pkg/user"
	"github.com/grulex/go-wishlist/pkg/wishlist"
	"time"
)

const (
	EventWishBookingUpdate eventmanager.EventName = "wish.booking.update"
)

func NewBookingUpdateEvent(payload BookingPayload) eventmanager.Event {
	return event{
		name:    EventWishBookingUpdate,
		payload: payload,
	}
}

type BookingPayload struct {
	ItemID      wishlist.ItemID
	WishOwner   user.ID
	OldBookedBy *user.ID
	NewBookedBy *user.ID
	EventBy     user.ID
	EventAt     time.Time
}

type event struct {
	name    eventmanager.EventName
	payload BookingPayload
}

func (e event) GetName() eventmanager.EventName {
	return e.name
}

func (e event) GetPayload() eventmanager.Payload {
	return e.payload
}
