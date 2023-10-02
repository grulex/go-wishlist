package user

import "gopkg.in/guregu/null.v4"

type ID string

type User struct {
	ID       ID
	FullName string
	Email    null.String
}
