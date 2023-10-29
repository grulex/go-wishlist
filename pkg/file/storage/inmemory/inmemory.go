package inmemory

import (
	"bytes"
	"context"
	"github.com/google/uuid"
	"github.com/grulex/go-wishlist/pkg/file"
	"image"
	_ "image/png"
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

func (s Storage) GetFileReader(_ context.Context, fileID file.ID) (io.ReadCloser, error) {
	s.Lock.RLock()
	content, ok := s.File[fileID]
	if !ok {
		return nil, file.ErrNotFound
	}
	s.Lock.RUnlock()
	reader := bytes.NewReader(content)
	return io.NopCloser(reader), nil
}

func (s Storage) UploadImageFile(_ context.Context, reader io.Reader) ([]file.ImageSize, error) {
	s.Lock.Lock()
	id := file.ID(uuid.NewString())
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	s.File[id] = content
	s.Lock.Unlock()
	img, _, err := image.Decode(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}

	sizes := []file.ImageSize{
		{
			Width:  uint(img.Bounds().Dx()),
			Height: uint(img.Bounds().Dy()),
			Link: file.Link{
				StorageType: s.GetStorageType(),
				ID:          id,
			},
		},
	}
	return sizes, nil
}

func (s Storage) GetStorageType() file.StorageType {
	return file.StorageTypeInMemory
}
