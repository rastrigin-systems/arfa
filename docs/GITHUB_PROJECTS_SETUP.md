# GitHub Projects Setup Guide

**Two-Board Strategy for Ubik Enterprise**

## Overview

We maintain **two separate GitHub Project boards** to separate technical execution from business strategy:

1. **Engineering Board** - Features, bugs, technical roadmap
2. **Business/Marketing Board** - GTM, content, customer acquisition, partnerships

---

## Prerequisites

```bash
# Ensure GitHub CLI has project permissions
gh auth refresh -s project --hostname github.com

# Verify authentication
gh auth status
```

---

## Board 1: Engineering Roadmap

### Create the Board

```bash
# Create project
gh project create \
  --owner sergei-rastrigin \
  --title "Ubik Engineering Roadmap" \
  --format board

# Expected output: Project URL and number
# Save the PROJECT_NUMBER for future commands
```

### Labels for Engineering Board

```bash
# Priority Labels (Business Value)
gh label create "priority/p0" --description "Critical - Revenue blocker / Security issue" --color "d73a4a" --force
gh label create "priority/p1" --description "High - Significant business impact" --color "ff6b6b" --force
gh label create "priority/p2" --description "Medium - Nice to have" --color "ffa500" --force
gh label create "priority/p3" --description "Low - Speculative / Future" --color "ffeb3b" --force

# Component Labels (Technical Area)
gh label create "area/api" --description "Backend API" --color "0366d6" --force
gh label create "area/cli" --description "CLI client" --color "0366d6" --force
gh label create "area/web" --description "Web dashboard" --color "0366d6" --force
gh label create "area/db" --description "Database/schema" --color "0366d6" --force
gh label create "area/docs" --description "Documentation" --color "0366d6" --force
gh label create "area/infra" --description "Infrastructure/DevOps" --color "0366d6" --force
gh label create "area/testing" --description "Testing infrastructure" --color "0366d6" --force

# Type Labels
gh label create "type/feature" --description "New feature" --color "a2eeef" --force
gh label create "type/bug" --description "Bug fix" --color "d73a4a" --force
gh label create "type/epic" --description "Large multi-issue feature" --color "3e4b9e" --force
gh label create "type/research" --description "Research/spike" --color "d4c5f9" --force
gh label create "type/refactor" --description "Code improvement" --color "fbca04" --force
gh label create "type/chore" --description "Maintenance/tooling" --color "fef2c0" --force

# Size Labels (T-shirt sizing for estimation)
gh label create "size/xs" --description "< 2 hours" --color "c5def5" --force
gh label create "size/s" --description "2-4 hours" --color "c5def5" --force
gh label create "size/m" --description "1-2 days" --color "c5def5" --force
gh label create "size/l" --description "3-5 days" --color "c5def5" --force
gh label create "size/xl" --description "> 1 week" --color "c5def5" --force

# Business Impact Labels
gh label create "impact/revenue" --description "Directly impacts revenue" --color "0e8a16" --force
gh label create "impact/retention" --description "Reduces churn" --color "0e8a16" --force
gh label create "impact/acquisition" --description "Helps win customers" --color "0e8a16" --force
gh label create "impact/efficiency" --description "Developer productivity" --color "0e8a16" --force
```

### Milestones for Engineering

```bash
# Technical milestones
gh milestone create \
  --title "v0.3.0 - Web UI MVP" \
  --due-date 2025-02-15 \
  --description "Ship web dashboard for agent configuration and usage tracking"

gh milestone create \
  --title "v0.4.0 - Beta Ready" \
  --due-date 2025-03-31 \
  --description "Complete employee management, cost tracking, approval workflows for beta launch"

gh milestone create \
  --title "v0.5.0 - Public Launch" \
  --due-date 2025-05-31 \
  --description "System prompts, MCP management, advanced analytics, production-ready"

gh milestone create \
  --title "v0.6.0 - Enterprise Features" \
  --due-date 2025-08-31 \
  --description "SSO, RBAC, audit logs, compliance features"
```

### Custom Fields for Engineering Board

**Note:** Custom fields are set via GitHub UI (Projects → Settings → Fields)

**Recommended fields:**

1. **Status** (Single select)
   - Backlog
   - Ready
   - In Progress
   - Blocked
   - In Review
   - Done

2. **Effort** (Number)
   - Story points or hours estimate

3. **Sprint** (Iteration field)
   - 2-week sprints

4. **Owner** (Assignees)

5. **Dependencies** (Text)
   - Issue numbers this depends on

---

## Board 2: Business & Marketing Roadmap

### Create the Board

```bash
# Create project
gh project create \
  --owner sergei-rastrigin \
  --title "Ubik Business & Marketing" \
  --format board

# Expected output: Project URL and number
```

### Labels for Business/Marketing Board

```bash
# Marketing Channel Labels
gh label create "channel/content" --description "Blog posts, case studies, guides" --color "8B5CF6" --force
gh label create "channel/social" --description "LinkedIn, Twitter, Reddit posts" --color "8B5CF6" --force
gh label create "channel/email" --description "Email campaigns, newsletters" --color "8B5CF6" --force
gh label create "channel/ads" --description "Paid advertising campaigns" --color "8B5CF6" --force
gh label create "channel/seo" --description "SEO optimization" --color "8B5CF6" --force
gh label create "channel/partnerships" --description "Partner integrations, co-marketing" --color "8B5CF6" --force

# Funnel Stage Labels
gh label create "funnel/awareness" --description "Top of funnel - Brand awareness" --color "10B981" --force
gh label create "funnel/consideration" --description "Mid funnel - Lead generation" --color "10B981" --force
gh label create "funnel/conversion" --description "Bottom funnel - Sales enablement" --color "10B981" --force
gh label create "funnel/retention" --description "Customer success, expansion" --color "10B981" --force

# Campaign Type Labels
gh label create "campaign/launch" --description "Product launch campaign" --color "F59E0B" --force
gh label create "campaign/feature" --description "Feature announcement" --color "F59E0B" --force
gh label create "campaign/education" --description "Educational content" --color "F59E0B" --force
gh label create "campaign/event" --description "Webinar, conference, demo" --color "F59E0B" --force

# Priority for Business (different meaning than engineering)
gh label create "biz-priority/revenue" --description "Direct revenue impact" --color "DC2626" --force
gh label create "biz-priority/brand" --description "Brand building" --color "F97316" --force
gh label create "biz-priority/pipeline" --description "Lead generation" --color "EAB308" --force
gh label create "biz-priority/research" --description "Market research, customer discovery" --color "3B82F6" --force

# Status Labels (for non-technical tasks)
gh label create "status/planning" --description "Planning/ideation phase" --color "6B7280" --force
gh label create "status/in-progress" --description "Actively working" --color "3B82F6" --force
gh label create "status/review" --description "In review" --color "8B5CF6" --force
gh label create "status/launched" --description "Live/published" --color "10B981" --force
gh label create "status/monitoring" --description "Launched, monitoring results" --color "059669" --force
```

### Milestones for Business/Marketing

```bash
# Business milestones (aligned with product milestones)
gh milestone create \
  --title "Beta Launch Campaign" \
  --due-date 2025-03-01 \
  --description "Acquire 5-10 beta customers, launch product messaging"

gh milestone create \
  --title "Public Launch Campaign" \
  --due-date 2025-05-31 \
  --description "Product Hunt launch, press coverage, first 50 customers"

gh milestone create \
  --title "Growth Campaign Q3" \
  --due-date 2025-09-30 \
  --description "Scale to 100+ customers, establish content engine"
```

### Custom Fields for Business Board

**Recommended fields (set via UI):**

1. **Status** (Single select)
   - Backlog
   - Planning
   - In Progress
   - In Review
   - Launched
   - Monitoring

2. **Channel** (Multi-select)
   - Blog
   - Social
   - Email
   - Paid
   - Partnerships

3. **Target Audience** (Single select)
   - IT Admins
   - Developers
   - CTOs
   - Security Teams
   - Finance/Procurement

4. **Expected Impact** (Single select)
   - High
   - Medium
   - Low

5. **Launch Date** (Date field)

6. **Metrics** (Text)
   - KPIs to track (e.g., "500 signups", "10 MQLs", "50% email open rate")

---

## Initial Issues - Engineering Board

Create these high-priority issues to populate the engineering board:

### P0 Issues (Critical)

```bash
# Issue 1: Web UI - Agent Configuration Dashboard
gh issue create \
  --title "Feature: Web UI - Agent Configuration Dashboard" \
  --label "priority/p0,area/web,type/feature,impact/revenue,size/xl" \
  --milestone "v0.3.0 - Web UI MVP" \
  --body "$(cat <<'EOF'
## Business Value
Enables IT admins to configure agents for their organization - core value proposition and revenue blocker for beta launch.

## Expected Impact
- Unblock first paying customer ($5-20K ARR)
- Enable compelling 3-minute demo
- 10x easier customer acquisition vs. CLI-only approach

## User Story
As an IT admin, I can log in and configure Claude Code agent for my organization, assign it to teams, and set policies - all in < 5 minutes.

## Acceptance Criteria
- [ ] Login page (uses /auth/login API)
- [ ] Agent catalog page (GET /agents)
- [ ] Org agent config page (CRUD via /organizations/current/agent-configs)
- [ ] Team assignment UI (assign config to teams)
- [ ] Responsive design (works on desktop)
- [ ] Error handling for all API calls
- [ ] Loading states for all async operations

## Technical Notes
- Framework: Next.js 14 (App Router)
- API: All endpoints exist in v0.1.0
- Auth: JWT token in localStorage/cookies
- Styling: Tailwind CSS
- Components: shadcn/ui (recommended)

## Dependencies
- None (all APIs ready!)

## Estimated Effort
Size: XL (2 weeks / 80 hours)

## Test Plan
- [ ] Unit tests for components
- [ ] Integration tests for API calls
- [ ] E2E test for complete flow (login → configure → assign)
EOF
)"

# Issue 2: Docker Images - Build and Publish
gh issue create \
  --title "Infra: Build and publish Docker images to registry" \
  --label "priority/p0,area/infra,type/chore,impact/acquisition,size/s" \
  --milestone "v0.3.0 - Web UI MVP" \
  --body "$(cat <<'EOF'
## Business Value
CLI cannot work without published Docker images. Blocking beta customers from using the product.

## Expected Impact
- Unblock CLI usage for beta customers
- Enable developers to sync and use agents immediately
- Reduce support burden (no manual image building)

## Acceptance Criteria
- [ ] Claude Code image built from Dockerfile
- [ ] MCP filesystem image built
- [ ] MCP git image built
- [ ] Images published to Docker Hub or GitHub Container Registry
- [ ] Images tagged with version (v0.2.0)
- [ ] Images tested locally (pull and run)
- [ ] Documentation updated with image URLs

## Technical Notes
- Dockerfiles exist in docker/ directory
- Use GitHub Actions for automated builds
- Multi-arch support (amd64, arm64) if possible

## Estimated Effort
Size: S (4 hours)
EOF
)"
```

### P1 Issues (High Priority)

```bash
# Issue 3: API - Usage Tracking Endpoint
gh issue create \
  --title "Feature: POST /usage-records endpoint for CLI telemetry" \
  --label "priority/p1,area/api,type/feature,impact/retention,size/m" \
  --milestone "v0.3.0 - Web UI MVP" \
  --body "$(cat <<'EOF'
## Business Value
Powers usage dashboard - visibility is killer feature for IT admins. Enables cost tracking and chargeback.

## User Story
As a developer, when I use `ubik`, the CLI automatically reports my usage to the platform so IT can track activity.

## Acceptance Criteria
- [ ] POST /usage-records endpoint implemented
- [ ] Accepts: employee_id, agent_config_id, started_at, ended_at, duration_seconds
- [ ] Validates JWT token
- [ ] Scopes to org_id via middleware
- [ ] Returns 201 Created on success
- [ ] TDD: Write tests first!
- [ ] 85%+ test coverage

## Technical Notes
- Table: usage_records (already exists in schema)
- SQL query: CreateUsageRecord (needs to be added to sqlc/queries/)

## Dependencies
- None

## Estimated Effort
Size: M (2 days)
EOF
)"

# Issue 4: Web UI - Usage Dashboard
gh issue create \
  --title "Feature: Web UI - Usage Dashboard" \
  --label "priority/p1,area/web,type/feature,impact/retention,size/l" \
  --milestone "v0.3.0 - Web UI MVP" \
  --body "$(cat <<'EOF'
## Business Value
Visibility = killer feature. IT admins need to see who's using what, when, and how much it costs.

## User Story
As an IT admin, I can see real-time usage data (which employees are using which agents) and filter by team, agent, or time period.

## Acceptance Criteria
- [ ] Usage table with: employee, agent, last used, total duration
- [ ] Filters: by team, by agent, by date range
- [ ] Export to CSV button
- [ ] Pagination (for large datasets)
- [ ] Real-time updates (optional - nice to have)

## Dependencies
- #3 (Usage tracking API endpoint)
- #1 (Web UI foundation/navigation)

## Estimated Effort
Size: L (1 week)
EOF
)"
```

---

## Initial Issues - Business/Marketing Board

Create these for the business board:

```bash
# Business Issue 1: Beta Customer Outreach Campaign
gh issue create \
  --title "Campaign: Beta Customer Outreach (Target: 5-10 signups)" \
  --label "biz-priority/pipeline,channel/email,channel/social,funnel/conversion,campaign/launch" \
  --milestone "Beta Launch Campaign" \
  --body "$(cat <<'EOF'
## Goal
Acquire 5-10 beta customers by March 1, 2025 to validate product-market fit and pricing.

## Target Audience
- Tech companies with 50-500 employees
- 10-100 developers
- Already using AI coding assistants
- Budget: $5K-20K/year

## Tactics
- [ ] Personal network outreach (20 warm intros)
- [ ] LinkedIn posts announcing beta (3 posts)
- [ ] Email to warm leads list (50 emails)
- [ ] Twitter announcement thread
- [ ] Post in r/devops, r/programming (where allowed)

## Assets Needed
- [ ] Beta landing page (web form for signups)
- [ ] Email template (personal pitch)
- [ ] 5-minute demo video
- [ ] One-pager PDF (product overview)

## Success Metrics
- 50 conversations started
- 10 beta signups
- 5 active weekly users
- 3 willing to convert to paid

## Timeline
- Week 1-2: Create assets
- Week 3-4: Outreach blitz
- Week 5-8: Nurture and onboard

## Owner
@sergei-rastrigin
EOF
)"

# Business Issue 2: Product Messaging & Positioning
gh issue create \
  --title "Content: Define product messaging and positioning" \
  --label "biz-priority/brand,channel/content,funnel/awareness,status/planning" \
  --milestone "Beta Launch Campaign" \
  --body "$(cat <<'EOF'
## Goal
Create clear, compelling product messaging that resonates with IT admins and CTOs.

## Deliverables
- [ ] Value proposition statement (1 sentence)
- [ ] Positioning statement (vs. DIY, vs. competitors)
- [ ] Key benefits (3-5 bullets)
- [ ] Feature list (grouped by persona)
- [ ] Pricing page copy
- [ ] FAQ section (10-15 questions)

## Key Messages (Draft)
- "Okta for AI agents" - Centralized access control
- "Reduce AI chaos" - Stop shadow IT
- "Compliance without friction" - Balance control and productivity

## Research Needed
- [ ] Interview 5 potential customers (IT admins)
- [ ] Analyze competitor messaging (Claude, Cursor, Windsurf)
- [ ] Study analogous products (Okta, 1Password, BrowserStack)

## Success Metrics
- Messaging tested with 10 prospects
- 70%+ understand value prop in < 30 seconds

## Timeline
2 weeks

## Owner
@sergei-rastrigin
EOF
)"

# Business Issue 3: Demo Video Production
gh issue create \
  --title "Content: Create 5-minute product demo video" \
  --label "biz-priority/pipeline,channel/content,funnel/consideration,campaign/launch" \
  --milestone "Beta Launch Campaign" \
  --body "$(cat <<'EOF'
## Goal
Create a compelling 5-minute demo video showing the full workflow: IT configures → dev uses → IT tracks usage.

## Script Outline
1. Problem (30s): AI chaos, shadow IT, security risks
2. Solution (30s): Ubik Enterprise - centralized control
3. Demo - IT Admin (2min):
   - Login to dashboard
   - Configure Claude Code for organization
   - Assign to Engineering team
   - Set policies (path restrictions, cost limits)
4. Demo - Developer (1.5min):
   - Run `ubik sync`
   - Start interactive session with `ubik`
   - Claude Code uses MCP servers (filesystem, git)
   - Exit session
5. Demo - IT Admin (30s):
   - View usage dashboard
   - See developer activity
   - Export cost report
6. CTA (30s): Sign up for beta

## Production
- [ ] Write script
- [ ] Record voiceover
- [ ] Record screen captures (1080p)
- [ ] Edit in Final Cut / Premiere
- [ ] Add captions
- [ ] Export and upload to YouTube/Vimeo

## Success Metrics
- 100+ views in first week
- 5+ demo requests from video

## Timeline
1 week

## Owner
@sergei-rastrigin
EOF
)"
```

---

## Workflow & Best Practices

### Engineering Board Workflow

```
Backlog → Ready → In Progress → In Review → Done
   ↓         ↓          ↓            ↓         ↓
 Triage   Assigned   Working     PR Open   Merged
```

**Weekly rituals:**
- Monday: Sprint planning (move issues to Ready)
- Wednesday: Mid-sprint check-in (unblock issues)
- Friday: Review/demo completed work

**Issue hygiene:**
- All issues must have: Priority, Area, Type, Size, Milestone
- P0 issues reviewed daily
- Blocked issues escalated immediately
- Closed issues require: PR link, release notes

---

### Business Board Workflow

```
Backlog → Planning → In Progress → In Review → Launched → Monitoring
   ↓         ↓           ↓            ↓          ↓           ↓
 Ideas   Research    Execution    Review    Published   Measure KPIs
```

**Weekly rituals:**
- Monday: Review metrics from previous week
- Wednesday: Campaign check-ins
- Friday: Plan next week's content/outreach

**Issue hygiene:**
- All campaigns must have: Target audience, metrics, timeline
- Launched campaigns tracked for 30 days (monitoring phase)
- Close issues with results summary (actual vs. expected metrics)

---

## Views Configuration

### Engineering Board Views

**View 1: Sprint Board** (Default)
- Filter: Status = "Ready" OR "In Progress" OR "In Review"
- Group by: Status
- Sort: Priority → Size

**View 2: Roadmap Timeline**
- Filter: Type = "feature" OR "epic"
- Group by: Milestone
- Sort: Priority

**View 3: Bug Triage**
- Filter: Type = "bug"
- Group by: Priority
- Sort: Created date (newest first)

**View 4: By Component**
- Filter: None (show all)
- Group by: Area label
- Sort: Priority

---

### Business Board Views

**View 1: Active Campaigns** (Default)
- Filter: Status = "In Progress" OR "Launched" OR "Monitoring"
- Group by: Status
- Sort: Launch date

**View 2: By Channel**
- Filter: None
- Group by: Channel label
- Sort: Priority

**View 3: By Funnel Stage**
- Filter: None
- Group by: Funnel label
- Sort: Priority

**View 4: Revenue Impact**
- Filter: Priority = "biz-priority/revenue" OR "biz-priority/pipeline"
- Group by: Status
- Sort: Launch date

---

## Integration Between Boards

**Cross-board dependencies:**

When a marketing campaign depends on a product feature:
```bash
# In marketing issue, reference engineering issue:
"Depends on: sergei-rastrigin/ubik-enterprise#1 (Web UI Dashboard)"

# In engineering issue, add label:
gh issue edit 1 --add-label "impact/marketing"
```

**Linking strategy:**
- Engineering issues use `impact/` labels to show business value
- Business issues reference engineering issues in description
- Weekly sync between engineering and business roadmaps

---

## Next Steps

1. **Run label creation commands** (copy/paste from sections above)
2. **Create milestones** (engineering + business)
3. **Create projects via UI** (GitHub Projects v2 requires UI for initial setup)
   - Go to: https://github.com/users/sergei-rastrigin/projects
   - Click "New project"
   - Choose "Board" layout
   - Name: "Ubik Engineering Roadmap"
   - Repeat for "Ubik Business & Marketing"
4. **Create initial issues** (run issue create commands above)
5. **Configure custom fields** (via Projects → Settings → Fields in UI)
6. **Set up views** (via Projects → Views in UI)

---

## GitHub CLI Reference

```bash
# List all projects
gh project list --owner sergei-rastrigin

# Add issue to project (requires project number from UI)
gh project item-add PROJECT_NUMBER --owner sergei-rastrigin --url https://github.com/sergei-rastrigin/ubik-enterprise/issues/1

# Create issue with all metadata
gh issue create \
  --title "Title" \
  --label "priority/p0,area/web" \
  --milestone "v0.3.0 - Web UI MVP" \
  --assignee "@me" \
  --body "Description"

# Update issue
gh issue edit 1 --add-label "impact/revenue" --milestone "v0.3.0 - Web UI MVP"

# Close issue
gh issue close 1 --comment "Completed in v0.3.0"

# View issues by label
gh issue list --label "priority/p0" --state open

# View issues by milestone
gh issue list --milestone "v0.3.0 - Web UI MVP" --state open
```

---

## Summary

**Two-Board Strategy:**
- **Engineering Board**: Focus on technical execution, sprints, code quality
- **Business Board**: Focus on GTM, customer acquisition, brand building

**Why separate boards?**
- Different workflows (Agile sprints vs. campaign cycles)
- Different stakeholders (developers vs. marketing/sales)
- Different metrics (test coverage vs. conversion rates)
- Prevents noise (engineers don't need to see social media posts)

**Integration points:**
- Weekly sync between boards
- Cross-reference dependencies
- Shared milestones (Beta Launch, Public Launch)

**Next action:** Run the label/milestone commands, then create the projects via GitHub UI.
