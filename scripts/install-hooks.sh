#!/bin/bash
#
# Install Git hooks for Ubik Enterprise
#
# Usage: ./scripts/install-hooks.sh
#

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}📦 Installing Git hooks...${NC}"

# Check if we're in a Git repository
if [ ! -d ".git" ]; then
    echo -e "${RED}❌ Not a Git repository${NC}"
    echo "Please run this script from the repository root"
    exit 1
fi

# Create hooks directory if it doesn't exist
mkdir -p .git/hooks

# Install pre-commit hook
if [ -f ".git/hooks/pre-commit" ]; then
    echo -e "${YELLOW}⚠️  Backing up existing pre-commit hook...${NC}"
    mv .git/hooks/pre-commit .git/hooks/pre-commit.backup
fi

cp scripts/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

echo -e "${GREEN}✓ Pre-commit hook installed${NC}"

# Show what the hook does
echo ""
echo -e "${BLUE}📝 Pre-commit hook features:${NC}"
echo "  - Automatically regenerates ERD docs when source models change"
echo "  - Checks: schema.sql, openapi/spec.yaml, sqlc/queries/*.sql"
echo "  - Generates: docs/ERD.md, docs/README.md, docs/schema.json, etc."
echo "  - Adds docs/ to your commit automatically"
echo ""
echo -e "${BLUE}ℹ️  Note:${NC}"
echo "  - Go code (generated/) is NOT committed - run 'make generate' locally"
echo ""
echo -e "${BLUE}💡 To skip the hook (not recommended):${NC}"
echo "  git commit --no-verify"
echo ""
echo -e "${GREEN}✅ Git hooks installed successfully!${NC}"
