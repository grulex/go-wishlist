package service

import (
	"context"
	"github.com/grulex/go-wishlist/pkg/file"
)

type storage interface {
	Store(ctx context.Context, content []byte) (file.ID, error)
	Get(ctx context.Context, id file.ID) ([]byte, error)
	GetStorageType() file.StorageType
}

type Service struct {
	storages map[file.StorageType]storage
}

func New(storages []storage) *Service {
	storagesMap := make(map[file.StorageType]storage)
	for _, s := range storages {
		storagesMap[s.GetStorageType()] = s
	}

	return &Service{
		storages: storagesMap,
	}
}

func (s *Service) Upload(ctx context.Context, content []byte, storageType file.StorageType) (file.Link, error) {
	storage, ok := s.storages[storageType]
	if !ok {
		return file.Link{}, file.ErrStorageNotDefined
	}

	id, err := storage.Store(ctx, content)
	if err != nil {
		return file.Link{}, err
	}

	return file.Link{
		ID:          id,
		StorageType: storage.GetStorageType(),
	}, nil
}

func (s *Service) Download(ctx context.Context, link file.Link) ([]byte, error) {
	storage, ok := s.storages[link.StorageType]
	if !ok {
		return nil, file.ErrStorageNotDefined
	}

	return storage.Get(ctx, link.ID)
}
