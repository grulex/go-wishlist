package service

import (
	"context"
	"github.com/google/uuid"
	productPkg "github.com/grulex/go-wishlist/pkg/product"
	"time"
)

type storage interface {
	Upsert(ctx context.Context, product productPkg.Product) error
	Get(ctx context.Context, id productPkg.ID) (productPkg.Product, error)
	GetMany(ctx context.Context, ids []productPkg.ID) ([]productPkg.Product, error)
}

type Service struct {
	storage storage
}

func New(storage storage) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) Create(ctx context.Context, product productPkg.Product) error {
	product.ID = productPkg.ID(uuid.NewString())
	product.CreatedAt = time.Now().UTC()
	product.UpdatedAt = product.CreatedAt
	return s.storage.Upsert(ctx, product)
}

func (s *Service) Get(ctx context.Context, id productPkg.ID) (productPkg.Product, error) {
	return s.storage.Get(ctx, id)
}

func (s *Service) GetMany(ctx context.Context, ids []productPkg.ID) ([]productPkg.Product, error) {
	return s.storage.GetMany(ctx, ids)
}

func (s *Service) Update(ctx context.Context, product productPkg.Product) error {
	product.UpdatedAt = time.Now().UTC()
	return s.storage.Upsert(ctx, product)
}
