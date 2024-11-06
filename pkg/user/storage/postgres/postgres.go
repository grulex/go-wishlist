package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/grulex/go-wishlist/pkg/notify"
	userPkg "github.com/grulex/go-wishlist/pkg/user"
	"github.com/jmoiron/sqlx"
	"time"
)

type userPersistent struct {
	ID              string    `db:"id"`
	FullName        string    `db:"fullname"`
	Language        string    `db:"lang"`
	NotifyType      *string   `db:"notify_type"`
	NotifyChannelID *string   `db:"notify_channel_id"`
	CreatedAt       time.Time `db:"created_at"`
}

type Storage struct {
	db *sqlx.DB
}

func NewUserStorage(db *sqlx.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) Upsert(ctx context.Context, u *userPkg.User) error {
	query := `INSERT INTO users (
		id,
		fullname,
		created_at,
		lang,
		notify_type,
		notify_channel_id
	) VALUES (
		:id,
		:fullname,
		:created_at,
		:lang,
		:notify_type,
		:notify_channel_id
	) ON CONFLICT (id) DO UPDATE SET
		fullname = :fullname,
		lang = :lang,
		notify_type = :notify_type,
		notify_channel_id = :notify_channel_id`

	var notifyType *string
	if u.NotifyType != nil {
		typeString := string(*u.NotifyType)
		notifyType = &typeString
	}

	userPersistent := userPersistent{
		ID:              string(u.ID),
		FullName:        u.FullName,
		CreatedAt:       u.CreatedAt,
		Language:        string(u.Language),
		NotifyType:      notifyType,
		NotifyChannelID: u.NotifyChannelID,
	}
	_, err := s.db.NamedExecContext(ctx, query, userPersistent)
	return err
}

func (s *Storage) Get(ctx context.Context, id userPkg.ID) (*userPkg.User, error) {
	query := `SELECT * FROM users WHERE id = $1`
	userPersistent := userPersistent{}
	err := s.db.GetContext(ctx, &userPersistent, query, string(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, userPkg.ErrNotFound
		}
		return nil, err
	}

	var notifyType *notify.Type
	if userPersistent.NotifyType != nil {
		typedType := notify.Type(*userPersistent.NotifyType)
		notifyType = &typedType
	}
	return &userPkg.User{
		ID:              userPkg.ID(userPersistent.ID),
		FullName:        userPersistent.FullName,
		CreatedAt:       userPersistent.CreatedAt,
		Language:        userPkg.Language(userPersistent.Language),
		NotifyType:      notifyType,
		NotifyChannelID: userPersistent.NotifyChannelID,
	}, nil
}

func (s *Storage) GetDailyStats(ctx context.Context, duration time.Duration) ([]*userPkg.Stats, error) {
	query := `Select substr(date_trunc('day', created_at::date)::text, 0, 11) as period, COUNT(*)
		FROM users
		WHERE created_at > $1
		GROUP BY period
		ORDER BY period;`
	var statsPersistent []*struct {
		Period string `db:"period"`
		Count  int    `db:"count"`
	}

	err := s.db.SelectContext(ctx, &statsPersistent, query, time.Now().Add(-duration))
	if err != nil {
		return nil, err
	}

	stats := make([]*userPkg.Stats, 0, len(statsPersistent))
	for _, s := range statsPersistent {
		stats = append(stats, &userPkg.Stats{
			Day:   s.Period,
			Count: s.Count,
		})
	}
	return stats, nil
}
