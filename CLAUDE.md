# run-tbot - Educational Telegram Bot

## Project Overview

**run-tbot** is an educational Telegram bot project designed for learning Go programming language and Google Cloud Platform (GCP) Cloud Run deployment. The project emphasizes clean code, detailed educational comments, and real-world deployment practices.

### Key Learning Objectives

- Understanding Go fundamentals and idioms
- Working with Telegram Bot API
- Environment-based configuration management
- HTTP server and webhook handling
- Container-based deployment with Docker
- CI/CD pipelines with GitHub Actions
- Cloud deployment on GCP Cloud Run
- Unit testing and integration testing
- Security best practices for bot development

---

## Architecture

### High-Level Design

```
Telegram Server → Webhook (HTTPS) → Cloud Run → Bot Handler → Response
```

The bot uses **webhook mode** instead of polling for cost efficiency on Cloud Run:
- **Polling**: Bot continuously checks for updates (requires always-on server)
- **Webhook**: Telegram sends updates to our endpoint (scales to zero when idle)

### Layer Architecture

```
┌─────────────────────────────────────┐
│         HTTP Server (main.go)       │  ← Entry point, webhook endpoint
├─────────────────────────────────────┤
│      Router (handlers/router.go)    │  ← Routes commands & button clicks
├─────────────────────────────────────┤
│     Handlers (handlers/*.go)        │  ← Business logic (dice, help, etc)
├─────────────────────────────────────┤
│       Bot Layer (bot/bot.go)        │  ← Telegram API wrapper + keyboard
├─────────────────────────────────────┤
│      External APIs (ovh/*.go)       │  ← OVH API client (optional)
├─────────────────────────────────────┤
│    Config Layer (config/config.go)  │  ← Environment configuration
└─────────────────────────────────────┘
```

**Update Flow**:
1. Telegram → main.go webhook endpoint
2. Router examines update (command vs button click)
3. Routes to appropriate handler
4. Handler may call external APIs (OVH)
5. Handler sends response via Bot layer
6. Telegram delivers to user

---

## File Structure

```
run-tbot/
├── .github/
│   └── workflows/
│       ├── ci.yml              # Continuous Integration (tests, lint)
│       └── deploy.yml          # Continuous Deployment to Cloud Run
├── bot/
│   └── bot.go                  # Bot initialization and ReplyKeyboard helpers
├── config/
│   └── config.go               # Configuration management (env vars)
├── handlers/
│   ├── dice.go                 # Dice roll handler
│   ├── dice_test.go            # Unit tests for dice handler
│   ├── doubledice.go           # Double dice roll handler
│   ├── doubledice_test.go      # Unit tests for double dice handler
│   ├── twister.go              # Twister game move generator handler
│   ├── twister_test.go         # Unit tests for twister handler
│   ├── ovhcheck.go             # OVH server availability handler (private)
│   ├── ovhcheck_test.go        # Unit tests for OVH handler
│   ├── start.go                # /start command handler
│   ├── start_test.go           # Unit tests for start handler
│   ├── help.go                 # /help command handler (with auth)
│   ├── help_test.go            # Unit tests for help handler
│   ├── router.go               # Central routing logic (commands + buttons)
│   └── integration_test.go     # Integration tests
├── logger/
│   └── logger.go               # Structured logging for Cloud Run
├── ovh/
│   ├── client.go               # OVH API client wrapper
│   └── client_test.go          # Unit tests for OVH client
├── docs/
│   └── DEPLOYMENT.md           # Detailed deployment guide
├── .env.example                # Environment variables template
├── .gitignore                  # Git ignore rules
├── CLAUDE.md                   # This file - project documentation
├── Dockerfile                  # Multi-stage Docker build
├── Makefile                    # Development automation
├── README.md                   # Project overview and setup
├── go.mod                      # Go module definition
├── go.sum                      # Go dependencies lock file
└── main.go                     # Application entry point (HTTP server)
```

---

## Design Decisions

### 1. Webhook vs Polling

**Decision**: Use webhook mode

**Rationale**:
- Cloud Run charges for active CPU time
- Webhook allows scaling to zero (no requests = no cost)
- Telegram delivers updates instantly
- More complex initial setup, but better for production

**Trade-offs**:
- Requires HTTPS endpoint (Cloud Run provides this)
- Requires public URL (Cloud Run provides this)
- More complex local development (use ngrok for testing)

### 2. Authorization Strategy

**Decision**: Environment variable with comma-separated user IDs

**Rationale**:
- Simple to configure (`ALLOWED_USERS=123456,789012`)
- No database required
- Secure (not in code, stored as Cloud Run secret)
- Easy to update via Cloud Run console

**Implementation**:
- Empty list = all users can access public functions only
- Check authorization in handlers (not middleware) for granular control

### 3. One File Per Commit

**Decision**: Each commit modifies/creates only one logical file or feature

**Rationale**:
- Educational clarity - easy to understand each change
- Better git history for learning
- Easier to review and explain
- Simple rollback if needed

**Example**:
```
Commit 1: .gitignore
Commit 2: README.md
Commit 3: CLAUDE.md
Commit 4: Extended config/config.go
```

### 4. English Comments Only

**Decision**: All code, comments, and documentation in English

**Rationale**:
- Industry standard practice
- Better for collaboration with international developers
- Better for AI tools and code analysis
- Chat can be in Russian, but code remains universal

### 5. ReplyKeyboard vs InlineKeyboard

**Decision**: Use ReplyKeyboard for main bot interface

**Rationale**:
- **Persistent Interface**: Buttons remain visible at bottom of screen, no need to scroll back to previous messages
- **Better Mobile UX**: ReplyKeyboard is optimized for mobile keyboards with ResizeKeyboard option
- **Simplified Routing**: Message-based routing is simpler than CallbackQuery handling (no need for AnswerCallbackQuery)
- **User Convenience**: Users can quickly access all features without typing commands

**Trade-offs**:
- Buttons take screen space (minimized with ResizeKeyboard=true)
- Button text must be synchronized between keyboard definition (bot.go) and router logic
- Cannot have dynamic button text (InlineKeyboard supports this)
- All users see the same buttons (OVH authorization check happens in handler)

**Implementation**:
- 2x2 button layout for visual balance and mobile-friendliness
- Row 1: `[🎲 Dice] [🎲🎲 Double Dice]`
- Row 2: `[🌀 Twister] [🖥️ OVH Servers]`
- OneTimeKeyboard=false (buttons stay persistent)

**Migration from InlineKeyboard**:
- Changed handler signatures from `CallbackQuery` to `Message`
- Removed CallbackQuery routing from router.go
- Added button message routing with exact text matching

### 6. Button Text Routing

**Decision**: Exact text matching for button routing

**Rationale**:
- Simple and explicit - easy to understand routing logic
- Easy to debug (button text visible in logs)
- No callback_data encoding needed
- Emojis in button text make buttons visually distinctive

**Implementation**:
```go
switch buttonText {
case "🎲 Dice":
    HandleDice(bot, message)
case "🎲🎲 Double Dice":
    HandleDoubleDice(bot, message)
case "🌀 Twister":
    HandleTwister(bot, message)
case "🖥️ OVH Servers":
    HandleOVHCheck(bot, message, cfg)
}
```

**Trade-off**: Button text must be kept in sync between:
- `bot.GetMainKeyboard()` (keyboard definition)
- `handlers.routeButtonMessage()` (routing logic)

**Alternative Considered**: Callback data with InlineKeyboard
- Pros: Dynamic button text, more flexible
- Cons: More complex routing, requires AnswerCallbackQuery, harder to debug

### 7. OVH Package Organization

**Decision**: Separate `ovh/` package for OVH API wrapper

**Rationale**:
- **Separation of Concerns**: API integration logic separate from Telegram handler logic
- **Reusability**: OVH client can be used by multiple handlers or future features
- **Testability**: Pure functions can be tested independently without Telegram mocks
- **Go Best Practices**: Follows standard package organization patterns

**Package Structure**:
- `ovh/client.go`: API types, GetTopOffers(), FormatOfferForTelegram()
- `ovh/client_test.go`: Unit tests for formatting and helper functions
- `handlers/ovhcheck.go`: Telegram-specific handler with authorization

**API Configuration**:
- Subsidiary: `FR` (France) for EUR pricing
- Datacenter: `lon` (London) for location
- This combination provides EUR prices for servers in London datacenter
- Top 3 cheapest servers displayed

**Error Handling**:
- Network errors → User-friendly message ("Please try again later")
- Empty results → Clear message ("No available servers found")
- Full error logging for debugging

### 8. MarkdownV2 Escaping Pattern

**Decision**: Escape formatted output before inserting into MarkdownV2 messages

**Rationale**:
- Telegram's MarkdownV2 requires escaping many special characters: `. - ( ) [ ] etc.`
- Prices like "15.99" contain periods that must be escaped as "15\\.99"
- Number formatting (sprintf) happens before escaping

**Implementation Pattern**:
```go
// 1. Format the value
priceStr := fmt.Sprintf("%.2f", offer.Price)

// 2. Escape for MarkdownV2
escapedPrice := escapeMarkdownV2(priceStr)

// 3. Insert into message
message += fmt.Sprintf("*%s %s/mo*", escapedPrice, currency)
```

**Critical Bug Fixed**: Initially forgot to escape formatted prices, causing Telegram API to reject messages with error: "Character '.' is reserved and must be escaped"

**Lesson Learned**: All user-facing text (including formatted numbers) must be escaped when using MarkdownV2 ParseMode

---

## Development Workflow

### Local Development

1. **Setup Environment**
   ```bash
   cp .env.example .env
   # Edit .env with your BOT_TOKEN
   ```

2. **Run Tests**
   ```bash
   make test
   ```

3. **Run Locally**
   ```bash
   make run
   ```

4. **Test with Webhook** (requires ngrok)
   ```bash
   ngrok http 8080
   # Use ngrok URL to set webhook
   ```

### Commit Workflow

1. Make changes to ONE file
2. Run tests: `make test`
3. Run linters: `make lint`
4. Commit with descriptive message
5. Push to GitHub
6. CI runs automatically
7. Merge to main → automatic deployment

### Testing Strategy

**Unit Tests** (`*_test.go`):
- Test individual functions in isolation
- Mock external dependencies (Telegram API)
- Table-driven tests for multiple scenarios
- Coverage target: >80%

**Integration Tests** (`integration_test.go`):
- Test routing and handler interactions
- Test authorization flow
- End-to-end scenarios

**Manual Testing**:
- Local testing with real Telegram bot (ngrok)
- Staging deployment before production
- Test all commands and callbacks

---

## Security Practices

### Secrets Management

✅ **DO**:
- Store secrets in environment variables
- Use `.env.example` as template (no real secrets)
- Use GitHub Secrets for CI/CD
- Use Cloud Run Secret Manager for production

❌ **DON'T**:
- Never commit `.env` files
- Never hardcode BOT_TOKEN in code
- Never log sensitive data (tokens, user messages)
- Never expose internal errors to users

### Authorization

```go
// Check user authorization in handlers
if cfg.IsUserAllowed(message.From.ID) {
    // Private function access
} else {
    // Public functions only
}
```

### Input Validation

- Validate all user inputs
- Sanitize callback data
- Use type-safe parsing (strconv, not casting)
- Handle edge cases (nil checks, empty strings)

### Error Handling

```go
// Log errors with context, don't crash
if err != nil {
    log.Printf("[ERROR] Failed to send message: %v", err)
    // Send user-friendly error to Telegram
    return
}
```

---

## Deployment Architecture

### Free Tier Optimization

**GCP Cloud Run**:
- Min instances: 0 (scale to zero = no idle cost)
- Max instances: 1 (limit concurrent requests)
- Memory: 256Mi (within free tier)
- CPU: 1 (minimum, sufficient for bot)
- **Free tier**: 2M requests/month, 360K GB-seconds/month

**GitHub Actions**:
- Public repo: unlimited minutes
- Private repo: 2000 minutes/month
- Our CI: ~2 minutes per run
- Our CD: ~5 minutes per deployment

**Telegram Bot API**:
- Completely free
- No rate limits for webhooks
- Rate limit: 30 messages/second per bot

### Deployment Flow

```
Git Push → GitHub Actions → Build Docker → Push to GCR → Deploy to Cloud Run
```

1. Developer pushes to `main` branch
2. GitHub Actions triggers deploy workflow
3. Workflow builds Docker image
4. Image pushed to Google Container Registry (GCR)
5. Cloud Run deploys new revision
6. Zero-downtime deployment (gradual traffic shift)
7. Old revision kept for rollback

### Environment Variables

**Local Development** (`.env`):
```bash
BOT_TOKEN=your_token_here
PORT=8080
ENVIRONMENT=development
ALLOWED_USERS=123456789
```

**Production** (Cloud Run):
- Set via GitHub Secrets
- Injected during deployment
- Encrypted at rest
- Never visible in logs

---

## Code Style and Conventions

### Go Best Practices

1. **Error Wrapping**: Use `%w` for error context
   ```go
   return fmt.Errorf("failed to create bot: %w", err)
   ```

2. **Exported vs Unexported**: Capital = public, lowercase = private
   ```go
   func NewBot() {}    // Exported - can be used by other packages
   func rollDice() {}  // Unexported - internal to package
   ```

3. **Pointer Receivers**: Use pointers for methods that modify state
   ```go
   func (c *Config) IsDevelopment() bool
   ```

4. **Table-Driven Tests**: Test multiple scenarios efficiently
   ```go
   tests := []struct {
       name    string
       input   int
       want    bool
   }{
       {"valid", 5, true},
       {"invalid", 0, false},
   }
   ```

### Documentation

- Every exported function has a comment
- Comments start with function name
- Explain **why**, not just **what**
- Include parameter and return value descriptions

### Naming Conventions

- **Files**: `lowercase_underscore.go` (e.g., `dice_test.go`)
- **Functions**: `CamelCase` for exported, `camelCase` for unexported
- **Variables**: Descriptive names, avoid single letters except in loops
- **Constants**: `UPPER_SNAKE_CASE` for package-level constants

---

## Technology Stack

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| Language | Go | 1.25.5 | Backend development |
| Bot API | go-telegram-bot-api | v5.5.1 | Telegram integration |
| Container | Docker | latest | Containerization |
| Runtime | Cloud Run | N/A | Serverless deployment |
| CI/CD | GitHub Actions | N/A | Automation |
| Registry | GCR | N/A | Container storage |

---

## Future Enhancements

### Planned Features (Post-MVP)

1. **Database Integration**
   - PostgreSQL on Cloud SQL
   - User preferences storage
   - Command history

2. **Advanced Commands**
   - `/stats` - Usage statistics
   - `/settings` - User preferences
   - Custom admin commands

3. **Monitoring**
   - Prometheus metrics
   - Cloud Monitoring dashboards
   - Alert notifications

4. **Caching**
   - Redis for frequently accessed data
   - Reduce API calls
   - Improve response time

5. **Localization**
   - Multi-language support
   - User language preferences
   - Localized error messages

6. **Rate Limiting**
   - Per-user rate limits
   - Anti-spam protection
   - Graceful degradation

---

## Learning Resources

### Go Programming

- [Official Go Tour](https://go.dev/tour/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go by Example](https://gobyexample.com/)

### Telegram Bots

- [Telegram Bot API Documentation](https://core.telegram.org/bots/api)
- [go-telegram-bot-api Examples](https://github.com/go-telegram-bot-api/telegram-bot-api/tree/master/examples)

### Cloud Run

- [Cloud Run Documentation](https://cloud.google.com/run/docs)
- [Cloud Run Quickstart](https://cloud.google.com/run/docs/quickstarts)
- [Cloud Run Pricing](https://cloud.google.com/run/pricing)

### GitHub Actions

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Workflow Syntax](https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions)

---

## Troubleshooting

### Common Issues

**Issue**: Bot doesn't respond to messages
- Check webhook is set correctly: `getWebhookInfo` API method
- Check Cloud Run logs for errors
- Verify BOT_TOKEN is correct
- Ensure Cloud Run service is deployed

**Issue**: Tests failing locally
- Run `go mod download` to ensure dependencies are installed
- Check Go version matches go.mod (1.25.5)
- Verify test files are in correct package

**Issue**: Deployment fails
- Check GitHub Secrets are set correctly
- Verify GCP service account has necessary permissions
- Check Cloud Run quota limits
- Review GitHub Actions logs

**Issue**: 403 Forbidden from Telegram
- BOT_TOKEN is invalid or revoked
- Recreate bot with @BotFather if needed

---

## Contributing

This is an educational project. When making changes:

1. Follow the "one file per commit" rule
2. Add detailed comments explaining concepts
3. Write unit tests for new functions
4. Update documentation (README, CLAUDE.md)
5. Test locally before pushing
6. Ensure CI passes before merging

---

## License

Educational project - feel free to learn from and modify.

---

## Support

For questions or issues:
- Review code comments (detailed explanations)
- Check `docs/DEPLOYMENT.md` for deployment help
- Review GitHub Actions logs for CI/CD issues
- Test locally with ngrok before deploying

**Remember**: This project is designed for learning. Take time to understand each component before moving to the next!
