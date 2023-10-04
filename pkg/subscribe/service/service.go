package service

import (
	"context"
	"errors"
	subscribePkg "github.com/grulex/go-wishlist/pkg/subscribe"
	userPkg "github.com/grulex/go-wishlist/pkg/user"
	"github.com/grulex/go-wishlist/pkg/wishlist"
	"time"
)

type storage interface {
	Upsert(ctx context.Context, subscribe *subscribePkg.Subscribe) error
	Get(ctx context.Context, id userPkg.ID, wishlist wishlist.ID) (*subscribePkg.Subscribe, error)
	GetByUser(ctx context.Context, id userPkg.ID) ([]*subscribePkg.Subscribe, error)
	Delete(ctx context.Context, id userPkg.ID, wishlist wishlist.ID) error
}

type Service struct {
	storage storage
}

func NewSubscribeService(storage storage) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) Subscribe(ctx context.Context, userID userPkg.ID, wishlistID wishlist.ID) error {
	subscribe := &subscribePkg.Subscribe{
		UserID:     userID,
		WishlistID: wishlistID,
		CreatedAt:  time.Now().UTC(),
	}
	return s.storage.Upsert(ctx, subscribe)
}

func (s *Service) Get(ctx context.Context, userID userPkg.ID, wishlistID wishlist.ID) (*subscribePkg.Subscribe, error) {
	subscribe, err := s.storage.Get(ctx, userID, wishlistID)
	if err != nil {
		if errors.Is(err, subscribePkg.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return subscribe, nil
}

func (s *Service) GetByUser(ctx context.Context, userID userPkg.ID) ([]*subscribePkg.Subscribe, error) {
	return s.storage.GetByUser(ctx, userID)
}

func (s *Service) Unsubscribe(ctx context.Context, userID userPkg.ID, wishlistID wishlist.ID) error {
	return s.storage.Delete(ctx, userID, wishlistID)
}
