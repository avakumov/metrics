
# ==================== VARIABLES ====================
MIGRATIONS_PATH := ./migrations

# переменные из .env делаем доступными для команд
include .env
export

# ==================== HELP ====================
.PHONY: help
help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s%s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""


# ==================== GO COMMANDS ====================
.PHONY: help test format build build-client build-server clean lint dev 
test: ## Run tests
	gotest -v ./...

format: ## Run format all go files
	go fmt	./...

build: build-client build-server ## Build server and client

build-client: ## Build client
	@echo "Building client..."
	go build -o cmd/agent/main ./cmd/agent

build-server: ## Build server
	@echo "Building server..."
	go build -o cmd/server/main ./cmd/server

clean: ## Remove temp files
	@echo "Cleaning..."
	rm -f cmd/server/main cmd/agent/main

lint: ## Run linting
	@echo "Linting..."
	golangci-lint run

dev: ## Start app with development options
	@echo "RUN DEV SERVER AND CLIENT..."
	air

# ==================== MIGRATIONS ====================
.PHONY: migrate migrate-up migrate-down migrate-down-all migrate-create migrate-version 
migrate: migrate-up ## Alias for migrate-up

migrate-up: ## Apply all pending migrations
	@echo "Applying migrations to '$(DATABASE_DSN)'..."
	@migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_DSN)" up
	@echo "✅ Migrations applied"

migrate-down: ## Rollback last migration
	@echo "Rolling back last migration..."
	@migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_DSN)" down 1
	@echo "✅ Migration rolled back"

migrate-down-all: ## Rollback all migrations
	@echo "Rolling back ALL migrations..."
	@migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_DSN)" down

migrate-create: ## Create new migration
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $$name
	@echo "✅ Migration created: $(MIGRATIONS_PATH)/*_$$name.sql"

migrate-version: ## Show current migration version
	@echo "Current migration version:"
	@migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_DSN)" version

