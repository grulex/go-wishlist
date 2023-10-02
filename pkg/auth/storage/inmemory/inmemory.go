package inmemory

import (
	"context"
	"github.com/grulex/go-wishlist/pkg/auth"
	"sync"
)

type Storage struct {
	Auths map[auth.Method]map[auth.SocialID]auth.Auth
	Lock  *sync.RWMutex
}

func New() *Storage {
	return &Storage{
		Auths: map[auth.Method]map[auth.SocialID]auth.Auth{},
	}
}

func (s *Storage) Upsert(_ context.Context, a auth.Auth) error {
	s.Lock.Lock()
	if _, ok := s.Auths[a.Method]; !ok {
		s.Auths[a.Method] = map[auth.SocialID]auth.Auth{}
	}
	s.Auths[a.Method][a.SocialID] = a
	s.Lock.Unlock()
	return nil
}

func (s *Storage) Get(_ context.Context, method auth.Method, socialID auth.SocialID) (auth.Auth, error) {
	byMethod, ok := s.Auths[method]
	if !ok {
		return auth.Auth{}, auth.ErrNotFound
	}
	a, ok := byMethod[socialID]
	if !ok {
		return auth.Auth{}, auth.ErrNotFound
	}
	return a, nil
}
