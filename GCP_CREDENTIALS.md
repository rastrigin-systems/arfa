# Ubik Enterprise - GCP Deployment Credentials

**⚠️ SENSITIVE - Keep Secure - Do Not Commit to Git**

---

## Project Information

- **Project ID**: `ubik-enterprise-prod`
- **Project Number**: `754414213269`
- **Region**: `us-central1`
- **Google Account**: `srastrigin@gmail.com`

---

## Cloud SQL Database

- **Instance Name**: `ubik-postgres`
- **Connection Name**: `ubik-enterprise-prod:us-central1:ubik-postgres`
- **Database**: `ubik`
- **User**: `ubik`
- **Password**: `HPR7WCIsQu30mxyTcKKLOppY+TPpe7Tx`
- **Public IP**: `35.222.140.138`

**Connection String (Cloud Run):**
```
postgres://ubik:HPR7WCIsQu30mxyTcKKLOppY+TPpe7Tx@/ubik?host=/cloudsql/ubik-enterprise-prod:us-central1:ubik-postgres
```

**Connection String (Local with Cloud SQL Proxy):**
```
postgres://ubik:HPR7WCIsQu30mxyTcKKLOppY+TPpe7Tx@localhost:5432/ubik
```

---

## Secret Manager Secrets

### JWT Secret
- **Secret Name**: `jwt-secret`
- **Value**: `QXvIRN7/w6fzNqRZS/dLDKaUNCb1atyU2ZCfJoTMzPk=`

### Database Password
- **Secret Name**: `db-password`
- **Value**: `HPR7WCIsQu30mxyTcKKLOppY+TPpe7Tx`

### Database URL
- **Secret Name**: `database-url`
- **Value**: `postgres://ubik:HPR7WCIsQu30mxyTcKKLOppY+TPpe7Tx@/ubik?host=/cloudsql/ubik-enterprise-prod:us-central1:ubik-postgres`

---

## Cloud Run Services

### API Service
- **Service Name**: `ubik-api`
- **URL**: https://ubik-api-754414213269.us-central1.run.app
- **Health Check**: https://ubik-api-754414213269.us-central1.run.app/api/v1/health
- **Docker Image**: `us-central1-docker.pkg.dev/ubik-enterprise-prod/ubik-images/api:latest`

---

## Artifact Registry

- **Repository**: `ubik-images`
- **Location**: `us-central1`
- **Full Path**: `us-central1-docker.pkg.dev/ubik-enterprise-prod/ubik-images`

---

## GitHub Actions Service Account

- **Service Account Email**: `github-actions@ubik-enterprise-prod.iam.gserviceaccount.com`
- **Key File**: `github-actions-key.json` (⚠️ LOCAL ONLY - DO NOT COMMIT)
- **Roles**:
  - `roles/cloudbuild.builds.builder` - Build Docker images
  - `roles/run.admin` - Deploy to Cloud Run
  - `roles/iam.serviceAccountUser` - Act as service accounts
  - `roles/artifactregistry.writer` - Push to Artifact Registry

**GitHub Secret Setup:**
1. Go to https://github.com/sergei-rastrigin/ubik-enterprise/settings/secrets/actions
2. Create secret named `GCP_SA_KEY`
3. Value: Entire contents of `github-actions-key.json`

---

## Important Commands

### Connect to Cloud SQL (via Cloud SQL Proxy)
```bash
cloud_sql_proxy -instances=ubik-enterprise-prod:us-central1:ubik-postgres=tcp:5432
```

### View Cloud Run Logs
```bash
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=ubik-api" \
  --limit=50 --project=ubik-enterprise-prod
```

### Deploy New Version
```bash
# Build new image
gcloud builds submit --config=cloudbuild.yaml --project=ubik-enterprise-prod .

# Or redeploy existing image
gcloud run deploy ubik-api \
  --image=us-central1-docker.pkg.dev/ubik-enterprise-prod/ubik-images/api:latest \
  --region=us-central1 \
  --project=ubik-enterprise-prod
```

### Access Cloud SQL Database
```bash
gcloud sql connect ubik-postgres --user=ubik --project=ubik-enterprise-prod
# Then enter password: HPR7WCIsQu30mxyTcKKLOppY+TPpe7Tx
```

---

## Cost Estimate

**Monthly (Estimated):**
- Cloud Run API: $5-10 (2M requests free tier)
- Cloud SQL (db-f1-micro): $10
- Artifact Registry: $1 (10 GB storage)
- Secret Manager: $0.06
- **Total**: ~$16-21/month

---

**Last Updated**: 2025-11-05
**Deployment Status**: ✅ API Deployed Successfully
**Next Steps**: Initialize database schema, deploy Web UI
