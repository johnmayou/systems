package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/johnmayou/systems/bitly/internal/dbx"
	"github.com/stretchr/testify/require"
)

func NewUser(t *testing.T, db *dbx.DB) dbx.User {
	t.Helper()
	u, err := db.CreateUser(context.Background())
	require.NoError(t, err)
	return u
}

type urlOpts struct {
	userID   string
	short    string
	long     string
	expireAt *time.Time
}

type UrlOpt func(*urlOpts)

func UrlWithUserID(id string) UrlOpt     { return func(o *urlOpts) { o.userID = id } }
func UrlWithShort(s string) UrlOpt       { return func(o *urlOpts) { o.short = s } }
func UrlWithLong(l string) UrlOpt        { return func(o *urlOpts) { o.long = l } }
func UrlWithExpireAt(t time.Time) UrlOpt { return func(o *urlOpts) { o.expireAt = &t } }

func NewUrl(t *testing.T, db *dbx.DB, opts ...UrlOpt) dbx.Url {
	t.Helper()
	o := &urlOpts{
		short: gofakeit.LetterN(6),
		long:  gofakeit.URL(),
	}
	for _, opt := range opts {
		opt(o)
	}
	if o.userID == "" {
		o.userID = NewUser(t, db).ID
	}
	u, err := db.CreateUrl(context.Background(), o.userID, o.short, o.long, o.expireAt)
	require.NoError(t, err)
	return u
}
