package dbx_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/johnmayou/systems/bitly/internal/testutil"
)

func TestGetUrlByID(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		created := testutil.NewUrl(t, db)

		got, err := db.GetUrlByID(ctx, created.ID)
		require.NoError(t, err)

		assert.Equal(t, created.ID, got.ID)
		assert.Equal(t, created.Short, got.Short)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := db.GetUrlByID(ctx, "1")
		require.Error(t, err)
	})
}

func TestGetUrlByShort(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		testutil.NewUrl(t, db, testutil.UrlWithShort("go"), testutil.UrlWithLong("https://go.dev"))

		got, err := db.GetUrlByShort(ctx, "go")
		require.NoError(t, err)

		assert.Equal(t, "go", got.Short)
		assert.Equal(t, "https://go.dev", got.Long)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := db.GetUrlByShort(ctx, "doesnotexist")
		require.Error(t, err)
	})
}

func TestCreateUrl(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	t.Run("creates url with fields", func(t *testing.T) {
		userID := testutil.NewUser(t, db).ID
		u, err := db.CreateUrl(ctx, userID, "abc", "https://example.com", nil)
		require.NoError(t, err)

		assert.NotEmpty(t, u.ID)
		assert.Equal(t, userID, u.UserID)
		assert.Equal(t, "abc", u.Short)
		assert.Equal(t, "https://example.com", u.Long)
		assert.Nil(t, u.ExpireAt)
	})

	t.Run("creates url with expiry", func(t *testing.T) {
		userID := testutil.NewUser(t, db).ID
		expire := time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second)
		u, err := db.CreateUrl(ctx, userID, "xyz", "https://example.com", &expire)

		require.NoError(t, err)
		require.NotNil(t, u.ExpireAt)
		assert.WithinDuration(t, expire, *u.ExpireAt, time.Second)
	})

	t.Run("duplicate short returns error", func(t *testing.T) {
		testutil.NewUrl(t, db, testutil.UrlWithShort("dup"))
		userID := testutil.NewUser(t, db).ID
		_, err := db.CreateUrl(ctx, userID, "dup", "https://b.com", nil)
		require.Error(t, err)
	})
}

func TestGetUrls(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	userID := testutil.NewUser(t, db).ID

	testutil.NewUrl(t, db, testutil.UrlWithUserID(userID), testutil.UrlWithShort("first"))
	testutil.NewUrl(t, db, testutil.UrlWithUserID(userID), testutil.UrlWithShort("second"))
	testutil.NewUrl(t, db, testutil.UrlWithUserID(userID), testutil.UrlWithShort("third"))

	t.Run("ordered by created_at desc", func(t *testing.T) {
		urls, err := db.GetUrls(ctx, 10)
		require.NoError(t, err)
		require.Len(t, urls, 3)
		assert.Equal(t, "third", urls[0].Short)
		assert.Equal(t, "first", urls[2].Short)
	})

	t.Run("respects limit", func(t *testing.T) {
		urls, err := db.GetUrls(ctx, 2)
		require.NoError(t, err)
		assert.Len(t, urls, 2)
	})
}

func TestUpdateUrl(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	u := testutil.NewUrl(t, db)
	u.Long = "https://new.com"
	err := db.UpdateUrl(ctx, u)
	require.NoError(t, err)

	got, err := db.GetUrlByID(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, "https://new.com", got.Long)
}
