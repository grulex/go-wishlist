package inmemory

import (
	"context"
	"github.com/grulex/go-wishlist/pkg/image"
)

type Storage struct {
	Images map[image.ID]image.Image
}

func New() *Storage {
	return &Storage{
		Images: map[image.ID]image.Image{},
	}
}

func (s *Storage) Create(_ context.Context, image image.Image) error {
	s.Images[image.ID] = image
	return nil
}

func (s *Storage) Get(_ context.Context, id image.ID) (image.Image, error) {
	return s.Images[id], nil
}
