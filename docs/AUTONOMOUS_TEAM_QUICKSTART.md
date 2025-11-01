# Autonomous AI Development Team - Quick Start Guide

**Transform any software project into an autonomous AI development team in 15 minutes.**

---

## What Is This?

A reproducible system for building software with AI agents that:
- ✅ Self-organize around GitHub Projects
- ✅ Work in parallel (no conflicts)
- ✅ Communicate via GitHub Issues
- ✅ Ship features end-to-end
- ✅ Require minimal human intervention

**Built on:** Claude Code + GitHub + Docker + Git worktrees

---

## Prerequisites

1. **Claude Code CLI** installed
2. **GitHub CLI** (`gh`) authenticated
3. **Docker** installed and running
4. **Git** repository with GitHub Projects enabled
5. **Qdrant** (optional, for knowledge retention)

---

## Installation (5 minutes)

### Step 1: Copy Agent Configurations

```bash
# Clone Ubik repository (or download agent configs)
git clone https://github.com/sergei-rastrigin/ubik-enterprise.git
cd ubik-enterprise

# Copy agent configs to your Claude Code setup
mkdir -p ~/.claude/agents
cp docs/agent-configs/*.md ~/.claude/agents/

# Agents installed:
# - coordinator.md          (orchestrator)
# - product-strategist.md   (product manager)
# - tech-lead.md            (architect)
# - go-backend-developer.md (backend engineer)
# - frontend-developer.md   (frontend engineer)
# - pr-reviewer.md          (code reviewer)
```

### Step 2: Set Up GitHub Project

```bash
# Create GitHub Project for your repository
gh project create --owner @me --title "AI Dev Team" --format table

# Add status columns
gh project field-list <PROJECT_NUMBER> --owner @me

# Required labels (create if not exists)
gh label create "status/ready" --color "0e8a16"
gh label create "status/in-progress" --color "fbca04"
gh label create "status/waiting-for-review" --color "ffa500"
gh label create "status/blocked" --color "d93f0b"
gh label create "status/done" --color "6f42c1"

gh label create "backend" --color "0052cc"
gh label create "frontend" --color "5319e7"
gh label create "epic" --color "3e4b9e"
```

### Step 3: Initialize Qdrant (Optional)

```bash
# Start Qdrant for knowledge retention
docker run -d --name ubik-qdrant -p 6333:6333 qdrant/qdrant:latest

# Configure Qdrant MCP in Claude Code
# (See Qdrant MCP setup docs)
```

---

## Usage (10 minutes)

### Test 1: Single Task Execution

**1. Create a task:**
```bash
gh issue create \
  --title "Add health check endpoint" \
  --body "Implement GET /health endpoint that returns server status" \
  --label "backend,status/ready"
```

**2. Start Coordinator:**
```bash
# In terminal 1 (Coordinator)
claude-code

# When prompted, say:
"Start the coordinator agent and monitor for ready tasks"
```

**Coordinator will:**
1. Detect issue in `status/ready`
2. Assign to `go-backend-developer` (based on `backend` label)
3. Agent creates worktree, implements feature, creates PR
4. Agent marks as `status/waiting-for-review`
5. Coordinator invokes `pr-reviewer`
6. PR reviewer merges → issue closed → Done!

**Expected time:** 10-20 minutes (fully autonomous)

### Test 2: Multi-Agent Parallel Development

**1. Create multiple tasks:**
```bash
# Backend task
gh issue create \
  --title "Implement user authentication API" \
  --body "Add POST /auth/login and POST /auth/register endpoints" \
  --label "backend,status/ready"

# Frontend task
gh issue create \
  --title "Build login page UI" \
  --body "Create login form with email/password fields" \
  --label "frontend,status/ready"

# Another backend task
gh issue create \
  --title "Add database migration for users table" \
  --body "Create migration: users table with email, password_hash" \
  --label "backend,status/ready"
```

**2. Start Coordinator (same as before)**

**Coordinator will:**
1. Assign Issue #1 → `go-backend-developer` (worktree: `../project-issue-1`)
2. Assign Issue #2 → `frontend-developer` (worktree: `../project-issue-2`)
3. Assign Issue #3 → `go-backend-developer` (queued, waits for #1)
4. Agents work in parallel
5. PRs created → Reviewed → Merged → Done

**Expected time:** 20-40 minutes (parallel execution)

### Test 3: Agent Communication & Dependencies

**1. Create frontend task that needs backend API:**
```bash
gh issue create \
  --title "Build user profile page" \
  --body "Display user profile with edit capability

  **Note:** Requires GET /users/me and PUT /users/me endpoints" \
  --label "frontend,status/ready"
```

**2. Start Coordinator**

**Coordinator will:**
1. Assign Issue #4 → `frontend-developer`
2. Frontend agent starts work
3. Frontend agent discovers missing API
4. Frontend agent comments on issue:
   ```
   Agent: frontend-developer
   Type: request
   To: go-backend-developer

   Need API endpoints:
   - GET /users/me (get current user)
   - PUT /users/me (update user profile)

   Blocking: Yes
   ```
5. **Coordinator detects message**
6. Coordinator creates subtask:
   ```
   Issue #5: "API: GET/PUT /users/me (for #4)"
   Labels: backend, status/ready, parent:#4
   ```
7. Coordinator blocks Issue #4 (`status/blocked`)
8. Coordinator assigns Issue #5 → `go-backend-developer`
9. Backend implements API → PR merged
10. **Coordinator unblocks Issue #4**
11. Frontend continues → PR merged → Done!

**Expected time:** 30-60 minutes (with dependency coordination)

---

## Advanced Usage

### Epic Breakdown by Tech Lead

**1. Create an epic:**
```bash
gh issue create \
  --title "Epic: Complete User Management System" \
  --body "Full CRUD for users + authentication + authorization" \
  --label "epic,status/ready" \
  --milestone "v1.0"
```

**2. Start Coordinator**

**Coordinator will:**
1. Detect `epic` label
2. Assign to `tech-lead`
3. Tech lead breaks down epic:
   - Issue #10: Database schema
   - Issue #11: Authentication API
   - Issue #12: User CRUD API
   - Issue #13: Login UI
   - Issue #14: User management UI
   - Issue #15: E2E tests
4. Tech lead sets dependencies:
   - #11 depends on #10
   - #12 depends on #11
   - #13 depends on #11
   - #14 depends on #12
   - #15 depends on #13, #14
5. All subtasks labeled `status/ready`
6. Coordinator begins assigning in dependency order
7. **Epic auto-ships!**

### Milestone Management

**1. Create milestone:**
```bash
gh milestone create "v1.0" --description "MVP launch"
```

**2. Create issues for milestone:**
```bash
gh issue create \
  --title "Feature: User auth" \
  --label "epic,status/ready" \
  --milestone "v1.0"

gh issue create \
  --title "Feature: Team management" \
  --label "epic,status/ready" \
  --milestone "v1.0"
```

**3. Start Coordinator**

**Coordinator will:**
1. Work through all tasks in milestone
2. Track progress (e.g., "v1.0: 45% complete")
3. When 100% → Notify `product-strategist` to plan v1.1

---

## Monitoring & Observability

### View Coordinator Logs
```bash
tail -f ~/.ubik/coordinator.log
```

**Example output:**
```
2025-11-01 20:00:00 ASSIGN #123 go-backend-developer in-progress
2025-11-01 20:05:00 DETECT-MESSAGE #234 frontend→backend api-request
2025-11-01 20:05:30 CREATE-SUBTASK #235 backend ready
2025-11-01 20:06:00 BLOCK #234 frontend blocked
2025-11-01 20:15:00 ASSIGN #235 go-backend-developer in-progress
2025-11-01 20:45:00 COMPLETE #235 backend merged
2025-11-01 20:45:30 UNBLOCK #234 frontend ready
2025-11-01 20:46:00 ASSIGN #234 frontend-developer in-progress
```

### View Agent Activity
```bash
# Active agents (tasks in-progress)
gh issue list --label="status/in-progress"

# Pending review
gh pr list --label="status/waiting-for-review"

# Blocked tasks
gh issue list --label="status/blocked"

# Today's completions
gh issue list --state=closed --search "closed:>=$(date -v-1d +%Y-%m-%d)"
```

### GitHub Project Board
```bash
# View project in browser
gh project view <PROJECT_NUMBER> --owner @me --web
```

---

## Troubleshooting

### Agent is stuck (no progress >60 minutes)

**Coordinator automatically detects and reassigns:**
```
2025-11-01 20:30:00 DETECT-STUCK #456 backend-developer reassign
```

**Manual intervention:**
```bash
# Force reset to ready
gh issue edit 456 \
  --remove-label "status/in-progress" \
  --add-label "status/ready"
```

### CI checks failed

**Agent will see failure and fix automatically.**

**If agent doesn't fix:**
```bash
# View failure logs
gh pr checks <PR_NUMBER>

# Comment with guidance
gh pr comment <PR_NUMBER> --body "Fix test failure in auth_test.go:42"
```

### Agent communication not working

**Ensure message format is correct:**
```markdown
Agent: <agent-name>
Type: <request|response|update>
To: <target-agent> (optional)

<message body>
```

**Check coordinator is running:**
```bash
ps aux | grep coordinator
```

---

## Configuration

### Coordinator Settings

Edit `~/.claude/agents/coordinator.md`:

```yaml
polling_interval: 300  # 5 minutes (default)
max_parallel_agents: 3  # Max concurrent agents
stuck_threshold: 3600  # 60 minutes
```

### Agent Customization

Each agent config is in `~/.claude/agents/<name>.md`.

**Customize:**
- Model selection (`model: sonnet` vs `model: haiku`)
- Workflows (TDD strictness, coverage targets)
- Communication style
- Integration patterns

---

## Reproducibility

### Use with ANY Project

**This setup is fully portable:**

```bash
# Project A
cd ~/projects/my-saas-app
gh repo clone my-saas-app
ubik init  # (future CLI command)

# Project B
cd ~/projects/client-project
gh repo clone client-project
ubik init

# Coordinator manages both!
```

### Package as Dockerfile

```dockerfile
# Dockerfile.ai-dev-team
FROM ubuntu:22.04

# Install dependencies
RUN apt-get update && apt-get install -y \
  git curl docker.io gh claude-code

# Copy agent configs
COPY .claude/agents/ /root/.claude/agents/

# Start coordinator
CMD ["claude-code", "--agent", "coordinator"]
```

**Deploy to cloud:**
```bash
docker build -t ai-dev-team .
docker run -d --name team -v /path/to/project:/project ai-dev-team
```

---

## What Makes This Work?

### 1. **Git Worktrees = Parallel Workspaces**
- Each agent works in isolated directory
- No file conflicts
- True parallelism

### 2. **GitHub = Communication Bus**
- Issues = Task queue
- Comments = Agent messages
- Labels = Status signals
- Projects = Visibility

### 3. **Specialized Agents = Division of Labor**
- Product → Business decisions
- Tech Lead → Architecture
- Backend → Implementation
- Frontend → UI
- Reviewer → Quality gate

### 4. **Coordinator = Operating System**
- Task assignment
- Agent communication
- Health monitoring
- Dependency resolution

### 5. **Fully Reproducible = Zero Lock-In**
- Copy agent configs → Works anywhere
- No vendor dependencies
- Open source possible
- Extensible

---

## Success Metrics

### Individual Developer
- **Time to ship feature:** 2-4 hours → 30-60 minutes (autonomous)
- **Context switching:** 10 switches/day → 0 (agents don't switch)
- **Code review wait:** 2-24 hours → 5-10 minutes (auto-review)
- **Bug introduction:** Reduced (mandatory TDD)

### Team
- **Velocity:** 10 story points/week → 30+ (parallel agents)
- **WIP limit:** Ignored → Enforced (visible on board)
- **Deploy frequency:** 1/week → Multiple/day
- **Lead time:** 5 days → <1 day

### Product
- **Feature delivery:** Predictable (autonomous)
- **Quality:** Consistent (TDD + CI)
- **Documentation:** Always current (agents update)
- **Technical debt:** Minimal (agents follow standards)

---

## Next Steps

1. **✅ Test basic task execution** (10 minutes)
2. **✅ Test parallel agents** (20 minutes)
3. **✅ Test agent communication** (30 minutes)
4. **⏳ Test epic breakdown** (60 minutes)
5. **⏳ Test milestone completion** (days)
6. **⏳ Use on real project** (weeks)
7. **⏳ Share with team** (months)

---

## Resources

- [Development Setup Snapshot](./DEV_SETUP_SNAPSHOT.md) - Your current config
- [Coordinator Agent Spec](./COORDINATOR_AGENT_SPEC.md) - Full architecture
- [Product Direction Decision](./PRODUCT_DIRECTION_DECISION.md) - Strategic context
- [Ubik Repository](https://github.com/sergei-rastrigin/ubik-enterprise) - Source code

---

## Vision

**Today:** You manually coordinate 2-3 Claude Code instances
**Tomorrow:** AI agents autonomously ship your roadmap
**Future:** Every software project has an AI dev team

**This is not a tool. This is a new way of building software.**

---

**Questions?** Open an issue on GitHub or consult the documentation.

**Ready to start?** Run the first test above and watch your AI team work! 🚀
