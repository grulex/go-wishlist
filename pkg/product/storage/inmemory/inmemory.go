package inmemory

import (
	"context"
	"github.com/grulex/go-wishlist/pkg/product"
	"sync"
)

type Storage struct {
	products map[product.ID]product.Product
	Lock     *sync.RWMutex
}

func New() *Storage {
	return &Storage{
		products: map[product.ID]product.Product{},
		Lock:     &sync.RWMutex{},
	}
}

func (s *Storage) Upsert(_ context.Context, p product.Product) error {
	s.Lock.Lock()
	s.products[p.ID] = p
	s.Lock.Unlock()
	return nil
}

func (s *Storage) Get(_ context.Context, id product.ID) (product.Product, error) {
	s.Lock.RLock()
	p, ok := s.products[id]
	if !ok {
		return product.Product{}, product.ErrNotFound
	}
	s.Lock.RUnlock()
	return p, nil
}

func (s *Storage) GetMany(_ context.Context, ids []product.ID) (products []product.Product, err error) {
	s.Lock.RLock()
	for _, id := range ids {
		p, ok := s.products[id]
		if !ok {
			return nil, product.ErrNotFound
		}
		products = append(products, p)
	}
	s.Lock.RUnlock()
	return products, nil
}
