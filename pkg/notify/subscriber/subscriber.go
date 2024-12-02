package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/grulex/go-wishlist/pkg/eventmanager"
	"github.com/grulex/go-wishlist/pkg/events/wish"
)

type eventManager interface {
	Subscribe(eventName eventmanager.EventName, handler eventmanager.EventHandler)
}

type Subscriber struct {
}

func NewSubscriberForNotify() *Subscriber {
	return &Subscriber{}
}

func (s *Subscriber) Subscribe(manager eventManager) {
	manager.Subscribe(wish.EventWishBookingUpdate, s.onWishBookingUpdate())
}

func (s *Subscriber) onWishBookingUpdate() eventmanager.EventHandler {
	return func(ctx context.Context, payload json.RawMessage) error {
		var bookingPayload wish.BookingPayload
		err := json.Unmarshal(payload, &bookingPayload)
		if err != nil {
			return eventmanager.ErrInvalidPayload
		}

		if bookingPayload.NewBookedBy != nil {
			// handle booking
			if *bookingPayload.NewBookedBy != bookingPayload.WishOwner {
				// notify initiator and owner
			}
		} else if bookingPayload.OldBookedBy != nil {
			// handle unbooking
			if bookingPayload.EventBy == bookingPayload.WishOwner && bookingPayload.EventBy == *bookingPayload.OldBookedBy {
				return nil
			}
			if *bookingPayload.OldBookedBy != bookingPayload.WishOwner {
				// notify old booker about unbooking by owner
			}
			if bookingPayload.EventBy == *bookingPayload.OldBookedBy {
				return nil
			}

		}

		fmt.Printf("Call onWishBookingUpdate %+v \n", bookingPayload)

		return nil
	}
}
