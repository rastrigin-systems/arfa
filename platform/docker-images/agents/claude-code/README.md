# Claude Code Docker Image

Docker image for running Claude Code CLI in a containerized environment.

## Base Image

- Node.js 20 (Bookworm/Debian 12)
- Claude Code CLI installed via npm

## Build

```bash
cd docker/agents/claude-code
docker build -t ubik/claude-code:latest .
```

## Usage

### Basic Run

```bash
docker run -it \
  -v $(pwd):/workspace \
  ubik/claude-code:latest
```

### With Configuration

```bash
docker run -it \
  -v $(pwd):/workspace \
  -e AGENT_CONFIG='{"model":"claude-3-5-sonnet-20241022","temperature":0.2}' \
  -e ANTHROPIC_API_KEY="sk-ant-..." \
  ubik/claude-code:latest
```

### With MCP Servers

```bash
docker run -it \
  -v $(pwd):/workspace \
  -e AGENT_CONFIG='{"model":"claude-3-5-sonnet-20241022"}' \
  -e MCP_CONFIG='{"filesystem":{"url":"http://mcp-fs:8001"},"git":{"url":"http://mcp-git:8002"}}' \
  -e ANTHROPIC_API_KEY="sk-ant-..." \
  --network ubik-network \
  ubik/claude-code:latest
```

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `AGENT_CONFIG` | JSON config for Claude Code | No |
| `MCP_CONFIG` | JSON config for MCP servers | No |
| `ANTHROPIC_API_KEY` | Anthropic API key | Yes (unless using OAuth) |

## Authentication

Claude Code supports multiple authentication methods:

1. **API Key** (Recommended for containers)
   - Set `ANTHROPIC_API_KEY` environment variable
   - Easiest for automated/containerized usage

2. **OAuth** (Interactive)
   - Requires browser access
   - Not suitable for container environments

For containerized usage, **API key authentication is recommended**.

## Configuration

Configuration is injected via environment variables:

```json
{
  "model": "claude-3-5-sonnet-20241022",
  "temperature": 0.2,
  "max_tokens": 8192
}
```

## Volumes

- `/workspace` - Mounted project directory

## Ports

None (Claude Code CLI doesn't expose ports)

## Notes

- Claude Code stores credentials in `~/.claude/`
- Config is written to `~/.claude/config.json`
- MCP config is written to `~/.claude/mcp.json`
- Runs in interactive mode (stdin/stdout)

## Verification

Test the image:

```bash
# Build
docker build -t ubik/claude-code:latest .

# Run with version check
docker run ubik/claude-code:latest --version

# Interactive session
docker run -it \
  -v $(pwd):/workspace \
  -e ANTHROPIC_API_KEY="your-key" \
  ubik/claude-code:latest
```

## Troubleshooting

### API Key Issues

If you see authentication errors:
- Verify API key is valid
- Check if key has correct permissions
- Ensure ANTHROPIC_API_KEY is set

### Workspace Access

If Claude Code can't read files:
- Check volume mount is correct
- Verify file permissions
- Ensure workspace path is `/workspace`
