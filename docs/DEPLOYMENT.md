# Deployment Guide - run-tbot

Complete step-by-step guide for deploying the run-tbot Telegram bot to Google Cloud Run.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Local Development Setup](#local-development-setup)
3. [Google Cloud Platform Setup](#google-cloud-platform-setup)
4. [GitHub Repository Setup](#github-repository-setup)
5. [First Deployment](#first-deployment)
6. [Webhook Registration](#webhook-registration)
7. [Monitoring and Logs](#monitoring-and-logs)
8. [Troubleshooting](#troubleshooting)
9. [Cost Optimization](#cost-optimization)

---

## Prerequisites

### Required Accounts

1. **Telegram Account**
   - Mobile phone number
   - Telegram app installed

2. **Google Cloud Platform Account**
   - Credit card required (for verification)
   - Free tier: $300 credit for 90 days
   - Free tier includes: 2M Cloud Run requests/month

3. **GitHub Account**
   - Free account sufficient
   - Repository can be public or private

### Required Tools (Local Development)

- **Go 1.25+**: [Download](https://go.dev/dl/)
- **Git**: [Download](https://git-scm.com/downloads)
- **gcloud CLI**: [Install](https://cloud.google.com/sdk/docs/install)
- **ngrok** (for local testing): [Download](https://ngrok.com/download)

### Getting Your Telegram Bot Token

1. Open Telegram and search for `@BotFather`
2. Send `/newbot` command
3. Follow the prompts:
   - Choose bot name (display name): e.g., "My Test Bot"
   - Choose bot username (must end in 'bot'): e.g., "my_test_run_bot"
4. **Save the token** - looks like: `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`
5. **IMPORTANT**: Never share this token publicly!

### Getting Your Telegram User ID

You need your user ID for the `ALLOWED_USERS` environment variable.

**Method 1: Using @userinfobot**
1. Open Telegram and search for `@userinfobot`
2. Send `/start`
3. Copy your ID (will be a number like `123456789`)

**Method 2: Using @raw_data_bot**
1. Search for `@raw_data_bot`
2. Send any message
3. Look for `"from": {"id": 123456789}`

---

## Local Development Setup

### 1. Clone and Setup Project

```bash
# Clone repository
git clone https://github.com/yourusername/run-tbot.git
cd run-tbot

# Create .env file from example
cp .env.example .env

# Edit .env with your values
nano .env
```

**`.env` file contents:**
```bash
BOT_TOKEN=123456789:ABCdefGHIjklMNOpqrsTUVwxyz
PORT=8080
ENVIRONMENT=development
ALLOWED_USERS=123456789,987654321
```

### 2. Download Dependencies

```bash
go mod download
go mod verify
```

### 3. Run Tests

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 4. Run Locally (Polling Mode)

For initial testing, you can run the bot in polling mode without webhook.

**Option A: Using webhook with ngrok (recommended)**

```bash
# Terminal 1: Start ngrok
ngrok http 8080

# Copy the HTTPS URL (e.g., https://abc123.ngrok.io)

# Terminal 2: Set webhook URL and run bot
export WEBHOOK_URL=https://abc123.ngrok.io/webhook
go run .
```

Then test by sending a message to your bot in Telegram.

**Option B: Disable webhook for local testing**

Modify `main.go` temporarily to skip webhook setup, or use polling mode (not included in current version).

---

## Google Cloud Platform Setup

### 1. Create GCP Project

```bash
# Authenticate with Google Cloud
gcloud auth login

# Set your project ID (must be globally unique)
PROJECT_ID="run-tbot-12345"

# Create new project
gcloud projects create $PROJECT_ID --name="Telegram Bot"

# Set as default project
gcloud config set project $PROJECT_ID

# Link billing account (required for Cloud Run)
# List billing accounts
gcloud billing accounts list

# Link billing account to project
gcloud billing projects link $PROJECT_ID \
  --billing-account=ABCDEF-123456-789012
```

### 2. Enable Required APIs

```bash
# Enable Cloud Run API
gcloud services enable run.googleapis.com

# Enable Cloud Build API (for container builds)
gcloud services enable cloudbuild.googleapis.com

# Enable Container Registry API
gcloud services enable containerregistry.googleapis.com

# Verify APIs are enabled
gcloud services list --enabled
```

### 3. Create Service Account for GitHub Actions

This service account will be used by GitHub Actions to deploy to Cloud Run.

```bash
# Set variables
SA_NAME="github-actions-deployer"
SA_EMAIL="${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"

# Create service account
gcloud iam service-accounts create $SA_NAME \
  --display-name="GitHub Actions Deployer" \
  --project=$PROJECT_ID

# Grant Cloud Run Admin role
# Allows creating, updating, deleting Cloud Run services
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/run.admin"

# Grant Service Account User role
# Allows acting as the Cloud Run runtime service account
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/iam.serviceAccountUser"

# Grant Storage Admin role (for container images)
# Allows pushing Docker images to Container Registry
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/storage.admin"

# Create JSON key for authentication
gcloud iam service-accounts keys create key.json \
  --iam-account=$SA_EMAIL

# Display key (you'll copy this to GitHub Secrets)
cat key.json

# IMPORTANT: Delete key file after copying to GitHub Secrets!
# Store it securely if you need a backup
rm key.json
```

### 4. Choose Cloud Run Region

```bash
# List available regions
gcloud run regions list

# Recommended regions for free tier:
# - us-central1 (Iowa)
# - us-east1 (South Carolina)
# - us-west1 (Oregon)

# Set default region
gcloud config set run/region us-central1
```

---

## GitHub Repository Setup

### 1. Create GitHub Repository

1. Go to https://github.com/new
2. Repository name: `run-tbot`
3. Choose Public or Private
4. **Don't** initialize with README (you already have one)
5. Click "Create repository"

### 2. Push Code to GitHub

```bash
# Add remote
git remote add origin https://github.com/yourusername/run-tbot.git

# Push code
git push -u origin main
```

### 3. Configure GitHub Secrets

Go to repository Settings → Secrets and variables → Actions → New repository secret

Add these secrets:

| Secret Name | Value | How to Get |
|-------------|-------|------------|
| `BOT_TOKEN` | `123456789:ABC...` | From @BotFather |
| `ALLOWED_USERS` | `123456789,987654321` | From @userinfobot (comma-separated) |
| `GCP_SA_KEY` | `{"type": "service_account"...}` | Contents of `key.json` from step 3 above |
| `GCP_REGION` | `us-central1` | Your chosen region |

**Important**:
- `GCP_SA_KEY` should be the **entire JSON file contents**
- Copy the JSON exactly as-is, including all braces and quotes
- Don't add extra formatting or line breaks

### 4. Verify Secrets

After adding secrets, you should see:
- ✅ BOT_TOKEN
- ✅ ALLOWED_USERS
- ✅ GCP_SA_KEY
- ✅ GCP_REGION

---

## First Deployment

### Understanding the Deployment Flow

```
Push to main → CI runs → Deploy triggers → Cloud Run updates
```

1. **CI Workflow** (`.github/workflows/ci.yml`):
   - Runs tests
   - Runs linters
   - Verifies build
   - Takes ~2-3 minutes

2. **Deploy Workflow** (`.github/workflows/deploy.yml`):
   - Only runs if CI succeeds
   - Builds Docker image
   - Pushes to GitHub Container Registry
   - Deploys to Cloud Run
   - Takes ~3-5 minutes

### Trigger First Deployment

```bash
# Make sure all files are committed
git status

# If go.mod and go.sum are not committed yet
git add go.mod go.sum
git commit -m "Add Go module files"

# Push to main branch
git push origin main
```

### Monitor Deployment

1. Go to your GitHub repository
2. Click "Actions" tab
3. You should see two workflows running:
   - ✅ CI (runs first)
   - ⏳ Deploy to Cloud Run (waits for CI)

4. Click on "Deploy to Cloud Run" to see progress
5. Wait for all steps to complete (~5-8 minutes total)

### Get Your Service URL

After successful deployment, the workflow will display:

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Deployment successful!
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🌐 Service URL: https://run-tbot-xyz123-uc.a.run.app
🔗 Webhook URL: https://run-tbot-xyz123-uc.a.run.app/webhook
```

**Save this URL** - you'll need it for webhook registration!

### Verify Deployment

```bash
# Test health check endpoint
curl https://run-tbot-xyz123-uc.a.run.app/

# Should return:
# {"status":"ok","service":"run-tbot"}
```

---

## Webhook Registration

After deployment, you need to tell Telegram where to send updates.

### Method 1: Using curl (Recommended)

```bash
# Set your bot token
BOT_TOKEN="123456789:ABCdefGHIjklMNOpqrsTUVwxyz"

# Set your service URL (from deployment output)
SERVICE_URL="https://run-tbot-xyz123-uc.a.run.app"

# Register webhook
curl -X POST "https://api.telegram.org/bot${BOT_TOKEN}/setWebhook" \
  -H "Content-Type: application/json" \
  -d "{\"url\": \"${SERVICE_URL}/webhook\"}"

# Expected response:
# {"ok":true,"result":true,"description":"Webhook was set"}
```

### Method 2: Using Browser

Open this URL in your browser (replace with your values):

```
https://api.telegram.org/bot123456789:ABCdefGHIjklMNOpqrsTUVwxyz/setWebhook?url=https://run-tbot-xyz123-uc.a.run.app/webhook
```

### Verify Webhook

```bash
# Check webhook status
curl "https://api.telegram.org/bot${BOT_TOKEN}/getWebhookInfo"

# Should show:
# {
#   "ok": true,
#   "result": {
#     "url": "https://run-tbot-xyz123-uc.a.run.app/webhook",
#     "has_custom_certificate": false,
#     "pending_update_count": 0,
#     "max_connections": 40
#   }
# }
```

### Delete Webhook (if needed)

```bash
# Remove webhook (useful for local testing)
curl "https://api.telegram.org/bot${BOT_TOKEN}/deleteWebhook"
```

---

## Testing Your Bot

### 1. Find Your Bot in Telegram

Search for your bot username (e.g., `@my_test_run_bot`)

### 2. Test Commands

**Currently implemented:**
- Just the dice button (from `bot/bot.go`)

**After implementing handlers (Phase 2-3):**
- `/start` - Welcome message with dice button
- `/help` - Show available commands
- Click "🎲 Roll Dice" - Get random number 1-6

### 3. Test Authorization

If your user ID is in `ALLOWED_USERS`:
- You should see private commands in `/help`

If not in `ALLOWED_USERS`:
- You'll only see public commands

---

## Monitoring and Logs

### View Cloud Run Logs

**Method 1: GCP Console**
1. Go to https://console.cloud.google.com/run
2. Click on `run-tbot` service
3. Click "Logs" tab
4. See real-time logs

**Method 2: gcloud CLI**
```bash
# Stream logs in real-time
gcloud run services logs read run-tbot \
  --region=us-central1 \
  --limit=50 \
  --follow

# Search for errors
gcloud run services logs read run-tbot \
  --region=us-central1 \
  --filter="severity=ERROR" \
  --limit=100
```

### Understanding Logs

Logs use structured JSON format (from `slog`):

```json
{
  "time": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "msg": "Server started",
  "port": "8080",
  "environment": "production"
}
```

**Common log levels:**
- `INFO` - Normal operation (startup, requests)
- `WARN` - Non-critical issues
- `ERROR` - Errors that need attention

### View Metrics

1. Go to Cloud Run service page
2. Click "Metrics" tab
3. View:
   - Request count
   - Request latency
   - Container instances
   - CPU utilization
   - Memory utilization

---

## Troubleshooting

### Bot Doesn't Respond to Messages

**Check 1: Webhook is set**
```bash
curl "https://api.telegram.org/bot${BOT_TOKEN}/getWebhookInfo"
```
- Verify `url` matches your Cloud Run service URL
- Check `pending_update_count` - should be 0 or low

**Check 2: Service is running**
```bash
curl https://your-service-url.run.app/
```
- Should return `{"status":"ok"}`

**Check 3: View logs**
```bash
gcloud run services logs read run-tbot --region=us-central1 --limit=50
```
- Look for errors or webhook requests

**Check 4: Webhook endpoint is correct**
- URL should be: `https://your-service.run.app/webhook`
- Must be HTTPS (not HTTP)
- Must end with `/webhook`

### CI Workflow Fails

**Test failures:**
```bash
# Run tests locally
go test -v ./...

# Fix failing tests
# Commit and push
```

**Linting failures:**
```bash
# Check formatting
gofmt -l -s .

# Fix formatting
gofmt -w -s .

# Run go vet
go vet ./...
```

### Deploy Workflow Fails

**Authentication errors:**
- Verify `GCP_SA_KEY` secret is correct
- Check service account has required permissions

**Image push errors:**
- Verify repository is public OR
- Check GitHub Container Registry permissions

**Cloud Run deployment errors:**
```bash
# Check Cloud Run quota
gcloud run services list --region=us-central1

# Check service account permissions
gcloud projects get-iam-policy $PROJECT_ID
```

### High Costs / Unexpected Charges

**Check current usage:**
```bash
# View Cloud Run revisions
gcloud run revisions list --service=run-tbot --region=us-central1

# Check if service is scaling correctly
gcloud run services describe run-tbot --region=us-central1 --format="value(spec.template.spec.containers[0].resources.limits)"
```

**Verify free tier settings:**
- min-instances: 0 ✅
- max-instances: 1 ✅
- memory: 256Mi ✅

**Check billing:**
- Go to https://console.cloud.google.com/billing
- View usage and costs
- Set up budget alerts

### Bot Token Invalid

**Symptoms:**
- 401 Unauthorized errors
- "Unauthorized" in logs

**Solution:**
1. Get new token from @BotFather
2. Update `BOT_TOKEN` secret in GitHub
3. Redeploy (push to main)

### Service Returns 403 Forbidden

**Symptoms:**
- Telegram can't reach your webhook
- 403 errors in Cloud Run logs

**Solutions:**
- Verify `--allow-unauthenticated` in deploy.yml
- Check Cloud Run IAM permissions:
```bash
gcloud run services get-iam-policy run-tbot --region=us-central1
```

---

## Cost Optimization

### Free Tier Limits (2024)

**Cloud Run:**
- 2,000,000 requests/month
- 360,000 GB-seconds of memory
- 180,000 vCPU-seconds
- **Our usage**: ~10,000 requests/month (well within free tier)

**GitHub Actions:**
- Public repos: unlimited minutes
- Private repos: 2,000 minutes/month
- **Our usage**: ~10 minutes/deploy (200 deploys/month within limit)

**Container Registry:**
- GitHub: unlimited for public images
- Google: 0.5 GB free storage

### Minimizing Costs

1. **Scale to zero**: `min-instances: 0`
   - No cost when bot is idle
   - Cold start: ~2-3 seconds (acceptable for bot)

2. **Limit concurrency**: `max-instances: 1`
   - Prevents runaway costs
   - Sufficient for personal bot

3. **Optimize memory**: `memory: 256Mi`
   - Enough for Go bot
   - Lower memory = lower cost

4. **Clean old images**:
```bash
# List container images
gcloud container images list

# Delete old images
gcloud container images delete gcr.io/$PROJECT_ID/run-tbot:old-tag
```

### Setting Budget Alerts

```bash
# Create budget alert (via GCP Console)
# 1. Go to Billing → Budgets & alerts
# 2. Create budget
# 3. Set amount: $5/month
# 4. Set alert at 50%, 90%, 100%
# 5. Add email notification
```

---

## Next Steps

After successful deployment:

1. ✅ Bot is running on Cloud Run
2. ✅ Webhook is registered
3. ✅ CI/CD pipeline is active

**Continue development:**
- Implement handlers (Phase 2): `/start`, `/help`, dice button
- Add tests (Phase 5)
- Create Makefile (Phase 5)
- Add structured logging (Phase 6)

**Production checklist:**
- [ ] Set up budget alerts
- [ ] Configure custom domain (optional)
- [ ] Set up monitoring/alerting
- [ ] Document bot commands for users
- [ ] Add error tracking (e.g., Sentry)

---

## Additional Resources

**Google Cloud Run:**
- [Documentation](https://cloud.google.com/run/docs)
- [Pricing](https://cloud.google.com/run/pricing)
- [Best Practices](https://cloud.google.com/run/docs/best-practices)

**Telegram Bot API:**
- [Official Documentation](https://core.telegram.org/bots/api)
- [Webhook Guide](https://core.telegram.org/bots/webhooks)
- [Best Practices](https://core.telegram.org/bots/best-practices)

**GitHub Actions:**
- [Documentation](https://docs.github.com/en/actions)
- [Workflow Syntax](https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions)

---

## Support

If you encounter issues:

1. Check logs: `gcloud run services logs read run-tbot`
2. Review GitHub Actions workflow runs
3. Verify webhook: `getWebhookInfo`
4. Check this troubleshooting guide

For educational project questions, review:
- `CLAUDE.md` - Architecture and design decisions
- `README.md` - Project overview
- Code comments - Detailed explanations
