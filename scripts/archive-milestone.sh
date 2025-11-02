#!/bin/bash
set -e

# Archive Milestone Script
# Closes and archives all issues from a completed milestone
# Usage: ./scripts/archive-milestone.sh --milestone v0.3.0

MILESTONE=""
PROJECT_ID="PVT_kwHOAGhClM4BG_A3"  # Ubik Engineering Roadmap
PROJECT_NUMBER=3

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --milestone)
      MILESTONE="$2"
      shift 2
      ;;
    --help)
      echo "Usage: $0 --milestone <version>"
      echo "Example: $0 --milestone v0.3.0"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

# Validate inputs
if [ -z "$MILESTONE" ]; then
  echo "❌ Error: --milestone is required"
  echo "Usage: $0 --milestone <version>"
  exit 1
fi

echo "📦 Archiving Milestone: $MILESTONE"
echo "=================================="

# Check if milestone exists
if ! gh api "/repos/sergei-rastrigin/ubik-enterprise/milestones" --jq ".[] | select(.title==\"$MILESTONE\")" > /dev/null 2>&1; then
  echo "❌ Milestone '$MILESTONE' not found"
  exit 1
fi

# Get all issues in milestone
echo -e "\n📋 Fetching issues in milestone..."
ISSUES=$(gh issue list --milestone "$MILESTONE" --state all --json number,title,state --jq '.[] | "\(.number)|\(.title)|\(.state)"')

if [ -z "$ISSUES" ]; then
  echo "✓ No issues found in milestone"
  exit 0
fi

# Count issues
TOTAL=$(echo "$ISSUES" | wc -l | tr -d ' ')
CLOSED=$(echo "$ISSUES" | grep -c "|CLOSED" || true)
OPEN=$(echo "$ISSUES" | grep -c "|OPEN" || true)
OPEN=${OPEN:-0}
CLOSED=${CLOSED:-0}

echo "✓ Found $TOTAL issues:"
echo "  - Closed: $CLOSED"
echo "  - Open: $OPEN"

# Warn if there are open issues
if [ "$OPEN" -gt 0 ]; then
  echo ""
  echo "⚠️  WARNING: Milestone has $OPEN open issues:"
  echo "$ISSUES" | grep "|OPEN" | while IFS='|' read -r num title state; do
    echo "  #$num: $title"
  done
  echo ""
  read -p "Continue archiving? (y/N): " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 0
  fi
fi

# Archive issues
echo -e "\n📦 Archiving issues..."

while IFS='|' read -r issue_num title state; do
  echo -n "  #$issue_num: "

  # Add archived label
  if gh issue edit "$issue_num" --add-label "archived" 2>/dev/null; then
    echo "✓ Labeled as archived"
  else
    echo "⚠️  Failed to add label"
  fi

  # Close if still open
  if [ "$state" = "OPEN" ]; then
    gh issue close "$issue_num" --comment "Archived as part of milestone $MILESTONE completion." 2>/dev/null || true
  fi
done <<< "$ISSUES"

# Archive items from project board
echo -e "\n📦 Archiving items from project board..."

while IFS='|' read -r issue_num title state; do
  echo -n "  #$issue_num: "

  # Get project item ID for this issue
  ITEM_ID=$(gh api graphql -f query="
    query {
      node(id: \"$PROJECT_ID\") {
        ... on ProjectV2 {
          items(first: 100) {
            nodes {
              id
              content {
                ... on Issue {
                  number
                }
              }
            }
          }
        }
      }
    }
  " 2>/dev/null | jq -r ".data.node.items.nodes[] | select(.content.number == $issue_num) | .id" 2>/dev/null)

  if [ -n "$ITEM_ID" ] && [ "$ITEM_ID" != "null" ]; then
    # Archive the item
    if gh api graphql -f query="
      mutation {
        archiveProjectV2Item(input: {
          projectId: \"$PROJECT_ID\"
          itemId: \"$ITEM_ID\"
        }) {
          item { id }
        }
      }
    " > /dev/null 2>&1; then
      echo "✓ Archived from board"
    else
      echo "⚠️  Failed to archive from board"
    fi
  else
    echo "⊘ Not on board (skipped)"
  fi
done <<< "$ISSUES"

# Close milestone
echo -e "\n🎯 Closing milestone..."
MILESTONE_NUMBER=$(gh api "/repos/sergei-rastrigin/ubik-enterprise/milestones" \
  --jq ".[] | select(.title==\"$MILESTONE\") | .number")

gh api -X PATCH "/repos/sergei-rastrigin/ubik-enterprise/milestones/$MILESTONE_NUMBER" \
  -f state="closed" > /dev/null

echo "✓ Milestone closed"

# Update docs
echo -e "\n📝 Updating documentation..."
ARCHIVE_DATE=$(date +%Y-%m-%d)
cat >> docs/MILESTONES_ARCHIVE.md <<EOF

## $MILESTONE (Completed: $ARCHIVE_DATE)

- Issues: $TOTAL ($CLOSED closed, $OPEN remaining)
- Status: Archived
- Release: https://github.com/sergei-rastrigin/ubik-enterprise/releases/tag/$MILESTONE

EOF

echo "✓ Updated docs/MILESTONES_ARCHIVE.md"

# Summary
echo -e "\n✅ Milestone $MILESTONE archived successfully!"
echo ""
echo "Summary:"
echo "  - $TOTAL issues archived"
echo "  - Items removed from project board"
echo "  - Milestone closed"
echo "  - Documentation updated"
echo ""
echo "🎉 Ready to start the next milestone!"
