package handlers

import (
	"fmt"
	"log/slog"
	"math/rand"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleDice handles the "🎲 Dice" button click from reply keyboard.
// When user clicks "🎲 Dice" button, Telegram sends a regular Message update
// with message text matching the button text.
//
// ReplyKeyboard behavior:
//   - Sends regular Message (not CallbackQuery like InlineKeyboard)
//   - Message.Text contains button text ("🎲 Dice")
//   - No callback_data field (that's only for InlineKeyboard)
//   - No need to call AnswerCallbackQuery (only needed for InlineKeyboard)
//
// Flow:
//  1. Generate random number 1-6
//  2. Send message with dice result
//
// Parameters:
//   - bot: Telegram Bot API instance for sending messages
//   - message: Message from Telegram containing button click
func HandleDice(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	// Step 1: Generate random dice number (1-6)
	result := rollDice()

	// Log the dice roll for debugging/monitoring
	// In production, this helps track bot usage and debug issues
	slog.Info("Dice rolled",
		"user_id", message.From.ID,
		"username", message.From.UserName,
		"result", result)

	// Step 2: Send dice result message
	// Create message text with dice emoji and result
	// Unicode dice emoji: 🎲 (U+1F3B2)
	messageText := fmt.Sprintf("🎲 You rolled: %d", result)

	// NewMessage creates a MessageConfig
	// Parameters: chatID (where to send), text (message content)
	msg := tgbotapi.NewMessage(message.Chat.ID, messageText)

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
			"chat_id", message.Chat.ID,
			"result", result)
		return
	}

	slog.Info("Dice result sent successfully",
		"chat_id", message.Chat.ID,
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
