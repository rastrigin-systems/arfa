#!/bin/bash
set -e

# Seed Claude Code configuration from .claude/ directory
# This script:
# 1. Generates seed SQL from .claude/ directory
# 2. Applies seed SQL to database

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SEED_FILE="$PROJECT_ROOT/shared/schema/seeds/002_claude_config.sql"

# Database connection (use env var or default)
DATABASE_URL="${DATABASE_URL:-postgres://ubik:ubik_dev_password@localhost:5432/ubik}"

echo "🔧 Generating seed SQL from .claude/ directory..."
python3 "$SCRIPT_DIR/generate-claude-seed.py"

if [ ! -f "$SEED_FILE" ]; then
    echo "❌ Error: Seed file not generated: $SEED_FILE"
    exit 1
fi

echo "📊 Applying seed data to database..."

# Check if running in Docker environment
if command -v psql &> /dev/null; then
    # Use psql directly
    psql "$DATABASE_URL" -f "$SEED_FILE"
    exit_code=$?
else
    # Use docker exec
    docker exec -i ubik-postgres psql -U ubik -d ubik < "$SEED_FILE"
    exit_code=$?
fi

if [ $exit_code -eq 0 ]; then
    echo "✅ Seed data applied successfully!"
    echo ""
    echo "Verify with:"
    echo "  docker exec ubik-postgres psql -U ubik -d ubik -c 'SELECT COUNT(*) FROM agents;'"
    echo "  docker exec ubik-postgres psql -U ubik -d ubik -c 'SELECT COUNT(*) FROM skill_catalog;'"
    echo "  docker exec ubik-postgres psql -U ubik -d ubik -c \"SELECT name, docker_image FROM mcp_catalog WHERE docker_image IS NOT NULL;\""
else
    echo "❌ Error applying seed data"
    exit 1
fi
