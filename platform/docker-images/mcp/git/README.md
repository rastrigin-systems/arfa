# MCP Git Server Docker Image

Docker image for running MCP Git Server in a containerized environment.

## Important Note

⚠️ **The official MCP Git server is Python-based**, not Node.js like the filesystem server.

Package: `mcp-server-git` (PyPI, not npm)

## Base Image

- Python 3.11 Slim
- Git installed
- mcp-server-git installed via pip

## Build

```bash
cd docker/mcp/git
docker build -t ubik/mcp-git:latest .
```

## Usage

### Basic Run

```bash
docker run -d \
  --name mcp-git \
  -v $(pwd):/workspace \
  -p 8002:8002 \
  ubik/mcp-git:latest
```

### With Custom Repository Path

```bash
docker run -d \
  --name mcp-git \
  -v /path/to/repo:/repo \
  -e REPOSITORY="/repo" \
  -p 8002:8002 \
  ubik/mcp-git:latest
```

### With Configuration

```bash
docker run -d \
  --name mcp-git \
  -v $(pwd):/workspace \
  -e MCP_CONFIG='{"features":["commit","push","pull"]}' \
  -e REPOSITORY="/workspace" \
  -p 8002:8002 \
  ubik/mcp-git:latest
```

### In Docker Network (for ubik CLI)

```bash
docker network create ubik-network

docker run -d \
  --name mcp-git \
  --network ubik-network \
  -v $(pwd):/workspace \
  -e REPOSITORY="/workspace" \
  ubik/mcp-git:latest
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MCP_CONFIG` | JSON configuration | None |
| `REPOSITORY` | Git repository path | `/workspace` |

## Volumes

- `/workspace` - Primary workspace/repository directory
- Any git repository can be mounted

## Ports

- `8002` - MCP Git Server port

## Git Operations

The MCP Git Server provides access to:

- `git status` - Check repository status
- `git diff` - View changes
- `git commit` - Create commits
- `git branch` - Manage branches
- `git log` - View history
- `git push` - Push to remote
- `git pull` - Pull from remote
- And more...

## Security

The MCP Git Server:

- Only operates on mounted repository
- Requires Git credentials if pushing to remote
- Container isolation provides sandboxing
- No access outside container filesystem

## Git Configuration

For operations that require authentication (push/pull), mount Git credentials:

```bash
docker run -d \
  --name mcp-git \
  -v $(pwd):/workspace \
  -v ~/.gitconfig:/root/.gitconfig:ro \
  -v ~/.ssh:/root/.ssh:ro \
  -e REPOSITORY="/workspace" \
  ubik/mcp-git:latest
```

## Verification

Test the image:

```bash
# Build
docker build -t ubik/mcp-git:latest .

# Create a test repo
mkdir -p /tmp/test-repo
cd /tmp/test-repo
git init
echo "test" > README.md
git add README.md
git commit -m "Initial commit"

# Run MCP Git server
docker run -d \
  --name mcp-git-test \
  -v /tmp/test-repo:/workspace \
  -p 8002:8002 \
  ubik/mcp-git:latest

# Check logs
docker logs mcp-git-test

# The server should show it's running on port 8002

# Cleanup
docker stop mcp-git-test
docker rm mcp-git-test
rm -rf /tmp/test-repo
```

## Troubleshooting

### Git Not Found

If git commands fail:
- Verify git is installed in container
- Check repository is a valid git repo
- Ensure .git directory exists

### Permission Issues

If git operations fail with permissions:
- Check file ownership in container
- Mount volumes with correct uid/gid
- Use `--user` flag in docker run

### Authentication Issues

If push/pull fail:
- Mount ~/.gitconfig and ~/.ssh
- Set up SSH keys or credentials
- Use HTTPS with credentials

### Connection Issues

If agent can't connect:
- Check container is running: `docker ps`
- Verify network connectivity
- Check logs: `docker logs mcp-git`
