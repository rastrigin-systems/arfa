# Coordinator Agent Specification

**Purpose:** Enable autonomous AI development team operation

**Vision:** Agents self-organize, communicate, and deliver software without human intervention

---

## Core Concept

The Coordinator Agent is the "operating system" that orchestrates multiple specialized AI agents to work together as a cohesive development team.

### Key Principle: **GitHub is the Communication Bus**

All agent coordination happens through GitHub:
- **Issues** = Task queue
- **Labels** = Status signals
- **Comments** = Agent messages
- **Project Boards** = Team visibility
- **Assignees** = Task ownership

**Why GitHub?**
- ✅ Already integrated
- ✅ Visible audit trail
- ✅ Human-readable
- ✅ No new infrastructure
- ✅ Works with existing workflows

---

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                    Coordinator Agent                         │
│  (Polls GitHub every 5 minutes, orchestrates agents)         │
└────────────┬─────────────────────────────────────────────────┘
             │
             ├─── Monitors GitHub Projects
             ├─── Assigns tasks to agents
             ├─── Detects agent failures
             └─── Coordinates agent communication

             ↓ (invokes via Task tool)

┌────────────┴─────────────────────────────────────────────────┐
│                    Specialized Agents                         │
├───────────────────────────────────────────────────────────────┤
│  product-strategist    │  Prioritizes features               │
│  tech-lead             │  Breaks down epics                  │
│  go-backend-developer  │  Implements backend                 │
│  frontend-developer    │  Implements frontend                │
│  pr-reviewer           │  Reviews & merges PRs               │
└───────────────────────────────────────────────────────────────┘

             ↓ (update)

┌────────────┴─────────────────────────────────────────────────┐
│                    GitHub Projects                            │
│  (Single source of truth for all coordination)               │
└───────────────────────────────────────────────────────────────┘
```

---

## Coordinator Agent Responsibilities

### 1. Task Discovery
```bash
# Every 5 minutes:
gh issue list --state=open --label="status/ready" --json number,title,labels,assignees
```

**Find tasks that:**
- Have label `status/ready`
- Are not assigned
- Are not blocked (no `status/blocked` label)

### 2. Task Assignment
```bash
# For each ready task:
LABELS=$(gh issue view $ISSUE --json labels -q '.labels[].name')

if [[ $LABELS == *"backend"* ]]; then
  AGENT="go-backend-developer"
elif [[ $LABELS == *"frontend"* ]]; then
  AGENT="frontend-developer"
elif [[ $LABELS == *"product"* ]]; then
  AGENT="product-strategist"
else
  AGENT="tech-lead"  # Let tech-lead decide
fi

# Mark as assigned
gh issue edit $ISSUE --add-assignee "@me" --add-label "status/in-progress"

# Invoke agent via Claude Code Task tool
# Agent will create worktree, implement feature, create PR
```

### 3. Agent Communication Detection
```bash
# Monitor issue comments for agent messages
gh issue comment list $ISSUE --json author,body,createdAt

# Detect patterns like:
# "Agent: frontend-developer"
# "Needs: POST /api/v1/teams API endpoint"
# "Blocking: Yes"

# Parse message → create subtask → assign to backend agent
```

### 4. PR Review Triggering
```bash
# Detect when PR is ready for review
gh pr list --state=open --label="status/waiting-for-review" --json number

# For each PR waiting for review:
# Invoke pr-reviewer agent
# Agent reviews, merges, updates status to "Done"
```

### 5. Health Monitoring
```bash
# Detect stuck agents
ACTIVE=$(gh issue list --state=open --label="status/in-progress" --json number,updatedAt)

for issue in $ACTIVE; do
  LAST_UPDATE=$(gh issue view $issue --json updatedAt -q .updatedAt)
  TIME_SINCE=$(date diff between now and $LAST_UPDATE)

  if [[ $TIME_SINCE > 60 minutes ]]; then
    # Agent is stuck!
    gh issue comment $issue --body "⚠️ Agent appears stuck. Reassigning..."
    gh issue edit $issue --remove-label "status/in-progress" --add-label "status/ready"
    # Will be picked up on next cycle
  fi
done
```

### 6. Dependency Resolution
```bash
# Parse issue bodies for dependencies
DEPENDS=$(gh issue view $ISSUE --json body -q .body | grep -oP 'Depends on #\K\d+')

for dep in $DEPENDS; do
  DEP_STATUS=$(gh issue view $dep --json state,labels)

  if [[ $DEP_STATUS != "closed" ]]; then
    # Dependency not complete!
    gh issue edit $ISSUE --add-label "status/blocked"
    gh issue comment $ISSUE --body "🚫 Blocked by #$dep (not complete yet)"
  fi
done
```

---

## Agent Communication Protocol

### Message Format (GitHub Issue Comments)

```markdown
---
agent: <agent-name>
type: <request|response|update|blocked>
to: <target-agent-name> (optional)
---

<message content>
```

**Examples:**

#### Example 1: Frontend Needs API
```markdown
---
agent: frontend-developer
type: request
to: go-backend-developer
---

Working on issue #234 (Team Management UI).

Need new API endpoint:
- POST /api/v1/teams/{id}/members
- Add team member to team

Request: Can you implement this API endpoint?
Blocking: Yes (can't proceed without it)
```

**Coordinator detects this:**
1. Parses comment → identifies API request
2. Creates subtask: "API: POST /teams/{id}/members (for #234)"
3. Labels: `backend`, `status/ready`, `parent:#234`
4. Assigns to go-backend-developer
5. Adds `status/blocked` to #234

#### Example 2: Backend Completes API
```markdown
---
agent: go-backend-developer
type: response
to: frontend-developer
---

✅ API endpoint complete: POST /api/v1/teams/{id}/members

Implemented in PR #245
Merged to main

Details:
- Endpoint: POST /api/v1/teams/{id}/members
- Auth: Requires JWT
- Request body: { "employee_id": "uuid" }
- Response: 200 OK with updated team

Frontend can now integrate!
```

**Coordinator detects this:**
1. Parses comment → identifies completion
2. Removes `status/blocked` from #234
3. Adds comment to #234: "Dependency #245 complete, unblocking"
4. Frontend agent can now proceed

---

## Workflow Example: Full Cycle

### Scenario: User wants to build "Team Management Feature"

#### Step 1: User Creates Epic
```bash
gh issue create \
  --title "Epic: Team Management Feature" \
  --body "Add full CRUD for teams + member assignment" \
  --label "enhancement,epic" \
  --milestone "v0.4.0"
```

#### Step 2: Tech Lead Breaks Down Epic
**Coordinator detects:** New epic created
**Action:** Invokes `tech-lead` agent

**Tech Lead:**
1. Analyzes epic scope
2. Consults product-strategist for priority
3. Creates subtasks:
   - Issue #301: DB schema for team_members table
   - Issue #302: API endpoints (CRUD teams)
   - Issue #303: Frontend components (TeamList, TeamForm)
   - Issue #304: E2E tests
4. Labels each: `status/ready`, `area/backend` or `area/frontend`
5. Sets dependencies (e.g., #303 depends on #302)

#### Step 3: Backend Agent Picks Task
**Coordinator polls GitHub:**
- Finds #302 in `status/ready`
- Label: `backend` → assigns to `go-backend-developer`
- Marks: `status/in-progress`

**Backend Agent:**
1. Creates worktree: `../ubik-issue-302`
2. Writes tests (TDD)
3. Implements API endpoints
4. Runs tests (all pass)
5. Commits + pushes
6. Creates PR #350
7. Waits for CI
8. Marks issue `status/waiting-for-review`

#### Step 4: PR Reviewer Merges
**Coordinator detects:** PR #350 ready for review
**Action:** Invokes `pr-reviewer` agent

**PR Reviewer:**
1. Reviews code
2. Checks CI (all green)
3. Merges PR
4. Deletes branch
5. Marks issue #302 as `Done`
6. Comments on #303: "API complete, unblocking"

#### Step 5: Frontend Agent Picks Task
**Coordinator detects:** #303 unblocked (dependency #302 done)
**Action:** Assigns to `frontend-developer`

**Frontend Agent:**
1. Creates worktree: `../ubik-issue-303`
2. Implements UI components
3. Integrates with API
4. Writes tests
5. Creates PR #351
6. Marks `status/waiting-for-review`

#### Step 6: Cycle Continues
- PR reviewer merges #351
- E2E tests (#304) can now run
- All subtasks → Done
- Epic #300 → Done
- **Feature shipped!**

**Total human intervention:** Created one epic issue. That's it.

---

## Implementation Plan

### Phase 1: Basic Coordination (Week 1)
```bash
# coordinator.md agent

You are the Coordinator Agent.

## Workflow:
Every 5 minutes:
1. Check for tasks in status/ready
2. Assign to appropriate agent based on labels
3. Invoke agent via Task tool
4. Monitor progress via GitHub

## Commands:
gh issue list --label="status/ready"
gh issue edit --add-label "status/in-progress"
# Invoke: Task tool with agent + task description
```

**Test:** Manually create 2 issues → Coordinator assigns → Agents implement

### Phase 2: Agent Communication (Week 2)
- Parse issue comments for agent messages
- Detect dependency requests (e.g., "Need API endpoint")
- Create subtasks automatically
- Notify blocking/unblocking

**Test:** Frontend requests API → Backend implements → Frontend unblocked

### Phase 3: Health Monitoring (Week 3)
- Detect stuck agents (no update >60 minutes)
- Reassign tasks
- Report failures to user

**Test:** Kill agent mid-task → Coordinator detects → Reassigns

### Phase 4: Dependency Resolution (Week 4)
- Parse "Depends on #X" in issue bodies
- Block tasks until dependencies complete
- Auto-unblock when dependencies close

**Test:** Create task chain (#1 → #2 → #3) → All execute in order

---

## Configuration

### coordinator-agent.md
```yaml
---
name: coordinator
description: Orchestrates multi-agent development team
model: sonnet
polling_interval: 300  # 5 minutes
---

# Coordinator Agent

You orchestrate AI development teams.

## Your Loop:
1. Poll GitHub Projects (every 5 minutes)
2. Find ready tasks → Assign to agents
3. Detect agent messages → Create subtasks
4. Monitor health → Reassign stuck tasks
5. Resolve dependencies → Unblock waiting tasks

## Tools:
- gh CLI for GitHub operations
- Task tool to invoke specialized agents
- GitHub comments for agent communication

## Rules:
- Only assign tasks labeled status/ready
- Never assign same task to multiple agents
- Detect stuck agents (no update >60 min)
- Parse agent messages in issue comments
- Update project board after each action
```

### Running the Coordinator

**Option A: Manual (Development)**
```bash
# Start coordinator in terminal
ubik-coordinator start

# Runs polling loop
# Logs all actions
# Ctrl+C to stop
```

**Option B: Daemon (Production)**
```bash
# Run as background service
ubik-coordinator daemon

# Logs to ~/.ubik/coordinator.log
# Automatically restarts on failure
```

**Option C: GitHub Actions (Cloud)**
```yaml
# .github/workflows/coordinator.yml
on:
  schedule:
    - cron: '*/5 * * * *'  # Every 5 minutes

jobs:
  coordinate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run Coordinator
        run: |
          # Invoke coordinator agent via Claude Code
          # Agent checks for ready tasks, assigns to workers
```

---

## Benefits of This Design

### 1. **Fully Reproducible**
- Copy agent configs → Any project can use
- Works with any GitHub repo
- No vendor lock-in
- Open source possible

### 2. **Human-Readable**
- All coordination visible in GitHub
- Comments show agent communication
- Project board shows progress
- Easy to debug

### 3. **Gradual Adoption**
- Start with manual invocation (current)
- Add coordinator for automation
- Scale to multiple projects
- Each step adds value

### 4. **Multi-Project**
- One coordinator can manage multiple repos
- Agents work on different projects in parallel
- Shared knowledge base (Qdrant)

### 5. **Extensible**
- Add new agent types easily
- Custom workflows per project
- Plugin architecture possible

---

## Success Criteria

### v1.0 (Minimal Viable Coordinator)
- ✅ Auto-assign tasks based on labels
- ✅ Invoke specialized agents via Task tool
- ✅ Update issue status automatically
- ✅ Detect and reassign stuck agents
- ✅ Run continuously (polling loop)

### v1.1 (Agent Communication)
- ✅ Parse agent messages in comments
- ✅ Create subtasks for dependencies
- ✅ Block/unblock tasks automatically
- ✅ Notify agents of status changes

### v1.2 (Advanced Features)
- ✅ Dependency graph resolution
- ✅ Parallel task execution
- ✅ Resource limits (max N agents)
- ✅ Agent prioritization
- ✅ Health monitoring dashboard

---

## Next Steps

1. **Create coordinator-agent.md** (this week)
2. **Test with Ubik project** (next week)
3. **Package as Ubik CLI command** (`ubik coordinate start`)
4. **Document setup guide** (for reproducibility)
5. **Test with 2nd project** (validate reproducibility)
6. **Open source** (GitHub repo with examples)

---

## The Vision

**With Coordinator Agent:**

```bash
# User creates a project
mkdir my-saas-app && cd my-saas-app
git init

# User installs Ubik
ubik init

# User creates epic
gh issue create --title "Build user auth system" --label "epic"

# User starts coordinator
ubik coordinate start

# ✨ MAGIC HAPPENS ✨
# - Tech lead breaks down epic
# - Backend implements API
# - Frontend builds UI
# - PR reviewer merges PRs
# - Feature ships

# User checks status
ubik status
# Output: "Auth system: 5/5 tasks complete ✅"

# User deploys
git push heroku main
```

**Zero human intervention between epic creation and deployment.**

That's the vision. That's what Ubik is.

---

**Document Status:** 📋 Implementation Specification
**Next Action:** Create coordinator-agent.md
**Target:** v1.0 coordinator in 1 week
