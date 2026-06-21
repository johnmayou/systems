.PHONY: up
up:
	docker compose up

.PHONY: down
down:
	docker compose down

.PHONY: fmt
fmt:
	@go fmt

.PHONY: lint
lint:
	@bash ./scripts/ci-lint.sh

.PHONY: flint
flint: fmt lint

.PHONY: test
test:
	@bash ./scripts/ci-test.sh

.PHONY: test-update
test-update:
	@UPDATE=1 bash ./scripts/ci-test.sh