package inmemory

import (
	"bytes"
	"context"
	"github.com/google/uuid"
	"github.com/grulex/go-wishlist/pkg/file"
	"io"
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

func (s Storage) GetPhotoReader(_ context.Context, fileID file.ID) (io.ReadCloser, error) {
	s.Lock.RLock()
	content, ok := s.File[fileID]
	if !ok {
		return nil, file.ErrNotFound
	}
	s.Lock.RUnlock()
	reader := bytes.NewReader(content)
	return io.NopCloser(reader), nil
}

func (s Storage) UploadPhoto(_ context.Context, reader io.Reader) (file.ID, error) {
	s.Lock.Lock()
	id := file.ID(uuid.NewString())
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	s.File[id] = content
	s.Lock.Unlock()
	return id, nil
}

func (s Storage) GetStorageType() file.StorageType {
	return file.StorageTypeInMemory
}
