package user

import "errors"

var ErrNotFound = errors.New("user not found")

type ID string

type User struct {
	ID       ID
	FullName string
}
