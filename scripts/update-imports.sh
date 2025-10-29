#!/bin/bash
set -e

echo "Updating import paths for monorepo structure..."

# Update API server imports
echo "Updating services/api imports..."
find services/api -name "*.go" -type f -exec sed -i '' \
    -e 's|github.com/sergeirastrigin/ubik-enterprise/internal/auth|github.com/sergeirastrigin/ubik-enterprise/services/api/internal/auth|g' \
    -e 's|github.com/sergeirastrigin/ubik-enterprise/internal/handlers|github.com/sergeirastrigin/ubik-enterprise/services/api/internal/handlers|g' \
    -e 's|github.com/sergeirastrigin/ubik-enterprise/internal/middleware|github.com/sergeirastrigin/ubik-enterprise/services/api/internal/middleware|g' \
    -e 's|github.com/sergeirastrigin/ubik-enterprise/internal/service|github.com/sergeirastrigin/ubik-enterprise/services/api/internal/service|g' \
    -e 's|github.com/sergeirastrigin/ubik-enterprise/tests/testutil|github.com/sergeirastrigin/ubik-enterprise/services/api/tests/testutil|g' \
    {} \;

# Update CLI client imports
echo "Updating services/cli imports..."
find services/cli -name "*.go" -type f -exec sed -i '' \
    -e 's|github.com/sergeirastrigin/ubik-enterprise/internal/cli|github.com/sergeirastrigin/ubik-enterprise/services/cli/internal|g' \
    {} \;

echo "Import paths updated successfully!"
