#!/bin/bash
set -e

# Start Milestone Script
# Creates a new milestone and populates it with prioritized issues from backlog
# Usage: ./scripts/start-milestone.sh --version v0.4.0 --description "Analytics" --due-date "2026-01-31"

VERSION=""
DESCRIPTION=""
DUE_DATE=""
AUTO_SPLIT=false

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --version)
      VERSION="$2"
      shift 2
      ;;
    --description)
      DESCRIPTION="$2"
      shift 2
      ;;
    --due-date)
      DUE_DATE="$2"
      shift 2
      ;;
    --auto-split)
      AUTO_SPLIT=true
      shift
      ;;
    --help)
      echo "Usage: $0 --version <vX.Y.Z> [options]"
      echo ""
      echo "Options:"
      echo "  --version       Version (required, e.g., v0.4.0)"
      echo "  --description   Milestone description"
      echo "  --due-date      Due date (YYYY-MM-DD)"
      echo "  --auto-split    Automatically split large tasks"
      echo ""
      echo "Example:"
      echo "  $0 --version v0.4.0 --description 'Analytics & Approvals' --due-date '2026-01-31' --auto-split"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

# Validate inputs
if [ -z "$VERSION" ]; then
  echo "❌ Error: --version is required"
  exit 1
fi

echo "🚀 Starting Milestone: $VERSION"
echo "================================="

# Step 1: Create milestone
echo -e "\n📅 Creating milestone..."

MILESTONE_DATA="{\"title\":\"$VERSION\""

if [ -n "$DESCRIPTION" ]; then
  MILESTONE_DATA="$MILESTONE_DATA,\"description\":\"$DESCRIPTION\""
fi

if [ -n "$DUE_DATE" ]; then
  # Convert to ISO 8601 format
  DUE_ISO="${DUE_DATE}T00:00:00Z"
  MILESTONE_DATA="$MILESTONE_DATA,\"due_on\":\"$DUE_ISO\""
fi

MILESTONE_DATA="$MILESTONE_DATA}"

gh api -X POST /repos/sergei-rastrigin/ubik-enterprise/milestones \
  --input - <<< "$MILESTONE_DATA" > /dev/null

echo "✓ Milestone '$VERSION' created"

# Step 2: Query backlog for high-priority issues
echo -e "\n📋 Querying backlog..."

BACKLOG_ISSUES=$(gh issue list \
  --label "priority/p0,priority/p1" \
  --state open \
  --json number,title,labels \
  --jq '.[] | select(.labels | map(.name) | index("status/done") | not) | "\(.number)|\(.title)|\(.labels | map(select(.name | startswith("size/"))) | .[].name // "size/unknown")"')

if [ -z "$BACKLOG_ISSUES" ]; then
  echo "⚠️  No backlog issues found"
  echo "   Create some issues with priority/p0 or priority/p1 labels first"
  exit 0
fi

BACKLOG_COUNT=$(echo "$BACKLOG_ISSUES" | wc -l | tr -d ' ')
echo "✓ Found $BACKLOG_COUNT backlog issues"

# Step 3: Display issues for review
echo -e "\n🎯 Backlog Issues (priority/p0, p1):"
echo "-----------------------------------"

LARGE_TASKS=()

echo "$BACKLOG_ISSUES" | while IFS='|' read -r num title size; do
  echo "  #$num [$size]: $title"

  if [ "$size" = "size/xl" ] || [ "$size" = "size/l" ]; then
    LARGE_TASKS+=("$num")
  fi
done

# Count large tasks
LARGE_COUNT=$(echo "$BACKLOG_ISSUES" | grep -E "size/xl|size/l" | wc -l | tr -d ' ')

if [ "$LARGE_COUNT" -gt 0 ]; then
  echo ""
  echo "⚠️  Found $LARGE_COUNT large tasks (size/l or size/xl)"
  if [ "$AUTO_SPLIT" = true ]; then
    echo "   Will automatically split after adding to milestone"
  else
    echo "   Tip: Use --auto-split to automatically split them"
  fi
fi

# Step 4: Ask for confirmation
echo ""
read -p "Add these $BACKLOG_COUNT issues to milestone $VERSION? (Y/n): " -n 1 -r
echo

if [[ $REPLY =~ ^[Nn]$ ]]; then
  echo "Cancelled. Milestone created but no issues added."
  exit 0
fi

# Step 5: Add issues to milestone
echo -e "\n📌 Adding issues to milestone..."

echo "$BACKLOG_ISSUES" | while IFS='|' read -r num title size; do
  gh issue edit "$num" --milestone "$VERSION" 2>/dev/null && echo "  ✓ Added #$num"
done

echo "✓ All issues added to milestone"

# Step 6: Move issues to Todo status
echo -e "\n📊 Updating project board..."

echo "$BACKLOG_ISSUES" | while IFS='|' read -r num title size; do
  ./scripts/update-project-status.sh --issue "$num" --status "Todo" 2>/dev/null && echo "  ✓ Moved #$num to Todo"
done

echo "✓ Project board updated"

# Step 7: Split large tasks (if enabled)
if [ "$AUTO_SPLIT" = true ] && [ "$LARGE_COUNT" -gt 0 ]; then
  echo -e "\n✂️  Splitting large tasks..."

  LARGE_ISSUES=$(echo "$BACKLOG_ISSUES" | grep -E "size/xl|size/l" | cut -d'|' -f1)

  for issue_num in $LARGE_ISSUES; do
    echo -n "  Splitting #$issue_num... "

    # Note: Actual splitting requires manual analysis or AI
    # For now, just flag them for manual splitting
    gh issue comment "$issue_num" --body "⚠️ This task is marked as large (size/l or size/xl).

Please review and split into subtasks using the GitHub Task Manager skill:
1. Analyze the task
2. Break into logical subtasks (size/s or size/m)
3. Create sub-issues with proper parent-child relationship
4. Update this issue with subtask checklist

See: \`.claude/skills/github-task-manager/SKILL.md\` for workflow." 2>/dev/null

    gh issue edit "$issue_num" --add-label "needs-splitting" 2>/dev/null

    echo "✓ Flagged for splitting"
  done

  echo "✓ Large tasks flagged (manual splitting required)"
  echo "  Tip: Use .claude/skills/github-task-manager to split tasks"
fi

# Step 8: Create milestone kickoff issue
echo -e "\n📝 Creating milestone kickoff issue..."

KICKOFF_BODY="# Milestone $VERSION - Kickoff

## Goals

$DESCRIPTION

## Timeline

- **Start Date:** $(date +%Y-%m-%d)
"

if [ -n "$DUE_DATE" ]; then
  KICKOFF_BODY="$KICKOFF_BODY- **Due Date:** $DUE_DATE"
fi

KICKOFF_BODY="$KICKOFF_BODY

## Issues

This milestone includes $BACKLOG_COUNT issues:

\`\`\`bash
gh issue list --milestone '$VERSION'
\`\`\`

## Success Criteria

- [ ] All milestone issues completed
- [ ] Tests passing
- [ ] Documentation updated
- [ ] Ready to release $VERSION

## Resources

- **Release Manager:** \`.claude/skills/release-manager/SKILL.md\`
- **Task Manager:** \`.claude/skills/github-task-manager/SKILL.md\`
- **Dev Workflow:** \`.claude/skills/development-workflow/SKILL.md\`

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)
"

gh issue create \
  --title "Milestone $VERSION - Kickoff" \
  --milestone "$VERSION" \
  --label "type/epic,priority/p0" \
  --body "$KICKOFF_BODY" > /dev/null

echo "✓ Kickoff issue created"

# Summary
echo -e "\n✅ Milestone $VERSION ready!"
echo ""
echo "Summary:"
echo "  - Milestone created: $VERSION"
echo "  - Issues added: $BACKLOG_COUNT"
echo "  - Todo status: $BACKLOG_COUNT issues"
if [ "$LARGE_COUNT" -gt 0 ]; then
  echo "  - Large tasks: $LARGE_COUNT (flagged for splitting)"
fi
echo ""
echo "🎯 Start working on issues:"
echo "   gh issue list --milestone '$VERSION' --assignee ''"
echo ""
echo "📊 View on project board:"
echo "   https://github.com/users/sergei-rastrigin/projects/3"
