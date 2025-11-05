# Railway Web Service Setup

**Next.js UI Deployment Guide**

---

## 🎯 Overview

This guide covers deploying the Next.js web UI (`services/web/`) to Railway as a separate service alongside the API service.

**Architecture:**
```
Railway Project: empowering-rejoicing
├── ubik-api (services/api/) - Go API server
├── ubik-web (services/web/) - Next.js UI [TO BE CREATED]
└── Postgres - Shared database
```

---

## 📋 Prerequisites

- ✅ Railway account authenticated
- ✅ Project linked: `empowering-rejoicing`
- ✅ API service configured (ubik-api)
- ✅ PostgreSQL database provisioned
- ✅ API domain generated (needed for web env vars)

---

## 🚀 Create Web Service

### Step 1: Create New Service

1. **Go to Railway dashboard:**
   https://railway.app/project/8eb7d506-4c2c-4ec8-9079-d4be70dd389f

2. **Click "+ New" → "Service"**

3. **Select "GitHub Repo" → `sergei-rastrigin/ubik-enterprise`**

4. **Configure Service:**
   - **Service Name:** `ubik-web`
   - **Branch:** `main`

### Step 2: Configure Root Directory

**⚠️ CRITICAL:** Set root directory to avoid build errors

1. Go to **Settings → General**
2. **Root Directory:** `services/web`
3. Click **Save**

### Step 3: Configure Build Settings

Railway will auto-detect Next.js, but verify:

1. **Settings → Build**
   - **Builder:** NIXPACKS (auto-detected)
   - **Nixpacks Config Path:** `services/web/nixpacks.toml`
   - **Build Command:** `npm run generate:api && npm run build`
   - **Start Command:** `npm start`

---

## 🔧 Environment Variables

### Step 4: Set Environment Variables

**Settings → Variables** - Add these:

#### Required Variables

```bash
# API URL (get from ubik-api service domain)
NEXT_PUBLIC_API_URL=https://ubik-api-production.up.railway.app

# Node Environment
NODE_ENV=production

# Next.js Port (Railway auto-assigns, but specify for clarity)
PORT=3000
```

#### Optional Variables (if using NextAuth)

```bash
# NextAuth Configuration
NEXTAUTH_URL=${{RAILWAY_PUBLIC_DOMAIN}}
NEXTAUTH_SECRET=<generate-with-openssl-rand-base64-32>
```

### Generate Secrets

```bash
# For NextAuth secret (if needed)
openssl rand -base64 32
```

---

## 🌐 Networking

### Step 5: Generate Domain

1. **Settings → Networking**
2. **Click "Generate Domain"**
3. **Note the URL** (e.g., `ubik-web-production.up.railway.app`)

### Step 6: Update API CORS

The API service needs to allow requests from the web domain:

1. Go to `ubik-api` service
2. **Settings → Variables**
3. Add or update:
   ```bash
   CORS_ORIGINS=https://ubik-web-production.up.railway.app
   ```
4. Redeploy API service

---

## 📦 Deployment

### Step 7: Deploy

Railway will auto-deploy when you push to `main` branch.

**Manual deploy:**
1. Go to service
2. Click **"Deploy"** button
3. Select latest commit

### Expected Build Output

```
[Region: asia-southeast1]
Using Nixpacks
╔════════════════════ Nixpacks v1.38.0 ═══════════════════╗
║ setup      │ nodejs_20                                  ║
║─────────────────────────────────────────────────────────║
║ install    │ npm ci                                     ║
║─────────────────────────────────────────────────────────║
║ build      │ npm run generate:api                       ║
║            │ npm run build                              ║
║─────────────────────────────────────────────────────────║
║ start      │ npm start                                  ║
╚═════════════════════════════════════════════════════════╝

Installing dependencies...
Generating API types from OpenAPI spec...
Building Next.js application...
✅ Build successful
Deploying...
✅ Service deployed
```

---

## ✅ Verification

### Step 8: Test Web UI

1. **Check deployment status:**
   ```bash
   railway status --service ubik-web
   ```

2. **View logs:**
   ```bash
   railway logs --service ubik-web
   ```

3. **Access web UI:**
   ```bash
   open https://ubik-web-production.up.railway.app
   ```

4. **Verify API connection:**
   - Login should work
   - Check browser console for API errors
   - Verify Network tab shows requests to API domain

---

## 🔍 Troubleshooting

### Build Fails: "Cannot find module"

**Solution:** Ensure root directory is set to `services/web`

### Build Fails: "OpenAPI spec not found"

**Cause:** `npm run generate:api` cannot find `../../shared/openapi/spec.yaml`

**Solution:**
1. Verify root directory is `services/web` (not `/`)
2. Check `shared/openapi/spec.yaml` exists in repo
3. Verify nixpacks build command includes `npm run generate:api`

### Runtime Error: "NEXT_PUBLIC_API_URL is not defined"

**Solution:** Add `NEXT_PUBLIC_API_URL` environment variable

### CORS Errors in Browser

**Solution:**
1. Verify API service has `CORS_ORIGINS` configured
2. Ensure web domain is included in CORS_ORIGINS
3. Redeploy API service after updating

### Port Already in Use

**Solution:** Railway auto-assigns ports. Ensure `PORT` env var is not hardcoded in code.

---

## 📊 Service Configuration Summary

### ubik-web Service

**General:**
- Service Name: `ubik-web`
- Root Directory: `services/web`
- Branch: `main`

**Build:**
- Builder: NIXPACKS
- Build Command: `npm run generate:api && npm run build`
- Start Command: `npm start`

**Environment:**
```env
NEXT_PUBLIC_API_URL=https://ubik-api-production.up.railway.app
NODE_ENV=production
PORT=3000
```

**Networking:**
- Public Domain: Generated
- Internal Domain: `ubik-web.railway.internal`

---

## 📚 Related Documentation

- [RAILWAY_MANUAL_SETUP.md](./RAILWAY_MANUAL_SETUP.md) - API service setup
- [RAILWAY_DEPLOYMENT.md](./RAILWAY_DEPLOYMENT.md) - Complete deployment guide
- [services/web/README.md](../services/web/README.md) - Web UI documentation

---

## 🎉 Success Checklist

- [ ] Web service created in Railway
- [ ] Root directory set to `services/web`
- [ ] Environment variables configured
- [ ] Domain generated
- [ ] API CORS updated with web domain
- [ ] Build successful
- [ ] Service running
- [ ] UI accessible via domain
- [ ] Login works
- [ ] API calls successful

---

**Last Updated:** 2025-11-05
**Status:** Ready for deployment
**Next Step:** Create web service in Railway dashboard
