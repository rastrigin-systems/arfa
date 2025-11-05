# Railway Deployment - Complete Checklist

**Project:** Ubik Enterprise
**Railway Project:** empowering-rejoicing (8eb7d506-4c2c-4ec8-9079-d4be70dd389f)
**Services:** API + Web UI + PostgreSQL

---

## 📊 Deployment Status

### ✅ Completed (Automated)

- [x] Railway CLI installed and authenticated
- [x] Project linked to repository
- [x] PostgreSQL database provisioned
- [x] API environment variables configured
- [x] Generated code committed
- [x] Nixpacks configurations created
- [x] All changes pushed to main branch
- [x] Documentation created

### ⚠️ Pending (Manual Steps Required)

- [ ] **API Service:** Set root directory
- [ ] **Web Service:** Create new service
- [ ] **Web Service:** Configure and deploy
- [ ] **Both Services:** Generate domains
- [ ] **Verification:** Test complete stack

---

## 🎯 Step-by-Step Guide

### 1. Configure API Service (ubik-api)

**Location:** https://railway.app/project/8eb7d506-4c2c-4ec8-9079-d4be70dd389f

#### Service: ubik-enterprise → ubik-api

1. **Click service** → **Settings** → **General**
2. **Root Directory:** `services/api`
3. **Save**
4. **Deploy** (will auto-deploy on save)

**Expected result:** API builds successfully from `services/api/`

**Verify:**
```bash
railway logs --service ubik-api
# Look for: "Server started on :8080"
```

**Reference:** [docs/RAILWAY_MANUAL_SETUP.md](./RAILWAY_MANUAL_SETUP.md)

---

### 2. Generate API Domain

1. **API Service** → **Settings** → **Networking**
2. **Click "Generate Domain"**
3. **Copy URL** (e.g., `https://ubik-api-production.up.railway.app`)
4. **Save for next steps**

**Test API:**
```bash
curl https://YOUR-API-DOMAIN.up.railway.app/health
# Expected: {"status":"ok"}
```

---

### 3. Create Web Service

1. **Dashboard** → **Click "+ New"** → **"Service"**
2. **Select "GitHub Repo"** → **`sergei-rastrigin/ubik-enterprise`**
3. **Service Name:** `ubik-web`
4. **Branch:** `main`
5. **Click "Add Service"**

---

### 4. Configure Web Service

#### 4a. Set Root Directory

1. **ubik-web Service** → **Settings** → **General**
2. **Root Directory:** `services/web`
3. **Save**

#### 4b. Set Environment Variables

**Settings** → **Variables** → **Add variables:**

```env
NEXT_PUBLIC_API_URL=https://YOUR-API-DOMAIN.up.railway.app
NODE_ENV=production
PORT=3000
```

**Replace `YOUR-API-DOMAIN` with actual API domain from Step 2**

#### 4c. Deploy

Service will auto-deploy. Monitor in dashboard.

**Expected build time:** 3-5 minutes

**Reference:** [docs/RAILWAY_WEB_SERVICE.md](./RAILWAY_WEB_SERVICE.md)

---

### 5. Generate Web Domain

1. **Web Service** → **Settings** → **Networking**
2. **Click "Generate Domain"**
3. **Copy URL** (e.g., `https://ubik-web-production.up.railway.app`)

---

### 6. Update API CORS

1. **API Service** → **Settings** → **Variables**
2. **Add or update:**
   ```env
   CORS_ORIGINS=https://YOUR-WEB-DOMAIN.up.railway.app
   ```
3. **Redeploy API** (automatic or manual trigger)

---

### 7. Verify Complete Stack

#### API Verification
```bash
# Health check
curl https://YOUR-API-DOMAIN/health

# Database connection
curl https://YOUR-API-DOMAIN/api/v1/health
```

#### Web UI Verification
```bash
# Open in browser
open https://YOUR-WEB-DOMAIN

# Check in browser:
1. Page loads without errors
2. Login page displays
3. No CORS errors in console
4. API requests work
```

---

## 📝 Configuration Reference

### API Service (ubik-api)

**Settings:**
- Root Directory: `services/api`
- Builder: NIXPACKS
- Nixpacks Config: `services/api/nixpacks.toml`

**Environment Variables:**
```env
DATABASE_URL=${{Postgres.DATABASE_URL}}
JWT_SECRET=<generated-secret>
PORT=8080
GIN_MODE=release
LOG_LEVEL=info
LOG_FORMAT=json
CORS_ORIGINS=https://YOUR-WEB-DOMAIN
```

### Web Service (ubik-web)

**Settings:**
- Root Directory: `services/web`
- Builder: NIXPACKS
- Nixpacks Config: `services/web/nixpacks.toml`

**Environment Variables:**
```env
NEXT_PUBLIC_API_URL=https://YOUR-API-DOMAIN
NODE_ENV=production
PORT=3000
```

### PostgreSQL Database

**Auto-configured by Railway**
- Provides: `DATABASE_URL`, `POSTGRES_HOST`, etc.
- Shared between services
- Automatically backed up

---

## 🚨 Common Issues & Solutions

### Issue: API Build Fails "go.work file"

**Solution:** Set root directory to `services/api`
**Status:** Documented in [RAILWAY_MANUAL_SETUP.md](./RAILWAY_MANUAL_SETUP.md)

### Issue: Web Build Fails "Cannot find OpenAPI spec"

**Solution:** Set root directory to `services/web`
**Status:** Documented in [RAILWAY_WEB_SERVICE.md](./RAILWAY_WEB_SERVICE.md)

### Issue: CORS Errors in Web UI

**Solution:** Update API `CORS_ORIGINS` with web domain
**Status:** Step 6 above

### Issue: Login Doesn't Work

**Check:**
1. API is running: `curl https://API-DOMAIN/health`
2. Database connected: Check API logs
3. JWT_SECRET is set in API env vars
4. Web can reach API: Check browser Network tab

---

## 📚 Documentation Index

1. **[RAILWAY_MANUAL_SETUP.md](./RAILWAY_MANUAL_SETUP.md)** - API service manual configuration
2. **[RAILWAY_WEB_SERVICE.md](./RAILWAY_WEB_SERVICE.md)** - Web service setup guide
3. **[RAILWAY_DEPLOYMENT.md](./RAILWAY_DEPLOYMENT.md)** - Complete deployment guide
4. **[.env.railway.example](../.env.railway.example)** - Environment variables reference

---

## ✅ Final Verification Checklist

### API Service
- [ ] Root directory set to `services/api`
- [ ] All environment variables configured
- [ ] Build successful
- [ ] Service running
- [ ] Domain generated
- [ ] Health endpoint responds
- [ ] CORS configured for web domain

### Web Service
- [ ] Service created
- [ ] Root directory set to `services/web`
- [ ] Environment variables configured
- [ ] Domain generated
- [ ] Build successful
- [ ] Service running
- [ ] UI accessible
- [ ] Login works
- [ ] API calls successful

### Database
- [ ] PostgreSQL provisioned
- [ ] Connected to API service
- [ ] Schema initialized
- [ ] Seed data loaded (if applicable)

---

## 🎉 Success!

Once all checklist items are complete, you'll have:

✅ **Fully deployed Ubik Enterprise platform on Railway**
- API service running at: `https://ubik-api-production.up.railway.app`
- Web UI running at: `https://ubik-web-production.up.railway.app`
- PostgreSQL database connected
- Auto-deploy on push to `main`
- Production-ready configuration

---

**Last Updated:** 2025-11-05
**Next Step:** Follow this checklist step-by-step
**Estimated Time:** 15-20 minutes
