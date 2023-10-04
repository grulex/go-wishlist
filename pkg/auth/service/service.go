package service

import (
	"context"
	"github.com/grulex/go-wishlist/pkg/auth"
	"time"
)

type storage interface {
	Upsert(ctx context.Context, auth *auth.Auth) error
	Get(ctx context.Context, method auth.Method, socialID auth.SocialID) (*auth.Auth, error)
}

type Service struct {
	storage storage
}

func NewAuthService(storage storage) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) Create(ctx context.Context, auth *auth.Auth) error {
	auth.CreatedAt = time.Now().UTC()
	auth.UpdatedAt = auth.CreatedAt
	return s.storage.Upsert(ctx, auth)
}

func (s *Service) Get(ctx context.Context, method auth.Method, socialID auth.SocialID) (*auth.Auth, error) {
	return s.storage.Get(ctx, method, socialID)
}
