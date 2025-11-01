# GitHub Workflow Guide - How Everything Works Together

**Your Setup:** 2 Project Boards + Milestones + Labels + Issues

---

## 🎯 The Big Picture

```
STRATEGY LEVEL (What & When)
├─ Milestones = Time-based releases/campaigns (v0.3.0, Beta Launch, etc.)
│
EXECUTION LEVEL (How & Who)
├─ Project Boards = Current work visualization (Kanban)
│  ├─ Engineering Roadmap = Technical tasks
│  └─ Business & Marketing = GTM tasks
│
DETAIL LEVEL (Track & Discuss)
└─ Issues/Tasks = Individual work items with discussions
   └─ Labels = Categorize and filter (priority, area, type)
```

---

## 📊 Component Breakdown

### 1. Milestones = "What are we shipping and when?"

**Purpose:** Time-boxed deliverables (like Epics)

**Your Milestones:**
- **v0.3.0 - Web UI MVP** (Feb 15, 2025) - Engineering
- **v0.4.0 - Beta Ready** (Mar 31, 2025) - Engineering
- **Beta Launch Campaign** (Mar 1, 2025) - Marketing
- **Product Hunt Launch** (May 31, 2025) - Marketing

**What goes in a milestone:**
- All issues/tasks that contribute to that release/campaign
- Cross-functional (engineering + marketing)
- Has clear deadline and success criteria

**Example:** v0.3.0 milestone contains:
- [Engineering] Build Web UI dashboard
- [Engineering] Usage tracking API
- [Engineering] Docker images published
- [Marketing] Demo video
- [Marketing] Beta landing page
- [Marketing] Pricing page

### 2. Project Boards = "What are we working on now?"

**Purpose:** Visualize current work progress (Kanban style)

**Your Boards:**
- **Engineering Roadmap** - Technical work (features, bugs, infra)
- **Business & Marketing** - GTM work (campaigns, content, sales)

**Status workflow:**
```
Backlog → Ready → In Progress → (Launched/Done) → Monitoring
```

**What shows on boards:**
- ALL tasks/issues (from any milestone)
- Grouped by current status
- Filtered by labels, assignee, milestone, etc.

### 3. Issues/Tasks = "What specifically needs to be done?"

**Purpose:** Individual actionable work items

**Types:**
- **GitHub Issues** - Connected to code (features, bugs, PRs)
- **Draft Tasks** - Standalone tasks (marketing, operations)

**What goes in an issue:**
- Clear title and description
- Acceptance criteria
- Labels (priority, area, type, size)
- Milestone assignment
- Comments/discussion
- Linked PRs

### 4. Labels = "How do we categorize work?"

**Purpose:** Organize, filter, and prioritize

**Your Label System:**
- **Priority:** p0, p1, p2, p3 (business urgency)
- **Area:** api, cli, web, db, infra (technical component)
- **Type:** feature, bug, epic, research, chore
- **Size:** xs, s, m, l, xl (effort estimate)
- **Impact:** revenue, retention, acquisition (business value)
- **Channel:** content, social, email (marketing channel)

---

## 🔄 Complete Workflow

### Scenario 1: Planning a New Feature

**1. Strategic Decision (Monthly/Quarterly)**
```
Product Strategist: "We need Web UI for beta launch"
↓
Create Milestone: v0.3.0 - Web UI MVP (Due: Feb 15)
↓
Define success criteria:
- IT admin can configure agent in dashboard
- Usage tracking visible
- Beta customers can self-onboard
```

**2. Break Down into Tasks**
```
Engineering tasks:
- Issue #1: Build agent config UI (priority/p0, area/web, size/xl)
- Issue #2: Add usage tracking API (priority/p1, area/api, size/m)
- Issue #3: Publish Docker images (priority/p0, area/infra, size/s)

Marketing tasks:
- Task #1: Create demo video (channel/social, impact/high)
- Task #2: Beta landing page (channel/social, impact/high)
- Task #3: Pricing page (channel/social, impact/high)

All assigned to milestone: v0.3.0
```

**3. Add to Project Boards**
```
Engineering Board:
├─ Backlog: #2 (usage API), #3 (Docker images)
└─ Ready: #1 (config UI) - all dependencies resolved

Marketing Board:
├─ Backlog: #3 (pricing page)
└─ Ready: #1 (demo video), #2 (landing page)
```

**4. Start Working (Weekly)**
```
Developer: Move #1 to "In Progress" on Engineering Board
↓
Create branch: feature/web-ui-agent-config
↓
Write tests (TDD)
↓
Implement feature
↓
Create PR, link to #1
↓
Code review
↓
Merge PR → #1 automatically closes
↓
Move #1 to "Done" on board
```

**5. Ship & Monitor**
```
When all v0.3.0 issues complete:
↓
Create release tag: v0.3.0
↓
Deploy to production
↓
Close milestone
↓
Review: Did we meet success criteria?
↓
Retrospective: What went well? What to improve?
```

---

## 🎭 Workflow by Role

### As a Developer

**Daily:**
1. Check **Engineering Board** → "In Progress" column (your tasks)
2. Update task status as you work
3. Comment on issues with progress/blockers
4. Link PRs to issues

**Weekly:**
1. Review **Ready** column → Pick next task
2. Move to "In Progress" when starting
3. Participate in sprint planning
4. Update estimates (if task is bigger/smaller than expected)

**Monthly:**
1. Review upcoming milestone
2. Provide effort estimates for new tasks
3. Suggest technical improvements

### As Product Owner/Strategist

**Daily:**
1. Review both boards for blockers
2. Prioritize new requests (add labels, set milestones)
3. Answer questions on issues

**Weekly:**
1. Sprint planning: Move tasks from Backlog → Ready
2. Review milestone progress
3. Adjust priorities based on feedback

**Monthly:**
1. Define next milestone goals
2. Break down epics into tasks
3. Balance engineering vs marketing work
4. Review completed milestones

### As Marketer

**Daily:**
1. Check **Marketing Board** → "In Progress" and "Monitoring"
2. Update campaign metrics in task comments
3. Move tasks through workflow (Ready → In Progress → Launched → Monitoring)

**Weekly:**
1. Review "Launched" tasks → Move to Monitoring
2. Review "Monitoring" tasks → Check metrics, optimize
3. Pick new tasks from "Ready" column

**Monthly:**
1. Report on milestone progress
2. Plan next campaign milestone
3. Review what worked/didn't work

---

## 🔧 Practical Examples

### Example 1: New Feature Request

```
User requests: "Can we add SSO for enterprise customers?"

Step 1: Triage
→ Create issue: "Feature: SSO integration (Google, Okta)"
→ Add labels: priority/p2, area/api, type/feature, size/xl
→ Assign milestone: v0.6.0 - Enterprise Features
→ Add to Engineering Board → Backlog

Step 2: Discussion
→ Tech lead comments with design approach
→ Product strategist comments with business value
→ Estimate effort: 2-3 weeks
→ Identify dependencies: Need user management refactor first

Step 3: Prioritization (Monthly planning)
→ Product strategist: "High value for enterprise, but v0.6.0"
→ Keep in Backlog until v0.5.0 complete
→ Update priority if enterprise deal pending

Step 4: Execution (When ready)
→ Move to Ready column (all deps resolved)
→ Developer assigned, moves to In Progress
→ Work tracked via PRs and comments
→ Completed, moved to Done
→ Shipped in v0.6.0 release
```

### Example 2: Marketing Campaign

```
Campaign: "Product Hunt Launch"

Step 1: Planning
→ Create task: "Product Hunt launch strategy"
→ Add to milestone: Public Launch Campaign (May 31)
→ Add to Marketing Board → Backlog
→ Set fields: channel/social, impact/high, audience/developers

Step 2: Preparation (4 weeks before)
→ Move to Ready (all assets prepared)
→ Break down into subtasks:
  - Write PH description
  - Create demo video
  - Recruit supporters
  - Schedule launch day
→ Assign each subtask to milestone

Step 3: Execution (Launch day)
→ Move to In Progress (launch day morning)
→ Comment updates throughout day:
  - "Submitted at 12:01 AM PT"
  - "50 upvotes after 2 hours"
  - "Top 5 at noon!"
→ Track metrics in comments

Step 4: Launched (After launch day)
→ Move to Launched
→ Monitor for 7 days
→ Comment with results: "Top 3 Product, 500 upvotes, 150 signups"

Step 5: Monitoring (7-30 days)
→ Move to Monitoring
→ Track conversions: signups → trials → paid
→ Write retrospective
→ Archive or keep in Monitoring if ongoing
```

### Example 3: Bug Fix

```
Bug discovered: "CLI sync fails with proxy environments"

Step 1: Report
→ Create issue: "Bug: CLI sync fails behind corporate proxy"
→ Add labels: priority/p1, area/cli, type/bug, size/m
→ No milestone (bugs get fixed ASAP)
→ Add to Engineering Board → Ready (urgent)

Step 2: Investigation
→ Developer assigned, moves to In Progress
→ Comments with findings: "Docker network issue"
→ Links to similar issues
→ Provides workaround in comment

Step 3: Fix
→ Creates PR with fix + test
→ Links PR to issue
→ Code review
→ Merged → Issue auto-closes

Step 4: Communication
→ Comment on issue: "Fixed in v0.2.1"
→ Update changelog
→ Notify affected users
```

---

## 📋 Board Views to Create

### Engineering Board Views

**1. Sprint Board** (Default)
- Filter: Status = Ready OR In Progress OR In Review
- Group by: Status
- Sort: Priority
- **Use for:** Daily standup, seeing current work

**2. This Milestone**
- Filter: Milestone = "v0.3.0 - Web UI MVP"
- Group by: Status
- Sort: Priority
- **Use for:** Tracking milestone progress

**3. My Tasks**
- Filter: Assignee = @me
- Group by: Status
- Sort: Priority
- **Use for:** Personal work queue

**4. Bugs**
- Filter: Type = bug
- Group by: Priority
- Sort: Created date (newest first)
- **Use for:** Bug triage

**5. Roadmap**
- Filter: Type = feature OR epic
- Group by: Milestone
- Sort: Priority
- **Use for:** Long-term planning

### Marketing Board Views

**1. This Week** (Default)
- Filter: Status = Ready OR In Progress
- Group by: Status
- Sort: Expected Impact
- **Use for:** Daily work planning

**2. Active Campaigns**
- Filter: Status = Launched OR Monitoring
- Group by: Channel
- Sort: Launch date
- **Use for:** Campaign performance tracking

**3. High Impact**
- Filter: Expected Impact = High
- Group by: Status
- Sort: Priority
- **Use for:** Focusing on high-value work

**4. By Milestone**
- Filter: None
- Group by: Milestone
- Sort: Priority
- **Use for:** Campaign planning

---

## 🎯 Best Practices

### 1. Keep Milestones Focused

**Good milestone:**
- Clear deliverable: "v0.3.0 - Web UI MVP"
- Time-boxed: Due Feb 15, 2025
- 5-15 tasks
- Can be shipped independently

**Bad milestone:**
- Vague: "Improve product"
- Open-ended: No deadline
- Too many tasks (50+)
- Dependent on other milestones

### 2. Use Labels Consistently

**Every issue should have:**
- 1 priority label (p0, p1, p2, p3)
- 1+ area label (api, cli, web, etc.)
- 1 type label (feature, bug, epic, etc.)
- 1 size label (xs, s, m, l, xl)

**Marketing tasks should have:**
- 1 channel label (if applicable)
- 1 impact label (high, medium)
- 1 audience label (who this is for)

### 3. Update Status Regularly

**Engineering:**
- Move to In Progress when you start working
- Update at least daily (comment with progress)
- Move to Done when merged (or closed)

**Marketing:**
- Move to In Progress when actively working
- Move to Launched when published
- Move to Monitoring immediately after launch
- Add metrics in comments weekly

### 4. Keep In Progress Minimal

**Rule:** Max 3 tasks in progress per person
**Why:** Context switching kills productivity
**How:** Finish what you start before starting new tasks

### 5. Close Milestones Properly

**When milestone deadline arrives:**
1. Review all tasks in milestone
2. Complete what you can
3. Move incomplete tasks to next milestone OR backlog
4. Write milestone retrospective
5. Close milestone
6. Create release notes

---

## 🚦 Decision Flowchart

### "Where should I put this work?"

```
Is this a new piece of work?
├─ Yes → Create Issue or Draft Task
│   ├─ Is it code-related? → GitHub Issue (shows in Engineering Board)
│   └─ Is it marketing/ops? → Draft Task (shows in Marketing Board)
└─ No → Update existing issue/task

Does it have a deadline?
├─ Yes → Assign to Milestone
│   ├─ Engineering release? → v0.X.0 milestone
│   └─ Marketing campaign? → Campaign milestone
└─ No → Leave milestone empty, add to Backlog

Is it urgent?
├─ Yes → priority/p0 or p1, move to Ready column
└─ No → priority/p2 or p3, keep in Backlog

Is it ready to work on?
├─ Yes → Move to Ready column on Project Board
└─ No → Keep in Backlog until dependencies resolved

Are you working on it RIGHT NOW?
├─ Yes → Move to In Progress
└─ No → Leave in Ready
```

---

## 🎓 Training Scenarios

### Scenario: "I just joined the team, what do I do?"

1. **Understand the strategy**
   - Read: CLAUDE.md (project overview)
   - Read: IMPLEMENTATION_ROADMAP.md (what we're building)
   - Read: MARKETING.md (go-to-market strategy)

2. **Find your work**
   - Go to Projects: https://github.com/users/sergei-rastrigin/projects
   - Select your board (Engineering or Marketing)
   - Look at "Ready" column → Pick a task
   - Check labels: Start with priority/p1 or p2, size/s or m

3. **Start working**
   - Assign task to yourself
   - Move to "In Progress"
   - Read task description and acceptance criteria
   - Comment if you have questions
   - Do the work
   - Update status when done

### Scenario: "I have an idea for a new feature"

1. **Create issue**
   - Title: Clear, concise (e.g., "Feature: Export usage data to CSV")
   - Description: Problem, solution, why it matters
   - Labels: priority/p2, area/api, type/feature, size/m (estimate)

2. **Get feedback**
   - Tag relevant people in comments
   - Product strategist prioritizes (changes priority label)
   - Tech lead provides design feedback
   - Refined and ready to build

3. **Add to roadmap**
   - Assigned to milestone (if time-sensitive)
   - Added to Engineering Board → Backlog
   - Moves to Ready when prioritized

### Scenario: "Customer reported a bug"

1. **Create bug issue**
   - Title: "Bug: CLI sync fails with proxy"
   - Description: Steps to reproduce, expected vs actual
   - Labels: priority/p1, area/cli, type/bug, size/m

2. **Triage**
   - If critical (data loss, security) → priority/p0
   - Add to Engineering Board → Ready (skip Backlog)
   - Assign to developer immediately

3. **Fix**
   - Developer moves to In Progress
   - Fixes bug + adds test
   - Creates PR, links to issue
   - Merges → Issue auto-closes
   - Comment on issue when fix is deployed

### Scenario: "Planning a marketing campaign"

1. **Create campaign task**
   - Title: "Campaign: Product Hunt Launch"
   - Description: Goals, tactics, timeline, metrics
   - Fields: channel/social, impact/high, audience/developers
   - Milestone: Public Launch Campaign

2. **Break down subtasks**
   - Subtask 1: Write PH description
   - Subtask 2: Create demo video
   - Subtask 3: Recruit supporters
   - Each subtask is a separate task, linked to parent

3. **Execute**
   - All subtasks move through workflow
   - Parent task stays in Planning until all subtasks done
   - Launch day: Move parent to In Progress
   - After launch: Move to Launched → Monitoring
   - Track metrics in comments

---

## 📝 Quick Reference Commands

```bash
# Create issue
gh issue create --title "Title" --label "priority/p1,area/web" --milestone "v0.3.0"

# Create project task (draft)
./scripts/create-project-task.sh --project marketing --title "Task" --status "Ready"

# Update task fields
./scripts/batch-update-marketing-fields.sh

# List issues by milestone
gh issue list --milestone "v0.3.0 - Web UI MVP"

# List issues by label
gh issue list --label "priority/p0" --state open

# View project board
open https://github.com/users/sergei-rastrigin/projects
```

---

## 🎯 Summary: The Complete Loop

```
PLAN (Monthly)
↓
Define Milestones (What & When)
  → v0.3.0 - Web UI MVP (Feb 15)
  → Beta Launch Campaign (Mar 1)
↓
Create Issues/Tasks (How)
  → Engineering: Build UI, Add API, Publish images
  → Marketing: Demo video, Landing page, Outreach
↓
Add to Project Boards (Visualize)
  → Engineering Board: Ready → In Progress → Done
  → Marketing Board: Ready → In Progress → Launched → Monitoring
↓
EXECUTE (Daily)
↓
Work on tasks, update status, comment progress
↓
SHIP (When milestone complete)
↓
Deploy, launch campaign, close milestone
↓
REVIEW (Retrospective)
↓
What worked? What didn't? What's next?
↓
PLAN (Next milestone)
```

---

**Your Question:** "How should we work with milestones, boards, and issues?"

**Answer:**
- **Milestones** = Time-based releases/campaigns (WHEN)
- **Boards** = Current work visualization (NOW)
- **Issues** = Individual tasks with details (WHAT)
- **Labels** = Categorization and filtering (HOW)

They all work together to give you strategy → execution → tracking in one system!
