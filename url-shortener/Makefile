APP_NAME := url_shortener
PKG_LIST := $(shell go list ./... | grep -v /vendor/)

# Run latest upwards migration
.PHONY: migrate_up
migrate_up:
	@echo "==> Running upward migrations..."
	migrate -database $(POSTGRES_URL) -path db/migrations up 1

# Run all upwards migration
.PHONY: migrate_up_all
migrate_up_all:
	@echo "==> Running upward migrations..."
	migrate -database $(POSTGRES_URL) -path db/migrations up

# Run latest backwards migration
.PHONY: migrate_down
migrate_down:
	@echo "==> Running latest downward migrations..."
	migrate -database $(POSTGRES_URL) -path db/migrations down 1

# Run all backwards migration
.PHONY: migrate_down_all
migrate_down_all:
	@echo "==> Running latest downward migrations..."
	migrate -database $(POSTGRES_URL) -path db/migrations down

# Drop latest dirty migration state
.PHONY: drop_dirty_migration
drop_dirty_migration:
	@echo "==> Running upward migrations..."
	migrate -database $(POSTGRES_URL) -path db/migrations force $(LAST_MIGRATION_STATE)

# Generate sqlc code
.PHONY: generate_sql
generate_sql:
	@echo "==> Generating sqlc code..."
	sqlc generate

## Format the code
.PHONY: format
format:
	@echo "==> Formatting code..."
	go fmt $(PKG_LIST)
	gofumpt -w .
	golines -w --max-len=120 .

# Run all tests
.PHONY: test
test:
	@echo "==> Running tests..."
	go test -timeout 30s -coverprofile=/tmp/coverage.out -v -race ./...

# Start docker container
.PHONY: compose
compose:
	@echo "==> Starting containers..."
	docker compose -f .vscode/compose.yml -p $(APP_NAME) up
