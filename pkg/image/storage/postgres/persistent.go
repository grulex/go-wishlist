package postgrestype

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/grulex/go-wishlist/pkg/file"
	"github.com/grulex/go-wishlist/pkg/image"
	"strings"
	"time"
)

type imagePersistent struct {
	ID          string    `db:"id"`
	StorageType string    `db:"storage_type"`
	FileId      string    `db:"file_id"`
	Width       uint      `db:"width"`
	Height      uint      `db:"height"`
	Hash        string    `db:"hash"`
	Sizes       sizes     `db:"sizes"`
	CreatedAt   time.Time `db:"created_at"`
}

func (p imagePersistent) toDomain() *image.Image {
	sizes := make([]image.Size, len(p.Sizes))
	for i, s := range p.Sizes {
		sizes[i] = image.Size{
			Width:  s.Width,
			Height: s.Height,
			FileLink: file.Link{
				ID:          file.ID(s.FileId),
				StorageType: file.StorageType(s.StorageType),
			},
		}
	}
	hash := image.Hash{}
	hashes := strings.Split(p.Hash, ";")
	if len(hashes) == 3 {
		hash.AHash = hashes[0]
		hash.DHash = hashes[1]
		hash.PHash = hashes[2]
	}

	return &image.Image{
		ID:        image.ID(p.ID),
		Width:     p.Width,
		Height:    p.Height,
		FileLink:  file.Link{ID: file.ID(p.FileId), StorageType: file.StorageType(p.StorageType)},
		Hash:      hash,
		Sizes:     sizes,
		CreatedAt: p.CreatedAt,
	}
}

func (p imagePersistent) fromDomain(image *image.Image) *imagePersistent {
	sizes := make(sizes, len(image.Sizes))
	for i, s := range image.Sizes {
		sizes[i] = size{
			Width:       s.Width,
			Height:      s.Height,
			StorageType: string(s.FileLink.StorageType),
			FileId:      string(s.FileLink.ID),
		}
	}
	hash := image.Hash.AHash + ";" + image.Hash.DHash + ";" + image.Hash.PHash

	return &imagePersistent{
		ID:          string(image.ID),
		StorageType: string(image.FileLink.StorageType),
		FileId:      string(image.FileLink.ID),
		Width:       image.Width,
		Height:      image.Height,
		Hash:        hash,
		CreatedAt:   image.CreatedAt,
		Sizes:       sizes,
	}
}

type size struct {
	Width       uint   `json:"width,omitempty"`
	Height      uint   `json:"height,omitempty"`
	StorageType string `json:"storage_type,omitempty"`
	FileId      string `json:"file_id,omitempty"`
}

type sizes []size

func (s sizes) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *sizes) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	switch src := src.(type) {
	case []byte:
		return json.Unmarshal(src, s)
	case string:
		return json.Unmarshal([]byte(src), s)
	default:
		return errors.New("incompatible type for sizes")
	}
}
