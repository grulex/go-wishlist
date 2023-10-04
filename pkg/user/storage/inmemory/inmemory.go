package inmemory

import (
	"context"
	"github.com/grulex/go-wishlist/pkg/user"
	"sync"
)

type Storage struct {
	Users map[user.ID]*user.User
	Lock  *sync.RWMutex
}

func NewUserInMemory() *Storage {
	return &Storage{
		Users: map[user.ID]*user.User{},
		Lock:  &sync.RWMutex{},
	}
}

func (s *Storage) Upsert(_ context.Context, u *user.User) error {
	s.Lock.Lock()
	s.Users[u.ID] = u
	s.Lock.Unlock()
	return nil
}

func (s *Storage) Get(_ context.Context, id user.ID) (*user.User, error) {
	s.Lock.RLock()
	u, ok := s.Users[id]
	if !ok {
		return nil, user.ErrNotFound
	}
	s.Lock.RUnlock()
	return u, nil
}
