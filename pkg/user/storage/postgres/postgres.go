package postgres

import (
	"context"
	"database/sql"
	"errors"
	userPkg "github.com/grulex/go-wishlist/pkg/user"
	"github.com/jmoiron/sqlx"
	"time"
)

type userPersistent struct {
	ID        string    `db:"id"`
	FullName  string    `db:"full_name"`
	CreatedAt time.Time `db:"created_at"`
	Language  string    `db:"lang"`
}

type Storage struct {
	db *sqlx.DB
}

func NewUserStorage(db *sqlx.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) Upsert(_ context.Context, u *userPkg.User) error {
	query := `INSERT INTO users (
		id,
		fullname,
		created_at,
		lang
	) VALUES (
		:id,
		:full_name,
		:created_at,
		:lang
	) ON CONFLICT (id) DO UPDATE SET
		fullname = :full_name,
		lang = :lang`
	userPersistent := userPersistent{
		ID:        string(u.ID),
		FullName:  u.FullName,
		CreatedAt: u.CreatedAt,
		Language:  string(u.Language),
	}
	_, err := s.db.NamedExec(query, userPersistent)
	return err
}

func (s *Storage) Get(_ context.Context, id userPkg.ID) (*userPkg.User, error) {
	query := `SELECT * FROM users WHERE id = $1`
	userPersistent := userPersistent{}
	err := s.db.Get(&userPersistent, query, string(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, userPkg.ErrNotFound
		}
		return nil, err
	}
	return &userPkg.User{
		ID:        userPkg.ID(userPersistent.ID),
		FullName:  userPersistent.FullName,
		CreatedAt: userPersistent.CreatedAt,
		Language:  userPkg.Language(userPersistent.Language),
	}, nil
}
