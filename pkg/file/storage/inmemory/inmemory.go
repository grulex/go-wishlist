package inmemory

import (
	"context"
	"github.com/google/uuid"
	"github.com/grulex/go-wishlist/pkg/file"
)

type Storage struct {
	File map[file.ID][]byte
}

func New() *Storage {
	return &Storage{
		File: make(map[file.ID][]byte),
	}
}

func (s *Storage) Store(_ context.Context, content []byte) (file.ID, error) {
	id := file.ID(uuid.NewString())
	s.File[id] = content
	return id, nil
}

func (s *Storage) Get(_ context.Context, id file.ID) ([]byte, error) {
	content, ok := s.File[id]
	if !ok {
		return nil, file.ErrNotFound
	}
	return content, nil
}

func (s *Storage) GetStorageType() file.StorageType {
	return file.StorageTypeInMemory
}
