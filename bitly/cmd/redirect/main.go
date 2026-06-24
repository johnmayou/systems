package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/gin-gonic/gin"
	"github.com/johnmayou/systems/bitly/internal/dbx"
	"github.com/johnmayou/systems/bitly/internal/healthx"
	"gorm.io/gorm"
)

type config struct {
	Port        string `env:"PORT" envDefault:"8080"`
	DatabaseUrl string `env:"DATABASE_URL,required"`
}

type urlStore interface {
	GetUrlByShort(ctx context.Context, short string) (dbx.Url, error)
}

func redirectHandler(urlStore urlStore, now func() time.Time) gin.HandlerFunc {
	return func(c *gin.Context) {
		url, err := urlStore.GetUrlByShort(c.Request.Context(), c.Param("short"))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.Status(http.StatusNotFound)
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if url.ExpireAt != nil && url.ExpireAt.Before(now()) {
			c.Status(http.StatusNotFound)
			return
		}

		c.Redirect(http.StatusTemporaryRedirect, url.Long)
	}
}

func newRouter(
	urlStore urlStore,
	now func() time.Time,
) *gin.Engine {
	r := gin.Default()
	r.GET("/up", healthx.HealthCheckHandler)
	r.GET("/:short", redirectHandler(urlStore, now))
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

	router := newRouter(db, time.Now)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
