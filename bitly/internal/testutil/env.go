package testutil

import "os"

const defaultTestDatabaseUrl = "postgres://postgres:postgres@localhost:5432/bitly_test?sslmode=disable"

func DatabaseURL() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	return defaultTestDatabaseUrl
}
