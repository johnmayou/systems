package dbx_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/johnmayou/systems/bitly/internal/testutil"
)

func TestCreateUser(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	t.Run("creates user with id", func(t *testing.T) {
		u, err := db.CreateUser(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, u.ID)
	})

	t.Run("each user gets unique id", func(t *testing.T) {
		a, err := db.CreateUser(ctx)
		require.NoError(t, err)
		b, err := db.CreateUser(ctx)
		require.NoError(t, err)
		assert.NotEqual(t, a.ID, b.ID)
	})
}
