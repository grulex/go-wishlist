package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/bojanz/currency"
	imagePkg "github.com/grulex/go-wishlist/pkg/image"
	productPkg "github.com/grulex/go-wishlist/pkg/product"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"
	"time"
)

type productPersistent struct {
	ID          string           `db:"id"`
	Title       string           `db:"title"`
	ImageID     *string          `db:"image_id"`
	Price       *currency.Amount `db:"price"`
	Description null.String      `db:"description"`
	Url         null.String      `db:"url"`
	CreatedAt   time.Time        `db:"created_at"`
	UpdatedAt   time.Time        `db:"updated_at"`
}

type Storage struct {
	db *sqlx.DB
}

func NewProductStorage(db *sqlx.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) Upsert(ctx context.Context, p *productPkg.Product) error {
	query := `INSERT INTO product (
		id,
		title,
		image_id,
		price,
		description,
		url,
		created_at,
		updated_at
	) VALUES (
		:id,
		:title,
		:image_id,
		:price,
		:description,
		:url,
		:created_at,
		:updated_at
	) ON CONFLICT (id) DO UPDATE SET
		title = :title,
		image_id = :image_id,
		price = :price,
		description = :description,
		url = :url,
		updated_at = :updated_at`

	var imageID *string
	if p.ImageID != nil {
		imageIDString := string(*p.ImageID)
		imageID = &imageIDString
	}

	productPersistent := productPersistent{
		ID:          string(p.ID),
		Title:       p.Title,
		ImageID:     imageID,
		Price:       p.Price,
		Description: p.Description,
		Url:         p.Url,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
	_, err := s.db.NamedExecContext(ctx, query, productPersistent)
	return err
}

func (s *Storage) Get(ctx context.Context, id productPkg.ID) (*productPkg.Product, error) {
	query := `SELECT * FROM product WHERE id = $1`
	p := &productPersistent{}
	err := s.db.GetContext(ctx, p, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, productPkg.ErrNotFound
		}
		return nil, err
	}

	var imageID *imagePkg.ID
	if p.ImageID != nil {
		imageIDString := imagePkg.ID(*p.ImageID)
		imageID = &imageIDString
	}

	product := &productPkg.Product{
		ID:          productPkg.ID(p.ID),
		Title:       p.Title,
		ImageID:     imageID,
		Price:       p.Price,
		Description: p.Description,
		Url:         p.Url,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
	return product, nil
}

func (s *Storage) GetMany(ctx context.Context, ids []productPkg.ID) (products []*productPkg.Product, err error) {
	if len(ids) == 0 {
		return nil, nil
	}

	idsString := make([]string, len(ids))
	for i, id := range ids {
		idsString[i] = string(id)
	}

	query := `SELECT * FROM product WHERE id IN (?)`
	query, args, err := sqlx.In(query, idsString)
	if err != nil {
		return nil, err
	}
	query = s.db.Rebind(query)

	rows, err := s.db.QueryxContext(ctx, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	defer func(rows *sqlx.Rows) {
		_ = rows.Close()
	}(rows)

	var buf productPersistent
	products = make([]*productPkg.Product, 0)
	for rows.Next() {
		err := rows.StructScan(&buf)
		if err != nil {
			return nil, err
		}

		var imageID *imagePkg.ID
		if buf.ImageID != nil {
			imageIDString := imagePkg.ID(*buf.ImageID)
			imageID = &imageIDString
		}
		product := &productPkg.Product{
			ID:          productPkg.ID(buf.ID),
			Title:       buf.Title,
			ImageID:     imageID,
			Price:       buf.Price,
			Description: buf.Description,
			Url:         buf.Url,
			CreatedAt:   buf.CreatedAt,
			UpdatedAt:   buf.UpdatedAt,
		}
		products = append(products, product)
	}

	return products, nil
}
