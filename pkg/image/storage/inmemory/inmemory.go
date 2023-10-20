package inmemory

import (
	"context"
	"github.com/grulex/go-wishlist/pkg/image"
)

type Storage struct {
	Images map[image.ID]*image.Image
}

func NewImageInMemory() *Storage {
	return &Storage{
		Images: map[image.ID]*image.Image{},
	}
}

func (s *Storage) Upsert(_ context.Context, image *image.Image) error {
	s.Images[image.ID] = image
	return nil
}

func (s *Storage) Get(_ context.Context, id image.ID) (*image.Image, error) {
	return s.Images[id], nil
}

func (s *Storage) GetMany(_ context.Context, ids []image.ID) ([]*image.Image, error) {
	var images []*image.Image
	for _, id := range ids {
		images = append(images, s.Images[id])
	}
	return images, nil
}
