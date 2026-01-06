package handlers

import (
	"fmt"
	"log/slog"
	"math/rand"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleDoubleDice handles the "🎲🎲 Double Dice" button click from reply keyboard.
// Rolls two dice and shows the sum (range 2-12).
//
// Probability distribution for sum of two dice:
//   - 2 or 12: 1/36 (~2.8%)  - one way each
//   - 3 or 11: 2/36 (~5.6%)  - two ways each
//   - 4 or 10: 3/36 (~8.3%)  - three ways each
//   - 5 or 9:  4/36 (~11.1%) - four ways each
//   - 6 or 8:  5/36 (~13.9%) - five ways each
//   - 7:       6/36 (~16.7%) - six ways (most likely)
//
// Flow:
//  1. Roll two dice (each 1-6)
//  2. Calculate sum
//  3. Send message with both dice values and sum
//
// Parameters:
//   - bot: Telegram Bot API instance for sending messages
//   - message: Message from Telegram containing button click
func HandleDoubleDice(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	// Step 1: Roll two dice
	dice1, dice2, sum := rollDoubleDice()

	// Log the roll for debugging/monitoring
	slog.Info("Double dice rolled",
		"user_id", message.From.ID,
		"username", message.From.UserName,
		"dice1", dice1,
		"dice2", dice2,
		"sum", sum)

	// Step 2: Create result message
	// Show both dice values and their sum
	// Format: "🎲🎲 You rolled: 3 + 5 = 8"
	messageText := fmt.Sprintf("🎲🎲 You rolled: %d + %d = *%d*", dice1, dice2, sum)

	// NewMessage creates a MessageConfig
	msg := tgbotapi.NewMessage(message.Chat.ID, messageText)

	// Enable Markdown formatting for bold sum
	msg.ParseMode = "Markdown"

	// Step 3: Send the message
	if _, err := bot.Send(msg); err != nil {
		slog.Error("Failed to send double dice result",
			"error", err,
			"chat_id", message.Chat.ID,
			"dice1", dice1,
			"dice2", dice2,
			"sum", sum)
		return
	}

	slog.Info("Double dice result sent successfully",
		"chat_id", message.Chat.ID,
		"sum", sum)
}

// rollDoubleDice rolls two dice and returns both values plus their sum.
// Each die is a standard 6-sided die (1-6).
//
// Implementation notes:
//   - Uses math/rand package (sufficient for games)
//   - rand.Intn(6) returns [0, 5], adding 1 gives [1, 6]
//   - Sum range is [2, 12] (min: 1+1, max: 6+6)
//
// Returns:
//   - int: first die value (1-6)
//   - int: second die value (1-6)
//   - int: sum of both dice (2-12)
func rollDoubleDice() (int, int, int) {
	dice1 := rand.Intn(6) + 1
	dice2 := rand.Intn(6) + 1
	sum := dice1 + dice2
	return dice1, dice2, sum
}
