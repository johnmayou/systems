#!/usr/bin/env bash

set -euo pipefail

go test -v ./... ${UPDATE:+-update}