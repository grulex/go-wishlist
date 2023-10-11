package service

import (
	"context"
	"github.com/grulex/go-wishlist/pkg/auth"
	"github.com/jmoiron/sqlx"
	"time"
)

type storage interface {
	StartCreateTransaction(ctx context.Context) (*sqlx.Tx, error)
	UpsertByTransaction(ctx context.Context, tx *sqlx.Tx, a *auth.Auth) error
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

func (s *Service) MakeCreateTransaction(ctx context.Context) (*sqlx.Tx, error) {
	return s.storage.StartCreateTransaction(ctx)
}

func (s *Service) CreateByTransaction(ctx context.Context, tx *sqlx.Tx, auth *auth.Auth) error {
	auth.CreatedAt = time.Now().UTC()
	auth.UpdatedAt = auth.CreatedAt
	return s.storage.UpsertByTransaction(ctx, tx, auth)
}

func (s *Service) Get(ctx context.Context, method auth.Method, socialID auth.SocialID) (*auth.Auth, error) {
	return s.storage.Get(ctx, method, socialID)
}
