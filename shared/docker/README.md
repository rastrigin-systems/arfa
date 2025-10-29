# Ubik Docker Images

Docker images for running Claude Code CLI and MCP servers in containerized environments.

## Overview

This directory contains Dockerfiles and build scripts for:

1. **Claude Code Agent** (`ubik/claude-code`) - Claude Code CLI container
2. **MCP Filesystem Server** (`ubik/mcp-filesystem`) - File operations MCP server
3. **MCP Git Server** (`ubik/mcp-git`) - Git operations MCP server

## Quick Start

### Build All Images

```bash
# From pivot/docker directory
make build-all

# Or manually:
cd agents/claude-code && docker build -t ubik/claude-code:latest .
cd ../../mcp/filesystem && docker build -t ubik/mcp-filesystem:latest .
cd ../git && docker build -t ubik/mcp-git:latest .
```

### Test All Images

```bash
make test-all
```

## Image Details

### 1. Claude Code Agent

**Image**: `ubik/claude-code:latest`
**Base**: Node.js 20 (Bookworm)
**Size**: ~400MB

**Features**:
- Claude Code CLI installed via npm
- Config injection via environment variables
- API key authentication
- Interactive stdin/stdout

**See**: [agents/claude-code/README.md](agents/claude-code/README.md)

---

### 2. MCP Filesystem Server

**Image**: `ubik/mcp-filesystem:latest`
**Base**: Node.js 20 Alpine
**Size**: ~150MB

**Features**:
- MCP Filesystem Server (npm package)
- Configurable allowed paths
- Secure file operations
- Lightweight Alpine base

**See**: [mcp/filesystem/README.md](mcp/filesystem/README.md)

---

### 3. MCP Git Server

**Image**: `ubik/mcp-git:latest`
**Base**: Python 3.11 Slim
**Size**: ~200MB

**Features**:
- MCP Git Server (Python package)
- Full git operations support
- Repository access controls
- SSH/HTTPS authentication support

**Note**: ⚠️ Official MCP Git server is Python-based, not Node.js

**See**: [mcp/git/README.md](mcp/git/README.md)

---

## Architecture

```
┌─────────────────────────────────────────────────┐
│ Docker Network: ubik-network                    │
│                                                 │
│  ┌──────────────────────────────────────────┐  │
│  │ claude-code (Node.js)                    │  │
│  │ Port: None (stdin/stdout)                │  │
│  │ Volume: /workspace                       │  │
│  └──────────────────────────────────────────┘  │
│               ↕                                  │
│  ┌──────────────────────────────────────────┐  │
│  │ mcp-filesystem (Node.js Alpine)          │  │
│  │ Port: 8001                               │  │
│  │ Volume: /workspace                       │  │
│  └──────────────────────────────────────────┘  │
│               ↕                                  │
│  ┌──────────────────────────────────────────┐  │
│  │ mcp-git (Python)                         │  │
│  │ Port: 8002                               │  │
│  │ Volume: /workspace                       │  │
│  └──────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

## Usage Example

### 1. Create Network

```bash
docker network create ubik-network
```

### 2. Start MCP Servers

```bash
# Filesystem server
docker run -d \
  --name mcp-filesystem \
  --network ubik-network \
  -v $(pwd):/workspace \
  ubik/mcp-filesystem:latest

# Git server
docker run -d \
  --name mcp-git \
  --network ubik-network \
  -v $(pwd):/workspace \
  ubik/mcp-git:latest
```

### 3. Start Claude Code Agent

```bash
docker run -it \
  --name claude-code \
  --network ubik-network \
  -v $(pwd):/workspace \
  -e ANTHROPIC_API_KEY="your-api-key" \
  -e MCP_CONFIG='{"filesystem":{"url":"http://mcp-filesystem:8001"},"git":{"url":"http://mcp-git:8002"}}' \
  ubik/claude-code:latest
```

## Docker Compose

For production use, the `ubik` CLI will auto-generate docker-compose.yml:

```yaml
version: '3.8'

services:
  claude-code:
    image: ubik/claude-code:latest
    volumes:
      - ./workspace:/workspace
    environment:
      - ANTHROPIC_API_KEY=${API_KEY}
      - MCP_CONFIG=${MCP_CONFIG}
    networks:
      - ubik-network
    depends_on:
      - mcp-filesystem
      - mcp-git

  mcp-filesystem:
    image: ubik/mcp-filesystem:latest
    volumes:
      - ./workspace:/workspace
    networks:
      - ubik-network

  mcp-git:
    image: ubik/mcp-git:latest
    volumes:
      - ./workspace:/workspace
    networks:
      - ubik-network

networks:
  ubik-network:
    driver: bridge
```

## Build Details

### Build Times

- Claude Code: ~5-10 minutes (first build)
- MCP Filesystem: ~2-3 minutes
- MCP Git: ~3-5 minutes

### Image Sizes

- Claude Code: ~400MB
- MCP Filesystem: ~150MB
- MCP Git: ~200MB
- **Total**: ~750MB

### Optimization

Images use:
- Multi-stage builds where applicable
- Alpine Linux for smaller footprint (filesystem)
- Minimal base images
- Layer caching for faster rebuilds

## Development

### Build Single Image

```bash
cd agents/claude-code
docker build -t ubik/claude-code:dev .
```

### Test Single Image

```bash
docker run -it ubik/claude-code:dev --version
```

### Debug Container

```bash
docker run -it --entrypoint /bin/bash ubik/claude-code:latest
```

## Makefile Commands

```bash
make build-all      # Build all images
make test-all       # Test all images
make clean          # Remove all images
make push-all       # Push to registry (future)
```

## Registry (Future)

Images will be pushed to:
- Docker Hub: `docker.io/ubik/*`
- GitHub Container Registry: `ghcr.io/ubik/*`

## Troubleshooting

### Build Failures

**Issue**: npm install fails
**Solution**: Check network connectivity, try with `--network=host`

**Issue**: Permission denied on entrypoint.sh
**Solution**: Ensure chmod +x is in Dockerfile

### Runtime Issues

**Issue**: Can't connect to MCP servers
**Solution**: Check all containers are on same network

**Issue**: Volume mount not working
**Solution**: Verify volume paths exist and permissions are correct

**Issue**: API key not working
**Solution**: Verify ANTHROPIC_API_KEY is set and valid

## CI/CD

Automated builds will be configured via GitHub Actions:
- Build on push to main
- Test all images
- Push to registry on tag
- Version tagging

## Version Tags

Images follow semantic versioning:
- `latest` - Latest stable build
- `v0.2.0` - Specific version
- `dev` - Development build

## Security

- Images run as non-root where possible
- Only necessary packages installed
- Regular security updates
- Secrets via environment variables (not baked in)

## Contributing

When adding new agent or MCP images:

1. Create directory under `agents/` or `mcp/`
2. Add Dockerfile, entrypoint.sh, README.md
3. Update this README
4. Add to Makefile
5. Test thoroughly

## License

See main project LICENSE file.

---

**Status**: Phase 0 - Docker Image Creation (v0.2.0)
**Last Updated**: 2025-10-29
