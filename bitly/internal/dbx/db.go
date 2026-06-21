package dbx

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func New(url string) (*DB, error) {
	g, err := gorm.Open(postgres.Open(url), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("db: connect: %w", err)
	}
	return &DB{g}, nil
}
