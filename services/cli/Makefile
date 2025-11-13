.PHONY: build install test test-unit test-integration test-e2e coverage clean help

# Build CLI binary
build:
	@echo "🔨 Building CLI binary..."
	mkdir -p ../../bin
	CGO_ENABLED=0 go build -ldflags="-s -w" -o ../../bin/ubik-cli ./cmd/ubik/main.go
	@echo "✅ CLI built: bin/ubik-cli"

# Install CLI to system
install: build
	@echo "📦 Installing ubik CLI to /usr/local/bin..."
	@if [ -w /usr/local/bin ]; then \
		cp ../../bin/ubik-cli /usr/local/bin/ubik; \
		chmod +x /usr/local/bin/ubik; \
	else \
		sudo cp ../../bin/ubik-cli /usr/local/bin/ubik; \
		sudo chmod +x /usr/local/bin/ubik; \
	fi
	@echo "✅ Installation complete!"
	@echo ""
	@echo "Try it out:"
	@echo "  ubik --version"
	@echo "  ubik --help"

# Uninstall CLI from system
uninstall:
	@echo "🗑️  Uninstalling ubik CLI..."
	@if [ -w /usr/local/bin ]; then \
		rm -f /usr/local/bin/ubik; \
	else \
		sudo rm -f /usr/local/bin/ubik; \
	fi
	@echo "✅ Uninstalled successfully"

# Run all tests
test: test-unit test-integration

# Run unit tests (tests alongside source code)
test-unit:
	@echo "🧪 Running CLI unit tests..."
	go test -v -race -coverprofile=../../coverage-cli-unit.out -covermode=atomic \
		./internal/... \
		./cmd/...

# Run integration tests
test-integration:
	@echo "🧪 Running CLI integration tests..."
	go test -v -race -coverprofile=../../coverage-cli-integration.out -covermode=atomic \
		./tests/integration/...

# Run e2e tests
test-e2e:
	@echo "🧪 Running CLI e2e tests..."
	@if [ -d ./tests/e2e ] && [ -n "$$(find ./tests/e2e -name '*_test.go' 2>/dev/null)" ]; then \
		go test -v -race ./tests/e2e/...; \
	else \
		echo "⚠️  No e2e tests found yet"; \
	fi

# Show test coverage
coverage:
	@echo "📊 Test coverage report..."
	@go tool cover -func=../../coverage-cli-unit.out 2>/dev/null || echo "No unit test coverage data"
	@go tool cover -func=../../coverage-cli-integration.out 2>/dev/null || echo "No integration test coverage data"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning CLI build artifacts..."
	rm -f ../../bin/ubik-cli
	rm -f ../../coverage-cli-*.out

# Show help
help:
	@echo "CLI Service Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build              Build CLI binary"
	@echo "  install            Install CLI to /usr/local/bin"
	@echo "  uninstall          Uninstall CLI from /usr/local/bin"
	@echo "  test               Run all tests (unit + integration)"
	@echo "  test-unit          Run unit tests only"
	@echo "  test-integration   Run integration tests only"
	@echo "  test-e2e           Run e2e tests only"
	@echo "  coverage           Show test coverage report"
	@echo "  clean              Clean build artifacts"
	@echo "  help               Show this help"
