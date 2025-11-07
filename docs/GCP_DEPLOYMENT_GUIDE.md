# Google Cloud Platform Deployment Guide

**Complete guide for deploying Ubik Enterprise to GCP Cloud Run**

---

## 🎯 Overview

This guide covers deploying both API and Web UI services to Google Cloud Platform using:
- **Cloud Run** - Serverless container platform
- **Cloud SQL** - Managed PostgreSQL database
- **Artifact Registry** - Docker image storage
- **Secret Manager** - Secure credential storage

**Estimated Setup Time:** 30-40 minutes
**Estimated Cost:** $20-30/month

---

## 📋 Prerequisites

- Google Cloud account
- gcloud CLI installed
- Docker installed
- Project repository cloned

---

## 🚀 Step-by-Step Deployment

### Step 1: Install and Authenticate GCP CLI

```bash
# Install gcloud CLI (macOS)
brew install --cask google-cloud-sdk

# Add to PATH
export PATH="/opt/homebrew/share/google-cloud-sdk/bin:$PATH"

# Authenticate
gcloud auth login

# Set application default credentials
gcloud auth application-default login
```

### Step 2: Create GCP Project

```bash
# Set project variables
export PROJECT_ID="ubik-enterprise-prod"
export REGION="us-central1"

# Create project
gcloud projects create $PROJECT_ID --name="Ubik Enterprise"

# Set default project
gcloud config set project $PROJECT_ID

# Enable billing (required - do this in console)
# https://console.cloud.google.com/billing
```

### Step 3: Enable Required APIs

```bash
# Enable all required APIs
gcloud services enable \
  cloudrun.googleapis.com \
  sqladmin.googleapis.com \
  artifactregistry.googleapis.com \
  secretmanager.googleapis.com \
  cloudbuild.googleapis.com
```

### Step 4: Create Artifact Registry Repository

```bash
# Create Docker repository
gcloud artifacts repositories create ubik-images \
  --repository-format=docker \
  --location=$REGION \
  --description="Docker images for Ubik Enterprise"

# Configure Docker authentication
gcloud auth configure-docker $REGION-docker.pkg.dev
```

### Step 5: Create Cloud SQL Database

```bash
# Create PostgreSQL instance
gcloud sql instances create ubik-postgres \
  --database-version=POSTGRES_15 \
  --tier=db-f1-micro \
  --region=$REGION \
  --root-password=<GENERATE_SECURE_PASSWORD>

# Create database
gcloud sql databases create ubik --instance=ubik-postgres

# Create user
gcloud sql users create ubik \
  --instance=ubik-postgres \
  --password=<GENERATE_SECURE_PASSWORD>

# Get connection name (save this)
gcloud sql instances describe ubik-postgres \
  --format='value(connectionName)'
```

### Step 6: Store Secrets

```bash
# Generate JWT secret
JWT_SECRET=$(openssl rand -base64 32)

# Store in Secret Manager
echo -n "$JWT_SECRET" | gcloud secrets create jwt-secret \
  --data-file=- \
  --replication-policy="automatic"

# Store database password
echo -n "<YOUR_DB_PASSWORD>" | gcloud secrets create db-password \
  --data-file=- \
  --replication-policy="automatic"
```

### Step 7: Build and Push Docker Images

**API Service:**
```bash
cd /path/to/ubik-enterprise

# Build from root (monorepo support)
docker build -t $REGION-docker.pkg.dev/$PROJECT_ID/ubik-images/api:latest \
  -f services/api/Dockerfile.gcp .

# Push to Artifact Registry
docker push $REGION-docker.pkg.dev/$PROJECT_ID/ubik-images/api:latest
```

**Web UI:**
```bash
# Build web service
docker build -t $REGION-docker.pkg.dev/$PROJECT_ID/ubik-images/web:latest \
  -f services/web/Dockerfile .

# Push to Artifact Registry
docker push $REGION-docker.pkg.dev/$PROJECT_ID/ubik-images/web:latest
```

### Step 8: Deploy API to Cloud Run

```bash
# Get Cloud SQL connection name
SQL_CONNECTION=$(gcloud sql instances describe ubik-postgres \
  --format='value(connectionName)')

# Deploy API service
gcloud run deploy ubik-api \
  --image=$REGION-docker.pkg.dev/$PROJECT_ID/ubik-images/api:latest \
  --platform=managed \
  --region=$REGION \
  --allow-unauthenticated \
  --add-cloudsql-instances=$SQL_CONNECTION \
  --set-env-vars="PORT=8080,GIN_MODE=release,LOG_LEVEL=info" \
  --set-secrets="JWT_SECRET=jwt-secret:latest,DATABASE_URL=db-password:latest" \
  --memory=512Mi \
  --cpu=1 \
  --min-instances=0 \
  --max-instances=10
```

### Step 9: Deploy Web UI to Cloud Run

```bash
# Get API URL
API_URL=$(gcloud run services describe ubik-api \
  --region=$REGION \
  --format='value(status.url)')

# Deploy web service
gcloud run deploy ubik-web \
  --image=$REGION-docker.pkg.dev/$PROJECT_ID/ubik-images/web:latest \
  --platform=managed \
  --region=$REGION \
  --allow-unauthenticated \
  --set-env-vars="NEXT_PUBLIC_API_URL=$API_URL,NODE_ENV=production,PORT=3000" \
  --memory=512Mi \
  --cpu=1 \
  --min-instances=0 \
  --max-instances=10
```

### Step 10: Initialize Database Schema

```bash
# Connect to Cloud SQL
gcloud sql connect ubik-postgres --user=postgres

# Run in psql:
\c ubik
\i /path/to/shared/schema/schema.sql
\q
```

---

## 🔧 Dockerfiles for GCP

### services/api/Dockerfile.gcp

```dockerfile
# Multi-stage build for Go API
FROM golang:1.24-alpine AS builder

WORKDIR /workspace

# Copy workspace files
COPY go.work go.work.sum ./
COPY services/api/go.mod services/api/go.sum ./services/api/
COPY services/cli/go.mod services/cli/go.sum ./services/cli/
COPY pkg/types/go.mod ./pkg/types/
COPY generated/go.mod generated/go.sum ./generated/

# Download dependencies
RUN go work sync
RUN cd services/api && go mod download

# Copy source code
COPY services/api/ ./services/api/
COPY pkg/ ./pkg/
COPY generated/ ./generated/

# Build binary
WORKDIR /workspace/services/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

### services/web/Dockerfile (already exists)

The existing `services/web/Dockerfile` should work as-is for GCP.

---

## 📊 Cost Estimation

### Monthly Costs (Approximate)

| Service | Usage | Cost |
|---------|-------|------|
| Cloud Run API | 1M requests, 512MB, 1 CPU | $5-10 |
| Cloud Run Web | 1M requests, 512MB, 1 CPU | $5-10 |
| Cloud SQL (db-f1-micro) | Always on | $10 |
| Artifact Registry | 10 GB storage | $1 |
| Secret Manager | 10 secrets | $0.06 |
| **Total** | | **~$21-31/month** |

### Free Tier

Cloud Run includes:
- 2 million requests/month free
- 360,000 GB-seconds free
- 180,000 vCPU-seconds free

---

## 🔍 Verification

### Test API Health

```bash
API_URL=$(gcloud run services describe ubik-api \
  --region=$REGION \
  --format='value(status.url)')

curl $API_URL/health
# Expected: {"status":"ok"}
```

### Test Web UI

```bash
WEB_URL=$(gcloud run services describe ubik-web \
  --region=$REGION \
  --format='value(status.url)')

open $WEB_URL
```

### View Logs

```bash
# API logs
gcloud run services logs read ubik-api --region=$REGION

# Web logs
gcloud run services logs read ubik-web --region=$REGION
```

---

## 🔄 CI/CD with Cloud Build

Create `cloudbuild.yaml`:

```yaml
steps:
  # Build API
  - name: 'gcr.io/cloud-builders/docker'
    args: ['build', '-t', '$_REGION-docker.pkg.dev/$PROJECT_ID/ubik-images/api:$SHORT_SHA',
           '-f', 'services/api/Dockerfile.gcp', '.']

  # Build Web
  - name: 'gcr.io/cloud-builders/docker'
    args: ['build', '-t', '$_REGION-docker.pkg.dev/$PROJECT_ID/ubik-images/web:$SHORT_SHA',
           '-f', 'services/web/Dockerfile', '.']

  # Push images
  - name: 'gcr.io/cloud-builders/docker'
    args: ['push', '$_REGION-docker.pkg.dev/$PROJECT_ID/ubik-images/api:$SHORT_SHA']

  - name: 'gcr.io/cloud-builders/docker'
    args: ['push', '$_REGION-docker.pkg.dev/$PROJECT_ID/ubik-images/web:$SHORT_SHA']

  # Deploy API
  - name: 'gcr.io/cloud-builders/gcloud'
    args: ['run', 'deploy', 'ubik-api',
           '--image', '$_REGION-docker.pkg.dev/$PROJECT_ID/ubik-images/api:$SHORT_SHA',
           '--region', '$_REGION']

  # Deploy Web
  - name: 'gcr.io/cloud-builders/gcloud'
    args: ['run', 'deploy', 'ubik-web',
           '--image', '$_REGION-docker.pkg.dev/$PROJECT_ID/ubik-images/web:$SHORT_SHA',
           '--region', '$_REGION']

substitutions:
  _REGION: us-central1

options:
  machineType: 'N1_HIGHCPU_8'
```

---

## 🚨 Troubleshooting

### Build Fails: "Cannot find generated/"

**Solution:** Build from repository root, not services/api/

```bash
# ✅ Correct
docker build -f services/api/Dockerfile.gcp .

# ❌ Wrong
cd services/api && docker build -f Dockerfile.gcp .
```

### Cloud Run Error: "Container failed to start"

**Check logs:**
```bash
gcloud run services logs read ubik-api --region=$REGION --limit=50
```

### Database Connection Failed

**Verify Cloud SQL connector:**
```bash
# Check instance is running
gcloud sql instances list

# Test connection
gcloud sql connect ubik-postgres --user=ubik
```

---

## 📚 Additional Resources

- [Cloud Run Documentation](https://cloud.google.com/run/docs)
- [Cloud SQL Documentation](https://cloud.google.com/sql/docs)
- [Artifact Registry](https://cloud.google.com/artifact-registry/docs)
- [GCP Pricing Calculator](https://cloud.google.com/products/calculator)

---

**Last Updated:** 2025-11-05
**Status:** Ready for deployment
**Next Step:** Follow step-by-step guide above
