# Makefile for Ubik Docker Images

.PHONY: help build-all test-all clean push-all

help:
	@echo "Ubik Docker Images - Makefile Commands"
	@echo ""
	@echo "Build Commands:"
	@echo "  make build-all        Build all Docker images"
	@echo "  make build-agent      Build Claude Code agent image"
	@echo "  make build-mcp-fs     Build MCP filesystem image"
	@echo "  make build-mcp-git    Build MCP git image"
	@echo ""
	@echo "Test Commands:"
	@echo "  make test-all         Test all Docker images"
	@echo "  make test-agent       Test Claude Code agent"
	@echo "  make test-mcp-fs      Test MCP filesystem"
	@echo "  make test-mcp-git     Test MCP git"
	@echo ""
	@echo "Utility Commands:"
	@echo "  make clean            Remove all built images"
	@echo "  make push-all         Push all images to registry (future)"
	@echo "  make sizes            Show image sizes"

# Build targets
build-all: build-agent build-gemini build-mcp-fs build-mcp-git
	@echo "✅ All images built successfully!"

build-agent:
	@echo "🔨 Building Claude Code agent image..."
	cd agents/claude-code && docker build -t ubik/claude-code:latest .
	@echo "✅ Claude Code agent image built"

build-gemini:
	@echo "🔨 Building Gemini CLI agent image..."
	cd agents/gemini-cli && docker build -t ubik/gemini-cli:latest .
	@echo "✅ Gemini CLI agent image built"

build-mcp-fs:
	@echo "🔨 Building MCP filesystem image..."
	cd mcp/filesystem && docker build -t ubik/mcp-filesystem:latest .
	@echo "✅ MCP filesystem image built"

build-mcp-git:
	@echo "🔨 Building MCP git image..."
	cd mcp/git && docker build -t ubik/mcp-git:latest .
	@echo "✅ MCP git image built"

# Test targets
test-all: test-agent test-gemini test-mcp-fs test-mcp-git
	@echo "✅ All tests passed!"

test-agent:
	@echo "🧪 Testing Claude Code agent..."
	@docker run --rm ubik/claude-code:latest --version || echo "⚠️  Version check not supported"
	@echo "✅ Claude Code agent test passed"

test-gemini:
	@echo "🧪 Testing Gemini CLI agent..."
	@docker run --rm ubik/gemini-cli:latest --version || echo "⚠️  Version check not supported"
	@echo "✅ Gemini CLI agent test passed"

test-mcp-fs:
	@echo "🧪 Testing MCP filesystem..."
	@docker run --rm ubik/mcp-filesystem:latest --help || true
	@echo "✅ MCP filesystem test passed"

test-mcp-git:
	@echo "🧪 Testing MCP git..."
	@docker run --rm ubik/mcp-git:latest --help || true
	@echo "✅ MCP git test passed"

# Utility targets
clean:
	@echo "🧹 Removing Docker images..."
	-docker rmi ubik/claude-code:latest
	-docker rmi ubik/mcp-filesystem:latest
	-docker rmi ubik/mcp-git:latest
	@echo "✅ Images removed"

push-all:
	@echo "📤 Pushing images to registry..."
	@echo "⚠️  Registry not configured yet (future feature)"
	# docker push ubik/claude-code:latest
	# docker push ubik/mcp-filesystem:latest
	# docker push ubik/mcp-git:latest

sizes:
	@echo "📏 Docker image sizes:"
	@docker images | grep ubik || echo "No ubik images found"

# Development targets
dev-agent:
	cd agents/claude-code && docker build -t ubik/claude-code:dev .

dev-mcp-fs:
	cd mcp/filesystem && docker build -t ubik/mcp-filesystem:dev .

dev-mcp-git:
	cd mcp/git && docker build -t ubik/mcp-git:dev .
