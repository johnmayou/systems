package testutil

import (
	"testing"

	"github.com/johnmayou/systems/bitly/internal/dbx"
	"github.com/stretchr/testify/require"
)

func NewTestDB(t *testing.T) *dbx.DB {
	db, err := dbx.New(DatabaseURL())
	require.NoError(t, err)
	reset(t, db)
	return db
}

func NewTestDBFromUrl(t *testing.T, url string) *dbx.DB {
	db, err := dbx.New(url)
	require.NoError(t, err)
	reset(t, db)
	return db
}

func reset(t *testing.T, db *dbx.DB) {
	err := db.Exec(`
        DO $$ DECLARE
            r RECORD;
        BEGIN
            FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename != 'goose_db_version')
            LOOP
                EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' RESTART IDENTITY CASCADE';
            END LOOP;
        END $$;
    `).Error
	require.NoError(t, err)
}
