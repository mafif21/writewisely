SERVER_DIR     := server
MIGRATIONS_DIR := $(SERVER_DIR)/migrations

# golang-migrate CLI, pinned and run via `go run` (no global install needed).
MIGRATE_VERSION := v4.19.1
MIGRATE         := go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)

# Use DB_* from the environment, else from server/.env.
ifndef DB_HOST
-include $(SERVER_DIR)/.env
endif
DB_PORT    ?= 5432
DB_SSLMODE ?= disable

# golang-migrate only accepts a URL, so it is built from the DB_* parts here.
# Unlike the server's keyword/value DSN, DB_PASSWORD must be URL-safe for the migrate targets.
MIGRATE_DB_URL = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

.PHONY: run test docs migrate-up migrate-down check-db-config

## run: start the API server (reads server/.env)
run:
	cd $(SERVER_DIR) && go run ./cmd/api

## test: run server tests
test:
	cd $(SERVER_DIR) && go test ./...

## docs: regenerate the OpenAPI spec into server/docs/ (swag init via go generate)
docs:
	cd $(SERVER_DIR) && go generate ./cmd/api

## migrate-up: apply all pending migrations
migrate-up: check-db-config
	@$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(MIGRATE_DB_URL)" up

## migrate-down: revert the most recent migration
migrate-down: check-db-config
	@$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(MIGRATE_DB_URL)" down 1

check-db-config:
	@test -n "$(DB_HOST)" -a -n "$(DB_USER)" -a -n "$(DB_PASSWORD)" -a -n "$(DB_NAME)" || \
		{ echo "DB_HOST, DB_USER, DB_PASSWORD and DB_NAME must be set (set them or create server/.env)"; exit 1; }
