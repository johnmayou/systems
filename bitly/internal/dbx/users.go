package dbx

import (
	"context"
	"fmt"
)

type User struct {
	ID string `gorm:"primarykey;default:gen_random_uuid()"`
}

func (db *DB) CreateUser(ctx context.Context) (User, error) {
	var u User
	if err := db.WithContext(ctx).Create(&u).Error; err != nil {
		return User{}, fmt.Errorf("db: create user: %w", err)
	}
	return u, nil
}
