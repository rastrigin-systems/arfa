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

# Setup Claude authentication (hybrid auth)
# Priority: CLAUDE_API_TOKEN > ANTHROPIC_API_KEY
if [ -n "$CLAUDE_API_TOKEN" ]; then
    echo "🔑 Setting up Claude token from platform..."
    # Use claude setup-token for long-lived authentication
    echo "$CLAUDE_API_TOKEN" | claude setup-token --non-interactive 2>/dev/null || {
        echo "⚠️  Failed to setup Claude token, trying as environment variable..."
        export ANTHROPIC_API_KEY="$CLAUDE_API_TOKEN"
    }
elif [ -n "$ANTHROPIC_API_KEY" ]; then
    echo "🔑 Setting API key (legacy)..."
    export ANTHROPIC_API_KEY="$ANTHROPIC_API_KEY"
else
    echo "⚠️  No authentication token provided"
    echo "   Set CLAUDE_API_TOKEN or ANTHROPIC_API_KEY environment variable"
fi

echo "✅ Claude Code ready!"
echo "📂 Workspace: /workspace"
echo ""

# Start Claude Code in interactive mode
# Pass any command-line arguments
exec claude "$@"
