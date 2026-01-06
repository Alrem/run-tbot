package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Alrem/run-tbot/bot"
	"github.com/Alrem/run-tbot/config"
	"github.com/Alrem/run-tbot/handlers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Step 1: Initialize structured logger with JSON output
	// slog is Go's standard structured logging library (since Go 1.21)
	// JSON format is perfect for Cloud Run - Google Cloud Logging parses it automatically
	// NewJSONHandler writes logs as JSON to stdout
	// Each log entry will have: time, level, msg, and any additional fields
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Set as default logger so slog.Info(), slog.Error() work globally
	slog.SetDefault(logger)

	slog.Info("Starting Telegram bot application")

	// Step 2: Load configuration from environment variables
	// Config contains: BotToken, Port, Environment, AllowedUsers
	cfg, err := config.Load()
	if err != nil {
		// Fatal error - can't proceed without valid config
		// This will log and exit with status code 1
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Log config (but never log the actual BOT_TOKEN for security!)
	slog.Info("Configuration loaded",
		"port", cfg.Port,
		"environment", cfg.Environment,
		"allowed_users_count", len(cfg.AllowedUsers))

	// Step 3: Initialize Telegram bot
	// cfg.IsDevelopment() enables debug mode which logs all HTTP requests/responses
	// Useful for learning and debugging, but disable in production (verbose)
	botAPI, err := bot.NewBot(cfg.BotToken, cfg.IsDevelopment())
	if err != nil {
		slog.Error("Failed to create bot", "error", err)
		os.Exit(1)
	}

	// Log bot info (bot.Self contains bot's username, ID, etc.)
	slog.Info("Bot authorized successfully",
		"bot_username", botAPI.Self.UserName,
		"bot_id", botAPI.Self.ID)

	// Step 4: Setup HTTP routes
	// http.ServeMux is Go's built-in HTTP request router
	mux := http.NewServeMux()

	// Route 1: Health check endpoint for Cloud Run
	// Cloud Run pings this to verify service is alive
	// Simply returns 200 OK
	mux.HandleFunc("/", healthCheckHandler)

	// Route 2: Telegram webhook endpoint
	// Telegram sends POST requests with Update JSON to this endpoint
	// We'll pass botAPI and cfg to the handler via closure
	mux.HandleFunc("/webhook", webhookHandler(botAPI, cfg))

	// Step 5: Create HTTP server with timeouts
	// Timeouts prevent hanging connections and DoS attacks
	server := &http.Server{
		Addr:    ":" + cfg.Port, // Listen on all interfaces, port from config
		Handler: mux,
		// ReadTimeout: max time to read request (headers + body)
		ReadTimeout: 15 * time.Second,
		// WriteTimeout: max time to write response
		WriteTimeout: 15 * time.Second,
		// IdleTimeout: max time to keep connection open between requests
		IdleTimeout: 60 * time.Second,
	}

	// Step 6: Start server in a goroutine so we can handle shutdown gracefully
	// goroutine = lightweight thread in Go
	go func() {
		slog.Info("Starting HTTP server", "port", cfg.Port)
		// ListenAndServe blocks until server stops or error occurs
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// http.ErrServerClosed is expected during graceful shutdown
			// Any other error is a problem
			slog.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("Bot is running. Press Ctrl+C to stop.")

	// Step 7: Wait for interrupt signal for graceful shutdown
	// Graceful shutdown = finish processing current requests before stopping
	// This is important for Cloud Run deployments

	// Create channel to receive OS signals
	// Buffered channel with size 1 prevents blocking
	quit := make(chan os.Signal, 1)

	// Notify channel on SIGINT (Ctrl+C) or SIGTERM (Cloud Run stop)
	// SIGINT = interrupt signal (Ctrl+C in terminal)
	// SIGTERM = termination signal (sent by Cloud Run on shutdown)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block here until we receive a signal
	sig := <-quit
	slog.Info("Received shutdown signal", "signal", sig.String())

	// Step 8: Graceful shutdown
	// Give server 30 seconds to finish existing requests
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel() // Ensure context is cancelled to free resources

	// Shutdown gracefully
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server stopped gracefully")
}

// healthCheckHandler handles GET / requests for Cloud Run health checks
// Returns 200 OK to indicate service is alive and ready
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept GET requests (health checks should be GET)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Return 200 OK with simple message
	w.WriteHeader(http.StatusOK)
	// Explicitly ignore write error - nothing useful to do if health check write fails
	_, _ = w.Write([]byte("OK"))
}

// webhookHandler creates a handler for POST /webhook requests from Telegram
// Uses closure to pass botAPI and cfg to the handler
// Returns http.HandlerFunc which can be registered with http.HandleFunc
func webhookHandler(botAPI *tgbotapi.BotAPI, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only accept POST requests (Telegram sends POST)
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse JSON body into Update struct
		// Update contains message, callback_query, etc.
		var update tgbotapi.Update

		// json.NewDecoder reads from request body
		// Decode(&update) parses JSON into update struct
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			slog.Error("Failed to decode update", "error", err)
			// IMPORTANT: Always return 200 OK to Telegram
			// If we return error, Telegram will retry the same update
			// This can cause duplicate processing
			w.WriteHeader(http.StatusOK)
			return
		}

		// Log the update (helpful for debugging)
		// update.UpdateID is unique identifier for each update
		var messageText string
		if update.Message != nil {
			messageText = update.Message.Text
		}
		slog.Info("Received update",
			"update_id", update.UpdateID,
			"has_message", update.Message != nil,
			"has_callback", update.CallbackQuery != nil,
			"message_text", messageText)

		// Process update with router
		// Router analyzes update type (Message, CallbackQuery, etc.)
		// and delegates to appropriate handler functions
		// Router implementation: handlers/router.go
		// Handler implementations: handlers/dice.go, handlers/start.go, handlers/help.go
		handlers.RouteUpdate(botAPI, update, cfg)

		// ALWAYS return 200 OK to Telegram
		// Even if processing failed, we don't want Telegram to retry
		// This prevents duplicate message delivery
		// Errors are logged by handlers, not returned to Telegram
		w.WriteHeader(http.StatusOK)
	}
}
