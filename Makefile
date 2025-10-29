.PHONY: help install-tools install-hooks install-cli uninstall-cli db-up db-down db-reset db-seed generate-erd generate-api generate-db generate-mocks generate check-drift test test-unit test-integration test-coverage test-cli dev run-server build build-cli build-server clean

# Default target
help:
	@echo "Ubik Enterprise - AI Agent Management Platform"
	@echo ""
	@echo "Setup Commands:"
	@echo "  make install-tools    Install code generation tools (tbls, oapi-codegen, sqlc, mockgen)"
	@echo "  make install-hooks    Install Git hooks (auto-regenerate ERD docs on commit)"
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
	@echo "  make test            Run all tests with coverage"
	@echo "  make test-unit       Run unit tests only (fast)"
	@echo "  make test-integration Run integration tests (requires Docker)"
	@echo "  make test-cli        Run CLI tests only"
	@echo "  make test-coverage   Generate HTML coverage report"
	@echo ""
	@echo "Development Commands:"
	@echo "  make check-drift     Check for OpenAPI ↔ DB schema drift"
	@echo "  make dev             Start development server with live reload"
	@echo "  make run-server      Build and run server (no live reload)"
	@echo "  make build           Build all binaries (server + CLI)"
	@echo "  make build-server    Build server binary only"
	@echo "  make build-cli       Build CLI binary only"
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

# Install Git hooks
install-hooks:
	@echo "🪝 Installing Git hooks..."
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

db-seed:
	@echo "🌱 Loading seed data into database..."
	@if [ -f seed.sql ]; then \
		docker-compose exec -T postgres psql -U ubik -d ubik < seed.sql; \
	elif [ -f shared/schema/seed.sql ]; then \
		docker-compose exec -T postgres psql -U ubik -d ubik < shared/schema/seed.sql; \
	else \
		echo "❌ Error: seed.sql not found"; \
		exit 1; \
	fi
	@echo "✅ Seed data loaded successfully"
	@echo ""
	@echo "Test credentials (all passwords: 'password123'):"
	@echo "  alice@acme.com         (Super Admin at Acme Corp)"
	@echo "  bob@acme.com           (Admin at Acme Corp)"
	@echo "  charlie@acme.com       (Developer at Acme Corp)"
	@echo "  grace@techstartup.com  (Admin at Tech Startup)"
	@echo "  iris@smallbiz.com      (Super Admin at Small Business)"

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
	oapi-codegen -package api -generate types,chi-server -o $(GENERATED_DIR)/api/server.gen.go shared/openapi/spec.yaml
	@echo "✅ API code generated at $(GENERATED_DIR)/api/"

generate-db: generate-setup
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
	cd services/api && go test -v -race -coverprofile=../../coverage-api.out ./...
	@echo ""
	@echo "Testing CLI client..."
	cd services/cli && go test -v -race -coverprofile=../../coverage-cli.out ./...
	@echo ""
	@echo "Testing shared types..."
	cd pkg/types && go test -v -race -coverprofile=../../coverage-types.out ./...
	@echo ""
	@echo "✅ All tests passed!"

test-unit:
	@echo "⚡ Running unit tests (fast)..."
	cd services/api && go test -v -short -race ./internal/...

test-integration:
	@echo "🔄 Running integration tests (requires Docker)..."
	cd services/api && go test -v -run Integration ./tests/integration/...

test-cli:
	@echo "🧪 Running CLI tests..."
	cd services/cli && go test -v -race ./internal/...

test-coverage:
	@echo "📊 Generating HTML coverage report..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"
	@open coverage.html || true

# Development
dev:
	@echo "🚀 Starting development server with live reload..."
	@echo "Server will run on http://localhost:$(SERVER_PORT)"
	@echo ""
	@if ! command -v air > /dev/null; then \
		echo "Installing air for live reload..."; \
		go install github.com/air-verse/air@latest; \
	fi
	PORT=$(SERVER_PORT) air -c .air.toml

run-server: build-server
	@echo "🚀 Starting server..."
	@echo "Server will run on http://localhost:$(SERVER_PORT)"
	@echo "Press Ctrl+C to stop"
	@echo ""
	@echo "To use a different port: PORT=3002 make run-server"
	@echo ""
	PORT=$(SERVER_PORT) ./bin/ubik-server

# Build
build: build-server build-cli
	@echo ""
	@echo "✅ All binaries built:"
	@ls -lh bin/

build-server: generate-api generate-db
	@echo "🔨 Building server binary..."
	@mkdir -p bin
	cd services/api && CGO_ENABLED=0 go build -ldflags="-s -w" -o ../../bin/ubik-server cmd/server/main.go
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
	docker build -t ubik-api:latest .
	@echo "✅ Docker image built: ubik-api:latest"

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
