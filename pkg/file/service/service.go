package service

import (
	"context"
	"github.com/grulex/go-wishlist/pkg/file"
	"io"
)

type FileStorage interface {
	GetFileReader(ctx context.Context, fileID file.ID) (io.ReadCloser, error)
	UploadFile(ctx context.Context, reader io.Reader) (file.ID, error)
	GetStorageType() file.StorageType
}

type Service struct {
	storages map[file.StorageType]FileStorage
	priority []file.StorageType
}

func NewFileService(storagesByPriority []FileStorage) *Service {
	storagesMap := make(map[file.StorageType]FileStorage)
	priority := make([]file.StorageType, 0, len(storagesByPriority))
	for _, s := range storagesByPriority {
		priority = append(priority, s.GetStorageType())
		storagesMap[s.GetStorageType()] = s
	}

	return &Service{
		storages: storagesMap,
		priority: priority,
	}
}

func (s *Service) UploadPhoto(ctx context.Context, reader io.Reader) (file.Link, error) {
	var lastErr error
	for _, storageType := range s.priority {
		storage, _ := s.storages[storageType]

		id, err := storage.UploadFile(ctx, reader)
		if err != nil {
			lastErr = err
			continue
		}

		return file.Link{
			StorageType: storage.GetStorageType(),
			ID:          id,
		}, nil
	}

	return file.Link{}, lastErr
}

func (s *Service) Download(ctx context.Context, link file.Link) (io.ReadCloser, error) {
	storage, ok := s.storages[link.StorageType]
	if !ok {
		return nil, file.ErrStorageNotDefined
	}

	return storage.GetFileReader(ctx, link.ID)
}
