# AI Agent Configurations

This directory contains project-specific AI agent configurations for autonomous development workflows.

## Agent Types

### Development Agents

#### `go-backend-developer.md`
**Purpose:** Backend API development, database queries, and Go code implementation

**Responsibilities:**
- Implement API endpoints following TDD
- Write database queries and migrations
- Fix backend bugs
- Create CLI commands
- Follow CI/CD workflow with automatic PR checks

**Key Features:**
- Consults tech-lead for architecture decisions
- Auto-waits for CI checks before marking tasks complete
- Updates GitHub Project status automatically
- Breaks down large tasks into subtasks

#### `frontend-developer.md`
**Purpose:** Next.js web UI development and frontend testing

**Responsibilities:**
- Implement admin panel pages and components
- Build responsive UI with React and Next.js
- Integrate with backend APIs
- Write E2E tests with Playwright
- Fix frontend bugs

**Key Features:**
- Coordinates with backend for API requirements
- Follows wireframe-first development
- Auto-waits for CI checks before completion
- Implements TDD for frontend code

### Management & Coordination

#### `coordinator.md`
**Purpose:** Orchestrates autonomous AI development team

**Responsibilities:**
- Monitors GitHub Projects for tasks
- Assigns work to specialized agents
- Enables agent-to-agent communication
- Ensures continuous progress on milestones

**Key Features:**
- Fully autonomous development workflow
- Cross-agent task coordination
- Milestone tracking and reporting

#### `tech-lead.md`
**Purpose:** Architecture decisions and technical leadership

**Responsibilities:**
- Major technical decisions affecting system design
- Cross-team coordination (frontend, backend, CLI)
- Product-technical alignment
- High-level project planning
- Technical governance and standards

**Key Features:**
- Reviews architectural changes
- Translates business requirements to technical specs
- Breaks down epics into work packages
- Maintains technical standards

### Product & Strategy

#### `product-strategist.md`
**Purpose:** Strategic guidance on feature prioritization and product direction

**Responsibilities:**
- Feature prioritization based on business value
- Product direction and MVP scope
- Business value assessment
- Sprint planning recommendations

**Key Features:**
- Proactive guidance after milestones
- Evaluates technical approaches by business impact
- Tracks MVP requirements
- Aligns features with business priorities

### Quality & Review

#### `pr-reviewer.md`
**Purpose:** Code review and quality assurance

**Responsibilities:**
- Review pull requests for quality
- Verify test coverage and TDD compliance
- Check architectural consistency
- Ensure documentation is updated

**Key Features:**
- Automated code quality checks
- Standards enforcement
- Security review
- Performance analysis

## Usage

These configurations are automatically loaded by Claude Code when using the Task tool with the appropriate `subagent_type`.

## Versioning

These agent configurations are:
- ✅ Versioned in Git (project-level)
- ✅ Tracked with the codebase
- ✅ Shared across team members
- ✅ Updated via pull requests

This ensures all developers use consistent agent workflows and improvements are tracked over time.

## Maintenance

When updating agent configurations:
1. Edit the `.md` file in this directory
2. Test the changes locally
3. Create a PR with the agent config updates
4. Document significant changes in commit messages

See [CLAUDE.md](../../CLAUDE.md) for complete project documentation.
