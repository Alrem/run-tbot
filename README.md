# run-tbot - Educational Telegram Bot

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.25.5-blue.svg)](https://go.dev/)
[![Cloud Run](https://img.shields.io/badge/GCP-Cloud%20Run-4285F4.svg)](https://cloud.google.com/run)

An educational Telegram bot project for learning Go programming and Google Cloud Platform deployment. Features webhook-based architecture, structured logging, and CI/CD automation.

## Features

### Interactive Features (ReplyKeyboard)
- 🎲 **Dice Roll**: Roll a single die (1-6)
- 🎲🎲 **Double Dice**: Roll two dice with sum (2-12)
- 🌀 **Twister**: Random Twister game move generator (hand/foot + color)
- 🖥️ **OVH Servers**: Check available OVH servers in London with EUR pricing (private feature)

### Technical Features
- 🔐 **Authorization**: User-based access control for private functions
- 📝 **Commands**: `/start`, `/help` with contextual help
- ⌨️ **ReplyKeyboard**: Persistent button interface at bottom of screen
- 🚀 **Cloud Native**: Deployed on GCP Cloud Run with auto-scaling
- 🔄 **CI/CD**: Automated deployment via GitHub Actions
- 📊 **Structured Logging**: JSON logs with slog for Cloud Run
- ✅ **Tested**: Unit and integration tests with >80% coverage
- 💰 **Free Tier**: Optimized to run within GCP free tier ($0/month)

## Table of Contents

- [Quick Start](#quick-start)
- [Prerequisites](#prerequisites)
- [Environment Variables](#environment-variables)
- [Local Development](#local-development)
- [Deployment](#deployment)
- [Project Structure](#project-structure)
- [Usage](#usage)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

## Quick Start

```bash
# 1. Clone the repository
git clone https://github.com/yourusername/run-tbot.git
cd run-tbot

# 2. Copy environment template
cp .env.example .env

# 3. Get bot token from @BotFather and add to .env
# Edit .env and set BOT_TOKEN=your_token_here

# 4. Install dependencies
go mod download

# 5. Run locally
go run main.go

# 6. (Optional) Test with ngrok for webhook testing
ngrok http 8080
```

## Prerequisites

### For Local Development

- **Go 1.25.5+**: [Download here](https://go.dev/dl/)
- **Telegram Bot Token**: Get from [@BotFather](https://t.me/BotFather)
- **ngrok** (optional): For local webhook testing

### For Cloud Deployment

- **Google Cloud Platform Account**: [Sign up](https://cloud.google.com/)
- **GitHub Account**: For CI/CD automation
- **gcloud CLI**: [Install guide](https://cloud.google.com/sdk/docs/install)

## Environment Variables

Create a `.env` file (use `.env.example` as template):

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BOT_TOKEN` | **Yes** | - | Telegram bot token from @BotFather |
| `PORT` | No | `8080` | HTTP server port (Cloud Run sets this automatically) |
| `ENVIRONMENT` | No | `production` | Environment mode (`development` or `production`) |
| `ALLOWED_USERS` | No | - | Comma-separated list of user IDs for private functions (e.g., `123456,789012`) |
| `WEBHOOK_URL` | No | - | Full webhook URL (set after Cloud Run deployment) |

### Getting Your Bot Token

1. Open Telegram and search for [@BotFather](https://t.me/BotFather)
2. Send `/newbot` command
3. Follow instructions to choose name and username
4. Copy the token (format: `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`)
5. Add to `.env`: `BOT_TOKEN=your_token_here`

### Finding Your User ID

1. Message [@userinfobot](https://t.me/userinfobot) in Telegram
2. Copy the `Id` number
3. Add to `.env`: `ALLOWED_USERS=your_user_id`

## Local Development

### Running the Bot

```bash
# Development mode with debug logging
export ENVIRONMENT=development
go run main.go

# Or use the Makefile
make run
```

The server will start on `http://localhost:8080` with these endpoints:
- `GET /` - Health check (returns "OK")
- `POST /webhook` - Telegram webhook endpoint

### Testing with Webhook (ngrok)

Since Telegram requires HTTPS, use ngrok for local webhook testing:

```bash
# In one terminal
go run main.go

# In another terminal
ngrok http 8080

# Copy the HTTPS URL (e.g., https://abc123.ngrok.io)
# Set webhook:
curl -X POST "https://api.telegram.org/bot${BOT_TOKEN}/setWebhook" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://abc123.ngrok.io/webhook"}'
```

**Note**: ngrok URLs change on each restart. For persistent development, consider ngrok paid plan or deploy to Cloud Run.

### Project Structure

```
run-tbot/
├── bot/
│   └── bot.go              # Bot initialization and keyboard helpers
├── config/
│   └── config.go           # Environment configuration management
├── handlers/
│   ├── dice.go             # Dice roll handler
│   ├── dice_test.go        # Unit tests for dice handler
│   ├── doubledice.go       # Double dice roll handler
│   ├── doubledice_test.go  # Unit tests for double dice handler
│   ├── twister.go          # Twister game move generator handler
│   ├── twister_test.go     # Unit tests for twister handler
│   ├── ovhcheck.go         # OVH server availability handler (private)
│   ├── ovhcheck_test.go    # Unit tests for OVH handler
│   ├── start.go            # /start command handler
│   ├── start_test.go       # Unit tests for start handler
│   ├── help.go             # /help command handler (with auth)
│   ├── help_test.go        # Unit tests for help handler
│   ├── router.go           # Central routing logic
│   └── integration_test.go # Integration tests
├── logger/
│   └── logger.go           # Structured logging (slog wrapper)
├── ovh/
│   ├── client.go           # OVH API client wrapper
│   └── client_test.go      # Unit tests for OVH client
├── .github/
│   └── workflows/
│       ├── ci.yml          # Continuous Integration
│       └── deploy.yml      # Continuous Deployment to Cloud Run
├── docs/
│   └── DEPLOYMENT.md       # Detailed deployment guide
├── .env.example            # Environment variables template
├── .gitignore              # Git ignore rules
├── CLAUDE.md               # Project architecture documentation
├── Dockerfile              # Multi-stage Docker build
├── Makefile                # Development automation
├── README.md               # This file
├── go.mod                  # Go module definition
├── go.sum                  # Go dependencies lock
└── main.go                 # Application entry point
```

## Deployment

### Quick Deploy to Cloud Run

Full deployment instructions are in [`docs/DEPLOYMENT.md`](docs/DEPLOYMENT.md).

**Summary**:

1. **Setup GCP**:
   ```bash
   gcloud auth login
   gcloud config set project YOUR_PROJECT_ID
   gcloud services enable run.googleapis.com artifactregistry.googleapis.com
   ```

2. **Create Service Account** for GitHub Actions (see [`docs/DEPLOYMENT.md`](docs/DEPLOYMENT.md))

3. **Configure GitHub Secrets**:
   - `GCP_PROJECT_ID`
   - `GCP_SA_KEY`
   - `BOT_TOKEN`
   - `GCP_REGION`

4. **Push to main**:
   ```bash
   git push origin main
   ```
   GitHub Actions will automatically build, push to Artifact Registry, and deploy to Cloud Run.

5. **Set Webhook**:
   ```bash
   # Get URL from GitHub Actions logs or Cloud Run console
   SERVICE_URL="https://your-service-url.run.app"

   curl -X POST "https://api.telegram.org/bot${BOT_TOKEN}/setWebhook" \
     -H "Content-Type: application/json" \
     -d "{\"url\": \"${SERVICE_URL}/webhook\"}"
   ```

### Deployment Architecture

```
Developer → Git Push → GitHub Actions → Build Docker → Artifact Registry → Cloud Run
                                                                              ↓
                                                                         Telegram Bot
```

**Cost**: $0/month (within free tier)
- Cloud Run: 2M requests/month free
- Artifact Registry: 0.5 GB storage free
- GitHub Actions: Unlimited for public repos

## Usage

### Bot Commands

- `/start` - Display welcome message with ReplyKeyboard showing all available buttons
- `/help` - Show available commands and features (context-aware based on authorization)

### Interactive Button Features

The bot provides a persistent ReplyKeyboard with 4 buttons at the bottom of your screen:

#### 🎲 Dice Roll
- Click the "🎲 Dice" button
- Receive a random number from 1 to 6
- Simple single die roll

#### 🎲🎲 Double Dice
- Click the "🎲🎲 Double Dice" button
- Roll two dice simultaneously
- Get individual results plus the sum (range: 2-12)
- Example: "You rolled: 4 + 5 = **9**"

#### 🌀 Twister
- Click the "🌀 Twister" button
- Generate a random Twister game move
- Returns: limb (Left/Right Hand/Foot) + color (Red/Blue/Green/Yellow)
- Example: "🔴 Right Hand Red"

#### 🖥️ OVH Servers (Private Feature)
- Click the "🖥️ OVH Servers" button
- **Authorization required**: Only available to users in `ALLOWED_USERS` list
- Shows top 3 cheapest available OVH servers in London datacenter
- Displays pricing in EUR with server specifications
- Uses OVH public API for real-time availability

### Private Functions

Set `ALLOWED_USERS` environment variable with comma-separated user IDs:

```bash
ALLOWED_USERS=123456789,987654321
```

Users in this list will:
- See OVH Servers button functionality (unauthorized users get an error message)
- See additional private features listed in `/help`
- Access future private commands

## Testing

### Run Tests

```bash
# Run all tests
make test

# Run tests with coverage
go test -v -cover ./...

# Run tests with race detector
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Linting and Formatting

```bash
# Run go vet
make lint

# Format code
gofmt -w .

# Check formatting
gofmt -l .
```

### CI/CD Pipeline

Every push triggers:
- ✅ Code formatting check (`gofmt`)
- ✅ Static analysis (`go vet`)
- ✅ Unit tests (`go test`)
- ✅ Build verification

Pushes to `main` also trigger:
- 🚀 Docker image build
- 📦 Push to Artifact Registry
- ☁️ Deploy to Cloud Run

## Architecture Decisions

### Why Webhook Instead of Polling?

- **Cost Efficiency**: Cloud Run charges per request, webhooks scale to zero when idle
- **Instant Delivery**: Telegram pushes updates immediately
- **Better for Production**: Industry standard for bot deployment

### Why ReplyKeyboard Instead of InlineKeyboard?

- **Persistent Interface**: Buttons stay visible at bottom of screen, no need to scroll up
- **Better Mobile UX**: ReplyKeyboard is optimized for mobile keyboards (ResizeKeyboard option)
- **Simplified Routing**: Message-based routing is simpler than CallbackQuery handling
- **User Convenience**: Users can quickly access all features without sending commands

**Trade-offs:**
- Buttons take screen space (minimized with ResizeKeyboard)
- Button text must be synchronized between keyboard definition and router
- Cannot have buttons with dynamic text (InlineKeyboard supports this)

**Layout:** 2x2 grid for visual balance and mobile-friendliness

### Why Structured Logging (slog)?

- **Cloud Integration**: Google Cloud Logging parses JSON automatically
- **Searchable Fields**: Filter logs by user ID, command, error type
- **Performance**: Efficient structured output format

### Why Separate OVH Package?

- **Separation of Concerns**: API wrapper logic separate from handler logic
- **Reusability**: OVH client can be used by multiple handlers in the future
- **Testability**: Pure functions can be tested independently
- **Go Best Practices**: Follows standard package organization patterns

### Why One File Per Commit?

- **Educational Clarity**: Easy to understand each change
- **Better History**: Clean git log for learning
- **Easy Rollback**: Simple to revert specific features

## Contributing

This is an educational project. Contributions are welcome!

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (one file per commit preferred)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

**Guidelines**:
- Follow existing code style
- Add unit tests for new features
- Update documentation (README, CLAUDE.md)
- Use English for all code, comments, and docs

## Troubleshooting

### Bot doesn't respond

- ✅ Check webhook is set: `curl "https://api.telegram.org/bot${BOT_TOKEN}/getWebhookInfo"`
- ✅ Verify BOT_TOKEN is correct
- ✅ Check Cloud Run logs in GCP Console
- ✅ Ensure Cloud Run service is deployed and running

### Local development issues

- ✅ Run `go mod download` to install dependencies
- ✅ Verify Go version: `go version` (should be 1.25+)
- ✅ Check `.env` file exists and BOT_TOKEN is set
- ✅ For webhook testing, use ngrok for HTTPS

### Deployment failures

- ✅ Verify GitHub Secrets are configured correctly
- ✅ Check Service Account has required permissions
- ✅ Review GitHub Actions logs for error details
- ✅ Ensure GCP APIs are enabled (Cloud Run, Artifact Registry)

## Resources

- [Telegram Bot API Documentation](https://core.telegram.org/bots/api)
- [go-telegram-bot-api Library](https://github.com/go-telegram-bot-api/telegram-bot-api)
- [Google Cloud Run Documentation](https://cloud.google.com/run/docs)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Go Programming Language](https://go.dev/)

## License

MIT License - see LICENSE file for details.

## Acknowledgments

- Built with [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)
- Deployed on [Google Cloud Run](https://cloud.google.com/run)
- CI/CD powered by [GitHub Actions](https://github.com/features/actions)

---

**Educational Project** - Perfect for learning Go, Telegram bots, and cloud deployment!

For detailed architecture and development guidelines, see [CLAUDE.md](CLAUDE.md).
For deployment instructions, see [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md).
