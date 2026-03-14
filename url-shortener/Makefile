APP_NAME := url_shortener
PKG_LIST := $(shell go list ./... | grep -v /vendor/)

# Run upwards migration
.PHONY: migrate_up
migrate_up:
	@echo "==> Running latest upward migrations..."
	migrate -database $(POSTGRES_URL) -path db/migrations up 1

# Run backwards migration
.PHONY: migrate down
migrate_down:
	@echo "==> Running latest downward migrations..."
	migrate -database $(POSTGRES_URL) -path db/migrations down 1

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

# Start docker container
.PHONY: compose
compose:
	@echo "==> Starting containers..."
	docker compose -f .vscode/compose.yml -p $(APP_NAME) up
