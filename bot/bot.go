// Package bot contains logic for working with Telegram Bot API
package bot

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// NewBot creates a new Telegram bot instance
// Parameters:
//   - token: token from @BotFather for API access
//   - debug: if true, library will log all requests/responses to API
//
// Returns pointer to BotAPI or error if token is invalid
func NewBot(token string, debug bool) (*tgbotapi.BotAPI, error) {
	// tgbotapi.NewBotAPI creates a new bot instance
	// Internally makes a request to Telegram API getMe method to verify token
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		// %w allows "wrapping" the original error for better tracing
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	// Enable debug mode if debug=true parameter is passed
	// In debug mode, library outputs all HTTP requests and responses to stdout
	// Useful for debugging, but DO NOT use in production!
	bot.Debug = debug

	return bot, nil
}

// GetMainKeyboard returns a reply keyboard with all bot features
// Reply keyboard - persistent buttons displayed at the bottom of the screen
// Unlike inline keyboard (buttons in messages), reply keyboard stays visible
// and sends regular messages when buttons are clicked
//
// Features:
//   - 🎲 Dice - Roll single die (1-6)
//   - 🎲🎲 Double Dice - Roll two dice (2-12)
//   - 🌀 Twister - Random Twister game move
//   - 🖥️ OVH Servers - Check OVH server availability (private)
//   - 💱 Kurs NBP - NBP average exchange rates to PLN (public)
//
// Returns ReplyKeyboardMarkup with 3-row button layout
func GetMainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	// Create keyboard with 3 rows
	keyboard := tgbotapi.NewReplyKeyboard(
		// Row 1: Dice features
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🎲 Dice"),
			tgbotapi.NewKeyboardButton("🎲🎲 Double Dice"),
		),
		// Row 2: Other features
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🌀 Twister"),
			tgbotapi.NewKeyboardButton("🖥️ OVH Servers"),
		),
		// Row 3: Finance features
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("💱 Kurs NBP"),
		),
	)

	// ResizeKeyboard optimizes button size for user's screen
	// Without this, keyboard may be too large on mobile devices
	keyboard.ResizeKeyboard = true

	// OneTimeKeyboard=false keeps keyboard visible after button click
	// If true, keyboard would hide after each use (not desired for bot features)
	keyboard.OneTimeKeyboard = false

	return keyboard
}
