// Package config manages application configuration
// Reads environment variables and validates them
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config stores application configuration
// All fields are public (capitalized) so they can be accessed from other packages
type Config struct {
	// BotToken - token for accessing Telegram Bot API, obtained from @BotFather
	BotToken string

	// Port - port on which HTTP server will listen
	// Cloud Run automatically sets PORT environment variable
	Port string

	// Environment - environment (development or production)
	// Used to enable debug mode in development
	Environment string

	// AllowedUsers - list of Telegram user IDs allowed to access private functions
	// Parsed from ALLOWED_USERS environment variable (comma-separated list)
	// Empty list means no users have access to private functions
	// Example: ALLOWED_USERS=123456789,987654321
	AllowedUsers []int64
}

// Load reads configuration from environment variables
// Returns pointer to Config or error if required variables are not set
func Load() (*Config, error) {
	// Read BOT_TOKEN from environment variable
	// os.Getenv returns empty string if variable is not set
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		// fmt.Errorf creates a new error with formatted message
		return nil, fmt.Errorf("BOT_TOKEN environment variable is required")
	}

	// Read PORT, use "8080" as default if not set
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port for local development
	}

	// Read ENVIRONMENT, use "production" as default
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "production"
	}

	// Read ALLOWED_USERS and parse comma-separated list of user IDs
	// strings.TrimSpace removes leading/trailing whitespace
	// If ALLOWED_USERS is empty or not set, allowedUsers will be empty slice
	allowedUsersStr := strings.TrimSpace(os.Getenv("ALLOWED_USERS"))
	var allowedUsers []int64
	if allowedUsersStr != "" {
		// strings.Split divides string by comma: "123,456" -> ["123", "456"]
		userIDs := strings.Split(allowedUsersStr, ",")
		for _, userIDStr := range userIDs {
			// strings.TrimSpace removes whitespace around each ID: " 123 " -> "123"
			userIDStr = strings.TrimSpace(userIDStr)
			if userIDStr == "" {
				continue // Skip empty strings (e.g., from "123,,456")
			}

			// strconv.ParseInt converts string to int64
			// Parameters: string, base (10 for decimal), bitSize (64 for int64)
			// Telegram user IDs are large numbers that require 64-bit integers
			userID, err := strconv.ParseInt(userIDStr, 10, 64)
			if err != nil {
				// If conversion fails, return error with context
				return nil, fmt.Errorf("invalid user ID in ALLOWED_USERS: %s: %w", userIDStr, err)
			}
			allowedUsers = append(allowedUsers, userID)
		}
	}

	// Create and return pointer to Config struct
	// & creates a pointer to the struct
	return &Config{
		BotToken:     botToken,
		Port:         port,
		Environment:  environment,
		AllowedUsers: allowedUsers,
	}, nil
}

// IsDevelopment checks if application is running in development mode
// Returns true if ENVIRONMENT = "development"
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsUserAllowed checks if a Telegram user ID is in the allowed users list
// Parameters:
//   - userID: Telegram user ID to check (from message.From.ID or callback.From.ID)
//
// Returns:
//   - true if user is in AllowedUsers list
//   - false if user is not in the list OR if AllowedUsers is empty
//
// Usage:
//
//	if cfg.IsUserAllowed(message.From.ID) {
//	    // User has access to private functions
//	} else {
//	    // Public functions only
//	}
func (c *Config) IsUserAllowed(userID int64) bool {
	// If AllowedUsers is empty, no users have access to private functions
	// This is a security-first approach: explicit > implicit
	if len(c.AllowedUsers) == 0 {
		return false
	}

	// Linear search through allowed users
	// This is fine for small lists (typically 1-10 users)
	// For larger lists, consider using a map for O(1) lookup
	for _, allowedID := range c.AllowedUsers {
		if allowedID == userID {
			return true
		}
	}

	return false
}
