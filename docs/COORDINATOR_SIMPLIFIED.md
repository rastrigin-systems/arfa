# Simplified Coordinator Logic (No Status Labels Needed)

**Question:** Do I need to create `status/ready`, `status/in-progress`, etc. labels?

**Answer:** **NO!** Use GitHub's built-in features instead.

---

## How to Determine Task Status (Without Labels)

### Status: Ready to Work
```bash
# Open issue + No assignee = Ready
gh issue list \
  --state=open \
  --json number,assignees \
  --jq '.[] | select(.assignees | length == 0) | .number'
```

### Status: In Progress
```bash
# Open issue + Has assignee = In Progress
gh issue list \
  --state=open \
  --json number,assignees \
  --jq '.[] | select(.assignees | length > 0) | .number'
```

### Status: In Review
```bash
# Has open PR = In Review
gh pr list --state=open --json number,title
```

### Status: Blocked
```bash
# Option 1: Has "blocked" label (optional - only if you want this)
gh issue list --label="blocked" --json number

# Option 2: Parse issue body for unmet dependencies
gh issue view $ISSUE --json body -q .body | grep "Depends on #"
# Then check if dependency is closed
```

### Status: Done
```bash
# Closed issue = Done
gh issue list --state=closed --json number
```

---

## Minimal Labels You Need

### Labels You Already Have ✅
- `area/api` → Tells coordinator: assign to go-backend-developer
- `area/cli` → Tells coordinator: assign to go-backend-developer
- `area/web` → Tells coordinator: assign to frontend-developer
- `area/db` → Tells coordinator: assign to go-backend-developer
- `type/epic` → Tells coordinator: assign to tech-lead
- `priority/p0`, `priority/p1`, etc. → Tells coordinator: urgency

**That's all you need!**

### Optional Label (If You Want Blocking)
```bash
# Only create this if you want explicit blocking
gh label create "blocked" --color "d93f0b" --description "Blocked by dependency"
```

**Usage:**
- Agent comments: "Blocked: Waiting for #123"
- Coordinator adds "blocked" label
- Coordinator monitors #123
- When #123 closes → Coordinator removes "blocked" label

**Alternative:** Don't use "blocked" label at all. Instead:
- Parse issue body for "Depends on #X"
- Skip assignment if #X is still open
- Auto-assign when #X closes

---

## Simplified Coordinator Workflow

### Main Loop (Every 5 Minutes)
```bash
while true; do
  # 1. Find ready tasks (open + no assignee)
  READY=$(gh issue list --state=open --json number,assignees,labels | \
    jq -r '.[] | select(.assignees | length == 0) | .number')

  for issue in $READY; do
    # Check if blocked
    LABELS=$(gh issue view $issue --json labels -q '.labels[].name | join(",")')
    if [[ $LABELS == *"blocked"* ]]; then
      echo "Skipping $issue (blocked)"
      continue
    fi

    # Check dependencies
    DEPENDS=$(gh issue view $issue --json body -q .body | grep -oP 'Depends on #\K\d+')
    if [ -n "$DEPENDS" ]; then
      for dep in $DEPENDS; do
        DEP_STATE=$(gh issue view $dep --json state -q .state 2>/dev/null)
        if [ "$DEP_STATE" != "closed" ]; then
          echo "Skipping $issue (waiting for #$dep)"
          gh issue edit $issue --add-label "blocked" 2>/dev/null || true
          continue 2
        fi
      done
      # All deps met - remove blocked label if present
      gh issue edit $issue --remove-label "blocked" 2>/dev/null || true
    fi

    # Determine which agent
    if [[ $LABELS == *"area/api"* ]] || [[ $LABELS == *"area/cli"* ]]; then
      AGENT="go-backend-developer"
    elif [[ $LABELS == *"area/web"* ]]; then
      AGENT="frontend-developer"
    elif [[ $LABELS == *"type/epic"* ]]; then
      AGENT="tech-lead"
    else
      AGENT="tech-lead"  # Default
    fi

    # Assign task
    gh issue edit $issue --add-assignee "@me"
    gh issue comment $issue --body "🤖 Assigned to $AGENT"

    # Invoke agent via Task tool
    # (User would do this in Claude Code)
    echo "Task: Work on issue #$issue" > ~/.ubik/agent-queue.txt
  done

  # 2. Check for PRs ready to review
  PRS=$(gh pr list --state=open --json number | jq -r '.[].number')
  for pr in $PRS; do
    # Check CI status
    CHECKS=$(gh pr checks $pr --json state | jq -r '.[] | select(.state != "SUCCESS") | .state')
    if [ -z "$CHECKS" ]; then
      echo "PR #$pr ready for review" > ~/.ubik/pr-queue.txt
    fi
  done

  # Sleep 5 minutes
  sleep 300
done
```

---

## What You DON'T Need

### ❌ status/ready
**Use:** Open + No assignee

### ❌ status/in-progress
**Use:** Open + Has assignee

### ❌ status/waiting-for-review
**Use:** Has open PR

### ❌ status/done
**Use:** Closed

---

## Summary

**Labels You Need:**
1. ✅ `area/*` (already have) - Tells coordinator which agent
2. ✅ `type/epic` (already have) - Tells coordinator to use tech-lead
3. ✅ `priority/*` (already have) - Tells coordinator urgency
4. ⚠️ `blocked` (optional) - Only if you want explicit blocking

**Labels You DON'T Need:**
- ❌ Any `status/*` labels - Use GitHub built-ins instead!

**Result:**
- Simpler system
- Less label clutter
- Works with any repo instantly (no setup)
- Uses GitHub's native features

---

## Migration from Docs

If documentation mentions `status/ready`, `status/in-progress`, etc., **ignore it**. Those were over-engineered.

**Use this simplified approach instead.**

---

**Bottom line:** You already have everything you need! No new labels required.
