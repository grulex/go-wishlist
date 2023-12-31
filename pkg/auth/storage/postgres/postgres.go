package postgres

import (
	"context"
	"database/sql"
	"errors"
	authPkg "github.com/grulex/go-wishlist/pkg/auth"
	"github.com/grulex/go-wishlist/pkg/user"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"
	"time"
)

type authPersistent struct {
	UserID    user.ID     `db:"user_id"`
	Method    string      `db:"method"`
	SocialID  null.String `db:"social_id"`
	CreatedAt time.Time   `db:"created_at"`
	UpdatedAt time.Time   `db:"updated_at"`
}

type Storage struct {
	db *sqlx.DB
}

func NewAuthStorage(db *sqlx.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) StartCreateTransaction(ctx context.Context) (*sqlx.Tx, error) {
	tx := s.db.MustBeginTx(ctx, nil)
	// lock table
	query := `LOCK TABLE auth IN EXCLUSIVE MODE`
	_, err := tx.ExecContext(ctx, query)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *Storage) UpsertByTransaction(ctx context.Context, tx *sqlx.Tx, a *authPkg.Auth) error {
	query := `
		INSERT INTO auth (
			user_id,
			method,
			social_id,
			created_at,
			updated_at
		) VALUES (
			:user_id,
			:method,
			:social_id,
			:created_at,
			:updated_at
		) ON CONFLICT (method, social_id) DO UPDATE SET
			updated_at = :updated_at`
	authPersistent := authPersistent{
		UserID:    a.UserID,
		Method:    string(a.Method),
		SocialID:  null.String(a.SocialID),
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
	_, err := tx.NamedExecContext(ctx, query, authPersistent)
	if err != nil {
		return err
	}

	return err
}

func (s *Storage) Get(ctx context.Context, method authPkg.Method, socialID authPkg.SocialID) (*authPkg.Auth, error) {
	query := `SELECT * FROM auth WHERE method = $1 AND social_id = $2 FOR UPDATE`
	a := &authPersistent{}
	err := s.db.GetContext(ctx, a, query, method, socialID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, authPkg.ErrNotFound
		}
		return nil, err
	}
	auth := &authPkg.Auth{
		UserID:    a.UserID,
		Method:    authPkg.Method(a.Method),
		SocialID:  authPkg.SocialID(a.SocialID),
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}

	return auth, nil
}
