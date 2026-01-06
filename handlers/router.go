package handlers

import (
	"log/slog"

	"github.com/Alrem/run-tbot/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// RouteUpdate routes incoming Telegram updates to appropriate handlers.
// This is the central routing logic that connects webhook endpoint to handler functions.
//
// Telegram Update structure can contain different types of updates:
//   - Message: regular message from user
//   - EditedMessage: user edited their previous message
//   - CallbackQuery: user clicked inline keyboard button
//   - InlineQuery: user typed @botname in any chat
//   - ChosenInlineResult: user selected inline query result
//   - ... and many more (see Telegram Bot API docs)
//
// Our routing strategy:
//  1. Check which field in Update is non-nil
//  2. Route based on update type
//  3. For messages: route by command
//  4. For callbacks: route by callback_data
//  5. Log and ignore unknown/unhandled updates
//
// Why this approach?
//   - Simple and explicit (easy to understand)
//   - Easy to extend (add new routes)
//   - Centralized routing logic (single source of truth)
//   - Good logging for debugging
//
// Parameters:
//   - bot: Telegram Bot API instance for sending responses
//   - update: Update from Telegram (contains message, callback, etc.)
//   - cfg: Application configuration (needed for authorization checks)
func RouteUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update, cfg *config.Config) {
	// Log incoming update for debugging
	// update.UpdateID is unique identifier for each update
	// Helps track update flow through the system
	slog.Debug("Routing update",
		"update_id", update.UpdateID,
		"has_message", update.Message != nil,
		"has_callback", update.CallbackQuery != nil,
		"has_edited_message", update.EditedMessage != nil)

	// Route 1: Handle regular messages (commands, text, etc.)
	// update.Message is non-nil when user sends a message
	if update.Message != nil {
		routeMessage(bot, update.Message, cfg)
		return
	}

	// Route 2: Handle callback queries (inline button clicks)
	// update.CallbackQuery is non-nil when user clicks inline keyboard button
	if update.CallbackQuery != nil {
		routeCallback(bot, update.CallbackQuery)
		return
	}

	// Route 3: Handle edited messages (optional)
	// update.EditedMessage is non-nil when user edits their message
	// For most bots, edited messages can be ignored or treated same as new messages
	// We log and ignore them for now
	if update.EditedMessage != nil {
		slog.Debug("Ignoring edited message",
			"update_id", update.UpdateID,
			"user_id", update.EditedMessage.From.ID,
			"chat_id", update.EditedMessage.Chat.ID)
		return
	}

	// Unknown/unhandled update type
	// This could be: InlineQuery, ChosenInlineResult, Poll, etc.
	// Log for debugging but don't crash
	slog.Warn("Received unhandled update type",
		"update_id", update.UpdateID)
}

// routeMessage routes Message updates to appropriate command handlers.
//
// Message routing logic:
//   - Check if message contains a command (starts with /)
//   - Extract command text
//   - Route to appropriate handler based on command
//   - Log unknown commands
//
// Command extraction:
//   - message.Command() returns command without / and bot username
//   - Example: "/start@mybot" -> "start"
//   - Example: "/help" -> "help"
//
// Parameters:
//   - bot: Telegram Bot API instance
//   - message: Message from Telegram
//   - cfg: Application configuration
func routeMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, cfg *config.Config) {
	// Check if message is a command
	// message.IsCommand() returns true if message starts with /
	// This handles both commands and text messages
	if !message.IsCommand() {
		// Not a command - log and ignore
		// In future, could handle regular text messages here
		slog.Debug("Ignoring non-command message",
			"user_id", message.From.ID,
			"chat_id", message.Chat.ID,
			"text", message.Text)
		return
	}

	// Extract command text
	// message.Command() returns command without / prefix
	// Also removes bot username if present (/start@botname -> start)
	command := message.Command()

	// Log command for monitoring
	slog.Info("Routing command",
		"command", command,
		"user_id", message.From.ID,
		"username", message.From.UserName,
		"chat_id", message.Chat.ID)

	// Route to appropriate handler based on command
	switch command {
	case "start":
		// /start command - welcome message + dice button
		HandleStart(bot, message)

	case "help":
		// /help command - show available commands (with authorization)
		HandleHelp(bot, message, cfg)

	default:
		// Unknown command - send friendly error message
		// This helps users understand the bot's capabilities
		sendUnknownCommandMessage(bot, message)
	}
}

// routeCallback routes CallbackQuery updates to appropriate handlers.
//
// CallbackQuery routing logic:
//   - Extract callback_data from button click
//   - Route to appropriate handler based on callback_data
//   - Log unknown callbacks
//
// Callback data format:
//   - When creating button: NewInlineKeyboardButtonData("text", "callback_data")
//   - When user clicks: callback.Data contains "callback_data"
//   - Can be any string, we use simple identifiers: "roll_dice", "settings_menu", etc.
//
// Parameters:
//   - bot: Telegram Bot API instance
//   - callback: CallbackQuery from Telegram
func routeCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	// callback.Data contains the callback_data from button
	// This is the string we set when creating the button
	data := callback.Data

	// Log callback for monitoring
	slog.Info("Routing callback",
		"callback_data", data,
		"user_id", callback.From.ID,
		"username", callback.From.UserName,
		"message_id", callback.Message.MessageID)

	// Route to appropriate handler based on callback_data
	switch data {
	case "roll_dice":
		// "Roll Dice" button click
		HandleDiceCallback(bot, callback)

	default:
		// Unknown callback - answer to remove loading spinner
		// Even if we don't handle it, we must answer to remove spinner
		slog.Warn("Unknown callback data",
			"callback_data", data,
			"user_id", callback.From.ID)

		// Answer callback query to remove loading spinner
		// Empty text = no notification shown to user
		callbackConfig := tgbotapi.NewCallback(callback.ID, "")
		if _, err := bot.Request(callbackConfig); err != nil {
			slog.Error("Failed to answer unknown callback",
				"error", err,
				"callback_data", data)
		}
	}
}

// sendUnknownCommandMessage sends a friendly error message for unknown commands.
// Helps users discover available commands without frustration.
//
// Parameters:
//   - bot: Telegram Bot API instance
//   - message: Original message with unknown command
func sendUnknownCommandMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	// Log unknown command for analytics
	// Helps identify which commands users expect but aren't implemented
	slog.Info("Unknown command received",
		"command", message.Command(),
		"user_id", message.From.ID,
		"chat_id", message.Chat.ID)

	// Create friendly error message
	// Don't just say "error" - guide user to /help
	errorText := "❓ Unknown command. Use /help to see available commands."

	msg := tgbotapi.NewMessage(message.Chat.ID, errorText)

	// Send error message
	if _, err := bot.Send(msg); err != nil {
		slog.Error("Failed to send unknown command message",
			"error", err,
			"chat_id", message.Chat.ID,
			"command", message.Command())
	}
}
