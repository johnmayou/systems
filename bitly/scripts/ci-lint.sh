#!/usr/bin/env bash

set -euo pipefail

GOLANGCI_VERSION="2.12.2"

go vet ./...

if ! command -v golangci-lint &>/dev/null || [[ "$(golangci-lint --version 2>&1)" != *"${GOLANGCI_VERSION}"* ]]; then
  echo "installing golangci-lint v${GOLANGCI_VERSION}..."
  go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v${GOLANGCI_VERSION}
fi
golangci-lint run ./...