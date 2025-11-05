# Railway Deployment Guide

Complete guide to deploying Ubik Enterprise to Railway.app.

**Estimated Setup Time:** 10-15 minutes
**Estimated Cost:** $10-15/month (Hobby plan)

---

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Detailed Setup](#detailed-setup)
- [Configuration](#configuration)
- [Deployment](#deployment)
- [Post-Deployment](#post-deployment)
- [Monitoring & Logs](#monitoring--logs)
- [Troubleshooting](#troubleshooting)
- [Cost Management](#cost-management)

---

## Prerequisites

- GitHub account with ubik-enterprise repository
- Railway account (free to create at [railway.app](https://railway.app))
- Git CLI installed locally
- Access to your repository settings

---

## Quick Start

**5-minute deployment:**

```bash
# 1. Sign up for Railway
open https://railway.app

# 2. Create new project from GitHub
# - Click "New Project"
# - Select "Deploy from GitHub repo"
# - Choose "sergei-rastrigin/ubik-enterprise"

# 3. Add PostgreSQL database
# - Click "+ New"
# - Select "Database" → "PostgreSQL"

  

# 5. Configure environment variables (see below)

# 6. Deploy!
git push origin main
```

**Important for Monorepo:** Railway needs to create two separate services pointing to different directories in the same repo.

Railway will auto-deploy on every push to `main`.

---

## Detailed Setup

### Step 1: Create Railway Account

1. Go to [railway.app](https://railway.app)
2. Click "Login" → "Login with GitHub"
3. Authorize Railway to access your repositories
4. Choose **Hobby Plan** ($5/month with $5 included resources)

### Step 2: Create New Project

1. Click **"New Project"**
2. Select **"Deploy from GitHub repo"**
3. Choose **`sergei-rastrigin/ubik-enterprise`**
4. Railway will detect your monorepo structure

### Step 3: Add Database

1. In your project, click **"+ New"**
2. Select **"Database"** → **"PostgreSQL"**
3. Railway provisions a PostgreSQL instance automatically
4. Note: Database URL is available as `$POSTGRES_URL`

### Step 4: Create API Service

1. Click **"+ New"** → **"Service"**
2. Select **"GitHub Repo"** → **`ubik-enterprise`**
3. Configure the service:

**Settings → General:**
- **Service Name:** `ubik-api`
- **Root Directory:** `services/api` (important for monorepo!)
- **Build Command:** Auto-detected from `nixpacks.toml`
- **Start Command:** `../../bin/ubik-server`

**Settings → Variables:**
Add these environment variables:

```bash
# Database (auto-populated from PostgreSQL service)
DATABASE_URL=${{Postgres.POSTGRES_URL}}

# JWT Secret (generate with: openssl rand -base64 32)
JWT_SECRET=<your-generated-secret>

# Server
PORT=8080
GIN_MODE=release
LOG_LEVEL=info
LOG_FORMAT=json
```

**Settings → Networking:**
- Click **"Generate Domain"**
- Note the URL (e.g., `https://ubik-api-production.up.railway.app`)

### Step 5: Create Next.js Service

1. Click **"+ New"** → **"Service"**
2. Select **"GitHub Repo"** → **`ubik-enterprise`**
3. Configure the service:

**Settings → General:**
- **Service Name:** `ubik-web`
- **Root Directory:** `web` (important for monorepo!)
- **Build Command:** Auto-detected from `nixpacks.toml`
- **Start Command:** `npm start`

**Settings → Variables:**
Add these environment variables:

```bash
# API URL (from Step 4)
NEXT_PUBLIC_API_URL=https://ubik-api-production.up.railway.app

# Node Environment
NODE_ENV=production

# Next.js (if using NextAuth)
NEXTAUTH_URL=${{RAILWAY_PUBLIC_DOMAIN}}
NEXTAUTH_SECRET=<your-generated-secret>
```

**Settings → Networking:**
- Click **"Generate Domain"**
- Note the URL (e.g., `https://ubik-web-production.up.railway.app`)

### Step 6: Configure Service Dependencies

**Important:** API depends on PostgreSQL

1. Go to **`ubik-api`** service
2. Click **"Settings"** → **"Service Dependencies"**
3. Add **PostgreSQL** as a dependency
4. This ensures database starts before API

---

## Configuration

### Environment Variables Reference

See [`.env.railway.example`](../.env.railway.example) for complete list.

**Critical Variables:**

| Variable | Service | Value | How to Generate |
|----------|---------|-------|----------------|
| `DATABASE_URL` | API | `${{Postgres.POSTGRES_URL}}` | Auto from Railway |
| `JWT_SECRET` | API | Random 32-byte string | `openssl rand -base64 32` |
| `NEXT_PUBLIC_API_URL` | Web | API service URL | From Railway dashboard |
| `PORT` | API | `8080` | Fixed |
| `NODE_ENV` | Web | `production` | Fixed |

### Generating Secrets

```bash
# JWT Secret (copy output)
openssl rand -base64 32

# NextAuth Secret (if needed)
openssl rand -base64 32
```

### Database Migration

Railway auto-runs migrations on deployment if configured.

**Option 1: Manual Migration (First Deploy)**

```bash
# Connect to Railway PostgreSQL
railway login
railway link
railway run psql $DATABASE_URL

# Run schema
\i shared/schema/schema.sql
\q
```

**Option 2: Auto-Migration (Recommended)**

Add to `services/api/nixpacks.toml`:

```toml
[phases.build]
cmds = [
  "cd services/api",
  "go mod download",
  "go build -o ../../bin/ubik-server ./cmd/server"
]

[phases.migrate]
cmds = [
  "psql $DATABASE_URL < shared/schema/schema.sql || true"
]

[start]
cmd = "./bin/ubik-server"
dependsOn = ["migrate"]
```

---

## Deployment

### Initial Deployment

Railway auto-deploys when you connect GitHub repo:

1. Push to `main` branch
2. Railway detects changes
3. Builds services
4. Runs migrations (if configured)
5. Deploys to production

**Monitor deployment:**
```bash
# Install Railway CLI
npm install -g @railway/cli

# Login
railway login

# Link to project
railway link

# Watch logs
railway logs
```

### Subsequent Deployments

**Automatic (Recommended):**
```bash
git push origin main
# Railway auto-deploys
```

**Manual Deploy:**
```bash
railway up
```

**Rollback:**
```bash
# Go to Railway dashboard → Deployments
# Click previous deployment → "Redeploy"
```

---

## Post-Deployment

### Verify Services

**1. Check API Health:**
```bash
curl https://ubik-api-production.up.railway.app/health

# Expected response:
{"status": "ok"}
```

**2. Check Next.js:**
```bash
open https://ubik-web-production.up.railway.app
```

**3. Test Authentication:**
```bash
curl -X POST https://ubik-api-production.up.railway.app/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password"}'
```

### Seed Database

```bash
# Connect to Railway
railway login
railway link

# Run seed script
railway run bash
./scripts/seed-production.sh
```

### Configure CLI to Use Railway API

Update your local CLI config:

```bash
# Edit ~/.ubik/config.json
{
  "api_url": "https://ubik-api-production.up.railway.app",
  "token": "<get-from-login>"
}

# Test
ubik login
ubik agents list
```

---

## Monitoring & Logs

### View Logs

**Railway Dashboard:**
1. Go to service (ubik-api or ubik-web)
2. Click "Deployments" tab
3. Click latest deployment
4. View real-time logs

**Railway CLI:**
```bash
# All services
railway logs

# Specific service
railway logs --service ubik-api

# Follow logs
railway logs --follow
```

### Metrics

**Railway provides built-in metrics:**
- CPU usage
- Memory usage
- Network I/O
- Request count

**Access:**
1. Go to service
2. Click "Metrics" tab
3. View graphs

### Alerts

**Set up alerts in Railway:**
1. Go to "Settings" → "Alerts"
2. Configure thresholds:
   - High CPU (>80%)
   - High memory (>80%)
   - Service down

---

## Troubleshooting

### Common Issues

**1. Build Fails**

```bash
# Check build logs in Railway dashboard
# Common causes:
- Missing dependencies in go.mod/package.json
- Incorrect nixpacks.toml path
- Environment variables not set

# Fix:
railway logs --service ubik-api
# Review error, fix, and push
```

**2. Service Won't Start**

```bash
# Check if DATABASE_URL is set
railway variables --service ubik-api

# Verify database is running
railway status

# Common fixes:
- Add service dependency (API → PostgreSQL)
- Check PORT matches (8080 for API)
- Verify start command in nixpacks.toml
```

**3. Database Connection Error**

```bash
# Verify DATABASE_URL format
railway variables --service ubik-api | grep DATABASE_URL

# Should be: postgresql://user:pass@host:port/db

# Test connection manually
railway run psql $DATABASE_URL
\dt  # List tables
\q
```

**4. CORS Errors**

```bash
# Update API CORS settings
railway variables set CORS_ORIGINS=https://ubik-web-production.up.railway.app

# Or add to services/api code:
// internal/middleware/cors.go
allowedOrigins := []string{
  os.Getenv("WEB_URL"),
  "https://ubik-web-production.up.railway.app",
}
```

**5. Next.js Build Timeout**

```bash
# Increase build timeout in Railway
# Settings → Build → Timeout → 10 minutes

# Or optimize build
cd web
npm run build -- --max-old-space-size=4096
```

### Debug Mode

**Enable verbose logging:**

```bash
# API service
railway variables set LOG_LEVEL=debug

# Next.js
railway variables set DEBUG=*
```

---

## Cost Management

### Estimated Costs (Hobby Plan)

```
Hobby Plan Subscription:        $5.00/month (includes $5 resources)

Resources:
- PostgreSQL (512MB RAM):       ~$2.56/month
- API Server (256MB, 0.25 CPU): ~$2.50/month
- Next.js (256MB, 0.25 CPU):    ~$2.50/month
- Network egress (1GB):         ~$0.10/month
----------------------------------------------------
TOTAL:                          ~$12.66/month

Actual charge:                  $7.66/month
(Subscription covers first $5 of resources)
```

### Monitor Usage

```bash
# Railway CLI
railway usage

# Or in dashboard:
# Project → Usage → View current month
```

### Optimize Costs

**1. Right-size Resources:**
```bash
# Start small (256MB RAM)
# Scale up only if needed
```

**2. Use Build Cache:**
```toml
# nixpacks.toml already configured for caching
# Faster builds = lower cost
```

**3. Scheduled Scaling (Pro Plan):**
```bash
# Scale down during off-hours (not available on Hobby)
```

### Upgrade to Pro Plan

**When to upgrade:**
- Usage consistently >$5/month
- Need custom domains
- Need more replicas
- Need staging environments

**Pro Plan:** $20/month (includes $20 resources)

---

## Best Practices

### 1. Use Environment Variables for All Config

```bash
# ✅ GOOD
DATABASE_URL=${{Postgres.POSTGRES_URL}}

# ❌ BAD
DATABASE_URL=hardcoded-value
```

### 2. Enable Health Checks

```go
// services/api/internal/handlers/health.go
func HealthCheck(w http.ResponseWriter, r *http.Request) {
    // Check database connection
    db.Ping()

    json.NewEncoder(w).Encode(map[string]string{
        "status": "ok",
        "version": "v0.2.0",
    })
}
```

### 3. Use Structured Logging

```go
// Already configured in your API
log.WithFields(log.Fields{
    "service": "api",
    "version": "v0.2.0",
}).Info("Server started")
```

### 4. Set Up GitHub Actions for Pre-Deploy Tests

```yaml
# .github/workflows/railway-deploy.yml
name: Railway Deploy
on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run tests
        run: make test

  # Railway auto-deploys after tests pass
```

### 5. Use Separate Environments

```bash
# Production (from main branch)
Railway Project: ubik-production

# Staging (from develop branch)
Railway Project: ubik-staging
```

---

## Next Steps

After successful deployment:

1. ✅ Set up custom domain (Pro plan)
2. ✅ Configure monitoring (Sentry, DataDog)
3. ✅ Set up staging environment
4. ✅ Enable automated backups
5. ✅ Add CI/CD with GitHub Actions
6. ✅ Configure alerts (uptime, errors)

---

## Resources

- **Railway Docs:** https://docs.railway.app
- **Railway Discord:** https://discord.gg/railway
- **Pricing Calculator:** https://railway.app/pricing
- **Status Page:** https://status.railway.app

---

## Support

**Railway Support:**
- Discord: https://discord.gg/railway
- Email: team@railway.app
- Docs: https://docs.railway.app

**Project-Specific:**
- GitHub Issues: https://github.com/sergei-rastrigin/ubik-enterprise/issues
- Email: your-email@example.com

---

**Last Updated:** 2025-11-05
**Railway Plan:** Hobby ($5/month)
**Estimated Monthly Cost:** $10-15/month
