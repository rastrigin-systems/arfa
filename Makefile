.PHONY: help install-tools db-up db-down db-reset generate-erd generate-api generate-db generate-mocks generate check-drift test test-unit test-integration test-coverage dev build clean

# Default target
help:
	@echo "Pivot - Enterprise AI Agent Management Platform"
	@echo ""
	@echo "Setup Commands:"
	@echo "  make install-tools    Install code generation tools (tbls, oapi-codegen, sqlc, mockgen)"
	@echo "  make db-up           Start PostgreSQL with Docker Compose"
	@echo "  make db-down         Stop PostgreSQL"
	@echo "  make db-reset        Reset database (drop and recreate)"
	@echo ""
	@echo "Generation Commands:"
	@echo "  make generate-erd    Generate ERD from PostgreSQL schema"
	@echo "  make generate-api    Generate Go code from OpenAPI spec"
	@echo "  make generate-db     Generate Go code from SQL queries"
	@echo "  make generate-mocks  Generate mocks from database interfaces"
	@echo "  make generate        Generate everything (ERD + API + DB + Mocks)"
	@echo ""
	@echo "Testing Commands:"
	@echo "  make test            Run all tests with coverage"
	@echo "  make test-unit       Run unit tests only (fast)"
	@echo "  make test-integration Run integration tests (requires Docker)"
	@echo "  make test-coverage   Generate HTML coverage report"
	@echo ""
	@echo "Development Commands:"
	@echo "  make check-drift     Check for OpenAPI ↔ DB schema drift"
	@echo "  make dev             Start development server with live reload"
	@echo "  make build           Build production binaries"
	@echo "  make clean           Clean generated files and build artifacts"

# Configuration
DATABASE_URL ?= postgres://pivot:pivot_dev_password@localhost:5432/pivot?sslmode=disable
GENERATED_DIR = generated
DOCS_DIR = docs

# Install code generation tools
install-tools:
	@echo "📦 Installing code generation tools..."
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/k1LoW/tbls@latest
	go install go.uber.org/mock/mockgen@latest
	@echo "✅ Tools installed successfully"
	@echo ""
	@echo "Verify installation:"
	@which oapi-codegen
	@which sqlc
	@which tbls
	@which mockgen

# Database management
db-up:
	@echo "🐘 Starting PostgreSQL..."
	docker-compose up -d postgres
	@echo "⏳ Waiting for PostgreSQL to be ready..."
	@sleep 3
	@docker-compose exec -T postgres pg_isready -U pivot || (echo "⚠️  PostgreSQL not ready yet, waiting..." && sleep 5)
	@echo "✅ PostgreSQL is ready"
	@echo ""
	@echo "Database connection:"
	@echo "  URL: $(DATABASE_URL)"
	@echo "  Web UI: http://localhost:8080 (Adminer)"

db-down:
	@echo "🛑 Stopping PostgreSQL..."
	docker-compose down
	@echo "✅ PostgreSQL stopped"

db-reset:
	@echo "⚠️  Resetting database (this will delete all data)..."
	docker-compose down -v
	docker-compose up -d postgres
	@echo "⏳ Waiting for PostgreSQL..."
	@sleep 5
	@echo "✅ Database reset complete"

# Code generation
generate-erd:
	@echo "📊 Generating ERD from database schema..."
	@mkdir -p $(DOCS_DIR)
	tbls doc $(DATABASE_URL) $(DOCS_DIR) --force
	@echo "✅ ERD generated at $(DOCS_DIR)/schema.md"

generate-api:
	@echo "🔧 Generating API code from OpenAPI spec..."
	@mkdir -p $(GENERATED_DIR)/api
	oapi-codegen -package api -generate types,chi-server -o $(GENERATED_DIR)/api/server.gen.go openapi/spec.yaml
	@echo "✅ API code generated at $(GENERATED_DIR)/api/"

generate-db:
	@echo "🔧 Generating database code from SQL queries..."
	@mkdir -p $(GENERATED_DIR)/db
	cd sqlc && sqlc generate
	@echo "✅ Database code generated at $(GENERATED_DIR)/db/"

generate-mocks:
	@echo "🎭 Generating mocks from database interfaces..."
	@mkdir -p $(GENERATED_DIR)/mocks
	mockgen -source=$(GENERATED_DIR)/db/querier.go \
		-destination=$(GENERATED_DIR)/mocks/db_mock.go \
		-package=mocks \
		-mock_names=Querier=MockQuerier
	@echo "✅ Mocks generated at $(GENERATED_DIR)/mocks/"

generate: generate-erd generate-api generate-db generate-mocks
	@echo ""
	@echo "✅ All code generation complete!"
	@echo ""
	@echo "Generated files:"
	@echo "  - ERD:   $(DOCS_DIR)/schema.md"
	@echo "  - API:   $(GENERATED_DIR)/api/"
	@echo "  - DB:    $(GENERATED_DIR)/db/"
	@echo "  - Mocks: $(GENERATED_DIR)/mocks/"

# Drift detection
check-drift:
	@echo "🔍 Checking for OpenAPI ↔ DB schema drift..."
	@node scripts/check-drift.js || echo "⚠️  Drift detected - review warnings above"

# Testing
test:
	@echo "🧪 Running all tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo ""
	@echo "Coverage summary:"
	@go tool cover -func=coverage.out | tail -1

test-unit:
	@echo "⚡ Running unit tests (fast)..."
	go test -v -short -race ./internal/...

test-integration:
	@echo "🔄 Running integration tests (requires Docker)..."
	go test -v -run Integration ./tests/integration/...

test-coverage:
	@echo "📊 Generating HTML coverage report..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"
	@open coverage.html || true

# Development
dev:
	@echo "🚀 Starting development server..."
	@if ! command -v air > /dev/null; then \
		echo "Installing air for live reload..."; \
		go install github.com/cosmtrek/air@latest; \
	fi
	air -c .air.toml

# Build
build:
	@echo "🔨 Building production binaries..."
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/pivot-server cmd/server/main.go
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/pivot-cli cmd/cli/main.go
	@echo "✅ Binaries built:"
	@ls -lh bin/

# Cleanup
clean:
	@echo "🧹 Cleaning generated files and build artifacts..."
	rm -rf $(GENERATED_DIR)
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "✅ Cleanup complete"

# Development helpers
lint:
	@echo "🔍 Running linters..."
	golangci-lint run ./...

format:
	@echo "✨ Formatting code..."
	gofmt -s -w .
	goimports -w .

# Docker build
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t pivot-api:latest .
	@echo "✅ Docker image built: pivot-api:latest"

# Initialize new project
init: install-tools db-up
	@echo ""
	@echo "🎉 Project initialized!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Create OpenAPI spec: vim openapi/spec.yaml"
	@echo "  2. Write SQL queries: vim sqlc/queries/employees.sql"
	@echo "  3. Generate code: make generate"
	@echo "  4. Start coding: vim internal/handlers/employees.go"
	@echo "  5. Run dev server: make dev"
