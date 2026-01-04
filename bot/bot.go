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

// GetDiceKeyboard returns an inline keyboard with "Roll Dice" button
// Inline keyboard - buttons that are displayed directly in the message
// Unlike custom keyboard (at the bottom of screen), inline buttons disappear after use
//
// Returns InlineKeyboardMarkup - structure with keyboard markup
func GetDiceKeyboard() tgbotapi.InlineKeyboardMarkup {
	// tgbotapi.NewInlineKeyboardMarkup accepts any number of button rows
	// Each row is created via NewInlineKeyboardRow
	return tgbotapi.NewInlineKeyboardMarkup(
		// NewInlineKeyboardRow creates one row with buttons
		// You can pass multiple buttons - they will be in one row
		tgbotapi.NewInlineKeyboardRow(
			// NewInlineKeyboardButtonData creates a button with callback data
			// Parameter 1: text on button (user sees)
			// Parameter 2: callback_data (sent to bot when clicked)
			tgbotapi.NewInlineKeyboardButtonData("🎲 Roll Dice", "roll_dice"),
		),
	)
}
