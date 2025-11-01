# GitHub Projects Setup - COMPLETE ✅

**Date:** 2025-11-01
**Status:** Fully Automated - Ready for Agent Integration

---

## 🎉 What's Working

### ✅ Two Project Boards Created
1. **Ubik Engineering Roadmap** (Project #3)
   - Status: Backlog, Todo, In Progress, Blocked, In Review, Done
   - Custom fields: Effort, Owner, Dependencies

2. **Ubik Business & Marketing** (Project #4)
   - Status: Backlog, Planning, In Progress, In Review, Launched, Monitoring
   - Custom fields: Channel, Target Audience, Expected Impact, Launch Date, Metrics

### ✅ Labels Created (49 labels)
- Priority: `priority/p0`, `priority/p1`, `priority/p2`, `priority/p3`
- Component: `area/api`, `area/cli`, `area/web`, `area/db`, `area/docs`, `area/infra`, `area/testing`
- Type: `type/feature`, `type/bug`, `type/epic`, `type/research`, `type/refactor`, `type/chore`
- Size: `size/xs`, `size/s`, `size/m`, `size/l`, `size/xl`
- Impact: `impact/revenue`, `impact/retention`, `impact/acquisition`, `impact/efficiency`
- Channel: `channel/content`, `channel/social`, `channel/email`, `channel/ads`, `channel/seo`, `channel/partnerships`
- Funnel: `funnel/awareness`, `funnel/consideration`, `funnel/conversion`, `funnel/retention`
- Campaign: `campaign/launch`, `campaign/feature`, `campaign/education`, `campaign/event`
- Business: `biz-priority/revenue`, `biz-priority/brand`, `biz-priority/pipeline`, `biz-priority/research`

### ✅ Milestones Created (7 milestones)
**Engineering:**
- v0.3.0 - Web UI MVP (Feb 15, 2025)
- v0.4.0 - Beta Ready (Mar 31, 2025)
- v0.5.0 - Public Launch (May 31, 2025)
- v0.6.0 - Enterprise Features (Aug 31, 2025)

**Marketing:**
- Beta Launch Campaign (Mar 1, 2025)
- Public Launch Campaign (May 31, 2025)
- Growth Campaign Q3 (Sep 30, 2025)

### ✅ Automation Scripts
- `scripts/setup-project-automation.sh` - Fetches project IDs and field IDs
- `scripts/create-project-task.sh` - Creates tasks programmatically
- `.github/project-config.json` - Configuration with all IDs

### ✅ Authentication
- GitHub Personal Access Token created with `project` scope
- Token saved in `.env.github` (gitignored)
- Verified working with test tasks

---

## 🚀 How to Use

### Create Engineering Task

```bash
./scripts/create-project-task.sh \
  --project engineering \
  --title "Feature: Add API endpoint" \
  --body "Implement GET /usage-records endpoint" \
  --status "Backlog"
```

### Create Marketing Task

```bash
./scripts/create-project-task.sh \
  --project marketing \
  --title "Campaign: Beta Launch" \
  --body "Acquire 5-10 beta customers" \
  --status "Planning"
```

### Use in Scripts

```bash
# Source token
source .env.github

# Create task
./scripts/create-project-task.sh \
  --project engineering \
  --title "Your task" \
  --status "Todo"
```

---

## 🤖 Agent Integration

### For AI Agents

Agents can now create tasks programmatically:

**go-backend-developer:**
```bash
./scripts/create-project-task.sh \
  --project engineering \
  --title "Feature: POST /usage-records endpoint" \
  --body "Implement usage tracking API for CLI telemetry" \
  --status "Backlog"
```

**product-strategist:**
```bash
./scripts/create-project-task.sh \
  --project marketing \
  --title "Content: Product positioning document" \
  --body "Define value proposition and key messaging" \
  --status "Planning"
```

**frontend-developer:**
```bash
./scripts/create-project-task.sh \
  --project engineering \
  --title "Component: Agent Configuration Form" \
  --body "Build React form for agent config with validation" \
  --status "Todo"
```

---

## 📁 Files Created

```
ubik-enterprise/
├── .env.github                           # GitHub token (gitignored)
├── .github/
│   └── project-config.json               # Project IDs and field IDs
├── scripts/
│   ├── setup-project-automation.sh       # Setup script
│   └── create-project-task.sh            # Task creation script
└── docs/
    ├── GITHUB_PROJECTS_SETUP.md          # Complete setup guide
    ├── GITHUB_PROJECTS_API.md            # API reference
    ├── PROJECTS_SETUP_COMPLETE.md        # Initial summary
    ├── SETUP_GITHUB_TOKEN.md             # Token creation guide
    └── GITHUB_PROJECTS_COMPLETE.md       # This file
```

---

## 🧪 Test Results

**Test 1: Engineering task**
- ✅ Created task: "Feature: Web UI - Agent Configuration Dashboard"
- ✅ Status set to "Backlog"
- ✅ Visible in project board

**Test 2: Marketing task**
- ✅ Created task: "Campaign: Beta Customer Outreach"
- ✅ Status set to "Planning"
- ✅ Visible in project board

**Test 3: Automation setup**
- ✅ Setup script ran successfully
- ✅ Configuration file generated
- ✅ All field IDs captured

---

## 📊 Project Structure

### Engineering Board Fields
- **Status** (Single select): Backlog, Todo, In Progress, Blocked, In Review, Done
- **Effort** (Number): Story points
- **Owner** (Text): Task owner
- **Dependencies** (Text): Issue dependencies

### Marketing Board Fields
- **Status** (Single select): Backlog, Planning, In Progress, In Review, Launched, Monitoring
- **Channel** (Single select): Blog, Social, Email, Paid, Partnerships
- **Target Audience** (Single select): IT Admins, Developers, CTOs, Security Teams, Finance
- **Expected Impact** (Single select): High, Medium
- **Launch Date** (Date): Campaign launch date
- **Metrics** (Text): KPIs to track

---

## 🔒 Security

**Token Security:**
- ✅ Token stored in `.env.github` (gitignored)
- ✅ Token has minimal scopes: `project`, `read:org`, `repo`, `workflow`
- ✅ Token can be revoked at: https://github.com/settings/tokens
- ⚠️ **Never commit token to git!**

**Best Practices:**
- Use `source .env.github` to load token when needed
- Rotate token every 90 days
- Don't share token in Slack/email
- Don't use token in CI/CD (use GitHub Actions secrets)

---

## 🎯 Next Steps

### 1. Populate Boards with Initial Tasks

**Engineering (v0.3.0 - Web UI MVP):**
```bash
# P0 - Critical
./scripts/create-project-task.sh --project engineering \
  --title "Feature: Web UI - Agent Configuration Dashboard" \
  --status "Backlog"

./scripts/create-project-task.sh --project engineering \
  --title "Infra: Build and publish Docker images" \
  --status "Backlog"

# P1 - High Priority
./scripts/create-project-task.sh --project engineering \
  --title "Feature: POST /usage-records API endpoint" \
  --status "Backlog"

./scripts/create-project-task.sh --project engineering \
  --title "Feature: Web UI - Usage Dashboard" \
  --status "Backlog"
```

**Marketing (Beta Launch Campaign):**
```bash
./scripts/create-project-task.sh --project marketing \
  --title "Campaign: Beta Customer Outreach" \
  --status "Planning"

./scripts/create-project-task.sh --project marketing \
  --title "Content: Product positioning and messaging" \
  --status "Planning"

./scripts/create-project-task.sh --project marketing \
  --title "Content: 5-minute product demo video" \
  --status "Planning"
```

### 2. Configure Agent Workflows

Add task creation to agent prompts:

**Example for go-backend-developer:**
```
When you identify a new feature/bug during development:
1. Create GitHub issue with labels
2. Create project task: ./scripts/create-project-task.sh --project engineering --title "..." --status "Backlog"
3. Link issue to task
```

### 3. Set Up Views in GitHub UI

**Engineering Board Views:**
1. Sprint Board (Status = Ready/In Progress/In Review)
2. Roadmap Timeline (Group by Milestone)
3. Bug Triage (Type = bug, Sort by Priority)
4. By Component (Group by Area label)

**Marketing Board Views:**
1. Active Campaigns (Status = In Progress/Launched/Monitoring)
2. By Channel (Group by Channel)
3. By Funnel Stage (Group by Funnel)
4. Revenue Impact (Priority = revenue/pipeline)

---

## 📚 Documentation

- **[GITHUB_PROJECTS_SETUP.md](./GITHUB_PROJECTS_SETUP.md)** - Complete setup guide with commands
- **[GITHUB_PROJECTS_API.md](./GITHUB_PROJECTS_API.md)** - GraphQL API reference
- **[SETUP_GITHUB_TOKEN.md](./SETUP_GITHUB_TOKEN.md)** - Token creation guide
- **[GITHUB_PROJECTS_COMPLETE.md](./GITHUB_PROJECTS_COMPLETE.md)** - This file

---

## 🐛 Troubleshooting

**Error: "authentication token is missing required scopes"**
- Solution: Source token: `source .env.github`

**Error: "Project not found"**
- Solution: Verify project exists: `gh project list`

**Error: "Field not found"**
- Solution: Re-run setup: `./scripts/setup-project-automation.sh`

**Script fails with jq error:**
- Solution: Install jq: `brew install jq`

---

## ✅ Success Criteria Met

- ✅ Two project boards created (Engineering + Marketing)
- ✅ 49 labels created and categorized
- ✅ 7 milestones created with deadlines
- ✅ Automation scripts working
- ✅ Configuration file generated
- ✅ Test tasks created successfully
- ✅ Token saved securely
- ✅ Documentation complete
- ✅ Ready for agent integration

---

## 🎉 Summary

**You now have:**
- **Two fully configured GitHub Project boards** (Engineering + Marketing)
- **Complete label taxonomy** (49 labels)
- **Milestone roadmap** (7 milestones through Aug 2025)
- **Automation scripts** for programmatic task creation
- **Agent-ready workflows** (any agent can create tasks via CLI)
- **Secure token management** (gitignored, scoped correctly)

**What agents can do now:**
- ✅ Create engineering tasks (features, bugs, epics)
- ✅ Create marketing tasks (campaigns, content, events)
- ✅ Set status, priority, and custom fields
- ✅ Link tasks to milestones
- ✅ All without GitHub UI!

**Next milestone:** Populate boards with v0.3.0 tasks and start building Web UI Dashboard (P0 - Revenue blocker)

---

**Questions?** See docs/GITHUB_PROJECTS_SETUP.md for complete reference.
