# CI/CD Setup Guide

## Overview

Automated deployment pipeline menggunakan Google Cloud Build yang trigger saat push ke GitHub repository.

**Workflow:**
1. Push code ke GitHub `main` branch
2. Cloud Build automatically triggered
3. Run tests (`go test ./...`)
4. Build all 6 services (parallel)
5. Push images to Artifact Registry
6. Deploy to Cloud Run (rolling update)
7. Services auto-updated with zero downtime

---

## Prerequisites

1. GitHub repository: `https://github.com/raflibima25/event-ticketing-platform`
2. Google Cloud project: `project-4aa96947-a91f-4413-a51`
3. Artifact Registry repository: `event-ticketing`
4. Cloud Run services already deployed

---

## Setup Instructions

### Step 1: Connect GitHub to Cloud Build

```bash
# 1. Enable Cloud Build API
gcloud services enable cloudbuild.googleapis.com

# 2. Connect GitHub repository (interactive)
# This will open a browser to authenticate with GitHub
gcloud alpha builds connections create github \
  --region=asia-southeast2 \
  github-connection

# Follow the prompts to:
# - Authorize Google Cloud Build
# - Select repository: raflibima25/event-ticketing-platform
```

**Alternative (Manual via Console):**
1. Go to: https://console.cloud.google.com/cloud-build/triggers
2. Click "Connect Repository"
3. Select "GitHub (Cloud Build GitHub App)"
4. Authenticate and select `raflibima25/event-ticketing-platform`

---

### Step 2: Create Cloud Build Trigger

```bash
# Create trigger for main branch
gcloud builds triggers create github \
  --name="deploy-production" \
  --repo-name="event-ticketing-platform" \
  --repo-owner="raflibima25" \
  --branch-pattern="^main$" \
  --build-config="backend/cloudbuild.yaml" \
  --description="Auto-deploy all services on push to main" \
  --region=asia-southeast2
```

**Verify trigger created:**
```bash
gcloud builds triggers list --region=asia-southeast2
```

---

### Step 3: Grant Cloud Build Permissions

Cloud Build needs permissions to deploy to Cloud Run:

```bash
# Get Cloud Build service account email
PROJECT_NUMBER=$(gcloud projects describe project-4aa96947-a91f-4413-a51 --format="value(projectNumber)")
CLOUD_BUILD_SA="${PROJECT_NUMBER}@cloudbuild.gserviceaccount.com"

# Grant Cloud Run Admin role
gcloud projects add-iam-policy-binding project-4aa96947-a91f-4413-a51 \
  --member="serviceAccount:${CLOUD_BUILD_SA}" \
  --role="roles/run.admin"

# Grant Service Account User role (needed to deploy with service account)
gcloud projects add-iam-policy-binding project-4aa96947-a91f-4413-a51 \
  --member="serviceAccount:${CLOUD_BUILD_SA}" \
  --role="roles/iam.serviceAccountUser"

# Grant Artifact Registry Writer (to push images)
gcloud projects add-iam-policy-binding project-4aa96947-a91f-4413-a51 \
  --member="serviceAccount:${CLOUD_BUILD_SA}" \
  --role="roles/artifactregistry.writer"
```

---

## Testing the Pipeline

### Test 1: Manual Trigger (Test Build Without Code Change)

```bash
# Trigger build manually
gcloud builds triggers run deploy-production \
  --branch=main \
  --region=asia-southeast2
```

### Test 2: Push to GitHub

```bash
# Make a small change
cd /path/to/event-ticketing-platform/backend
echo "# CI/CD Pipeline Enabled" >> README.md

# Commit and push
git add .
git commit -m "test: trigger CI/CD pipeline"
git push origin main
```

**Monitor build:**
```bash
# View build logs
gcloud builds list --limit=5

# Get specific build log
BUILD_ID="<build-id-from-list>"
gcloud builds log $BUILD_ID --region=asia-southeast2
```

**Or view in console:**
https://console.cloud.google.com/cloud-build/builds

---

## Build Configuration

### File: `backend/cloudbuild.yaml`

**Key sections:**

1. **Tests** - Run first, build fails if tests fail
   ```yaml
   - name: 'golang:1.25-alpine'
     entrypoint: 'sh'
     args: ['go test ./... -v -cover']
   ```

2. **Parallel Builds** - All 6 services build simultaneously
   ```yaml
   waitFor: ['run-tests']  # All wait for tests to pass
   ```

3. **Tagging** - Each image tagged with:
   - `$COMMIT_SHA` - Specific version
   - `latest` - Always points to latest deployment

4. **Zero Downtime** - Cloud Run rolling update (default)

---

## Environment Variables & Secrets

**Important:** Environment variables **are NOT** included in cloudbuild.yaml.

Cloud Build preserves existing Cloud Run environment variables during deployment.

**To update environment variables:**

```bash
# Update via gcloud (manual)
gcloud run services update SERVICE_NAME \
  --update-env-vars="KEY=VALUE" \
  --region=asia-southeast2
```

**Current services use these secrets:**
- `DB_PASSWORD` - PostgreSQL password
- `JWT_SECRET` - JWT signing key
- `XENDIT_API_KEY` - Xendit payment API
- `XENDIT_WEBHOOK_TOKEN` - Webhook verification
- `RESEND_API_KEY` - Email service
- `UPSTASH_REDIS_REST_TOKEN` - Redis cache

These are **automatically preserved** during CI/CD deployment.

---

## Troubleshooting

### Build Failing?

**Check logs:**
```bash
gcloud builds list --limit=1
gcloud builds log <BUILD_ID>
```

**Common issues:**

1. **Tests failing:**
   - Check test output in logs
   - Tests run without external dependencies (Redis/DB are skipped)

2. **Permission denied:**
   - Verify Cloud Build service account has correct IAM roles
   - Run permission grant commands above

3. **Image push failed:**
   - Verify Artifact Registry exists: `gcloud artifacts repositories list`
   - Check Cloud Build has `artifactregistry.writer` role

4. **Deployment failed:**
   - Check Cloud Run service exists
   - Verify `run.admin` role granted to Cloud Build

### Build Too Slow?

Current timeout: 30 minutes (1800s)

**Optimize:**
- Use `machineType: E2_HIGHCPU_8` (already configured)
- Enable concurrent builds (Google Cloud default)
- Use Docker layer caching (future improvement)

---

## Rollback

If deployment causes issues, rollback to previous revision:

```bash
# List revisions
gcloud run revisions list --service=SERVICE_NAME --region=asia-southeast2

# Rollback to previous revision
gcloud run services update-traffic SERVICE_NAME \
  --to-revisions=REVISION_NAME=100 \
  --region=asia-southeast2
```

---

## Monitoring Builds

### Email Notifications

Cloud Build can send email notifications:

```bash
# Configure via Cloud Console (recommended)
# Go to: Cloud Build > Settings > Notifications
```

### Slack Notifications (Optional)

Integrate with Slack using Cloud Pub/Sub:
https://cloud.google.com/build/docs/configuring-notifications/notifiers

---

## Cost Estimation

**Cloud Build pricing:**
- First 120 build-minutes/day: **FREE**
- After: $0.003 per build-minute

**Typical build time:** ~10-15 minutes
**Estimated monthly cost:** ~$5-10 (assuming 2-3 builds/day)

**Free tier covers:**
- ~4 builds per day
- Or ~120 builds per month

---

## Next Steps

1. ✅ Setup Cloud Build trigger
2. ✅ Test with manual trigger
3. ✅ Make small commit to test auto-trigger
4. 🔜 Add linting step (golangci-lint)
5. 🔜 Add integration tests
6. 🔜 Setup staging environment

---

## Advanced: Staging Environment (Future)

To add staging environment:

1. Create new trigger for `staging` branch:
   ```bash
   gcloud builds triggers create github \
     --name="deploy-staging" \
     --branch-pattern="^staging$" \
     --build-config="backend/cloudbuild-staging.yaml"
   ```

2. Create separate Cloud Run services with `-staging` suffix

3. Update `cloudbuild-staging.yaml` to deploy to staging services

---

## Support

**Build logs:** https://console.cloud.google.com/cloud-build/builds
**Cloud Run logs:** https://console.cloud.google.com/run
**Documentation:** https://cloud.google.com/build/docs
