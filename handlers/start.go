package handlers

import (
	"log/slog"

	"github.com/Alrem/run-tbot/bot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleStart handles the /start command.
// This is typically the first command users send when they start interacting with the bot.
//
// Telegram /start command behavior:
//   - Automatically sent when user clicks "Start" button in bot info
//   - Can include deep linking parameters (e.g., /start referral_code)
//   - Should provide welcome message and basic bot instructions
//
// Our implementation:
//  1. Sends welcome message explaining what the bot does
//  2. Attaches reply keyboard with all bot features (persistent buttons at bottom)
//  3. User can immediately try any feature via keyboard buttons
//
// Parameters:
//   - botAPI: Telegram Bot API instance for sending messages
//   - message: Message from Telegram containing the /start command
func HandleStart(botAPI *tgbotapi.BotAPI, message *tgbotapi.Message) {
	// Log the start command for monitoring
	// Track user_id to understand bot adoption
	// Track username (may be empty if user hasn't set it)
	slog.Info("/start command received",
		"user_id", message.From.ID,
		"username", message.From.UserName,
		"chat_id", message.Chat.ID)

	// Step 1: Create welcome message text
	// message.From.FirstName is user's first name from their Telegram profile
	// Using FirstName makes the message more personal and friendly
	welcomeText := formatStartMessage(message.From.FirstName)

	// Step 2: Create message configuration
	// NewMessage creates a MessageConfig structure
	// Parameters: chatID (where to send), text (message content)
	msg := tgbotapi.NewMessage(message.Chat.ID, welcomeText)

	// Step 3: Attach reply keyboard with all bot features
	// bot.GetMainKeyboard() returns ReplyKeyboardMarkup with 4 buttons:
	//   - 🎲 Dice, 🎲🎲 Double Dice, 🌀 Twister, 🖥️ OVH Servers
	// When user clicks button, we'll receive regular Message with button text
	// These messages will be routed by router.go to appropriate handlers
	msg.ReplyMarkup = bot.GetMainKeyboard()

	// Step 4: Send the message
	// bot.Send() returns (Message, error)
	// We ignore the returned Message (we don't need message_id for anything)
	if _, err := botAPI.Send(msg); err != nil {
		// If sending fails, log the error
		// Possible causes:
		//   - Bot was blocked by user
		//   - Chat doesn't exist
		//   - Network/API error
		slog.Error("Failed to send /start message",
			"error", err,
			"chat_id", message.Chat.ID,
			"user_id", message.From.ID)
		return
	}

	// Log successful send for monitoring
	// This helps track bot usage and successful interactions
	slog.Info("/start message sent successfully",
		"chat_id", message.Chat.ID,
		"user_id", message.From.ID)
}

// formatStartMessage creates the welcome message text for /start command.
// Extracted as separate function for easier testing and maintenance.
//
// The message should:
//   - Be friendly and welcoming
//   - Explain what the bot does
//   - List available features
//   - Encourage user to try the features
//
// Parameters:
//   - firstName: User's first name from Telegram profile
//
// Returns:
//   - string: Formatted welcome message
func formatStartMessage(firstName string) string {
	// Fallback to "there" if firstName is empty
	// This can happen if user hasn't set their first name in Telegram
	// (rare, but possible)
	name := firstName
	if name == "" {
		name = "there"
	}

	// Use multiline string for better readability
	// The message explains:
	//   1. What the bot does (educational project)
	//   2. Available features (4 buttons)
	//   3. Call to action (use the keyboard)
	return "👋 Hello, " + name + "!\n\n" +
		"Welcome to Run-Tbot - an educational Telegram bot built with Go.\n\n" +
		"Try these features using the keyboard below:\n" +
		"🎲 Dice - Roll a single die (1-6)\n" +
		"🎲🎲 Double Dice - Roll two dice (2-12)\n" +
		"🌀 Twister - Get a random Twister move\n" +
		"🖥️ OVH Servers - Check server availability"
}
