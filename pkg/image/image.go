package image

import (
	"errors"
	"github.com/grulex/go-wishlist/pkg/file"
	"time"
)

var ErrNotFound = errors.New("image not found")

type ID string

type Image struct {
	ID        ID
	FileLink  file.Link
	Width     uint
	Height    uint
	Hash      Hash
	Sizes     []Size
	CreatedAt time.Time
}

type Hash struct {
	AHash string
	DHash string
	PHash string
}

type Size struct {
	Width    uint
	Height   uint
	FileLink file.Link
}
