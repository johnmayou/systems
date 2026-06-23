package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/gin-gonic/gin"
	"github.com/johnmayou/systems/bitly/internal/auth"
	"github.com/johnmayou/systems/bitly/internal/counter"
	"github.com/johnmayou/systems/bitly/internal/dbx"
	"github.com/johnmayou/systems/bitly/internal/healthx"
)

type config struct {
	Port        string `env:"PORT" envDefault:"8080"`
	JwtSecret   string `env:"JWT_SECRET,required"`
	DatabaseUrl string `env:"DATABASE_URL,required"`
	CounterUrl  string `env:"COUNTER_URL,required"`
}

type handler struct {
	urlStore urlStore
	counter  *counter.Client
}

type urlStore interface {
	CreateUrl(ctx context.Context, userID, short, long string, expire *time.Time) (dbx.Url, error)
}

type createUrlRequest struct {
	Long   string     `json:"long" binding:"required"`
	Short  string     `json:"short"`
	Expire *time.Time `json:"expire"`
}

func (h *handler) createUrl(c *gin.Context) {
	var req createUrlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createUrl := func(short string) (dbx.Url, error) {
		return h.urlStore.CreateUrl(
			c.Request.Context(),
			auth.UserID(c),
			short,
			req.Long,
			req.Expire,
		)
	}

	if req.Short == "" {
		count, err := h.counter.GetCount(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		url, err := createUrl(strconv.Itoa(count))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, url)
	} else {
		url, err := createUrl(req.Short)
		if err != nil {
			if errors.Is(err, dbx.ErrDuplicateUrlShort) {
				c.JSON(http.StatusConflict, gin.H{"error": "short already exists"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		c.JSON(http.StatusCreated, url)
	}

}

func newRouter(h *handler, jwtSecret string) *gin.Engine {
	r := gin.Default()
	r.GET("/up", healthx.HealthCheckHandler())

	api := r.Group("/api")
	api.Use(auth.AuthMiddleware(jwtSecret))
	{
		api.POST("/urls", h.createUrl)
	}

	return r
}

func main() {
	var cfg config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to parse env: %v", err)
	}

	db, err := dbx.New(cfg.DatabaseUrl)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	router := newRouter(
		&handler{
			urlStore: db,
			counter:  counter.NewClient(cfg.CounterUrl),
		},
		cfg.JwtSecret,
	)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
