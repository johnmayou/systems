package dbx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type Url struct {
	ID        string `gorm:"primarykey;default:gen_random_uuid()"`
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    string
	Short     string
	Long      string
	ExpireAt  *time.Time
}

func (db *DB) GetUrlByID(ctx context.Context, id string) (Url, error) {
	var u Url
	if err := db.WithContext(ctx).First(&u, "id = ?", id).Error; err != nil {
		return Url{}, fmt.Errorf("db: get url by id: %w", err)
	}
	return u, nil
}

func (db *DB) GetUrlByShort(ctx context.Context, short string) (Url, error) {
	var u Url
	if err := db.WithContext(ctx).First(&u, "short = ?", short).Error; err != nil {
		return Url{}, fmt.Errorf("db: get url by short: %w", err)
	}
	return u, nil
}

func (db *DB) GetUrls(ctx context.Context, limit int) ([]Url, error) {
	var urls []Url
	if err := db.WithContext(ctx).Order("created_at DESC").Limit(limit).Find(&urls).Error; err != nil {
		return nil, fmt.Errorf("db: get urls: %w", err)
	}
	return urls, nil
}

var ErrDuplicateUrlShort = errors.New("short url already exists in database")

func (db *DB) CreateUrl(ctx context.Context, userID, short, long string, expire *time.Time) (Url, error) {
	u := Url{UserID: userID, Short: short, Long: long, ExpireAt: expire}
	if err := db.WithContext(ctx).Create(&u).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.ConstraintName == "urls_short_key" {
			return Url{}, ErrDuplicateUrlShort
		}
		return Url{}, fmt.Errorf("db: create url: %w", err)
	}
	return u, nil
}

func (db *DB) UpdateUrl(ctx context.Context, u Url) error {
	if err := db.WithContext(ctx).Save(&u).Error; err != nil {
		return fmt.Errorf("db: update url: %w", err)
	}
	return nil
}
