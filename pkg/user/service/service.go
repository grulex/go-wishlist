package service

import (
	"context"
	"github.com/google/uuid"
	userPkg "github.com/grulex/go-wishlist/pkg/user"
	"time"
)

type storage interface {
	Upsert(ctx context.Context, user *userPkg.User) error
	Get(ctx context.Context, id userPkg.ID) (*userPkg.User, error)
	GetDailyStats(ctx context.Context, duration time.Duration) ([]*userPkg.Stats, error)
}

type Service struct {
	storage storage
}

func NewUserService(storage storage) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) Create(ctx context.Context, user *userPkg.User) error {
	user.ID = userPkg.ID(uuid.NewString())
	user.CreatedAt = time.Now().UTC()
	return s.storage.Upsert(ctx, user)
}

func (s *Service) Update(ctx context.Context, user *userPkg.User) error {
	user.UpdatedAt = time.Now().UTC()
	return s.storage.Upsert(ctx, user)
}

func (s *Service) Get(ctx context.Context, id userPkg.ID) (*userPkg.User, error) {
	return s.storage.Get(ctx, id)
}

func (s *Service) GetDailyStats(ctx context.Context, duration time.Duration) ([]*userPkg.Stats, error) {
	return s.storage.GetDailyStats(ctx, duration)
}
