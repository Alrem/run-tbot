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
//   - CallbackQuery: user clicked inline keyboard button (not used - we use ReplyKeyboard)
//   - InlineQuery: user typed @botname in any chat
//   - ChosenInlineResult: user selected inline query result
//   - ... and many more (see Telegram Bot API docs)
//
// Our routing strategy:
//  1. Check which field in Update is non-nil
//  2. Route based on update type
//  3. For messages: route by command or button text
//  4. Log and ignore unknown/unhandled updates
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
		"has_edited_message", update.EditedMessage != nil)

	// Route 1: Handle regular messages (commands, button clicks, text)
	// update.Message is non-nil when user sends a message
	// This includes:
	//   - Commands (/start, /help)
	//   - ReplyKeyboard button clicks (sends Message with button text)
	//   - Regular text messages
	if update.Message != nil {
		routeMessage(bot, update.Message, cfg)
		return
	}

	// Route 2: Handle edited messages (optional)
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
	// This could be: InlineQuery, ChosenInlineResult, Poll, CallbackQuery, etc.
	// Note: We don't handle CallbackQuery anymore since we use ReplyKeyboard
	// Log for debugging but don't crash
	slog.Warn("Received unhandled update type",
		"update_id", update.UpdateID)
}

// routeMessage routes Message updates to appropriate handlers.
//
// Message routing logic:
//   - Check if message is a command (starts with /)
//   - If command: route to command handler
//   - If not command: check if it's a button click (ReplyKeyboard)
//   - If button: route to button handler
//   - Otherwise: log and ignore
//
// ReplyKeyboard vs InlineKeyboard:
//   - ReplyKeyboard: sends regular Message with button text
//   - InlineKeyboard: sends CallbackQuery with callback_data
//   - We use ReplyKeyboard, so button clicks arrive as Messages
//
// Parameters:
//   - bot: Telegram Bot API instance
//   - message: Message from Telegram
//   - cfg: Application configuration
func routeMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, cfg *config.Config) {
	// Route 1: Handle commands (messages starting with /)
	if message.IsCommand() {
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
			// /start command - welcome message + keyboard
			HandleStart(bot, message)

		case "help":
			// /help command - show available commands (with authorization)
			HandleHelp(bot, message, cfg)

		default:
			// Unknown command - send friendly error message
			sendUnknownCommandMessage(bot, message)
		}
		return
	}

	// Route 2: Handle button clicks from ReplyKeyboard
	// ReplyKeyboard buttons send regular messages with button text
	// We check if message text matches any of our button labels
	routeButtonMessage(bot, message, cfg)
}

// routeButtonMessage routes ReplyKeyboard button clicks to appropriate handlers.
//
// ReplyKeyboard button routing logic:
//   - Extract button text from message
//   - Match against known button labels
//   - Route to appropriate handler based on button text
//   - Log and ignore unknown button text
//
// Button text format:
//   - When creating button: NewKeyboardButton("🎲 Dice")
//   - When user clicks: message.Text contains "🎲 Dice"
//   - We match exact text (including emojis)
//
// Why exact text matching?
//   - Simple and explicit
//   - No need for callback_data encoding
//   - Easy to debug (see button text in logs)
//   - Emojis make buttons visually distinctive
//
// Trade-off:
//   - Must keep button text in sync between bot.GetMainKeyboard() and this router
//   - Changing button text requires updating both places
//   - But: this is explicit and easy to maintain
//
// Parameters:
//   - bot: Telegram Bot API instance
//   - message: Message from Telegram containing button click
//   - cfg: Application configuration (needed for authorization in OVH handler)
func routeButtonMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, cfg *config.Config) {
	// Extract and trim button text
	// strings.TrimSpace removes any accidental whitespace
	buttonText := message.Text

	// Log button click for monitoring
	slog.Info("Routing button click",
		"button_text", buttonText,
		"user_id", message.From.ID,
		"username", message.From.UserName,
		"chat_id", message.Chat.ID)

	// Route to appropriate handler based on button text
	// IMPORTANT: These strings must match button text in bot.GetMainKeyboard()
	switch buttonText {
	case "🎲 Dice":
		// Single dice roll (1-6)
		HandleDice(bot, message)

	case "🎲🎲 Double Dice":
		// Double dice roll (2-12)
		HandleDoubleDice(bot, message)

	case "🌀 Twister":
		// Twister game move
		HandleTwister(bot, message)

	case "🖥️ OVH Servers":
		// OVH server availability check (private)
		HandleOVHCheck(bot, message, cfg)

	default:
		// Unknown button or regular text message
		// Log but don't send error (could be user typing normally)
		slog.Debug("Ignoring unknown button text or regular message",
			"text", buttonText,
			"user_id", message.From.ID,
			"chat_id", message.Chat.ID)
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
