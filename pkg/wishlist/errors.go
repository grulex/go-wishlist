package wishlist

import "errors"

var ErrNotFound = errors.New("wishlist not found")
var ErrItemNotFound = errors.New("wishlist item not found")
var ErrBookingNotAvailable = errors.New("item's booking not available")
var ErrItemAlreadyBooked = errors.New("item already booked")
var ErrItemBookedByAnotherUser = errors.New("item booked by another user")
