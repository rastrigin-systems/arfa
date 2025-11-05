# Railway Deployment - Manual Configuration Required

**Status:** ⚠️ Deployment blocked - requires manual dashboard configuration

**Project:** empowering-rejoicing (8eb7d506-4c2c-4ec8-9079-d4be70dd389f)
**Service:** ubik-enterprise
**Issue:** Root directory must be set manually in Railway dashboard

---

## 🔴 Critical Step Required

Railway deployments are failing because the service is building from the repository root instead of `services/api`. This causes it to encounter the `go.work` file which references the `generated/` module.

### ✅ Solution: Set Root Directory in Railway Dashboard

**Required action:**

1. **Navigate to Railway dashboard:**
   https://railway.app/project/8eb7d506-4c2c-4ec8-9079-d4be70dd389f

2. **Select the `ubik-enterprise` service**

3. **Go to Settings → General**

4. **Find "Root Directory" field**

5. **Set value to:** `services/api`

6. **Click "Save"**

7. **Trigger redeploy:**
   - Option A: Push a commit to `main` branch
   - Option B: Click "Deploy" button in Railway dashboard

---

## 📊 Current Configuration Status

### ✅ Completed Configuration

- [x] Railway CLI installed and authenticated
- [x] Project linked: `empowering-rejoicing`
- [x] Service identified: `ubik-enterprise`
- [x] Environment variables configured:
  - `DATABASE_URL` (PostgreSQL connection)
  - `JWT_SECRET` (generated securely)
  - `PORT=8080`
  - `GIN_MODE=release`
  - `LOG_LEVEL=info`
  - `LOG_FORMAT=json`
- [x] Generated code committed to repository
- [x] Nixpacks configuration fixed (go_1_22)
- [x] Build commands configured in `services/api/nixpacks.toml`
- [x] Start command configured in `railway.json`

### ⚠️ Pending Configuration

- [ ] **Root Directory set to `services/api`** (blocks deployment)

---

## 🔧 Technical Details

### Why Root Directory Must Be Set

**The Problem:**
```
Repository root (/)
├── go.work              ← Contains workspace config
├── generated/           ← Gitignored module (committed for Railway)
├── services/
│   └── api/             ← Actual API service code
│       ├── go.mod
│       ├── cmd/server/
│       └── nixpacks.toml
```

When Railway builds from `/`:
1. It sees `go.work` file
2. Go workspace tries to load `generated/` module
3. Build commands run in wrong directory context
4. **Result:** `directory cmd/server is contained in a module that is not one of the workspace modules`

**The Solution:**

Building from `services/api/`:
1. No `go.work` file present
2. Uses local `go.mod` only
3. Build commands run in correct context
4. **Result:** Clean build ✅

### Why Configuration Files Don't Work

- `railway.json` - Does not support `rootDirectory` field
- `nixpacks.toml` - `nixpacksConfigPath` only references config, doesn't set build root
- Railway CLI - No `railway config set root` command exists
- **Only option:** Manual dashboard configuration

---

## 📝 Build Configuration Reference

### services/api/nixpacks.toml
```toml
[phases.setup]
nixPkgs = ["go_1_22", "postgresql"]

[phases.build]
cmds = [
  "go mod download",
  "go build -o ./bin/ubik-server ./cmd/server"
]

[start]
cmd = "./bin/ubik-server"

[variables]
GO111MODULE = "on"
CGO_ENABLED = "0"
```

### railway.json (project root)
```json
{
  "$schema": "https://railway.app/railway.schema.json",
  "build": {
    "builder": "NIXPACKS",
    "nixpacksConfigPath": "services/api/nixpacks.toml"
  },
  "deploy": {
    "numReplicas": 1,
    "restartPolicyType": "ON_FAILURE",
    "restartPolicyMaxRetries": 10,
    "startCommand": "./bin/ubik-server"
  }
}
```

---

## 🚀 After Setting Root Directory

Once the root directory is configured, the next deployment should:

1. **Build successfully** from `services/api/`
2. **Use committed generated code** from `generated/` directory
3. **Connect to PostgreSQL** using `DATABASE_URL`
4. **Start the API server** on port 8080
5. **Generate a domain** (e.g., `ubik-enterprise-production.up.railway.app`)

### Expected Build Output
```
[Region: asia-southeast1]
Using Nixpacks
╔════════════════════ Nixpacks v1.38.0 ═══════════════════╗
║ setup      │ go_1_22, postgresql                        ║
║─────────────────────────────────────────────────────────║
║ install    │ go mod download                            ║
║─────────────────────────────────────────────────────────║
║ build      │ go mod download                            ║
║            │ go build -o ./bin/ubik-server ./cmd/server ║
║─────────────────────────────────────────────────────────║
║ start      │ ./bin/ubik-server                          ║
╚═════════════════════════════════════════════════════════╝

Building...
✅ Build successful
Deploying...
✅ Service deployed
```

---

## 🔍 Verification Steps

After successful deployment:

### 1. Check Deployment Status
```bash
railway status
# Should show: Active deployment
```

### 2. Generate Domain
```bash
railway domain
# Or manually in dashboard: Settings → Networking → Generate Domain
```

### 3. Test API Health
```bash
curl https://your-domain.up.railway.app/health
# Expected: {"status":"ok"}
```

### 4. View Logs
```bash
railway logs
# Should show: Server started on :8080
```

---

## 📚 Related Documentation

- [RAILWAY_DEPLOYMENT.md](./RAILWAY_DEPLOYMENT.md) - Complete deployment guide
- [.env.railway.example](../.env.railway.example) - Environment variables reference
- [services/api/nixpacks.toml](../services/api/nixpacks.toml) - Build configuration

---

## 🆘 Troubleshooting

### If Build Still Fails After Setting Root Directory

1. **Check Railway dashboard shows:** `Root Directory: services/api`
2. **Verify latest commit** is deployed (check commit hash)
3. **Check build logs** for actual error
4. **Common issues:**
   - Go module cache issues → Clear Railway cache
   - Missing environment variables → Verify all vars set
   - Database connection failed → Check `DATABASE_URL`

### Get Support

- Railway Discord: https://discord.gg/railway
- GitHub Issues: https://github.com/sergei-rastrigin/ubik-enterprise/issues
- Railway Docs: https://docs.railway.app/guides/services#root-directory

---

**Last Updated:** 2025-11-05
**Status:** Awaiting manual dashboard configuration
**Next Step:** Set Root Directory to `services/api` in Railway dashboard
