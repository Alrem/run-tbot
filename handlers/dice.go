package handlers

import (
	"fmt"
	"log/slog"
	"math/rand"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleDiceCallback handles the "roll_dice" callback query from inline keyboard button.
// When user clicks "Roll Dice" button, Telegram sends a CallbackQuery update.
//
// CallbackQuery structure:
//   - ID: unique identifier for this query (required for AnswerCallbackQuery)
//   - From: user who clicked the button
//   - Message: original message with the button
//   - Data: callback_data from the button (in our case: "roll_dice")
//
// Important: ALWAYS call AnswerCallbackQuery, even if you don't show an alert.
// If you don't answer, Telegram will show a loading spinner for 30 seconds.
//
// Flow:
//  1. Generate random number 1-6
//  2. Answer callback query (removes loading spinner)
//  3. Send new message with dice result
//
// Parameters:
//   - bot: Telegram Bot API instance for sending messages
//   - callback: CallbackQuery from Telegram containing button click data
func HandleDiceCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	// Step 1: Generate random dice number (1-6)
	result := rollDice()

	// Log the dice roll for debugging/monitoring
	// In production, this helps track bot usage and debug issues
	slog.Info("Dice rolled",
		"user_id", callback.From.ID,
		"username", callback.From.UserName,
		"result", result)

	// Step 2: Answer the callback query
	// This is MANDATORY - tells Telegram to remove the loading spinner
	// If you don't call this, user sees loading spinner for 30 seconds
	//
	// NewCallback creates a CallbackConfig with:
	//   - CallbackQueryID: unique ID from callback.ID
	//   - Text: optional text to show in alert (empty = no alert, just remove spinner)
	//   - ShowAlert: false = small notification, true = popup alert
	callbackConfig := tgbotapi.NewCallback(callback.ID, "")
	if _, err := bot.Request(callbackConfig); err != nil {
		// If this fails, user will see loading spinner (bad UX)
		// Log error but continue - we still want to send the result message
		slog.Error("Failed to answer callback query",
			"error", err,
			"callback_id", callback.ID)
	}

	// Step 3: Send dice result message
	// Create message text with dice emoji and result
	// Unicode dice emoji: 🎲 (U+1F3B2)
	messageText := fmt.Sprintf("🎲 You rolled: %d", result)

	// NewMessage creates a MessageConfig
	// Parameters: chatID (where to send), text (message content)
	// callback.Message.Chat.ID = chat where button was clicked
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, messageText)

	// Send the message
	// bot.Send() returns Message and error
	// We ignore the returned Message (don't need it), but check error
	if _, err := bot.Send(msg); err != nil {
		// If sending fails, user won't see the result
		// This could happen if:
		//   - Bot was blocked by user
		//   - Chat was deleted
		//   - Network error
		//   - Telegram API is down
		slog.Error("Failed to send dice result",
			"error", err,
			"chat_id", callback.Message.Chat.ID,
			"result", result)
		return
	}

	slog.Info("Dice result sent successfully",
		"chat_id", callback.Message.Chat.ID,
		"result", result)
}

// rollDice generates a random number between 1 and 6 (inclusive).
// This simulates a standard 6-sided dice roll.
//
// Implementation notes:
//   - Uses math/rand package (not crypto/rand - we don't need cryptographic randomness for a game)
//   - rand.Intn(n) returns [0, n), so rand.Intn(6) returns [0, 5]
//   - Adding 1 shifts range to [1, 6]
//
// Why not crypto/rand?
//   - crypto/rand is for security-critical randomness (passwords, tokens, encryption keys)
//   - math/rand is faster and sufficient for games/simulations
//   - For dice rolls, predictability is not a security issue
//
// Note: math/rand is automatically seeded since Go 1.20
// Before Go 1.20, you had to call rand.Seed(time.Now().UnixNano())
//
// Returns:
//   - int: random number from 1 to 6 (inclusive)
func rollDice() int {
	// rand.Intn(6) returns 0, 1, 2, 3, 4, or 5
	// Adding 1 gives us 1, 2, 3, 4, 5, or 6
	return rand.Intn(6) + 1
}
