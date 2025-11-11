#!/bin/bash
set -e

echo "🔧 Starting MCP Git Server..."

# Write config from environment if provided
if [ -n "$MCP_CONFIG" ]; then
    echo "📝 Writing MCP configuration..."
    mkdir -p /etc/mcp
    echo "$MCP_CONFIG" > /etc/mcp/config.json
fi

# Default to /workspace if no repository specified
REPOSITORY="${REPOSITORY:-/workspace}"

echo "✅ MCP Git Server ready!"
echo "📂 Repository: $REPOSITORY"
echo "🌐 Listening on port 8002"
echo ""

# Start MCP git server
# The official mcp-server-git command
exec mcp-server-git --repository "$REPOSITORY"
