app_name := url_shortener

.PHONY: migrate_up
migrate_up:
	@echo "==> Running latest upward migrations..."
	migrate -database $(POSTGRES_URL) -path db/migrations up 1

.PHONY: migrate down
migrate_down:
	@echo "==> Running latest downward migrations..."
	migrate -database $(POSTGRES_URL) -path db/migrations down 1
