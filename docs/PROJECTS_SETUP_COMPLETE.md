# GitHub Projects Setup - Complete! ✅

**Date:** 2025-11-01
**Status:** Labels and Milestones Created

---

## What's Been Set Up

### ✅ Labels Created (50+ labels)

#### Engineering Labels

**Priority Labels:**
- `priority/p0` - Critical (Revenue blocker / Security)
- `priority/p1` - High (Significant business impact)
- `priority/p2` - Medium (Nice to have)
- `priority/p3` - Low (Speculative / Future)

**Component Labels:**
- `area/api` - Backend API
- `area/cli` - CLI client
- `area/web` - Web dashboard
- `area/db` - Database/schema
- `area/docs` - Documentation
- `area/infra` - Infrastructure/DevOps
- `area/testing` - Testing infrastructure

**Type Labels:**
- `type/feature` - New feature
- `type/bug` - Bug fix
- `type/epic` - Large multi-issue feature
- `type/research` - Research/spike
- `type/refactor` - Code improvement
- `type/chore` - Maintenance/tooling

**Size Labels (Estimation):**
- `size/xs` - < 2 hours
- `size/s` - 2-4 hours
- `size/m` - 1-2 days
- `size/l` - 3-5 days
- `size/xl` - > 1 week

**Business Impact Labels:**
- `impact/revenue` - Directly impacts revenue
- `impact/retention` - Reduces churn
- `impact/acquisition` - Helps win customers
- `impact/efficiency` - Developer productivity

#### Business/Marketing Labels

**Channel Labels:**
- `channel/content` - Blog posts, case studies, guides
- `channel/social` - LinkedIn, Twitter, Reddit posts
- `channel/email` - Email campaigns, newsletters
- `channel/ads` - Paid advertising
- `channel/seo` - SEO optimization
- `channel/partnerships` - Partner integrations

**Funnel Stage Labels:**
- `funnel/awareness` - Top of funnel
- `funnel/consideration` - Mid funnel (lead gen)
- `funnel/conversion` - Bottom funnel (sales)
- `funnel/retention` - Customer success

**Campaign Type Labels:**
- `campaign/launch` - Product launch campaign
- `campaign/feature` - Feature announcement
- `campaign/education` - Educational content
- `campaign/event` - Webinar, conference, demo

**Business Priority Labels:**
- `biz-priority/revenue` - Direct revenue impact
- `biz-priority/brand` - Brand building
- `biz-priority/pipeline` - Lead generation
- `biz-priority/research` - Market research

**Status Labels (for non-technical tasks):**
- `status/planning` - Planning/ideation phase
- `status/in-progress` - Actively working
- `status/review` - In review
- `status/launched` - Live/published
- `status/monitoring` - Monitoring results

---

### ✅ Milestones Created

#### Engineering Milestones

1. **v0.3.0 - Web UI MVP** (Due: Feb 15, 2025)
   - Ship web dashboard for agent configuration and usage tracking

2. **v0.4.0 - Beta Ready** (Due: Mar 31, 2025)
   - Complete employee management, cost tracking, approval workflows for beta launch

3. **v0.5.0 - Public Launch** (Due: May 31, 2025)
   - System prompts, MCP management, advanced analytics, production-ready

4. **v0.6.0 - Enterprise Features** (Due: Aug 31, 2025)
   - SSO, RBAC, audit logs, compliance features

#### Business/Marketing Milestones

5. **Beta Launch Campaign** (Due: Mar 1, 2025)
   - Acquire 5-10 beta customers, launch product messaging

6. **Public Launch Campaign** (Due: May 31, 2025)
   - Product Hunt launch, press coverage, first 50 customers

7. **Growth Campaign Q3** (Due: Sep 30, 2025)
   - Scale to 100+ customers, establish content engine

---

## Next Steps (Manual)

### 1. Create GitHub Project Boards

**You must create the boards manually via GitHub UI** (Projects v2 doesn't support CLI creation yet):

1. Go to: https://github.com/users/sergei-rastrigin/projects
2. Click **"New project"**
3. Choose **"Board"** layout
4. Name: **"Ubik Engineering Roadmap"**
5. Click **"Create project"**

Repeat for:
- Name: **"Ubik Business & Marketing"**

### 2. Configure Custom Fields (via UI)

For **Engineering Board**, add these custom fields (Projects → Settings → Fields):

- **Status** (Single select): Backlog, Ready, In Progress, Blocked, In Review, Done
- **Effort** (Number): Story points or hours estimate
- **Sprint** (Iteration): 2-week sprints
- **Dependencies** (Text): Issue numbers this depends on

For **Business Board**, add these custom fields:

- **Status** (Single select): Backlog, Planning, In Progress, In Review, Launched, Monitoring
- **Channel** (Multi-select): Blog, Social, Email, Paid, Partnerships
- **Target Audience** (Single select): IT Admins, Developers, CTOs, Security Teams, Finance
- **Expected Impact** (Single select): High, Medium, Low
- **Launch Date** (Date)
- **Metrics** (Text): KPIs to track

### 3. Create Views (via UI)

See [GITHUB_PROJECTS_SETUP.md](./GITHUB_PROJECTS_SETUP.md) for recommended view configurations.

### 4. Create Initial Issues

Run the issue creation commands from [GITHUB_PROJECTS_SETUP.md](./GITHUB_PROJECTS_SETUP.md).

---

## Quick Commands Reference

```bash
# View all labels
gh label list

# View all milestones
gh api repos/sergei-rastrigin/ubik-enterprise/milestones | jq '.[] | {number, title, due_on}'

# Create issue with labels and milestone
gh issue create \
  --title "Feature: Web UI - Agent Configuration Dashboard" \
  --label "priority/p0,area/web,type/feature,impact/revenue,size/xl" \
  --milestone "v0.3.0 - Web UI MVP" \
  --assignee "@me" \
  --body "Description here"

# List issues by priority
gh issue list --label "priority/p0" --state open

# List issues by milestone
gh issue list --milestone "v0.3.0 - Web UI MVP" --state open

# Add issue to project (after getting project number from UI)
gh project item-add PROJECT_NUMBER \
  --owner sergei-rastrigin \
  --url https://github.com/sergei-rastrigin/ubik-enterprise/issues/1
```

---

## Documentation Files

- **[GITHUB_PROJECTS_SETUP.md](./GITHUB_PROJECTS_SETUP.md)** - Complete setup guide with all commands and workflows
- **[PROJECTS_SETUP_COMPLETE.md](./PROJECTS_SETUP_COMPLETE.md)** - This file (summary of what's been done)

---

## Summary

**What's Ready:**
- ✅ 50+ labels created (engineering + marketing)
- ✅ 7 milestones created (4 engineering + 3 marketing)
- ✅ Complete documentation

**What's Manual (GitHub UI Required):**
- ⏳ Create 2 project boards (Engineering + Marketing)
- ⏳ Configure custom fields
- ⏳ Set up views
- ⏳ Create initial issues

**Estimated Time to Complete Manual Steps:** 30 minutes

**Next Action:** Go to https://github.com/users/sergei-rastrigin/projects and create the boards!

---

**Two-Board Strategy Recap:**

1. **Engineering Board** = Technical execution (features, bugs, sprints)
2. **Marketing Board** = Business growth (campaigns, content, GTM)

**Why separate?**
- Different workflows (Agile sprints vs. campaign cycles)
- Different stakeholders (developers vs. marketing/sales)
- Different metrics (test coverage vs. conversion rates)
- Prevents noise (devs don't need to see social media posts)

---

**Questions?** See [GITHUB_PROJECTS_SETUP.md](./GITHUB_PROJECTS_SETUP.md) for complete guide.
