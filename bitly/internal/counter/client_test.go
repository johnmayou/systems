package counter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetCount(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewEncoder(w).Encode(map[string]int{"value": 42}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}))
		defer srv.Close()

		c := NewClient(srv.URL)
		count, err := c.GetCount(context.Background())
		require.NoError(t, err)
		require.Equal(t, 42, count)
	})
	t.Run("non 200 status", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}))
		defer srv.Close()

		c := NewClient(srv.URL)
		_, err := c.GetCount(context.Background())
		require.Error(t, err)
		require.ErrorContains(t, err, "500")
	})

	t.Run("context cancelled", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewEncoder(w).Encode(map[string]int{"value": 42}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}))
		defer srv.Close()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		c := NewClient(srv.URL)
		_, err := c.GetCount(ctx)
		require.Error(t, err)
		require.ErrorIs(t, err, context.Canceled)
	})
}
