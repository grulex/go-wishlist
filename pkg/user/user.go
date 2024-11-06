package user

import (
	"errors"
	"github.com/grulex/go-wishlist/pkg/notify"
	"strconv"
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

type Stats struct {
	Day   string
	Count int
}

func (s *Stats) String() string {
	return s.Day + ": " + strconv.Itoa(s.Count)
}
