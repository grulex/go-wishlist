package inmemory

import (
	"context"
	"github.com/google/uuid"
	"github.com/grulex/go-wishlist/pkg/file"
	"sync"
)

type Storage struct {
	File map[file.ID][]byte
	Lock *sync.RWMutex
}

func NewFileInMemory() *Storage {
	return &Storage{
		File: make(map[file.ID][]byte),
		Lock: &sync.RWMutex{},
	}
}

func (s *Storage) Store(_ context.Context, content []byte) (file.ID, error) {
	s.Lock.Lock()
	id := file.ID(uuid.NewString())
	s.File[id] = content
	s.Lock.Unlock()
	return id, nil
}

func (s *Storage) Get(_ context.Context, id file.ID) ([]byte, error) {
	s.Lock.RLock()
	content, ok := s.File[id]
	if !ok {
		return nil, file.ErrNotFound
	}
	s.Lock.RUnlock()
	return content, nil
}

func (s *Storage) GetStorageType() file.StorageType {
	return file.StorageTypeInMemory
}
