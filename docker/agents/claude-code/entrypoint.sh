#!/bin/bash
set -e

echo "🚀 Starting Claude Code container..."

# Ensure .claude directory exists
mkdir -p ~/.claude

# Create minimal settings.json to skip first-run setup wizard
if [ ! -f ~/.claude/settings.json ]; then
    echo "📝 Creating default settings..."
    cat > ~/.claude/settings.json << 'EOF'
{
  "theme": "dark"
}
EOF
fi

# Write config from environment variable if provided
if [ -n "$AGENT_CONFIG" ]; then
    echo "📝 Writing agent configuration..."
    echo "$AGENT_CONFIG" > ~/.claude/config.json
fi

# Write MCP server configuration if provided
if [ -n "$MCP_CONFIG" ]; then
    echo "🔧 Writing MCP configuration..."
    echo "$MCP_CONFIG" > ~/.claude/mcp.json
fi

# Write API key if provided (for authentication)
if [ -n "$ANTHROPIC_API_KEY" ]; then
    echo "🔑 Setting API key..."
    export ANTHROPIC_API_KEY="$ANTHROPIC_API_KEY"
fi

echo "✅ Claude Code ready!"
echo "📂 Workspace: /workspace"
echo ""

# Start Claude Code in interactive mode
# Pass any command-line arguments
exec claude "$@"
