#!/bin/sh
set -e

echo "📁 Starting MCP Filesystem Server..."

# Write config from environment if provided
if [ -n "$MCP_CONFIG" ]; then
    echo "📝 Writing MCP configuration..."
    mkdir -p /etc/mcp
    echo "$MCP_CONFIG" > /etc/mcp/config.json
fi

# Default to /workspace if no paths specified
ALLOWED_PATHS="${ALLOWED_PATHS:-/workspace}"

echo "✅ MCP Filesystem Server ready!"
echo "📂 Allowed paths: $ALLOWED_PATHS"
echo "🌐 Listening on port 8001"
echo ""

# Start MCP filesystem server
# Use npx to run the server with allowed paths
exec npx -y @modelcontextprotocol/server-filesystem $ALLOWED_PATHS
