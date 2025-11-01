#!/bin/bash
#
# Update GitHub Project status for an issue
#
# Usage:
#   ./scripts/update-project-status.sh --issue 123 --status "In Review"
#   ./scripts/update-project-status.sh --issue 123 --status "In Review" --project engineering
#
# Arguments:
#   --issue NUM       Issue number to update
#   --status STATUS   Target status (e.g., "In Review", "Done", "In Progress")
#   --project NAME    Project name (engineering or marketing, default: engineering)
#

set -e

# Parse arguments
ISSUE_NUM=""
STATUS=""
PROJECT="engineering"

while [[ $# -gt 0 ]]; do
  case $1 in
    --issue)
      ISSUE_NUM="$2"
      shift 2
      ;;
    --status)
      STATUS="$2"
      shift 2
      ;;
    --project)
      PROJECT="$2"
      shift 2
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

# Validate arguments
if [ -z "$ISSUE_NUM" ]; then
  echo "Error: --issue is required"
  echo "Usage: $0 --issue NUM --status STATUS [--project NAME]"
  exit 1
fi

if [ -z "$STATUS" ]; then
  echo "Error: --status is required"
  echo "Usage: $0 --issue NUM --status STATUS [--project NAME]"
  exit 1
fi

# Load project configuration
CONFIG_FILE=".github/project-config.json"
if [ ! -f "$CONFIG_FILE" ]; then
  echo "Error: Project configuration not found: $CONFIG_FILE"
  exit 1
fi

# Extract project ID and status field ID
PROJECT_ID=$(jq -r ".projects.${PROJECT}.id" "$CONFIG_FILE")
STATUS_FIELD_ID=$(jq -r ".projects.${PROJECT}.fields[] | select(.name == \"Status\") | .id" "$CONFIG_FILE")

if [ "$PROJECT_ID" == "null" ] || [ -z "$PROJECT_ID" ]; then
  echo "Error: Project '${PROJECT}' not found in configuration"
  exit 1
fi

if [ "$STATUS_FIELD_ID" == "null" ] || [ -z "$STATUS_FIELD_ID" ]; then
  echo "Error: Status field not found for project '${PROJECT}'"
  exit 1
fi

# Get status option ID for the target status
STATUS_OPTION_ID=$(jq -r ".projects.${PROJECT}.fields[] | select(.name == \"Status\") | .options[] | select(.name == \"${STATUS}\") | .id" "$CONFIG_FILE")

if [ "$STATUS_OPTION_ID" == "null" ] || [ -z "$STATUS_OPTION_ID" ]; then
  echo "Error: Status '${STATUS}' not found in project '${PROJECT}'"
  echo "Available statuses:"
  jq -r ".projects.${PROJECT}.fields[] | select(.name == \"Status\") | .options[] | \"  - \(.name)\"" "$CONFIG_FILE"
  exit 1
fi

# Get issue content ID (required for project item lookup)
echo "🔍 Finding project item for issue #${ISSUE_NUM}..."

# First, get the issue node ID
ISSUE_NODE_ID=$(gh api graphql -f query='
  query($owner: String!, $repo: String!, $number: Int!) {
    repository(owner: $owner, name: $repo) {
      issue(number: $number) {
        id
      }
    }
  }
' -f owner="sergei-rastrigin" -f repo="ubik-enterprise" -F number="$ISSUE_NUM" -q '.data.repository.issue.id')

if [ "$ISSUE_NODE_ID" == "null" ] || [ -z "$ISSUE_NODE_ID" ]; then
  echo "Error: Issue #${ISSUE_NUM} not found"
  exit 1
fi

# Find the project item ID for this issue
ITEM_ID=$(gh api graphql -f query='
  query($projectId: ID!) {
    node(id: $projectId) {
      ... on ProjectV2 {
        items(first: 100) {
          nodes {
            id
            content {
              ... on Issue {
                id
              }
            }
          }
        }
      }
    }
  }
' -f projectId="$PROJECT_ID" -q ".data.node.items.nodes[] | select(.content.id == \"$ISSUE_NODE_ID\") | .id")

if [ "$ITEM_ID" == "null" ] || [ -z "$ITEM_ID" ]; then
  echo "Error: Issue #${ISSUE_NUM} not found in project '${PROJECT}'"
  echo "Make sure the issue is added to the project board first."
  exit 1
fi

# Update the status field
echo "📝 Updating issue #${ISSUE_NUM} status to '${STATUS}' in project '${PROJECT}'..."

gh api graphql -f query='
mutation($projectId: ID!, $itemId: ID!, $fieldId: ID!, $optionId: String!) {
  updateProjectV2ItemFieldValue(input: {
    projectId: $projectId
    itemId: $itemId
    fieldId: $fieldId
    value: {
      singleSelectOptionId: $optionId
    }
  }) {
    projectV2Item {
      id
    }
  }
}
' -f projectId="$PROJECT_ID" -f itemId="$ITEM_ID" -f fieldId="$STATUS_FIELD_ID" -f optionId="$STATUS_OPTION_ID" > /dev/null

echo "✅ Successfully updated issue #${ISSUE_NUM} to status '${STATUS}'"
