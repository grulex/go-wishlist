package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/grulex/go-wishlist/pkg/file"
	"github.com/jmoiron/sqlx"
	"image"
	"io"
	"time"
)

type filePersistent struct {
	ID        string    `db:"id"`
	Content   []byte    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}

type Storage struct {
	db *sqlx.DB
}

func NewPostgresStorage(db *sqlx.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) GetFileReader(ctx context.Context, fileID file.ID) (io.ReadCloser, error) {
	query := `SELECT * FROM file WHERE id = $1`
	var buf filePersistent
	err := s.db.GetContext(ctx, &buf, query, string(fileID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, file.ErrNotFound
		}
		return nil, err
	}
	readCloser := io.NopCloser(bytes.NewReader(buf.Content))
	return readCloser, nil
}

func (s *Storage) UploadImageFile(ctx context.Context, reader io.Reader) ([]file.ImageSize, error) {
	query := `
		INSERT INTO file (
			id,
			content,
			created_at
		) VALUES (
			:id,
			:content,
			:created_at
		)`
	id := file.ID(uuid.NewString())
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	filePersistent := filePersistent{
		ID:        string(id),
		Content:   content,
		CreatedAt: time.Now().UTC(),
	}
	_, err = s.db.NamedExecContext(ctx, query, filePersistent)
	if err != nil {
		return nil, err
	}

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

func (s *Storage) GetStorageType() file.StorageType {
	return file.StorageTypePostgres
}
