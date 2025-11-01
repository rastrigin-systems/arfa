#!/bin/bash
# Setup GitHub Projects API automation
# This script fetches project IDs and field IDs for automated task creation

set -e

echo "🔧 GitHub Projects Automation Setup"
echo "===================================="
echo ""

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo "❌ Error: GitHub CLI (gh) not found"
    echo "Install: https://cli.github.com/"
    exit 1
fi

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "❌ Error: jq not found"
    echo "Install: brew install jq"
    exit 1
fi

echo "✅ Prerequisites check passed"
echo ""

# Check authentication
echo "🔐 Checking GitHub authentication..."
if ! gh auth status &> /dev/null; then
    echo "❌ Not authenticated. Run: gh auth login"
    exit 1
fi

echo "✅ Authenticated"
echo ""
echo "📊 Fetching project information..."
echo ""

# Get projects
PROJECTS=$(gh api graphql -f query='
query {
  user(login: "sergei-rastrigin") {
    projectsV2(first: 10) {
      nodes {
        id
        number
        title
        url
      }
    }
  }
}')

echo "$PROJECTS" | jq '.data.user.projectsV2.nodes[] | "[\(.number)] \(.title)"'
echo ""

# Extract project IDs
ENG_PROJECT_ID=$(echo "$PROJECTS" | jq -r '.data.user.projectsV2.nodes[] | select(.title | contains("Engineering")) | .id')
MKT_PROJECT_ID=$(echo "$PROJECTS" | jq -r '.data.user.projectsV2.nodes[] | select(.title | contains("Marketing")) | .id')

if [ -z "$ENG_PROJECT_ID" ]; then
    echo "❌ Error: Engineering project not found"
    echo "   Create it at: https://github.com/users/sergei-rastrigin/projects"
    exit 1
fi

if [ -z "$MKT_PROJECT_ID" ]; then
    echo "❌ Error: Marketing project not found"
    echo "   Create it at: https://github.com/users/sergei-rastrigin/projects"
    exit 1
fi

echo "✅ Found projects:"
echo "   Engineering: $ENG_PROJECT_ID"
echo "   Marketing: $MKT_PROJECT_ID"
echo ""

# Get fields for Engineering project
echo "📋 Fetching Engineering project fields..."
ENG_FIELDS=$(gh api graphql -f query='
query {
  node(id: "'$ENG_PROJECT_ID'") {
    ... on ProjectV2 {
      fields(first: 20) {
        nodes {
          ... on ProjectV2Field {
            id
            name
          }
          ... on ProjectV2SingleSelectField {
            id
            name
            options {
              id
              name
            }
          }
        }
      }
    }
  }
}')

echo "$ENG_FIELDS" | jq -r '.data.node.fields.nodes[] | "  - \(.name) (\(.id))"'
echo ""

# Get fields for Marketing project
echo "📋 Fetching Marketing project fields..."
MKT_FIELDS=$(gh api graphql -f query='
query {
  node(id: "'$MKT_PROJECT_ID'") {
    ... on ProjectV2 {
      fields(first: 20) {
        nodes {
          ... on ProjectV2Field {
            id
            name
          }
          ... on ProjectV2SingleSelectField {
            id
            name
            options {
              id
              name
            }
          }
        }
      }
    }
  }
}')

echo "$MKT_FIELDS" | jq -r '.data.node.fields.nodes[] | "  - \(.name) (\(.id))"'
echo ""

# Create configuration file
CONFIG_FILE=".github/project-config.json"
mkdir -p .github

echo "📝 Creating configuration file: $CONFIG_FILE"

cat > "$CONFIG_FILE" <<EOF
{
  "projects": {
    "engineering": {
      "id": "$ENG_PROJECT_ID",
      "title": "Ubik Engineering Roadmap",
      "fields": $(echo "$ENG_FIELDS" | jq '.data.node.fields.nodes')
    },
    "marketing": {
      "id": "$MKT_PROJECT_ID",
      "title": "Ubik Business & Marketing",
      "fields": $(echo "$MKT_FIELDS" | jq '.data.node.fields.nodes')
    }
  }
}
EOF

echo "✅ Configuration saved to $CONFIG_FILE"
echo ""

# Test creating a draft item
echo "🧪 Testing task creation..."
TEST_ITEM=$(gh api graphql -f query='
mutation {
  addProjectV2DraftIssue(input: {
    projectId: "'$ENG_PROJECT_ID'"
    title: "[TEST] Automation Setup Complete"
    body: "This is a test task created by the automation setup script. You can delete it."
  }) {
    projectItem {
      id
    }
  }
}')

TEST_ITEM_ID=$(echo "$TEST_ITEM" | jq -r '.data.addProjectV2DraftIssue.projectItem.id')

if [ -n "$TEST_ITEM_ID" ]; then
    echo "✅ Test task created successfully!"
    echo "   Task ID: $TEST_ITEM_ID"
    echo "   View: https://github.com/users/sergei-rastrigin/projects/$(echo "$PROJECTS" | jq -r '.data.user.projectsV2.nodes[] | select(.title | contains("Engineering")) | .number')"
else
    echo "❌ Failed to create test task"
fi

echo ""
echo "🎉 Setup complete!"
echo ""
echo "Next steps:"
echo "1. Review configuration: cat $CONFIG_FILE"
echo "2. Use scripts/create-project-task.sh to create tasks"
echo "3. See docs/GITHUB_PROJECTS_API.md for API reference"
echo ""
