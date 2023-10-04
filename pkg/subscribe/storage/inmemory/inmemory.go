package inmemory

import (
	"context"
	subscribePkg "github.com/grulex/go-wishlist/pkg/subscribe"
	"github.com/grulex/go-wishlist/pkg/user"
	"github.com/grulex/go-wishlist/pkg/wishlist"
)

type Storage struct {
	subscribes map[user.ID]map[wishlist.ID]*subscribePkg.Subscribe
}

func NewSubscribeInMemory() *Storage {
	return &Storage{
		subscribes: map[user.ID]map[wishlist.ID]*subscribePkg.Subscribe{},
	}
}

func (s *Storage) Upsert(_ context.Context, subscribe *subscribePkg.Subscribe) error {
	if _, ok := s.subscribes[subscribe.UserID]; !ok {
		s.subscribes[subscribe.UserID] = map[wishlist.ID]*subscribePkg.Subscribe{}
	}
	s.subscribes[subscribe.UserID][subscribe.WishlistID] = subscribe
	return nil
}

func (s *Storage) Get(_ context.Context, userID user.ID, wishlistID wishlist.ID) (*subscribePkg.Subscribe, error) {
	if _, ok := s.subscribes[userID]; !ok {
		return nil, subscribePkg.ErrNotFound
	}
	if _, ok := s.subscribes[userID][wishlistID]; !ok {
		return nil, subscribePkg.ErrNotFound
	}
	subscribe := s.subscribes[userID][wishlistID]
	return subscribe, nil
}

func (s *Storage) GetByUser(_ context.Context, userID user.ID) ([]*subscribePkg.Subscribe, error) {
	var subscribes []*subscribePkg.Subscribe
	for _, subscribe := range s.subscribes[userID] {
		subscribes = append(subscribes, subscribe)
	}
	return subscribes, nil
}

func (s *Storage) Delete(_ context.Context, userID user.ID, wishlistID wishlist.ID) error {
	if _, ok := s.subscribes[userID]; !ok {
		return nil
	}
	delete(s.subscribes[userID], wishlistID)
	return nil
}
