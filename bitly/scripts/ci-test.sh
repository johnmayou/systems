#!/usr/bin/env bash

set -euo pipefail

docker compose -f docker-compose.test.yaml up -d --wait
trap 'docker compose -f docker-compose.test.yaml down' EXIT

goose -dir db/migrations postgres "postgres://postgres:postgres@localhost:5432/bitly_test" up

go test -v ./... ${UPDATE:+-update}