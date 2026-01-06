package handlers

import (
	"log/slog"

	"github.com/Alrem/run-tbot/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleHelp handles the /help command.
// Shows list of available commands, with different content for authorized vs public users.
//
// Authorization logic:
//   - All users see public commands (/start, /help, dice button)
//   - Only users in ALLOWED_USERS see private commands section
//   - Authorization is checked via cfg.IsUserAllowed(userID)
//
// Security note:
//   - We don't reveal that private commands exist to unauthorized users
//   - This prevents information disclosure
//   - Authorized users see "🔐 Private Commands" section
//
// Parameters:
//   - botAPI: Telegram Bot API instance for sending messages
//   - message: Message from Telegram containing the /help command
//   - cfg: Application configuration (contains AllowedUsers list)
func HandleHelp(botAPI *tgbotapi.BotAPI, message *tgbotapi.Message, cfg *config.Config) {
	// Check if user is authorized to see private commands
	// message.From.ID is the Telegram user ID
	// This is a unique int64 number assigned by Telegram
	isAuthorized := cfg.IsUserAllowed(message.From.ID)

	// Log the help command with authorization status
	// This helps track who is using the bot and whether they have access
	slog.Info("/help command received",
		"user_id", message.From.ID,
		"username", message.From.UserName,
		"chat_id", message.Chat.ID,
		"is_authorized", isAuthorized)

	// Step 1: Create help message text
	// Different content for authorized vs unauthorized users
	helpText := formatHelpMessage(isAuthorized)

	// Step 2: Create and send message
	msg := tgbotapi.NewMessage(message.Chat.ID, helpText)

	// ParseMode enables Markdown formatting in message text
	// This allows us to use **bold**, *italic*, `code`, etc.
	// Available modes: "Markdown" (legacy), "MarkdownV2" (recommended), "HTML"
	// We use MarkdownV2 for better control and escaping
	msg.ParseMode = "MarkdownV2"

	// Step 3: Send the message
	if _, err := botAPI.Send(msg); err != nil {
		// If sending fails, log the error
		slog.Error("Failed to send /help message",
			"error", err,
			"chat_id", message.Chat.ID,
			"user_id", message.From.ID,
			"is_authorized", isAuthorized)
		return
	}

	// Log successful send
	slog.Info("/help message sent successfully",
		"chat_id", message.Chat.ID,
		"user_id", message.From.ID,
		"is_authorized", isAuthorized)
}

// formatHelpMessage creates the help message text with command list.
// Returns different content based on user authorization status.
//
// MarkdownV2 formatting rules:
//   - *text* = italic
//   - **text** = bold
//   - `text` = monospace (code)
//   - Special characters must be escaped: _ * [ ] ( ) ~ ` > # + - = | { } . !
//   - Use \\ for literal backslash
//
// Parameters:
//   - isAuthorized: true if user is in AllowedUsers list
//
// Returns:
//   - string: Formatted help message with MarkdownV2 markup
func formatHelpMessage(isAuthorized bool) string {
	// Base message with public commands
	// Using MarkdownV2 for formatting
	// Note: Special characters like . - need escaping with \
	message := "*📖 Available Commands*\n\n" +
		"*Public Commands:*\n" +
		"/start \\- Start the bot and see welcome message\n" +
		"/help \\- Show this help message\n" +
		"🎲 Roll Dice \\- Click the button to roll a dice \\(1\\-6\\)\n"

	// Add private commands section only for authorized users
	if isAuthorized {
		message += "\n*🔐 Private Commands:*\n" +
			"_No private commands implemented yet\\._\n" +
			"_Future features will appear here\\._\n"
	}

	// Add footer with project info
	message += "\n" +
		"_This is an educational bot built with Go\\._\n" +
		"_Source code demonstrates best practices for Telegram bots\\._"

	return message
}
