# Docker Testing Checklist

**Use this checklist for ANY change that affects Docker deployment**

## When to Use This Checklist

- ✅ Changes to Dockerfile
- ✅ Changes to cloudbuild.yaml
- ✅ New API endpoints added
- ✅ New file resources needed (images, configs, specs)
- ✅ Environment variable changes
- ✅ New dependencies added

## Pre-Deployment Checklist

### 1. Local Build Test ⚠️ MANDATORY

```bash
# Build the Docker image locally
docker build -f services/api/Dockerfile.gcp -t ubik-api-test .

# Verify build succeeded
echo $?  # Should be 0
```

**✅ Build must succeed before proceeding**

### 2. Image Inspection

```bash
# Verify required files are in the image
docker run --rm ubik-api-test ls -la /app/

# Check for specific files
docker run --rm ubik-api-test ls -la /app/platform/api-spec/spec.yaml
docker run --rm ubik-api-test ls -la /app/server

# Verify environment variables
docker run --rm ubik-api-test env | grep -E "PROJECT_ROOT|PORT"
```

**✅ All required files must be present**
**✅ Environment variables must be set**

### 3. Local Container Test

```bash
# Start database
make db-up

# Run the container locally
docker run --rm -p 8080:8080 \
  -e DATABASE_URL="postgres://ubik:ubik_dev_password@host.docker.internal:5432/ubik?sslmode=disable" \
  ubik-api-test

# In another terminal, test endpoints
curl http://localhost:8080/api/v1/health
curl http://localhost:8080/api/docs/
curl http://localhost:8080/api/docs/spec.yaml
```

**✅ All endpoints must return 200 OK**
**✅ Swagger UI must load correctly**
**✅ No 404 errors**

### 4. Compare Local vs Docker Behavior

```bash
# Test locally (non-Docker)
make build-server && ./bin/ubik-server &
curl http://localhost:3001/api/docs/

# Test Docker
docker run -p 8080:8080 ubik-api-test &
curl http://localhost:8080/api/docs/

# Results should be identical
```

**✅ Behavior must match between local and Docker**

### 5. Review Dockerfile Changes

**For ANY Dockerfile change, verify:**
- [ ] All COPY commands include necessary files
- [ ] ENV variables are set correctly
- [ ] WORKDIR is appropriate
- [ ] Ports are exposed
- [ ] CMD/ENTRYPOINT is correct

**For new file dependencies:**
- [ ] Files are copied in both builder and runtime stages (if needed)
- [ ] File paths match what the code expects
- [ ] Permissions are correct

### 6. cloudbuild.yaml Validation

```bash
# Verify syntax
cat cloudbuild-api.yaml | grep -E "steps:|name:|args:"

# Check build context
grep -A 5 "docker build" cloudbuild-api.yaml
```

**✅ Build context must be `.` (project root)**
**✅ Dockerfile path must be correct**

### 7. Pre-Commit Verification

```bash
# Run all checks before committing
make build-server     # Local build works
docker build -f services/api/Dockerfile.gcp -t ubik-api-test .  # Docker build works
docker run --rm ubik-api-test ls -la /app/  # Files present
make test             # Tests pass
```

**✅ ALL checks must pass before commit**

## Common Mistakes to Avoid

### ❌ Mistake 1: Assuming local = Docker
**Problem:** Files in local filesystem may not be in Docker image

**Solution:** ALWAYS test Docker build locally

### ❌ Mistake 2: Not verifying file paths
**Problem:** COPY command uses wrong path or files not in context

**Solution:** Inspect the image after building

### ❌ Mistake 3: Wrong WORKDIR
**Problem:** Code expects files at `/workspace` but WORKDIR is `/app`

**Solution:** Verify paths match code expectations

### ❌ Mistake 4: Missing runtime files
**Problem:** Files copied to builder but not runtime stage

**Solution:** COPY files to final stage if needed at runtime

### ❌ Mistake 5: Environment variables not set
**Problem:** Code reads ENV vars but Dockerfile doesn't set them

**Solution:** Add ENV directives to Dockerfile

## Fast Testing Commands

```bash
# Quick Docker test (copy-paste)
docker build -f services/api/Dockerfile.gcp -t test . && \
docker run --rm -p 8080:8080 -e DATABASE_URL="postgres://ubik:ubik_dev_password@host.docker.internal:5432/ubik?sslmode=disable" test &
sleep 3 && curl http://localhost:8080/api/v1/health && \
curl -I http://localhost:8080/api/docs/
```

## CI/CD Integration

**The GitHub Actions workflow automatically:**
- ✅ Builds Docker image via Cloud Build
- ✅ Runs smoke tests after deployment
- ❌ Does NOT test Docker build locally first

**Manual testing is still required before pushing!**

## Checklist Template

Copy this for each Docker change:

```
Docker Testing Checklist for: [CHANGE DESCRIPTION]

Pre-Deployment:
- [ ] Local Docker build successful
- [ ] Files present in image (verified with ls)
- [ ] Environment variables set (verified with env)
- [ ] Container runs locally
- [ ] Health check returns 200
- [ ] New endpoints return correct responses
- [ ] Behavior matches non-Docker local build
- [ ] Dockerfile changes reviewed
- [ ] All tests pass

Post-Deployment:
- [ ] Production health check succeeds
- [ ] Production endpoints work
- [ ] No 404 errors in logs
- [ ] Smoke tests pass
```

## Emergency Rollback

If production fails:

```bash
# Check last successful deployment
gh run list --workflow "API Server" --limit 10

# Find last successful run
# Revert to that commit
git revert HEAD
git push

# Or roll back in Cloud Run console
# Cloud Run > ubik-api > Revisions > Select previous > Manage Traffic
```

---

**Remember:** 5 minutes of Docker testing saves 30 minutes of debugging production! 🐳
