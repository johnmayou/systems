-- +goose Up
CREATE TABLE urls (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id    UUID NOT NULL REFERENCES users(id),
    short      TEXT NOT NULL UNIQUE,
    long       TEXT NOT NULL,
    expire_at  TIMESTAMPTZ
);

-- +goose Down
DROP TABLE urls;
