#!/bin/bash
#
# Git Hooks Information for Ubik Enterprise
#
# NOTE: As of v0.3.0, we no longer use git pre-commit hooks for code generation.
# This script is kept for backward compatibility but does nothing.
#
# Usage: ./scripts/install-hooks.sh
#

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}ℹ️  Git Hooks Information${NC}"
echo ""
echo -e "${GREEN}✓ No git hooks to install!${NC}"
echo ""
echo -e "${BLUE}📝 Code Generation Workflow:${NC}"
echo "  - Code generation is NO LONGER automatic on commit"
echo "  - Run 'make generate' manually when you change:"
echo "    • schema.sql (database schema)"
echo "    • openapi/spec.yaml (API specification)"
echo "    • sqlc/queries/*.sql (SQL queries)"
echo ""
echo -e "${BLUE}🤖 CI/CD Handles Generation:${NC}"
echo "  - All CI/CD pipelines regenerate code automatically"
echo "  - This ensures consistency and catches errors early"
echo "  - generated/ directory is NOT committed to git"
echo ""
echo -e "${BLUE}💡 When to run 'make generate':${NC}"
echo "  - After pulling changes to source files"
echo "  - Before building or running tests locally"
echo "  - When you modify schema, API spec, or SQL queries"
echo ""
echo -e "${BLUE}📚 See docs/DEVELOPMENT.md for more details${NC}"
echo ""
