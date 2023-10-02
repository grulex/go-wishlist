package user

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("user not found")

type ID string

type User struct {
	ID        ID
	FullName  string
	CreatedAt time.Time
}
