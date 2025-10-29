# MCP Filesystem Server Docker Image

Docker image for running MCP Filesystem Server in a containerized environment.

## Base Image

- Node.js 20 Alpine
- @modelcontextprotocol/server-filesystem installed via npm

## Build

```bash
cd docker/mcp/filesystem
docker build -t ubik/mcp-filesystem:latest .
```

## Usage

### Basic Run

```bash
docker run -d \
  --name mcp-filesystem \
  -v $(pwd):/workspace \
  -p 8001:8001 \
  ubik/mcp-filesystem:latest
```

### With Custom Paths

```bash
docker run -d \
  --name mcp-filesystem \
  -v $(pwd):/workspace \
  -v /other/path:/other \
  -e ALLOWED_PATHS="/workspace /other" \
  -p 8001:8001 \
  ubik/mcp-filesystem:latest
```

### With Configuration

```bash
docker run -d \
  --name mcp-filesystem \
  -v $(pwd):/workspace \
  -e MCP_CONFIG='{"permissions":"read-write"}' \
  -e ALLOWED_PATHS="/workspace" \
  -p 8001:8001 \
  ubik/mcp-filesystem:latest
```

### In Docker Network (for ubik CLI)

```bash
docker network create ubik-network

docker run -d \
  --name mcp-filesystem \
  --network ubik-network \
  -v $(pwd):/workspace \
  -e ALLOWED_PATHS="/workspace" \
  ubik/mcp-filesystem:latest
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MCP_CONFIG` | JSON configuration | None |
| `ALLOWED_PATHS` | Space-separated allowed paths | `/workspace` |

## Volumes

- `/workspace` - Primary workspace directory
- Any additional paths can be mounted

## Ports

- `8001` - MCP Filesystem Server port

## Security

The MCP Filesystem Server only allows access to explicitly mounted and configured paths. This provides:

- **Path Restriction**: Only configured directories accessible
- **Sandboxing**: Container isolation
- **Controlled Access**: No access outside allowed paths

## Configuration Example

```json
{
  "permissions": "read-write",
  "maxFileSize": "10MB",
  "allowedExtensions": ["*"]
}
```

## Verification

Test the image:

```bash
# Build
docker build -t ubik/mcp-filesystem:latest .

# Run
docker run -d \
  --name mcp-fs-test \
  -v $(pwd):/workspace \
  -p 8001:8001 \
  ubik/mcp-filesystem:latest

# Check logs
docker logs mcp-fs-test

# Test access (from another container or host)
# The server communicates via MCP protocol

# Cleanup
docker stop mcp-fs-test
docker rm mcp-fs-test
```

## Troubleshooting

### Volume Mount Issues

If the server can't access files:
- Verify volume mounts are correct
- Check file permissions (uid/gid)
- Ensure paths in ALLOWED_PATHS match mounts

### Port Conflicts

If port 8001 is in use:
- Change port mapping: `-p 8002:8001`
- Or use Docker networks (no port mapping needed)

### Connection Issues

If agent can't connect:
- Check container is running: `docker ps`
- Verify network connectivity
- Check logs: `docker logs mcp-filesystem`
