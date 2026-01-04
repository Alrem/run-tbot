// Package config manages application configuration
// Reads environment variables and validates them
package config

import (
	"fmt"
	"os"
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

	// Create and return pointer to Config struct
	// & creates a pointer to the struct
	return &Config{
		BotToken:    botToken,
		Port:        port,
		Environment: environment,
	}, nil
}

// IsDevelopment checks if application is running in development mode
// Returns true if ENVIRONMENT = "development"
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}
