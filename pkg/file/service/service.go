package service

import (
	"context"
	"github.com/grulex/go-wishlist/pkg/file"
)

type FileStorage interface {
	Store(ctx context.Context, content []byte) (file.ID, error)
	Get(ctx context.Context, id file.ID) ([]byte, error)
	GetStorageType() file.StorageType
}

type Service struct {
	storages map[file.StorageType]FileStorage
}

func NewFileService(storages []FileStorage) *Service {
	storagesMap := make(map[file.StorageType]FileStorage)
	for _, s := range storages {
		storagesMap[s.GetStorageType()] = s
	}

	return &Service{
		storages: storagesMap,
	}
}

func (s *Service) Upload(ctx context.Context, content []byte, storageType file.StorageType) (*file.Link, error) {
	storage, ok := s.storages[storageType]
	if !ok {
		return nil, file.ErrStorageNotDefined
	}

	id, err := storage.Store(ctx, content)
	if err != nil {
		return nil, err
	}

	return &file.Link{
		ID:          id,
		StorageType: storage.GetStorageType(),
	}, nil
}

func (s *Service) Download(ctx context.Context, link *file.Link) ([]byte, error) {
	storage, ok := s.storages[link.StorageType]
	if !ok {
		return nil, file.ErrStorageNotDefined
	}

	return storage.Get(ctx, link.ID)
}
