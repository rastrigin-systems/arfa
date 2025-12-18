# Monorepo Migration Checklist

**Quick reference for executing the monorepo refactoring**

---

## Pre-Migration Validation

- [ ] All tests passing on main branch
- [ ] Clean working directory (`git status`)
- [ ] No pending PRs that would conflict
- [ ] Team notified of upcoming changes
- [ ] Backup branch created (`git branch backup/pre-migration`)

---

## Phase 1: Preparation (1-2 days)

### Create New Directories

```bash
# Platform infrastructure
mkdir -p platform/database
mkdir -p platform/api-spec
mkdir -p platform/docker-images

# Service build/docs
mkdir -p services/api/{build,docs,scripts}
mkdir -p services/cli/{build,docs,scripts}
mkdir -p services/web/{build,docs}

# Documentation organization
mkdir -p docs/architecture
mkdir -p docs/guides
mkdir -p docs/product
mkdir -p docs/operations

# Shared internal code
mkdir -p internal/testutil
```

### Move Files with Git History

```bash
# Database
git mv shared/schema/schema.sql platform/database/
git mv shared/schema/migrations platform/database/
git mv shared/schema/seeds platform/database/
git mv sqlc platform/database/

# API spec
git mv shared/openapi platform/api-spec

# Docker images
git mv shared/docker platform/docker-images

# Documentation
git mv docs/user-stories docs/product/
git mv docs/wireframes docs/product/
```

### Update Configuration Files

- [ ] `platform/database/sqlc/sqlc.yaml` - Update schema path to `../schema.sql`
- [ ] `Makefile` - Update all references from `shared/` to `platform/`
- [ ] `services/api/go.mod` - Verify replace directives
- [ ] `services/cli/go.mod` - Verify replace directives

### Update Imports in Code

```bash
# Find and replace shared/ references
find . -name "*.go" -type f -exec sed -i '' 's|shared/schema/|platform/database/|g' {} \;
find . -name "*.go" -type f -exec sed -i '' 's|shared/openapi/|platform/api-spec/|g' {} \;
```

### Validation

- [ ] `make generate` - succeeds
- [ ] `make test` - all tests pass
- [ ] `make build` - all services build
- [ ] Commit changes: `git commit -m "refactor: Move shared resources to platform/"`

---

## Phase 2: API Service Consolidation (2-3 days)

### Create Service-Specific Files

- [ ] Create `services/api/README.md` with service overview
- [ ] Create `services/api/Makefile` with build commands
- [ ] Create `services/api/docs/API.md` with design decisions
- [ ] Create `services/api/docs/DEPLOYMENT.md` with deployment guide

### Move Deployment Configs

```bash
# Deployment artifacts
mkdir -p services/api/build
git mv services/api/Dockerfile.gcp services/api/build/
```

- [ ] Create `services/api/build/Dockerfile` for local development
- [ ] Create `services/api/build/cloudbuild.yaml` for service-specific CI/CD

### Create Service Makefile

Create `services/api/Makefile`:

```makefile
.PHONY: build test test-unit test-integration clean

build:
	@echo "Building API server..."
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/server cmd/server/main.go

test:
	@echo "Running all API tests..."
	go test -v -race ./...

test-unit:
	@echo "Running unit tests..."
	go test -v -short -race ./internal/...

test-integration:
	@echo "Running integration tests..."
	go test -v ./tests/integration/...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
```

### Move Service-Specific Scripts

- [ ] Identify API-specific scripts in `scripts/`
- [ ] Move to `services/api/scripts/`
- [ ] Update script paths in documentation

### Remove Root Duplicates

**ONLY after confirming they're duplicates:**

```bash
# Compare first!
diff -r cmd/server/ services/api/cmd/server/
diff -r internal/ services/api/internal/
diff -r tests/ services/api/tests/

# If identical, remove root versions
git rm -r cmd/server/
git rm -r internal/auth/    # (if duplicated)
git rm -r tests/integration/ # (if duplicated)
```

### Update Root Makefile

Update root `Makefile` to delegate to service:

```makefile
build-api:
	@cd services/api && $(MAKE) build

test-api:
	@cd services/api && $(MAKE) test

test-api-integration:
	@cd services/api && $(MAKE) test-integration
```

### Validation

- [ ] `cd services/api && make build` - succeeds
- [ ] `cd services/api && make test` - all tests pass
- [ ] `docker build -f services/api/build/Dockerfile.gcp .` - succeeds
- [ ] Commit changes: `git commit -m "refactor: Consolidate API service"`

---

## Phase 3: CLI Service Consolidation (1-2 days)

### Create Service-Specific Files

- [ ] Create `services/cli/README.md`
- [ ] Create `services/cli/Makefile`
- [ ] Create `services/cli/docs/CLI_ARCHITECTURE.md`
- [ ] Create `services/cli/docs/CLI_USAGE.md`

### Create Service Makefile

Create `services/cli/Makefile`:

```makefile
.PHONY: build test install uninstall clean

build:
	@echo "Building CLI client..."
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/ubik cmd/ubik/main.go

test:
	@echo "Running CLI tests..."
	go test -v -race ./...

install: build
	@echo "Installing ubik CLI..."
	@if [ -w /usr/local/bin ]; then \
		cp bin/ubik /usr/local/bin/ubik; \
	else \
		sudo cp bin/ubik /usr/local/bin/ubik; \
	fi

uninstall:
	@echo "Uninstalling ubik CLI..."
	@if [ -w /usr/local/bin ]; then \
		rm -f /usr/local/bin/ubik; \
	else \
		sudo rm -f /usr/local/bin/ubik; \
	fi

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
```

### Move CLI Scripts

- [ ] Move CLI-specific scripts to `services/cli/scripts/`

### Update Root Makefile

```makefile
build-cli:
	@cd services/cli && $(MAKE) build

test-cli:
	@cd services/cli && $(MAKE) test

install-cli:
	@cd services/cli && $(MAKE) install
```

### Validation

- [ ] `cd services/cli && make build` - succeeds
- [ ] `cd services/cli && make test` - all tests pass
- [ ] `cd services/cli && make install` - installs CLI
- [ ] Commit changes: `git commit -m "refactor: Consolidate CLI service"`

---

## Phase 4: Web Service Consolidation (1 day)

### Create Service-Specific Files

- [ ] Create `services/web/README.md`
- [ ] Create `services/web/docs/WEB_ARCHITECTURE.md`

### Move Deployment Config

```bash
mkdir -p services/web/build
git mv services/web/build/Dockerfile services/web/build/  # Already moved
```

- [ ] Create `services/web/build/cloudbuild.yaml`

### Create Service Makefile (Optional)

Create `services/web/Makefile` (or rely on npm scripts):

```makefile
.PHONY: dev build test test-e2e clean

dev:
	@echo "Starting Next.js dev server..."
	npm run dev

build:
	@echo "Building Next.js production bundle..."
	npm run build

test:
	@echo "Running Next.js tests..."
	npm test

test-e2e:
	@echo "Running E2E tests..."
	npm run test:e2e

clean:
	@echo "Cleaning build artifacts..."
	rm -rf .next out node_modules/.cache
```

### Update Root Makefile

```makefile
build-web:
	@cd services/web && npm run build

test-web:
	@cd services/web && npm test

dev-web:
	@cd services/web && npm run dev
```

### Validation

- [ ] `cd services/web && npm test` - succeeds
- [ ] `cd services/web && npm run build` - succeeds
- [ ] `docker build -f services/web/build/Dockerfile .` - succeeds
- [ ] Commit changes: `git commit -m "refactor: Consolidate Web service"`

---

## Phase 5: Root Cleanup (1 day)

### Identify True Duplicates

```bash
# Check what's left at root that might be duplicates
ls -la cmd/
ls -la internal/
ls -la tests/
```

### Remove Confirmed Duplicates

**ONLY if confirmed duplicates:**

- [ ] Remove `cmd/` if duplicate of `services/api/cmd/`
- [ ] Remove `internal/` if duplicate of `services/api/internal/`
- [ ] Remove `tests/` if duplicate of `services/api/tests/`

### Check Root go.mod

```bash
# See what imports root module
rg "github.com/rastrigin-systems/ubik-enterprise\"" --type go

# If nothing imports root module, consider removing
# But KEEP go.work!
```

- [ ] Decision: Keep or remove root `go.mod`?
- [ ] Document decision in `docs/architecture/DECISIONS.md`

### Update Root README.md

Update `README.md` to point to service READMEs:

```markdown
## Services

- [API Server](./services/api/README.md) - Go API server
- [CLI Client](./services/cli/README.md) - Command-line interface
- [Web UI](./services/web/README.md) - Next.js web application

## Documentation

See [CLAUDE.md](./CLAUDE.md) for complete documentation.
```

### Create CODEOWNERS

Create `.github/CODEOWNERS`:

```
# Service ownership
/services/api/ @backend-team
/services/cli/ @cli-team
/services/web/ @frontend-team

# Platform infrastructure
/platform/ @platform-team

# Shared packages
/pkg/ @platform-team
/internal/ @platform-team

# Documentation
/docs/ @tech-writers

# Build and CI
/.github/ @platform-team
/scripts/ @platform-team
```

### Validation

- [ ] `make test-all` - all tests pass
- [ ] `make build-all` - all services build
- [ ] No broken imports (`go build ./...`)
- [ ] Commit changes: `git commit -m "refactor: Clean up root directory"`

---

## Phase 6: CI/CD Updates (2-3 days)

### Split GitHub Actions Workflows

Create `.github/workflows/api-ci.yml`:

```yaml
name: API CI

on:
  push:
    branches: [ main, develop ]
    paths:
      - 'services/api/**'
      - 'platform/**'
      - 'pkg/**'
      - 'generated/**'
  pull_request:
    branches: [ main, develop ]
    paths:
      - 'services/api/**'
      - 'platform/**'
      - 'pkg/**'

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Install tools
        run: make install-tools

      - name: Generate code
        run: make generate-db generate-api

      - name: Run API tests
        run: cd services/api && make test
```

- [ ] Create `.github/workflows/api-ci.yml`
- [ ] Create `.github/workflows/cli-ci.yml`
- [ ] Create `.github/workflows/web-ci.yml`
- [ ] Create `.github/workflows/monorepo-ci.yml` (for cross-cutting checks)

### Update Deployment Workflows

Option A: Per-service deployment workflows

- [ ] Create `.github/workflows/api-deploy.yml`
- [ ] Create `.github/workflows/web-deploy.yml`

Option B: Keep root `cloudbuild.yaml` but reference service configs

- [ ] Update `cloudbuild.yaml` to use service-specific Dockerfiles
- [ ] Reference `services/*/build/Dockerfile.gcp`

### Validation

- [ ] Push to feature branch - CI triggers
- [ ] Only affected services run CI
- [ ] All CI checks pass
- [ ] Commit changes: `git commit -m "ci: Split CI workflows per service"`

---

## Phase 7: Documentation Updates (2-3 days)

### Update Core Documentation

- [ ] Update `CLAUDE.md` with new structure diagrams
- [ ] Update file paths throughout documentation
- [ ] Update command examples

### Reorganize Documentation

```bash
# Move existing docs to organized structure
git mv docs/TESTING.md docs/guides/
git mv docs/DEVELOPMENT.md docs/guides/
git mv docs/DEBUGGING.md docs/guides/
git mv docs/QUICKSTART.md docs/guides/
git mv docs/DEV_WORKFLOW.md docs/guides/
```

### Create New Documentation

- [ ] Create `docs/architecture/DECISIONS.md` with ADRs
- [ ] Create `docs/architecture/MONOREPO.md` (move existing plan)
- [ ] Create `services/api/docs/API.md`
- [ ] Create `services/cli/docs/CLI_ARCHITECTURE.md`
- [ ] Create `services/web/docs/WEB_ARCHITECTURE.md`

### Update Path References

```bash
# Find all documentation references to old paths
grep -r "shared/schema" docs/
grep -r "sqlc/" docs/
grep -r "services/api/Dockerfile" docs/

# Update to new paths
# This will be a manual process
```

### Update CLAUDE.md Structure Section

Replace structure diagram in `CLAUDE.md` with new structure.

### Validation

- [ ] All docs render correctly
- [ ] No broken links
- [ ] Commands in docs work
- [ ] Commit changes: `git commit -m "docs: Update for new monorepo structure"`

---

## Post-Migration Validation

### Full System Test

- [ ] `make db-up` - database starts
- [ ] `make generate` - all code generation succeeds
- [ ] `make test-all` - all tests pass (API + CLI + Web)
- [ ] `make build-all` - all services build
- [ ] `make dev` - all services start via docker-compose

### Service Independence Test

```bash
# Test each service can build independently
cd services/api && make build
cd services/cli && make build
cd services/web && npm run build
```

- [ ] API service builds independently
- [ ] CLI service builds independently
- [ ] Web service builds independently

### Documentation Validation

- [ ] Read through `CLAUDE.md` - all paths correct
- [ ] Read through `docs/guides/QUICKSTART.md` - commands work
- [ ] Read through `docs/guides/TESTING.md` - examples work
- [ ] Check service READMEs exist and are helpful

### CI/CD Validation

- [ ] Push to feature branch
- [ ] Only affected services run CI
- [ ] All checks pass
- [ ] No broken workflows

---

## Rollback Plan

If migration fails at any phase:

1. **Immediate Rollback:**
   ```bash
   git reset --hard backup/pre-migration
   git push --force
   ```

2. **Partial Rollback:**
   ```bash
   # Revert specific phase commits
   git revert <commit-hash>
   ```

3. **Fix Forward:**
   - Identify broken component
   - Fix in place
   - Re-run validation

---

## Communication Checklist

### Before Migration

- [ ] Notify team of migration timeline
- [ ] Share migration plan documents
- [ ] Get approval from stakeholders
- [ ] Schedule migration window (ideally no active PRs)

### During Migration

- [ ] Post updates in team chat after each phase
- [ ] Mark main branch as "in migration" (if needed)
- [ ] Keep team informed of any issues

### After Migration

- [ ] Announce completion
- [ ] Share updated documentation
- [ ] Conduct team walkthrough of new structure
- [ ] Update onboarding docs

---

## Success Criteria

### Must Have

- ✅ All tests pass
- ✅ All services build independently
- ✅ No broken imports
- ✅ Documentation updated
- ✅ CI/CD working

### Should Have

- ✅ Per-service CI triggers only for affected services
- ✅ CODEOWNERS file in place
- ✅ Service-specific docs created
- ✅ Root Makefile delegates to services

### Nice to Have

- ✅ Architecture Decision Records documented
- ✅ Service template created for future services
- ✅ Migration lessons learned documented

---

## Timeline Estimate

| Phase | Duration | Cumulative |
|-------|----------|------------|
| Phase 1: Preparation | 1-2 days | 2 days |
| Phase 2: API Service | 2-3 days | 5 days |
| Phase 3: CLI Service | 1-2 days | 7 days |
| Phase 4: Web Service | 1 day | 8 days |
| Phase 5: Root Cleanup | 1 day | 9 days |
| Phase 6: CI/CD | 2-3 days | 12 days |
| Phase 7: Documentation | 2-3 days | 15 days |
| **Total** | **2-3 weeks** | **15 days** |

**Buffer:** Add 3-5 days for unexpected issues.

---

## Appendix: Quick Commands

### File Moves with History

```bash
# Preserve git history when moving files
git mv <old-path> <new-path>

# Move directory with history
git mv <old-dir> <new-dir>
```

### Find and Replace

```bash
# Find all files with old path
grep -r "shared/schema" . --exclude-dir=node_modules --exclude-dir=.git

# Replace in all Go files
find . -name "*.go" -type f -exec sed -i '' 's|old|new|g' {} \;

# Replace in all YAML files
find . -name "*.yaml" -type f -exec sed -i '' 's|old|new|g' {} \;
```

### Validation Commands

```bash
# Check for broken imports
go build ./...

# Check for missing files
git status

# Check for uncommitted changes
git diff
```

---

**See also:**
- [MONOREPO_REFACTORING_PLAN.md](./MONOREPO_REFACTORING_PLAN.md) - Complete design
- [MONOREPO_COMPARISON.md](./MONOREPO_COMPARISON.md) - Visual comparison
