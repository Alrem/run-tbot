package handlers

import (
	"fmt"
	"log/slog"
	"math/rand"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleTwister handles the "🌀 Twister" button click from reply keyboard.
// Generates a random Twister game move (limb + color).
//
// Twister game rules:
//   - 4 limbs: Left Hand, Right Hand, Left Foot, Right Foot
//   - 4 colors: Red, Blue, Green, Yellow
//   - Total combinations: 4 × 4 = 16 possible moves
//   - Each move has equal probability: 1/16 (6.25%)
//
// Flow:
//  1. Generate random limb (hand or foot, left or right)
//  2. Generate random color with matching emoji
//  3. Send formatted message with move instruction
//
// Parameters:
//   - bot: Telegram Bot API instance for sending messages
//   - message: Message from Telegram containing button click
func HandleTwister(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	// Step 1: Generate random Twister move
	limb, color, emoji := generateTwisterMove()

	// Log the move for debugging/monitoring
	slog.Info("Twister move generated",
		"user_id", message.From.ID,
		"username", message.From.UserName,
		"limb", limb,
		"color", color)

	// Step 2: Create result message
	// Format: "🌀 Twister Move
	//
	//          🔴 Right Hand Red"
	messageText := fmt.Sprintf("🌀 *Twister Move*\n\n%s %s %s", emoji, limb, color)

	// NewMessage creates a MessageConfig
	msg := tgbotapi.NewMessage(message.Chat.ID, messageText)

	// Enable Markdown formatting for bold header
	msg.ParseMode = "Markdown"

	// Step 3: Send the message
	if _, err := bot.Send(msg); err != nil {
		slog.Error("Failed to send Twister move",
			"error", err,
			"chat_id", message.Chat.ID,
			"limb", limb,
			"color", color)
		return
	}

	slog.Info("Twister move sent successfully",
		"chat_id", message.Chat.ID,
		"limb", limb,
		"color", color)
}

// generateTwisterMove generates a random Twister game move.
// Returns limb (e.g., "Left Hand"), color (e.g., "Red"), and color emoji.
//
// Limbs (4 options):
//   - Left Hand
//   - Right Hand
//   - Left Foot
//   - Right Foot
//
// Colors (4 options):
//   - Red (🔴)
//   - Blue (🔵)
//   - Green (🟢)
//   - Yellow (🟡)
//
// Implementation notes:
//   - Uses math/rand for random selection
//   - Color index is used to get matching emoji
//   - All combinations have equal probability
//
// Returns:
//   - string: limb name (e.g., "Left Hand")
//   - string: color name (e.g., "Red")
//   - string: color emoji (e.g., "🔴")
func generateTwisterMove() (string, string, string) {
	// Define all possible limbs
	limbs := []string{
		"Left Hand",
		"Right Hand",
		"Left Foot",
		"Right Foot",
	}

	// Define all possible colors and matching emojis
	// Index must match between colors and emojis arrays
	colors := []string{"Red", "Blue", "Green", "Yellow"}
	emojis := []string{"🔴", "🔵", "🟢", "🟡"}

	// Randomly select limb
	limb := limbs[rand.Intn(len(limbs))]

	// Randomly select color (and get matching emoji)
	colorIndex := rand.Intn(len(colors))
	color := colors[colorIndex]
	emoji := emojis[colorIndex]

	return limb, color, emoji
}
