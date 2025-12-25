# Debugging Guide

**Last Updated:** 2025-11-05

This guide covers debugging strategies, best practices, and common pitfalls when working with the Arfa platform.

---

## Table of Contents

- [The Golden Rule](#the-golden-rule)
- [Debugging Workflow](#debugging-workflow)
- [Common Debugging Techniques](#common-debugging-techniques)
- [Common Pitfalls](#common-pitfalls)
- [Integration Testing Debugging](#integration-testing-debugging)
- [CLI Debugging](#cli-debugging)
- [API Debugging](#api-debugging)
- [Database Debugging](#database-debugging)
- [Real-World Examples](#real-world-examples)

---

## The Golden Rule

### Check the Data, Not Just the Code

**When integration tests or CLI operations fail unexpectedly, the issue is often in the data, not the code.**

**The Debugging Hierarchy:**
```
1. Check the data (database, cache, config files)
2. Check the logs (request/response, errors)
3. Check the binaries (are they up-to-date?)
4. Check the code (logic errors)
```

**Why This Matters:**
- Foreign key errors → Missing seed data
- Authentication failures → Stale cache/tokens
- Unexpected behavior → Outdated binary
- Integration failures → Database state mismatch

---

## Debugging Workflow

### Standard Debugging Process

```
Issue Reported
    ↓
Add Debug Logging
    ↓
Check Database State
    ↓
Check Cache/Config
    ↓
Rebuild Binaries
    ↓
Verify Fix
    ↓
Remove Debug Logging
```

---

### Step-by-Step Guide

#### 1. Add Request/Response Logging

**Always add logging FIRST before making code changes.**

```go
// Temporary debug logging
fmt.Fprintf(os.Stderr, "[DEBUG] Request: %s %s\n", method, url)
fmt.Fprintf(os.Stderr, "[DEBUG] Request body: %s\n", string(jsonData))
fmt.Fprintf(os.Stderr, "[DEBUG] Response: status=%d body=%s\n", resp.StatusCode, string(respBody))
```

**Why logging first?**
- Reveals actual request/response data
- Shows what the system is actually doing
- Often reveals the root cause immediately

---

#### 2. Verify Database State

**Check if required records exist:**

```bash
# Check if required records exist
docker exec arfa-postgres psql -U arfa -d arfa -c "SELECT * FROM table_name WHERE id = '...'"

# Check foreign key relationships
docker exec arfa-postgres psql -U arfa -d arfa -c "\d table_name"

# Check table counts
docker exec arfa-postgres psql -U arfa -d arfa -c "SELECT COUNT(*) FROM table_name"
```

**Common database issues:**
- Missing seed data
- Foreign key constraint violations
- Orphaned records
- Incorrect org_id scoping

---

#### 3. Check for Stale Cache

**CLI cache locations:**

```bash
# Clear agent configs
rm -rf ~/.arfa/agents/

# Clear offline log queue
rm -rf ~/.arfa/log_queue/

# Clear entire config
rm -rf ~/.arfa/

# Force fresh sync
./bin/arfa-test sync
```

**API cache (if using Redis/memcached):**

```bash
# Flush Redis cache
docker exec arfa-redis redis-cli FLUSHALL

# Or restart Redis
docker restart arfa-redis
```

---

#### 4. Rebuild Binaries

**Always rebuild if source changed recently:**

```bash
# Check binary timestamp vs source file timestamp
ls -la ./bin/arfa-test
ls -la services/cli/cmd/arfa/main.go

# Rebuild if stale
cd services/cli && go build -o ../../bin/arfa-test ./cmd/arfa

# Or rebuild everything
make build
```

**Why binaries get stale:**
- Code changes not reflected in binary
- Old binary cached by IDE/test runner
- Build process interrupted
- Wrong binary being executed

---

#### 5. Verify Fix

**After fixing, verify:**

```bash
# Run affected tests
make test-integration

# Run full test suite
make test

# Manual verification
./bin/arfa-test <command>
```

---

## Common Debugging Techniques

### Request/Response Logging

**HTTP Client:**

```go
// Add logging wrapper
type loggingRoundTripper struct {
    next http.RoundTripper
}

func (l *loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
    // Log request
    body, _ := httputil.DumpRequest(req, true)
    log.Printf("[HTTP] Request:\n%s", body)

    // Execute request
    resp, err := l.next.RoundTrip(req)

    // Log response
    if resp != nil {
        body, _ := httputil.DumpResponse(resp, true)
        log.Printf("[HTTP] Response:\n%s", body)
    }

    return resp, err
}
```

---

### Database Query Logging

**Enable SQL logging in tests:**

```go
// In test setup
db := sqlx.MustConnect("postgres", dsn)
db.MapperFunc(strings.ToLower)

// Log all queries
db.DB.SetConnMaxLifetime(time.Minute * 3)
db.DB.SetMaxOpenConns(10)
db.DB.SetMaxIdleConns(10)

// Add query logger
db = db.Unsafe() // Don't use in production
```

**PostgreSQL logging:**

```bash
# Enable query logging
docker exec arfa-postgres psql -U arfa -d arfa -c "ALTER SYSTEM SET log_statement = 'all';"
docker exec arfa-postgres psql -U arfa -d arfa -c "SELECT pg_reload_conf();"

# View logs
docker logs arfa-postgres -f | grep "LOG:"

# Disable after debugging
docker exec arfa-postgres psql -U arfa -d arfa -c "ALTER SYSTEM SET log_statement = 'none';"
```

---

### Environment Variable Debugging

**Print all env vars:**

```bash
# In shell
env | grep ARFA
env | sort

# In Go code
fmt.Fprintf(os.Stderr, "[DEBUG] Environment:\n")
for _, e := range os.Environ() {
    fmt.Fprintf(os.Stderr, "  %s\n", e)
}
```

---

### Docker Container Debugging

**Inspect containers:**

```bash
# List all containers
docker ps -a

# Inspect container
docker inspect <container-name>

# View container logs
docker logs <container-name> -f

# Execute command in container
docker exec -it <container-name> sh

# Check container environment
docker exec <container-name> env
```

---

## Common Pitfalls

### 1. Assuming Code is Wrong When Data is Wrong

**Symptom:** Tests fail with foreign key errors, 404s, or unexpected data

**Common Causes:**
- Missing seed data
- Incorrect test fixtures
- Database not reset between tests
- Wrong org_id in test data

**Solution:**
```bash
# Reset database
make db-reset

# Re-run migrations
make db-migrate

# Verify seed data
docker exec arfa-postgres psql -U arfa -d arfa -c "SELECT COUNT(*) FROM organizations"
```

---

### 2. Not Checking Cache Invalidation

**Symptom:** Changes don't take effect, old data returned

**Common Causes:**
- Stale config files in `~/.arfa/`
- Old tokens in cache
- Browser cache (for web UI)
- API response cache

**Solution:**
```bash
# Clear all caches
rm -rf ~/.arfa/
docker restart arfa-redis  # If using Redis

# Force refresh
./bin/arfa-test sync --force
```

---

### 3. Testing with Stale Binaries

**Symptom:** Code changes don't affect behavior

**Common Causes:**
- Binary built before latest changes
- Wrong binary in PATH
- IDE running old binary
- Test runner cached binary

**Solution:**
```bash
# Rebuild everything
make clean
make build

# Verify binary timestamp
ls -la ./bin/arfa-test

# Run with absolute path
/Users/you/Projects/arfa/bin/arfa-test <command>
```

---

### 4. Missing Org-Level Configurations

**Symptom:** Foreign key errors on employee configs

**Common Causes:**
- `org_agent_configs` empty
- Employee config references non-existent org config
- Org config deleted but employee configs remain

**Solution:**
```bash
# Check org configs exist
docker exec arfa-postgres psql -U arfa -d arfa -c \
  "SELECT id, org_id, agent_id FROM org_agent_configs"

# Create org config if missing
# (via API or SQL)
```

---

## Integration Testing Debugging

### Common Integration Test Issues

#### Test Database Not Reset

**Problem:** Tests fail due to leftover data

**Solution:**
```go
// In test setup
func setupTestDB(t *testing.T) *sql.DB {
    // Use testcontainers to ensure clean state
    container := testcontainers.NewPostgresContainer(...)

    // Run migrations
    migrate.Up(db)

    // Add seed data
    seedTestData(db)

    return db
}
```

---

#### Flaky Tests

**Problem:** Tests pass sometimes, fail other times

**Common Causes:**
- Race conditions
- Time-dependent tests
- External dependencies
- Port conflicts

**Solution:**
```go
// Add proper synchronization
var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    // ... async work
}()
wg.Wait()

// Mock time
now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
timeMock := &TimeMock{Now: now}

// Use random ports
listener, _ := net.Listen("tcp", "127.0.0.1:0")
port := listener.Addr().(*net.TCPAddr).Port
```

---

## CLI Debugging

### CLI-Specific Issues

#### Authentication Failures

**Problem:** `arfa login` fails or returns 401

**Debug steps:**
```bash
# Check API is running
curl http://localhost:8080/health

# Verify credentials
cat ~/.arfa/config.json | jq .

# Check token expiration
# Decode JWT token
echo "<token>" | cut -d. -f2 | base64 -d | jq .

# Clear auth and re-login
rm ~/.arfa/config.json
./bin/arfa-test login
```

---

#### Docker Container Issues

**Problem:** Containers won't start or connect

**Debug steps:**
```bash
# Check Docker daemon
docker ps

# Check container logs
docker logs arfa-<container> -f

# Check network
docker network inspect arfa-network

# Recreate network
docker network rm arfa-network
docker network create arfa-network

# Remove stale containers
docker rm -f $(docker ps -aq --filter "name=arfa-")
```

---

## API Debugging

### API-Specific Issues

#### Middleware Issues

**Problem:** Requests fail with 401/403 even with valid auth

**Debug steps:**
```go
// Add middleware logging
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("[MIDDLEWARE] %s %s", r.Method, r.URL.Path)
        log.Printf("[MIDDLEWARE] Headers: %v", r.Header)

        next.ServeHTTP(w, r)

        log.Printf("[MIDDLEWARE] Complete")
    })
}
```

---

#### RLS (Row-Level Security) Issues

**Problem:** Queries return empty results

**Debug steps:**
```bash
# Check RLS is enabled
docker exec arfa-postgres psql -U arfa -d arfa -c \
  "SELECT schemaname, tablename, policyname FROM pg_policies WHERE tablename = 'employees'"

# Test query as specific org
docker exec arfa-postgres psql -U arfa -d arfa -c \
  "SET app.current_org_id = '<org-uuid>'; SELECT * FROM employees"

# Disable RLS temporarily for testing
docker exec arfa-postgres psql -U arfa -d arfa -c \
  "ALTER TABLE employees DISABLE ROW LEVEL SECURITY"
```

---

## Database Debugging

### Schema Issues

**Check schema version:**

```bash
# Check migrations table
docker exec arfa-postgres psql -U arfa -d arfa -c \
  "SELECT version, dirty FROM schema_migrations"

# Check table exists
docker exec arfa-postgres psql -U arfa -d arfa -c \
  "\dt"

# Check table structure
docker exec arfa-postgres psql -U arfa -d arfa -c \
  "\d employees"
```

---

### Data Integrity Issues

**Check foreign keys:**

```bash
# List all foreign keys
docker exec arfa-postgres psql -U arfa -d arfa -c \
  "SELECT tc.table_name, kcu.column_name, ccu.table_name AS foreign_table_name
   FROM information_schema.table_constraints AS tc
   JOIN information_schema.key_column_usage AS kcu ON tc.constraint_name = kcu.constraint_name
   JOIN information_schema.constraint_column_usage AS ccu ON ccu.constraint_name = tc.constraint_name
   WHERE tc.constraint_type = 'FOREIGN KEY'"

# Find orphaned records
docker exec arfa-postgres psql -U arfa -d arfa -c \
  "SELECT * FROM employee_agent_configs eac
   WHERE NOT EXISTS (SELECT 1 FROM org_agent_configs oac WHERE oac.id = eac.org_agent_config_id)"
```

---

## Real-World Examples

### Example 1: CLI Logging Foreign Key Error

**Problem:** Logs failing with HTTP 500

**Symptoms:**
```
2024-11-05 10:30:00 [ERROR] Failed to create log entry: HTTP 500
2024-11-05 10:30:00 [DEBUG] Response: foreign key constraint violation
```

**Debug Process:**

```bash
# 1. Add debug logging
# (Added request/response logging to client)

# 2. Revealed error
[DEBUG] Response: {"error": "agent_id violates foreign key constraint"}

# 3. Check database
docker exec arfa-postgres psql -U arfa -d arfa -c "SELECT COUNT(*) FROM org_agent_configs"
# Result: 0 rows (empty table!)

# 4. Root cause
# - Empty org_agent_configs table
# - Stale cache with test agent IDs
# - Employee configs referencing non-existent org configs

# 5. Fix
# - Create org_agent_config record
# - Clear cache: rm -rf ~/.arfa/agents/
# - Rebuild: make build
# - Re-sync: ./bin/arfa-test sync

# 6. Result
✅ Logs flowing successfully
```

**Lesson:** Check database state BEFORE assuming code is wrong.

**See:** `/tmp/CLI_LOGGING_FIX_COMPLETE.md` for complete details.

---

### Example 2: Flaky Integration Tests

**Problem:** Tests pass locally, fail in CI

**Debug Process:**

```bash
# 1. Add deterministic time
now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

# 2. Use random ports
listener, _ := net.Listen("tcp", "127.0.0.1:0")

# 3. Add proper cleanup
t.Cleanup(func() {
    container.Terminate(ctx)
    os.RemoveAll(tempDir)
})

# 4. Add synchronization
done := make(chan bool)
go func() {
    // ... async work
    done <- true
}()
<-done

# Result: Tests now deterministic
```

---

## See Also

- [TESTING.md](./TESTING.md) - Testing strategies
- [DEVELOPMENT.md](./DEVELOPMENT.md) - Development workflow
- [DATABASE.md](./DATABASE.md) - Database operations
- [MCP_SERVERS.md](./MCP_SERVERS.md) - MCP server debugging
- [QUICK_REFERENCE.md](./QUICK_REFERENCE.md) - Quick command reference
