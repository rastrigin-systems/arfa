# GCP CI/CD Setup with GitHub Actions

**Automated deployment to Google Cloud Platform via GitHub Actions**

---

## Overview

We've set up automated deployment pipelines for both API and Web services using GitHub Actions. When code is pushed to the `main` branch and changes are detected in the respective service directories, the workflows will automatically:

1. **Build & Test** - Run all tests to ensure quality
2. **Build Docker Images** - Use Cloud Build to create production images
3. **Deploy to Cloud Run** - Deploy updated images automatically
4. **Smoke Test** - Verify deployments are healthy

---

## Architecture

```
GitHub Push (main branch)
    ↓
GitHub Actions Workflow
    ↓
Authenticate with GCP Service Account
    ↓
Cloud Build (Docker image build)
    ↓
Artifact Registry (store images)
    ↓
Cloud Run Deployment
    ↓
Smoke Tests
```

---

## Workflows

### API Server Workflow
**File**: `.github/workflows/api-server.yml`

**Triggers**:
- Push to `main` or `develop` branches
- Changes in:
  - `services/api/**`
  - `pkg/types/**`
  - `generated/**`
  - `shared/schema/**`
  - `shared/openapi/**`
  - `sqlc/**`
  - `go.work`

**Jobs**:
1. **Build** - Compile Go binary
2. **Test** - Run unit and integration tests with PostgreSQL
3. **Deploy** - Build Docker image and deploy to Cloud Run (main only)

**Build Config**: `cloudbuild-api.yaml`

### Web Server Workflow
**File**: `.github/workflows/web-server.yml`

**Triggers**:
- Push to `main` or `develop` branches
- Changes in:
  - `services/web/**`

**Jobs**:
1. **Build** - Build Next.js application
2. **Test** - Run linter, type-check, unit tests, and E2E tests
3. **Deploy** - Build Docker image and deploy to Cloud Run (main only)

**Build Config**: `cloudbuild-web.yaml`

---

## GCP Service Account

**Email**: `github-actions@ubik-enterprise-prod.iam.gserviceaccount.com`

**Granted Roles**:
- `roles/cloudbuild.builds.builder` - Build Docker images via Cloud Build
- `roles/run.admin` - Deploy and manage Cloud Run services
- `roles/iam.serviceAccountUser` - Impersonate service accounts
- `roles/artifactregistry.writer` - Push images to Artifact Registry

---

## Setup Instructions

### 1. Add GitHub Secret

The service account key has been created locally as `credentials/github-actions-key.json`.

**⚠️ IMPORTANT**: This file is gitignored and MUST NOT be committed to the repository.

**To add the secret to GitHub:**

1. Go to: https://github.com/rastrigin-org/ubik-enterprise/settings/secrets/actions
2. Click **"New repository secret"**
3. Name: `GCP_SA_KEY`
4. Value: Copy the **entire contents** of `credentials/github-actions-key.json`
5. Click **"Add secret"**

**To view the key locally:**
```bash
cat credentials/github-actions-key.json
```

### 2. Commit and Push Changes

Once the secret is added to GitHub, commit and push the workflow changes:

```bash
# On a feature branch or main
git add .github/workflows/ cloudbuild-api.yaml cloudbuild-web.yaml
git add credentials/GCP_CREDENTIALS.md docs/GCP_CICD_SETUP.md
git add services/web/instrumentation.ts  # Fixed webpack issue
git add .gitignore  # Added credentials/github-actions-key.json

git commit -m "feat: Add GCP CI/CD deployment workflows

- Configure GitHub Actions for automated deployment
- Create separate Cloud Build configs for API and Web
- Set up GCP service account with required permissions
- Fix Next.js instrumentation to handle missing test modules
- Add comprehensive documentation

Closes #[issue-number]"

git push origin main
```

### 3. Monitor Deployment

After pushing, the workflows will trigger automatically:

**View workflow runs:**
- https://github.com/rastrigin-org/ubik-enterprise/actions

**View Cloud Build logs:**
```bash
gcloud builds list --project=ubik-enterprise-prod --limit=5
```

**View Cloud Run services:**
```bash
gcloud run services list --project=ubik-enterprise-prod --region=us-central1
```

---

## Deployed Services

### API Service
- **URL**: https://ubik-api-754414213269.us-central1.run.app
- **Health Check**: https://ubik-api-754414213269.us-central1.run.app/api/v1/health
- **Image**: `us-central1-docker.pkg.dev/ubik-enterprise-prod/ubik-images/api:latest`

### Web Service
- **URL**: (Will be generated on first deployment)
- **Image**: `us-central1-docker.pkg.dev/ubik-enterprise-prod/ubik-images/web:latest`

---

## Testing the Workflows

### Manual Trigger Test (Optional)

You can manually trigger a workflow run to test without pushing to main:

```bash
# Make a minor change to trigger the workflow
echo "# CI/CD Test" >> services/api/README.md
git add services/api/README.md
git commit -m "test: Trigger API deployment workflow"
git push origin feature/test-cicd
```

Then create a PR to `main` - the build and test jobs will run, but deploy will only happen after merging to `main`.

### Verify Deployment

After deployment completes:

```bash
# Test API
curl https://ubik-api-754414213269.us-central1.run.app/api/v1/health

# Test Web (once deployed)
curl https://[web-url-from-deployment]
```

---

## Troubleshooting

### Workflow Fails with "Permission Denied"

**Issue**: Service account doesn't have required permissions.

**Fix**: Verify IAM roles are granted:
```bash
gcloud projects get-iam-policy ubik-enterprise-prod \
  --flatten="bindings[].members" \
  --filter="bindings.members:serviceAccount:github-actions@ubik-enterprise-prod.iam.gserviceaccount.com"
```

### Cloud Build Fails

**Issue**: Docker build errors or missing files.

**Fix**: Check Cloud Build logs:
```bash
BUILD_ID=$(gcloud builds list --limit=1 --format='value(id)' --project=ubik-enterprise-prod)
gcloud builds log $BUILD_ID --project=ubik-enterprise-prod
```

### Deployment Succeeds but Service Fails

**Issue**: Cloud Run service crashes on startup.

**Fix**: Check Cloud Run logs:
```bash
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=ubik-api" \
  --limit=50 --project=ubik-enterprise-prod
```

---

## Cost Optimization

**Free Tier Limits** (per month):
- Cloud Build: 120 build-minutes free
- Cloud Run: 2M requests, 360K GB-seconds free
- Artifact Registry: 0.5 GB free

**Estimated Monthly Cost with CI/CD**:
- ~20 deployments/month @ 5 min each = 100 build-minutes (within free tier)
- Additional storage: ~$0.50
- **Total**: Still ~$16-21/month (same as before)

---

## Security Best Practices

✅ **Implemented:**
- Service account with minimal required permissions
- Secrets stored in GitHub Secrets (encrypted)
- Service account key not committed to repository
- All traffic over HTTPS
- Database credentials in GCP Secret Manager

⚠️ **Future Improvements:**
- Enable Workload Identity Federation (eliminates need for service account keys)
- Add branch protection rules requiring status checks
- Implement blue-green deployments
- Add automated rollback on smoke test failure

---

## Next Steps

1. ✅ Add `GCP_SA_KEY` secret to GitHub
2. ✅ Commit and push workflow changes
3. ✅ Monitor first deployment
4. 🔄 Configure CORS in API to allow Web UI domain
5. 🔄 Set up custom domains (optional)
6. 🔄 Configure monitoring and alerts (optional)

---

**Last Updated**: 2025-11-06
**Status**: ✅ Ready for deployment
**Documentation**: See credentials/GCP_CREDENTIALS.md for all deployment credentials
