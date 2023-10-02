package service

import (
	"context"
	"github.com/google/uuid"
	imagePkg "github.com/grulex/go-wishlist/pkg/image"
	"time"
)

type storage interface {
	Create(ctx context.Context, image imagePkg.Image) error
	Get(ctx context.Context, id imagePkg.ID) (imagePkg.Image, error)
}

type Service struct {
	storage storage
}

func New(storage storage) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) Create(ctx context.Context, image imagePkg.Image) error {
	image.ID = imagePkg.ID(uuid.NewString())
	image.CreatedAt = time.Now().UTC()
	return s.storage.Create(ctx, image)
}

func (s *Service) Get(ctx context.Context, id imagePkg.ID) (imagePkg.Image, error) {
	return s.storage.Get(ctx, id)
}
