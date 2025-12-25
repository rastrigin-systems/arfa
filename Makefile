.PHONY: help install-tools install-cli uninstall-cli db-up db-down db-reset generate-erd generate-api generate-db generate-mocks generate test test-unit test-integration test-cli test-web dev dev-down dev-logs dev-rebuild dev-api dev-web run-server build build-cli build-server build-web docker-build docker-test docker-run lint lint-api lint-cli format clean

# Default target
help:
	@echo "Arfa - AI Agent Security Platform"
	@echo ""
	@echo "Setup:"
	@echo "  make install-tools    Install code generation tools (oapi-codegen, sqlc, tbls, mockgen)"
	@echo "  make db-up            Start PostgreSQL with Docker Compose"
	@echo "  make db-down          Stop PostgreSQL"
	@echo "  make db-reset         Reset database (drop and recreate)"
	@echo ""
	@echo "Code Generation:"
	@echo "  make generate         Generate all (ERD + API + DB + Mocks)"
	@echo "  make generate-erd     Generate database ERD documentation"
	@echo "  make generate-api     Generate HTTP types/handlers from OpenAPI spec"
	@echo "  make generate-db      Generate type-safe database queries from SQL"
	@echo "  make generate-mocks   Generate mocks for testing"
	@echo ""
	@echo "Development:"
	@echo "  make dev              Start all services with Docker Compose"
	@echo "  make dev-down         Stop all Docker services"
	@echo "  make dev-logs         Follow Docker Compose logs"
	@echo "  make dev-rebuild      Rebuild and restart Docker services"
	@echo "  make dev-api          Start API server with live reload (local)"
	@echo "  make dev-web          Start Next.js dev server (local)"
	@echo "  make run-server       Build and run API server (local)"
	@echo ""
	@echo "Testing:"
	@echo "  make test             Run all tests (API + CLI + Web)"
	@echo "  make test-unit        Run API unit tests"
	@echo "  make test-integration Run API integration tests (requires Docker)"
	@echo "  make test-cli         Run CLI tests"
	@echo "  make test-web         Run Next.js tests"
	@echo ""
	@echo "Build:"
	@echo "  make build            Build all (server + CLI + web)"
	@echo "  make build-server     Build API server binary"
	@echo "  make build-cli        Build CLI binary"
	@echo "  make build-web        Build Next.js production bundle"
	@echo "  make install-cli      Install CLI to /usr/local/bin"
	@echo "  make uninstall-cli    Remove CLI from /usr/local/bin"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build     Build API Docker image"
	@echo "  make docker-test      Test API Docker image"
	@echo "  make docker-run       Run API in Docker locally"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint             Run all linters (API + CLI)"
	@echo "  make lint-api         Run linters on API"
	@echo "  make lint-cli         Run linters on CLI"
	@echo "  make format           Format Go code (gofmt, goimports)"
	@echo "  make clean            Remove generated files and build artifacts"

# Configuration
DATABASE_URL ?= postgres://arfa:arfa_dev_password@localhost:5432/arfa?sslmode=disable
SERVER_PORT ?= 8080
GENERATED_DIR = generated
DOCS_DIR = docs/database

# =============================================================================
# Setup
# =============================================================================

install-tools:
	@echo "Installing code generation tools..."
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/k1LoW/tbls@latest
	go install go.uber.org/mock/mockgen@latest
	@echo "Tools installed. Verify with: which oapi-codegen sqlc tbls mockgen"

# =============================================================================
# Database
# =============================================================================

db-up:
	@echo "Starting PostgreSQL..."
	docker-compose up -d postgres
	@sleep 3
	@docker-compose exec -T postgres pg_isready -U arfa || (sleep 5 && docker-compose exec -T postgres pg_isready -U arfa)
	@echo "PostgreSQL ready at localhost:5432"

db-down:
	docker-compose down

db-reset:
	docker-compose down -v
	docker-compose up -d postgres
	@sleep 5
	@echo "Database reset complete"

# =============================================================================
# Code Generation
# =============================================================================

generate-setup:
	@mkdir -p $(GENERATED_DIR)
	@if [ ! -f $(GENERATED_DIR)/go.mod ]; then \
		echo 'module github.com/rastrigin-systems/arfa/generated' > $(GENERATED_DIR)/go.mod; \
		echo '' >> $(GENERATED_DIR)/go.mod; \
		echo 'go 1.24.5' >> $(GENERATED_DIR)/go.mod; \
		echo '' >> $(GENERATED_DIR)/go.mod; \
		echo 'require (' >> $(GENERATED_DIR)/go.mod; \
		echo '	github.com/go-chi/chi/v5 v5.0.11' >> $(GENERATED_DIR)/go.mod; \
		echo '	github.com/google/uuid v1.6.0' >> $(GENERATED_DIR)/go.mod; \
		echo '	github.com/jackc/pgx/v5 v5.5.3' >> $(GENERATED_DIR)/go.mod; \
		echo '	github.com/oapi-codegen/runtime v1.1.1' >> $(GENERATED_DIR)/go.mod; \
		echo '	go.uber.org/mock v0.4.0' >> $(GENERATED_DIR)/go.mod; \
		echo ')' >> $(GENERATED_DIR)/go.mod; \
	fi

generate-erd:
	@echo "Generating database documentation..."
	@mkdir -p $(DOCS_DIR)
	tbls doc $(DATABASE_URL) $(DOCS_DIR) --force --er-format svg
	tbls doc $(DATABASE_URL) $(DOCS_DIR) --force --er-format mermaid
	@echo "ERD generated at $(DOCS_DIR)/"

generate-api: generate-setup
	@echo "Generating API code from OpenAPI spec..."
	@mkdir -p $(GENERATED_DIR)/api
	oapi-codegen -package api -generate types,chi-server -o $(GENERATED_DIR)/api/server.gen.go platform/api-spec/spec.yaml

generate-db: generate-setup
	@echo "Generating database code from SQL queries..."
	@mkdir -p $(GENERATED_DIR)/db
	cd platform/database/sqlc && sqlc generate

generate-mocks: generate-db
	@echo "Generating mocks..."
	@mkdir -p $(GENERATED_DIR)/mocks
	mockgen -source=$(GENERATED_DIR)/db/querier.go \
		-destination=$(GENERATED_DIR)/mocks/db_mock.go \
		-package=mocks \
		-mock_names=Querier=MockQuerier

generate: generate-erd generate-api generate-db generate-mocks
	@echo "All code generation complete"

# =============================================================================
# Development (Docker Compose)
# =============================================================================

dev:
	@echo "Starting all services..."
	docker-compose up -d
	@echo ""
	@echo "Services running:"
	@echo "  Web:      http://localhost:3000"
	@echo "  API:      http://localhost:8080"
	@echo "  Database: localhost:5432"
	@echo "  Adminer:  http://localhost:8081"
	@echo ""
	@echo "Logs: make dev-logs"
	@echo "Stop: make dev-down"

dev-down:
	docker-compose down

dev-logs:
	docker-compose logs -f

dev-rebuild:
	docker-compose up -d --build

# =============================================================================
# Development (Local)
# =============================================================================

dev-api:
	@echo "Starting API with live reload..."
	@if ! command -v air > /dev/null; then \
		echo "Installing air..."; \
		go install github.com/air-verse/air@latest; \
	fi
	PORT=$(SERVER_PORT) air -c .air.toml

dev-web:
	@echo "Starting Next.js dev server..."
	cd services/web && pnpm dev

run-server: build-server
	@echo "Starting API server on port $(SERVER_PORT)..."
	DATABASE_URL=$(DATABASE_URL) PORT=$(SERVER_PORT) ./bin/arfa-server

# =============================================================================
# Testing
# =============================================================================

test: test-unit test-cli test-web
	@echo "All tests passed"

test-unit:
	cd services/api && $(MAKE) test-unit

test-integration:
	cd services/api && $(MAKE) test-integration

test-cli:
	cd services/cli && go test -v -race ./...

test-web:
	cd services/web && pnpm test

# =============================================================================
# Build
# =============================================================================

build: build-server build-cli build-web
	@echo ""
	@echo "Build complete:"
	@ls -lh bin/

build-server: generate-api generate-db
	cd services/api && $(MAKE) build

build-cli:
	cd services/cli && $(MAKE) build

build-web:
	cd services/web && pnpm build

install-cli: build-cli
	cd services/cli && $(MAKE) install

uninstall-cli:
	cd services/cli && $(MAKE) uninstall

# =============================================================================
# Docker
# =============================================================================

docker-build:
	cd services/api && $(MAKE) docker-build

docker-test: docker-build
	cd services/api && $(MAKE) docker-test

docker-run: docker-build
	cd services/api && $(MAKE) docker-run

# =============================================================================
# Code Quality
# =============================================================================

lint: lint-api lint-cli

lint-api:
	@echo "Linting API..."
	cd services/api && golangci-lint run ./...

lint-cli:
	@echo "Linting CLI..."
	cd services/cli && golangci-lint run ./...

format:
	@echo "Formatting code..."
	gofmt -s -w .
	goimports -w .

# =============================================================================
# Cleanup
# =============================================================================

clean:
	rm -rf $(GENERATED_DIR)
	rm -rf bin/
	rm -f coverage*.out
	rm -rf services/web/.next
	rm -rf services/web/out
	@echo "Cleaned"
