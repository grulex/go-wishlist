package postgrestype

import (
	"context"
	"database/sql"
	"errors"
	"github.com/grulex/go-wishlist/pkg/file"
	"github.com/grulex/go-wishlist/pkg/image"
	"github.com/jmoiron/sqlx"
	"time"
)

type imagePersistent struct {
	ID          string    `db:"id"`
	StorageType string    `db:"storage_type"`
	FileId      string    `db:"file_id"`
	Width       uint      `db:"width"`
	Height      uint      `db:"height"`
	Hash        string    `db:"hash"`
	CreatedAt   time.Time `db:"created_at"`
}

type Storage struct {
	db *sqlx.DB
}

func NewImageStorage(db *sqlx.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) Upsert(ctx context.Context, image *image.Image) error {
	query := `
		INSERT INTO image (
			id,
			storage_type,
			file_id,
			width,
			height,
			hash,
			created_at
		) VALUES (
			:id,
			:storage_type,
			:file_id,
			:width,
			:height,
			:hash,
			:created_at
		)`
	imagePersistent := imagePersistent{
		ID:          string(image.ID),
		StorageType: string(image.FileLink.StorageType),
		FileId:      string(image.FileLink.ID),
		Width:       image.Width,
		Height:      image.Height,
		Hash:        image.Hash.AHash + ";" + image.Hash.DHash + ";" + image.Hash.PHash,
		CreatedAt:   image.CreatedAt,
	}
	_, err := s.db.NamedExecContext(ctx, query, imagePersistent)
	return err
}

func (s *Storage) Get(ctx context.Context, id image.ID) (*image.Image, error) {
	query := `SELECT * FROM image WHERE id = $1`
	a := &imagePersistent{}
	err := s.db.GetContext(ctx, a, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, image.ErrNotFound
		}
		return nil, err
	}
	return &image.Image{
		ID: image.ID(a.ID),
		FileLink: file.Link{
			StorageType: file.StorageType(a.StorageType),
			ID:          file.ID(a.FileId),
		},
		Width:     a.Width,
		Height:    a.Height,
		Hash:      image.Hash{AHash: a.Hash, DHash: a.Hash, PHash: a.Hash},
		CreatedAt: a.CreatedAt,
	}, nil
}
