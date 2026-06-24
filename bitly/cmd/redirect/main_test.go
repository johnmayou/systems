package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/johnmayou/systems/bitly/internal/dbx"
	"github.com/johnmayou/systems/bitly/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	if os.Getenv("RUN_AS_SERVER") == "1" {
		main()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestMainBinary(t *testing.T) {
	databaseUrl := testutil.DatabaseURL()
	testutil.NewTestDBFromUrl(t, databaseUrl)

	cmd := exec.Command(os.Args[0], "-test.run=^$")
	cmd.Env = []string{
		"RUN_AS_SERVER=1",
		"PORT=18080",
		"DATABASE_URL=" + databaseUrl,
	}
	require.NoError(t, cmd.Start())
	defer func() { _ = cmd.Process.Kill() }()

	require.Eventually(t, func() bool {
		resp, err := http.Get("http://localhost:18080/up")
		if err != nil {
			return false
		}
		require.NoError(t, resp.Body.Close())
		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 100*time.Millisecond, "server did not start")
}

type testEnv struct {
	router *gin.Engine
	db     *dbx.DB
	now    func() time.Time
}

func newTestEnv(t *testing.T, opts ...func(*testEnv)) *testEnv {
	t.Helper()

	db := testutil.NewTestDB(t)

	e := &testEnv{
		db:  db,
		now: time.Now,
	}
	for _, opt := range opts {
		opt(e)
	}
	e.router = newRouter(e.db, e.now)
	return e
}

func (e *testEnv) get(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	e.router.ServeHTTP(w, req)
	return w
}

func TestRedirect(t *testing.T) {
	t.Run("redirects to long url", func(t *testing.T) {
		env := newTestEnv(t)
		testutil.NewUrl(t, env.db, testutil.UrlWithShort("abc"), testutil.UrlWithLong("https://example.com"))

		w := env.get(t, "/abc")
		require.Equal(t, http.StatusTemporaryRedirect, w.Code)
		require.Equal(t, "https://example.com", w.Header().Get("Location"))
	})
	t.Run("unknown short returns 404", func(t *testing.T) {
		env := newTestEnv(t)

		w := env.get(t, "/doesnotexist")
		require.Equal(t, http.StatusNotFound, w.Code)
	})
	t.Run("expired url returns 404", func(t *testing.T) {
		past := time.Now().Add(-1 * time.Minute)
		env := newTestEnv(t)
		testutil.NewUrl(t, env.db, testutil.UrlWithShort("old"), testutil.UrlWithExpireAt(past))

		w := env.get(t, "/old")
		require.Equal(t, http.StatusNotFound, w.Code)
	})
	t.Run("non-expired url redirects", func(t *testing.T) {
		future := time.Now().Add(1 * time.Minute)
		env := newTestEnv(t)
		testutil.NewUrl(t, env.db, testutil.UrlWithShort("fresh"), testutil.UrlWithLong("https://example.com"), testutil.UrlWithExpireAt(future))

		w := env.get(t, "/fresh")
		require.Equal(t, http.StatusTemporaryRedirect, w.Code)
		require.Equal(t, "https://example.com", w.Header().Get("Location"))
	})
}
