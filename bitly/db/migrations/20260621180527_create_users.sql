-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

-- +goose Down
DROP TABLE users;
