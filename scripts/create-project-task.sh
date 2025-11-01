#!/bin/bash
# Create a task on GitHub Projects board
# Usage: ./create-project-task.sh [options]

set -e

# Default values
PROJECT="engineering"
TITLE=""
BODY=""
STATUS=""
PRIORITY=""
CONFIG_FILE=".github/project-config.json"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --project)
            PROJECT="$2"
            shift 2
            ;;
        --title)
            TITLE="$2"
            shift 2
            ;;
        --body)
            BODY="$2"
            shift 2
            ;;
        --status)
            STATUS="$2"
            shift 2
            ;;
        --priority)
            PRIORITY="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --project <engineering|marketing>  Target project board (default: engineering)"
            echo "  --title <string>                   Task title (required)"
            echo "  --body <string>                    Task description"
            echo "  --status <string>                  Status (Backlog, Ready, In Progress, etc.)"
            echo "  --priority <string>                Priority (p0, p1, p2, p3)"
            echo "  --help                             Show this help message"
            echo ""
            echo "Examples:"
            echo "  # Create engineering task"
            echo "  $0 --project engineering --title \"Feature: Add API endpoint\" --status \"Ready\" --priority \"p1\""
            echo ""
            echo "  # Create marketing task"
            echo "  $0 --project marketing --title \"Campaign: Beta Launch\" --body \"Acquire 5-10 beta customers\""
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Run '$0 --help' for usage"
            exit 1
            ;;
    esac
done

# Validate required arguments
if [ -z "$TITLE" ]; then
    echo "❌ Error: --title is required"
    echo "Run '$0 --help' for usage"
    exit 1
fi

# Check if config file exists
if [ ! -f "$CONFIG_FILE" ]; then
    echo "❌ Error: Configuration file not found: $CONFIG_FILE"
    echo "Run './scripts/setup-project-automation.sh' first"
    exit 1
fi

# Get project ID from config
PROJECT_ID=$(jq -r ".projects.$PROJECT.id" "$CONFIG_FILE")
if [ -z "$PROJECT_ID" ] || [ "$PROJECT_ID" = "null" ]; then
    echo "❌ Error: Project '$PROJECT' not found in configuration"
    exit 1
fi

echo "📝 Creating task on $PROJECT project..."
echo "   Title: $TITLE"

# Create draft item
ITEM_RESPONSE=$(gh api graphql -f query='
mutation {
  addProjectV2DraftIssue(input: {
    projectId: "'$PROJECT_ID'"
    title: "'"$TITLE"'"
    body: "'"$BODY"'"
  }) {
    projectItem {
      id
    }
  }
}')

ITEM_ID=$(echo "$ITEM_RESPONSE" | jq -r '.data.addProjectV2DraftIssue.projectItem.id')

if [ -z "$ITEM_ID" ] || [ "$ITEM_ID" = "null" ]; then
    echo "❌ Error: Failed to create task"
    echo "$ITEM_RESPONSE" | jq '.'
    exit 1
fi

echo "✅ Task created: $ITEM_ID"

# Set status if provided
if [ -n "$STATUS" ]; then
    echo "   Setting status: $STATUS"

    # Get Status field ID and option ID
    STATUS_FIELD_ID=$(jq -r ".projects.$PROJECT.fields[] | select(.name == \"Status\") | .id" "$CONFIG_FILE")
    STATUS_OPTION_ID=$(jq -r ".projects.$PROJECT.fields[] | select(.name == \"Status\") | .options[] | select(.name == \"$STATUS\") | .id" "$CONFIG_FILE")

    if [ -n "$STATUS_FIELD_ID" ] && [ -n "$STATUS_OPTION_ID" ] && [ "$STATUS_OPTION_ID" != "null" ]; then
        gh api graphql -f query='
        mutation {
          updateProjectV2ItemFieldValue(input: {
            projectId: "'$PROJECT_ID'"
            itemId: "'$ITEM_ID'"
            fieldId: "'$STATUS_FIELD_ID'"
            value: {
              singleSelectOptionId: "'$STATUS_OPTION_ID'"
            }
          }) {
            projectV2Item {
              id
            }
          }
        }' > /dev/null
        echo "   ✅ Status set"
    else
        echo "   ⚠️  Warning: Status field/option not found in config"
    fi
fi

# Set priority if provided
if [ -n "$PRIORITY" ]; then
    echo "   Setting priority: $PRIORITY"

    # Get Priority field ID and option ID
    PRIORITY_FIELD_ID=$(jq -r ".projects.$PROJECT.fields[] | select(.name == \"Priority\") | .id" "$CONFIG_FILE")
    PRIORITY_OPTION_ID=$(jq -r ".projects.$PROJECT.fields[] | select(.name == \"Priority\") | .options[] | select(.name == \"$PRIORITY\") | .id" "$CONFIG_FILE")

    if [ -n "$PRIORITY_FIELD_ID" ] && [ -n "$PRIORITY_OPTION_ID" ] && [ "$PRIORITY_OPTION_ID" != "null" ]; then
        gh api graphql -f query='
        mutation {
          updateProjectV2ItemFieldValue(input: {
            projectId: "'$PROJECT_ID'"
            itemId: "'$ITEM_ID'"
            fieldId: "'$PRIORITY_FIELD_ID'"
            value: {
              singleSelectOptionId: "'$PRIORITY_OPTION_ID'"
            }
          }) {
            projectV2Item {
              id
            }
          }
        }' > /dev/null
        echo "   ✅ Priority set"
    else
        echo "   ⚠️  Warning: Priority field/option not found in config"
    fi
fi

echo ""
echo "🎉 Task created successfully!"
echo "   Project: $PROJECT"
echo "   Task ID: $ITEM_ID"
echo "   View: https://github.com/users/sergei-rastrigin/projects"
echo ""
