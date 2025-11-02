#!/bin/bash
set -e

# Split Large Tasks Script
# Helper for breaking down large tasks (size/l, size/xl) into subtasks
# Usage: ./scripts/split-large-tasks.sh --issue 123 [--auto]

ISSUE_NUM=""
AUTO_MODE=false

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --issue)
      ISSUE_NUM="$2"
      shift 2
      ;;
    --auto)
      AUTO_MODE=true
      shift
      ;;
    --help)
      echo "Usage: $0 --issue <number> [--auto]"
      echo ""
      echo "Options:"
      echo "  --issue    Issue number to split (required)"
      echo "  --auto     Auto-generate subtasks with Claude (future)"
      echo ""
      echo "Example:"
      echo "  $0 --issue 123"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

# Validate inputs
if [ -z "$ISSUE_NUM" ]; then
  echo "❌ Error: --issue is required"
  echo "Usage: $0 --issue <number>"
  exit 1
fi

echo "✂️  Splitting Large Task: #$ISSUE_NUM"
echo "=================================="

# Fetch issue details
echo -e "\n📋 Fetching issue details..."
ISSUE_DATA=$(gh issue view "$ISSUE_NUM" --json title,body,labels,milestone)

TITLE=$(echo "$ISSUE_DATA" | jq -r '.title')
BODY=$(echo "$ISSUE_DATA" | jq -r '.body // ""')
SIZE=$(echo "$ISSUE_DATA" | jq -r '.labels[] | select(.name | startswith("size/")) | .name')
MILESTONE=$(echo "$ISSUE_DATA" | jq -r '.milestone.title // ""')

echo "✓ Issue: $TITLE"
echo "  Size: $SIZE"
if [ -n "$MILESTONE" ]; then
  echo "  Milestone: $MILESTONE"
fi

# Check if issue is large
if [ "$SIZE" != "size/l" ] && [ "$SIZE" != "size/xl" ]; then
  echo ""
  echo "⚠️  Issue is not marked as large (size/l or size/xl)"
  echo "   Current size: $SIZE"
  echo ""
  read -p "Continue anyway? (y/N): " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 0
  fi
fi

# Display task breakdown workflow
echo ""
echo "🧠 Task Breakdown Workflow:"
echo "-----------------------------------"
echo ""
echo "1. Analyze the task and identify logical components"
echo "2. Break into subtasks (size/s or size/m each)"
echo "3. Ensure subtasks are independent and testable"
echo "4. Create sub-issues with parent-child relationship"
echo "5. Update parent issue with subtask checklist"
echo ""

if [ "$AUTO_MODE" = true ]; then
  echo "🤖 Auto mode: Using github-task-manager skill..."
  echo ""
  echo "This will use Claude to:"
  echo "  - Analyze the task description"
  echo "  - Generate subtask breakdown"
  echo "  - Create sub-issues automatically"
  echo "  - Link to parent issue"
  echo ""
  read -p "Proceed with auto-split? (Y/n): " -n 1 -r
  echo

  if [[ ! $REPLY =~ ^[Nn]$ ]]; then
    echo ""
    echo "📝 Suggested approach for Claude:"
    echo ""
    echo "   Please use the github-task-manager skill to split issue #$ISSUE_NUM into subtasks."
    echo "   Each subtask should be size/s or size/m, independent, and testable."
    echo ""
    echo "Copy the above to Claude Code and it will handle the rest!"
    exit 0
  fi
fi

# Manual workflow
echo "📝 Manual Workflow:"
echo ""
echo "Step 1: Identify subtasks (example breakdown)"
echo "  - Subtask 1: [Component A] - size/s"
echo "  - Subtask 2: [Component B] - size/m"
echo "  - Subtask 3: [Component C] - size/s"
echo ""
echo "Step 2: Create sub-issues"
echo "  Run: ./scripts/create-sub-issue.sh --parent $ISSUE_NUM --title \"Subtask title\" --size s"
echo ""
echo "Step 3: Update parent with checklist"
echo "  Add to issue body:"
echo "  ## Subtasks"
echo "  - [ ] #XX - Subtask 1"
echo "  - [ ] #YY - Subtask 2"
echo "  - [ ] #ZZ - Subtask 3"
echo ""
echo "Or use github-task-manager skill for automated workflow:"
echo "  .claude/skills/github-task-manager/SKILL.md"
echo ""

# Offer to create first subtask
echo ""
read -p "Create first subtask now? (y/N): " -n 1 -r
echo

if [[ $REPLY =~ ^[Yy]$ ]]; then
  echo ""
  read -p "Subtask title: " SUBTASK_TITLE
  read -p "Size (s/m/l): " SUBTASK_SIZE

  if [ -z "$SUBTASK_TITLE" ]; then
    echo "❌ Title required"
    exit 1
  fi

  echo ""
  echo "Creating subtask..."

  # Create subtask using create-sub-issue.sh if it exists
  if [ -f "./scripts/create-sub-issue.sh" ]; then
    ./scripts/create-sub-issue.sh \
      --parent "$ISSUE_NUM" \
      --title "$SUBTASK_TITLE" \
      --size "$SUBTASK_SIZE"
  else
    # Fallback: create manually
    LABELS="size/$SUBTASK_SIZE,type/task"
    if [ -n "$MILESTONE" ]; then
      LABELS="$LABELS"
    fi

    SUB_BODY="Part of #$ISSUE_NUM

$SUBTASK_TITLE

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)"

    SUBTASK_NUM=$(gh issue create \
      --title "$SUBTASK_TITLE" \
      --body "$SUB_BODY" \
      --label "$LABELS" \
      $([ -n "$MILESTONE" ] && echo "--milestone \"$MILESTONE\"") \
      --json number -q '.number')

    echo "✓ Created subtask #$SUBTASK_NUM"

    # Add comment to parent
    gh issue comment "$ISSUE_NUM" --body "Created subtask: #$SUBTASK_NUM" 2>/dev/null || true
  fi

  echo ""
  echo "🎯 Next steps:"
  echo "  1. Create remaining subtasks"
  echo "  2. Update parent #$ISSUE_NUM with checklist"
  echo "  3. Close parent once all subtasks complete"
fi

echo ""
echo "✅ Ready to split large task!"
echo ""
echo "Resources:"
echo "  - github-task-manager: .claude/skills/github-task-manager/SKILL.md"
echo "  - create-sub-issue.sh: ./scripts/create-sub-issue.sh --help"
