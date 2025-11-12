.PHONY: help install-tools install-hooks install-cli uninstall-cli db-up db-down db-reset db-seed generate-erd generate-api generate-db generate-mocks generate check-drift test test-unit test-integration test-coverage test-cli test-web test-web-e2e dev dev-api dev-web dev-all run-server build build-cli build-server build-web clean

# Default target
help:
	@echo "Ubik Enterprise - AI Agent Management Platform"
	@echo ""
	@echo "Setup Commands:"
	@echo "  make install-tools    Install code generation tools (tbls, oapi-codegen, sqlc, mockgen)"
	@echo "  make install-hooks    Show code generation workflow info (no hooks installed)"
	@echo "  make db-up           Start PostgreSQL with Docker Compose"
	@echo "  make db-down         Stop PostgreSQL"
	@echo "  make db-reset        Reset database (drop and recreate)"
	@echo "  make db-seed         Load seed data into database"
	@echo ""
	@echo "Generation Commands:"
	@echo "  make generate-erd    Generate ERD from PostgreSQL schema"
	@echo "  make generate-api    Generate Go code from OpenAPI spec"
	@echo "  make generate-db     Generate Go code from SQL queries"
	@echo "  make generate-mocks  Generate mocks from database interfaces"
	@echo "  make generate        Generate everything (ERD + API + DB + Mocks)"
	@echo ""
	@echo "Testing Commands:"
	@echo "  make test            Run all tests (API + CLI + Web) with coverage"
	@echo "  make test-unit       Run API unit tests only (fast)"
	@echo "  make test-integration Run API integration tests (requires Docker)"
	@echo "  make test-cli        Run CLI tests only"
	@echo "  make test-web        Run Next.js tests (unit + integration)"
	@echo "  make test-web-e2e    Run Next.js E2E tests with Playwright"
	@echo "  make test-coverage   Generate HTML coverage report"
	@echo ""
	@echo "Development Commands:"
	@echo "  make check-drift     Check for OpenAPI ↔ DB schema drift"
	@echo "  make dev-all         Start all services (API + Next.js) in parallel"
	@echo "  make dev-api         Start API server with live reload"
	@echo "  make dev-web         Start Next.js development server"
	@echo "  make run-server      Build and run API server (no live reload)"
	@echo "  make build           Build all artifacts (server + CLI + web)"
	@echo "  make build-server    Build server binary only"
	@echo "  make build-cli       Build CLI binary only"
	@echo "  make build-web       Build Next.js production bundle"
	@echo "  make install-cli     Install ubik CLI to /usr/local/bin (requires sudo)"
	@echo "  make uninstall-cli   Uninstall ubik CLI from /usr/local/bin (requires sudo)"
	@echo "  make clean           Clean generated files and build artifacts"

# Configuration
DATABASE_URL ?= postgres://ubik:ubik_dev_password@localhost:5432/ubik?sslmode=disable
SERVER_PORT ?= 8080
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

# Show code generation workflow info (no hooks)
install-hooks:
	@echo "ℹ️  Code Generation Workflow Information..."
	@chmod +x scripts/install-hooks.sh
	@./scripts/install-hooks.sh

# Database management
db-up:
	@echo "🐘 Starting PostgreSQL..."
	docker-compose up -d postgres
	@echo "⏳ Waiting for PostgreSQL to be ready..."
	@sleep 3
	@docker-compose exec -T postgres pg_isready -U ubik || (echo "⚠️  PostgreSQL not ready yet, waiting..." && sleep 5)
	@echo "✅ PostgreSQL is ready"
	@echo ""
	@echo "Database connection:"
	@echo "  URL: $(DATABASE_URL)"
	@echo "  Web UI: http://localhost:8081 (Adminer)"

# Start all services with Docker Compose
dev:
	@echo "🚀 Starting all services with Docker Compose..."
	docker-compose up -d
	@echo ""
	@echo "✅ All services running:"
	@echo "  🌐 Web UI:      http://localhost:3000"
	@echo "  🔌 API Server:  http://localhost:8080"
	@echo "  🗄️  Database:    localhost:5432"
	@echo "  🔧 Adminer:     http://localhost:8081"
	@echo ""
	@echo "Useful commands:"
	@echo "  docker-compose logs -f web      # View Web UI logs"
	@echo "  docker-compose logs -f api      # View API logs"
	@echo "  docker-compose logs -f postgres # View DB logs"
	@echo "  docker-compose down             # Stop all services"
	@echo "  docker-compose restart web      # Restart Web UI"
	@echo "  docker-compose restart api      # Restart API"

# Stop all Docker Compose services
dev-down:
	@echo "🛑 Stopping all services..."
	docker-compose down
	@echo "✅ All services stopped"

# View logs (defaults to all services)
dev-logs:
	docker-compose logs -f

# View Web UI logs
web-logs:
	docker-compose logs -f web

# View API logs
api-logs:
	docker-compose logs -f api

# Rebuild and restart services
dev-rebuild:
	@echo "🔨 Rebuilding all services..."
	docker-compose up -d --build
	@echo "✅ All services rebuilt and restarted"

# Rebuild and restart Web UI only
web-rebuild:
	@echo "🔨 Rebuilding Web UI service..."
	docker-compose up -d --build web
	@echo "✅ Web UI rebuilt and restarted"

# Rebuild and restart API only
api-rebuild:
	@echo "🔨 Rebuilding API service..."
	docker-compose up -d --build api
	@echo "✅ API rebuilt and restarted"

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

db-seed:
	@echo "🌱 Loading seed data into database..."
	@./scripts/seed-claude-config.sh
	@echo ""
	@echo "Test credentials (all passwords: 'password123'):"
	@echo ""
	@echo "Acme Corp (Mature Enterprise):"
	@echo "  sarah.cto@acme.com      (Owner/CTO)"
	@echo "  alex.manager@acme.com   (Manager)"
	@echo "  maria.senior@acme.com   (Developer - has agents)"
	@echo "  emma.frontend@acme.com  (Developer)"
	@echo ""
	@echo "Other Companies:"
	@echo "  jane.founder@techco.com (TechCo Owner)"
	@echo "  tom.dev@techco.com      (TechCo Developer)"
	@echo "  owner@newcorp.com       (NewCorp Owner)"
	@echo "  john@solostartup.com    (Solo Startup Owner)"

# Code generation
generate-erd:
	@echo "📊 Generating ERD from database schema..."
	@mkdir -p $(DOCS_DIR)
	tbls doc $(DATABASE_URL) $(DOCS_DIR) --force --er-format svg
	tbls doc $(DATABASE_URL) $(DOCS_DIR) --force --er-format mermaid
	@echo "🔧 Generating ERD overview (ERD.md)..."
	python3 scripts/generate-erd-overview.py
	@echo ""
	@echo "✅ ERD generation complete:"
	@echo "   - Overview:  $(DOCS_DIR)/ERD.md (auto-generated Mermaid)"
	@echo "   - Per-table: $(DOCS_DIR)/public.*.md (27 files)"
	@echo "   - SVG:       $(DOCS_DIR)/schema.svg"
	@echo "   - JSON:      $(DOCS_DIR)/schema.json"

generate-setup:
	@echo "📦 Setting up generated module..."
	@mkdir -p $(GENERATED_DIR)
	@if [ ! -f $(GENERATED_DIR)/go.mod ]; then \
		echo 'module github.com/sergeirastrigin/ubik-enterprise/generated' > $(GENERATED_DIR)/go.mod; \
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
		echo "✅ Created $(GENERATED_DIR)/go.mod"; \
	fi

generate-api: generate-setup
	@echo "🔧 Generating API code from OpenAPI spec..."
	@mkdir -p $(GENERATED_DIR)/api
	oapi-codegen -package api -generate types,chi-server -o $(GENERATED_DIR)/api/server.gen.go platform/api-spec/spec.yaml
	@echo "✅ API code generated at $(GENERATED_DIR)/api/"

generate-db: generate-setup
	@echo "🔧 Generating database code from SQL queries..."
	@mkdir -p $(GENERATED_DIR)/db
	cd platform/database/sqlc && sqlc generate
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
	@if [ -f scripts/check-drift.js ]; then \
		node scripts/check-drift.js || echo "⚠️  Drift detected - review warnings above"; \
	else \
		echo "⚠️  Drift check script not yet implemented"; \
		echo "📝 TODO: Compare openapi/spec.yaml endpoints with database schema"; \
		echo "✅ Manual verification recommended"; \
	fi

# Testing
test:
	@echo "🧪 Running all tests with coverage..."
	@echo "Testing API server..."
	cd services/api && $(MAKE) test
	@echo ""
	@echo "Testing CLI client..."
	cd services/cli && go test -v -race -coverprofile=../../coverage-cli.out ./...
	@echo ""
	@echo "Testing shared types..."
	cd pkg/types && go test -v -race -coverprofile=../../coverage-types.out ./...
	@echo ""
	@echo "Testing Next.js web app..."
	cd services/web && npm test
	@echo ""
	@echo "✅ All tests passed!"

test-unit:
	@echo "⚡ Running API unit tests (fast)..."
	cd services/api && $(MAKE) test-unit

test-integration:
	@echo "🔄 Running API integration tests (requires Docker)..."
	cd services/api && $(MAKE) test-integration

test-cli:
	@echo "🧪 Running CLI tests..."
	cd services/cli && go test -v -race ./internal/...

test-web:
	@echo "🧪 Running Next.js tests..."
	cd services/web && npm test

test-web-e2e:
	@echo "🎭 Running Next.js E2E tests with Playwright..."
	@echo "⚠️  Requires API server running on port 8080"
	cd services/web && npm run test:e2e

test-coverage:
	@echo "📊 Generating HTML coverage report..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"
	@open coverage.html || true

# Development
dev-all:
	@echo "🚀 Starting all development services..."
	@echo ""
	@echo "Starting services in parallel:"
	@echo "  - API Server:  http://localhost:$(SERVER_PORT)"
	@echo "  - Next.js Web: http://localhost:3000"
	@echo ""
	@echo "Press Ctrl+C to stop all services"
	@echo ""
	@trap 'kill 0' EXIT; \
		(DATABASE_URL=$(DATABASE_URL) PORT=$(SERVER_PORT) ./bin/ubik-server) & \
		(cd services/web && npm run dev)

dev-api:
	@echo "🚀 Starting API development server with live reload..."
	@echo "Server will run on http://localhost:$(SERVER_PORT)"
	@echo ""
	@if ! command -v air > /dev/null; then \
		echo "Installing air for live reload..."; \
		go install github.com/air-verse/air@latest; \
	fi
	PORT=$(SERVER_PORT) air -c .air.toml

dev-web:
	@echo "🚀 Starting Next.js development server..."
	@echo "Web app will run on http://localhost:3000"
	@echo "API URL: $${NEXT_PUBLIC_API_URL:-http://localhost:8080/api/v1}"
	@echo ""
	@echo "Press Ctrl+C to stop"
	@echo ""
	cd services/web && npm run dev

run-server: build-server
	@echo "🚀 Starting API server..."
	@echo "Server will run on http://localhost:$(SERVER_PORT)"
	@echo "Press Ctrl+C to stop"
	@echo ""
	@echo "To use a different port: PORT=3002 make run-server"
	@echo ""
	DATABASE_URL=$(DATABASE_URL) PORT=$(SERVER_PORT) ./bin/ubik-server

# Build
build: build-server build-cli build-web
	@echo ""
	@echo "✅ All artifacts built:"
	@echo ""
	@echo "Binaries:"
	@ls -lh bin/
	@echo ""
	@echo "Next.js:"
	@ls -lh services/web/.next/ 2>/dev/null | head -5 || echo "  (build output in services/web/.next/)"

build-server: generate-api generate-db
	@echo "🔨 Building server binary..."
	cd services/api && $(MAKE) build
	@echo "✅ Server built: bin/ubik-server"

build-cli: generate-api generate-db
	@echo "🔨 Building CLI binary..."
	@mkdir -p bin
	cd services/cli && CGO_ENABLED=0 go build -ldflags="-s -w" -o ../../bin/ubik-cli cmd/ubik/main.go
	@echo "✅ CLI built: bin/ubik-cli"
	@echo ""
	@echo "Try it out:"
	@echo "  ./bin/ubik-cli --help"
	@echo "  ./bin/ubik-cli --version"
	@echo ""
	@echo "To install system-wide:"
	@echo "  make install-cli"

build-web:
	@echo "🔨 Building Next.js production bundle..."
	cd services/web && npm run build
	@echo "✅ Next.js built: services/web/.next/"
	@echo ""
	@echo "To run production build:"
	@echo "  cd services/web && npm start"

install-cli: build-cli
	@echo "📦 Installing ubik CLI to /usr/local/bin..."
	@if [ -w /usr/local/bin ]; then \
		cp bin/ubik-cli /usr/local/bin/ubik; \
		chmod +x /usr/local/bin/ubik; \
	else \
		sudo cp bin/ubik-cli /usr/local/bin/ubik; \
		sudo chmod +x /usr/local/bin/ubik; \
	fi
	@echo "✅ Installation complete!"
	@echo ""
	@echo "Try it out:"
	@echo "  ubik --version"
	@echo "  ubik --help"
	@echo "  ubik login"

uninstall-cli:
	@echo "🗑️  Uninstalling ubik CLI..."
	@if [ -w /usr/local/bin ]; then \
		rm -f /usr/local/bin/ubik; \
	else \
		sudo rm -f /usr/local/bin/ubik; \
	fi
	@echo "✅ Uninstalled successfully"

# Cleanup
clean:
	@echo "🧹 Cleaning generated files and build artifacts..."
	rm -rf $(GENERATED_DIR)
	rm -rf bin/
	rm -f coverage.out coverage.html coverage-*.out
	rm -rf services/web/.next
	rm -rf services/web/out
	@echo "✅ Cleanup complete"

# Development helpers
lint:
	@echo "🔍 Running linters..."
	golangci-lint run ./...

format:
	@echo "✨ Formatting code..."
	gofmt -s -w .
	goimports -w .

# Docker targets (delegate to service)
docker-build:
	@echo "🐳 Building Docker image..."
	cd services/api && $(MAKE) docker-build

docker-test:
	@echo "🧪 Testing Docker image..."
	cd services/api && $(MAKE) docker-test

docker-run:
	@echo "🚀 Running Docker container..."
	cd services/api && $(MAKE) docker-run

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
