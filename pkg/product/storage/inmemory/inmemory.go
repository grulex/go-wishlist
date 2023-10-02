package inmemory

import (
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

func (s *Storage) Upsert(p product.Product) error {
	s.Lock.Lock()
	s.products[p.ID] = p
	s.Lock.Unlock()
	return nil
}

func (s *Storage) Get(id product.ID) (product.Product, error) {
	s.Lock.RLock()
	p, ok := s.products[id]
	if !ok {
		return product.Product{}, product.ErrNotFound
	}
	s.Lock.RUnlock()
	return p, nil
}

func (s *Storage) GetMany(ids []product.ID) (products []product.Product, err error) {
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
