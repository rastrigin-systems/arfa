.PHONY: build test test-unit test-integration docker-build docker-test deploy clean help

# Build API binary
build:
	@echo "🔨 Building API server..."
	mkdir -p ../../bin
	go build -v -o ../../bin/arfa-server ./cmd/server

# Run all tests
test: test-unit test-integration

# Run unit tests (tests alongside source code)
test-unit:
	@echo "🧪 Running unit tests..."
	go test -v -race -coverprofile=../../coverage-unit.out -covermode=atomic \
		./internal/... \
		./pkg/... \
		./cmd/...

# Run integration tests
test-integration:
	@echo "🧪 Running integration tests..."
	go test -v -race -coverprofile=../../coverage-integration.out -covermode=atomic \
		./tests/integration/...

# Show test coverage
coverage:
	@echo "📊 Test coverage report..."
	@go tool cover -func=../../coverage-unit.out 2>/dev/null || echo "No unit test coverage data"
	@go tool cover -func=../../coverage-integration.out 2>/dev/null || echo "No integration test coverage data"

# Build Docker image
docker-build:
	@echo "🐳 Building Docker image..."
	cd ../.. && docker build -f services/api/build/Dockerfile.gcp -t arfa-api:latest .

# Test Docker image locally
docker-test: docker-build
	@echo "🧪 Testing Docker image..."
	@echo "Verifying files in container..."
	docker run --rm arfa-api:latest ls -la /app/
	docker run --rm arfa-api:latest ls -la /app/platform/api-spec/
	@echo "✅ Docker image structure verified"

# Run Docker container locally
docker-run: docker-build
	@echo "🚀 Running Docker container locally..."
	@echo "API will be available at http://localhost:8080"
	docker run --rm -p 8080:8080 \
		-e DATABASE_URL="postgres://arfa:arfa_dev_password@host.docker.internal:5432/arfa?sslmode=disable" \
		arfa-api:latest

# Deploy to GCP Cloud Run
deploy:
	@echo "🚀 Deploying to GCP Cloud Run..."
	cd ../.. && gcloud builds submit --config services/api/build/cloudbuild.yaml

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f ../../bin/arfa-server
	rm -f ../../coverage-*.out

# Show help
help:
	@echo "API Service Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build              Build API binary"
	@echo "  test               Run all tests"
	@echo "  test-unit          Run unit tests only"
	@echo "  test-integration   Run integration tests only"
	@echo "  coverage           Show test coverage report"
	@echo "  docker-build       Build Docker image"
	@echo "  docker-test        Test Docker image locally"
	@echo "  docker-run         Run Docker container locally"
	@echo "  deploy             Deploy to GCP Cloud Run"
	@echo "  clean              Clean build artifacts"
	@echo "  help               Show this help"
