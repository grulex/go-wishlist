package postgrestype

import (
	"context"
	"database/sql"
	"errors"
	"github.com/grulex/go-wishlist/pkg/image"
	"github.com/jmoiron/sqlx"
)

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
		    sizes,
			created_at
		) VALUES (
			:id,
			:storage_type,
			:file_id,
			:width,
			:height,
			:hash,
		    :sizes,
			:created_at
		)`

	_, err := s.db.NamedExecContext(ctx, query, imagePersistent{}.fromDomain(image))
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
	return a.toDomain(), nil
}

func (s *Storage) GetMany(ctx context.Context, ids []image.ID) ([]*image.Image, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	query := `SELECT * FROM image WHERE id IN (?)`
	query, args, err := sqlx.In(query, ids)
	if err != nil {
		return nil, err
	}
	query = s.db.Rebind(query)

	rows, err := s.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func(rows *sqlx.Rows) {
		_ = rows.Close()
	}(rows)
	var buf imagePersistent
	var returnImages []*image.Image
	for rows.Next() {
		err := rows.StructScan(&buf)
		if err != nil {
			return nil, err
		}
		returnImages = append(returnImages, buf.toDomain())
	}
	return returnImages, nil
}
