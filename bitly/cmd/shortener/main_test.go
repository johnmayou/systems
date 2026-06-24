package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/johnmayou/systems/bitly/internal/auth"
	"github.com/johnmayou/systems/bitly/internal/counter"
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
		"JWT_SECRET=secret",
		"DATABASE_URL=" + databaseUrl,
		"COUNTER_URL=",
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

const testJwtSecret = "secret"

type testEnv struct {
	router     *gin.Engine
	db         *dbx.DB
	bearer     string
	userID     string
	counterUrl string
}

func newTestEnv(t *testing.T, opts ...func(*testEnv)) *testEnv {
	t.Helper()

	db := testutil.NewTestDB(t)

	count := 1
	counterSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(map[string]int{"value": count}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		count++
	}))
	t.Cleanup(counterSrv.Close)

	userID := uuid.New().String()
	require.NoError(t, db.Create(&dbx.User{ID: userID}).Error)

	tok, err := auth.IssueJwtToken(userID, "test@example.com", testJwtSecret)
	require.NoError(t, err)

	e := &testEnv{
		db:         db,
		bearer:     "Bearer " + tok,
		userID:     userID,
		counterUrl: counterSrv.URL,
	}
	for _, opt := range opts {
		opt(e)
	}
	e.router = newRouter(
		e.db,
		counter.NewClient(e.counterUrl),
		testJwtSecret,
	)
	return e
}

func (e *testEnv) post(t *testing.T, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", e.bearer)
	w := httptest.NewRecorder()
	e.router.ServeHTTP(w, req)
	return w
}

func TestUrls(t *testing.T) {
	t.Run("happy path no alias", func(t *testing.T) {
		env := newTestEnv(t)

		w := env.post(t, "/api/urls", createUrlRequest{Long: "https://example.com"})
		require.Equal(t, http.StatusCreated, w.Code)

		var result dbx.Url
		require.NoError(t, env.db.First(&result, "long = ?", "https://example.com").Error)
		require.Equal(t, "https://example.com", result.Long)
		require.Equal(t, "1", result.Short)
	})
	t.Run("happy path alias", func(t *testing.T) {
		env := newTestEnv(t)

		w := env.post(t, "/api/urls", createUrlRequest{Long: "https://example.com", Short: "short"})
		require.Equal(t, http.StatusCreated, w.Code)

		var result dbx.Url
		require.NoError(t, env.db.First(&result, "long = ?", "https://example.com").Error)
		require.Equal(t, env.userID, result.UserID)
		require.Equal(t, "https://example.com", result.Long)
		require.Equal(t, "short", result.Short)
	})
	t.Run("happy path with expire", func(t *testing.T) {
		env := newTestEnv(t)
		expire := time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second)

		w := env.post(t, "/api/urls", createUrlRequest{Long: "https://example.com", Expire: &expire})
		require.Equal(t, http.StatusCreated, w.Code)

		var result dbx.Url
		require.NoError(t, env.db.First(&result, "long = ?", "https://example.com").Error)
		require.NotNil(t, result.ExpireAt)
		require.Equal(t, expire, result.ExpireAt.UTC().Truncate(time.Second))
	})
	t.Run("alias already exists returns 409", func(t *testing.T) {
		env := newTestEnv(t)

		w := env.post(t, "/api/urls", createUrlRequest{Long: "https://example.com", Short: "short"})
		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

		w = env.post(t, "/api/urls", createUrlRequest{Long: "https://example.com", Short: "short"})
		require.Equal(t, http.StatusConflict, w.Code)
		require.Equal(t, `{"error":"short already exists"}`, w.Body.String())
	})
	t.Run("missing long returns 400", func(t *testing.T) {
		env := newTestEnv(t)

		w := env.post(t, "/api/urls", createUrlRequest{Long: ""})
		require.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("missing auth returns 401", func(t *testing.T) {
		env := newTestEnv(t)
		env.bearer = ""

		w := env.post(t, "/api/urls", createUrlRequest{Long: "https://example.com"})
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("invalid token returns 401", func(t *testing.T) {
		env := newTestEnv(t)
		env.bearer = "Bearer invalid.token.here"

		w := env.post(t, "/api/urls", createUrlRequest{Long: "https://example.com"})
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("counter error returns 500", func(t *testing.T) {
		errSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "counter down", http.StatusServiceUnavailable)
		}))
		t.Cleanup(errSrv.Close)

		env := newTestEnv(t, func(e *testEnv) { e.counterUrl = errSrv.URL })

		w := env.post(t, "/api/urls", createUrlRequest{Long: "https://example.com"})
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
