package postgres

import (
	"context"
	subscribePkg "github.com/grulex/go-wishlist/pkg/subscribe"
	"github.com/grulex/go-wishlist/pkg/user"
	"github.com/grulex/go-wishlist/pkg/wishlist"
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	db *sqlx.DB
}

func NewSubscribeStorage(db *sqlx.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) Upsert(ctx context.Context, subscribe *subscribePkg.Subscribe) error {
	return nil
}

func (s *Storage) Get(ctx context.Context, userID user.ID, wishlistID wishlist.ID) (*subscribePkg.Subscribe, error) {
	return nil, nil
}

func (s *Storage) GetByUser(ctx context.Context, userID user.ID) ([]*subscribePkg.Subscribe, error) {
	return nil, nil
}

func (s *Storage) Delete(ctx context.Context, userID user.ID, wishlistID wishlist.ID) error {
	return nil
}
