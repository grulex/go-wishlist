package inmemory

import (
	"context"
	"github.com/grulex/go-wishlist/pkg/auth"
	"github.com/jmoiron/sqlx"
	"sync"
)

type Storage struct {
	Auths map[auth.Method]map[auth.SocialID]*auth.Auth
	Lock  *sync.RWMutex
}

func NewAuthInMemory() *Storage {
	return &Storage{
		Auths: map[auth.Method]map[auth.SocialID]*auth.Auth{},
		Lock:  &sync.RWMutex{},
	}
}

func (s *Storage) StartCreateTransaction(_ context.Context) (*sqlx.Tx, error) {
	return nil, nil
}

func (s *Storage) UpsertByTransaction(_ context.Context, _ *sqlx.Tx, a *auth.Auth) error {
	s.Lock.Lock()
	if _, ok := s.Auths[a.Method]; !ok {
		s.Auths[a.Method] = map[auth.SocialID]*auth.Auth{}
	}
	s.Auths[a.Method][a.SocialID] = a
	s.Lock.Unlock()
	return nil
}

func (s *Storage) Get(_ context.Context, method auth.Method, socialID auth.SocialID) (*auth.Auth, error) {
	byMethod, ok := s.Auths[method]
	if !ok {
		return nil, auth.ErrNotFound
	}
	a, ok := byMethod[socialID]
	if !ok {
		return nil, auth.ErrNotFound
	}
	return a, nil
}
