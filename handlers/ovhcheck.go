package handlers

import (
	"log/slog"

	"github.com/Alrem/run-tbot/config"
	"github.com/Alrem/run-tbot/ovh"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleOVHCheck handles the "🖥️ OVH Servers" button click from reply keyboard.
// Shows available OVH servers (private feature, only for authorized users).
//
// Authorization:
//   - Only users in ALLOWED_USERS can use this feature
//   - Other users get a "not authorized" message
//
// Functionality:
//   - Fetches OVH server availability from public API
//   - Filters by datacenter (London) and subsidiary (GB)
//   - Returns top 5 cheapest servers with prices
//   - Includes FQN (Fully Qualified Name) for each server
//
// Parameters:
//   - bot: Telegram Bot API instance for sending messages
//   - message: Message from Telegram containing button click
//   - cfg: Application configuration (needed for authorization check)
func HandleOVHCheck(bot *tgbotapi.BotAPI, message *tgbotapi.Message, cfg *config.Config) {
	// Step 1: Check authorization
	if !cfg.IsUserAllowed(message.From.ID) {
		// Log unauthorized access attempt
		slog.Info("Unauthorized OVH check attempt",
			"user_id", message.From.ID,
			"username", message.From.UserName,
			"chat_id", message.Chat.ID)

		// Send error message
		errorMsg := tgbotapi.NewMessage(message.Chat.ID,
			"⛔ This feature is only available to authorized users\\.")
		errorMsg.ParseMode = "MarkdownV2"

		if _, err := bot.Send(errorMsg); err != nil {
			slog.Error("Failed to send authorization error message",
				"error", err, "chat_id", message.Chat.ID)
		}
		return
	}

	// Step 2: Send status message
	statusMsg := tgbotapi.NewMessage(message.Chat.ID,
		"🖥️ Checking OVH server availability\\.\\.\\.\\nThis may take a few seconds\\.")
	statusMsg.ParseMode = "MarkdownV2"

	if _, err := bot.Send(statusMsg); err != nil {
		slog.Error("Failed to send OVH status message",
			"error", err, "chat_id", message.Chat.ID)
		return
	}

	// Step 3: Fetch OVH data
	// Parameters: GB (Great Britain), lon (London), top 5 servers
	slog.Info("Fetching OVH server availability",
		"user_id", message.From.ID,
		"subsidiary", "GB",
		"datacenter", "lon",
		"top", 5)

	offers, err := ovh.GetTopOffers("GB", "lon", 5)
	if err != nil {
		// Log error
		slog.Error("Failed to fetch OVH offers",
			"error", err,
			"user_id", message.From.ID,
			"chat_id", message.Chat.ID)

		// Send user-friendly error message
		errMsg := tgbotapi.NewMessage(message.Chat.ID,
			"❌ Failed to fetch server availability\\. Please try again later\\.")
		errMsg.ParseMode = "MarkdownV2"

		if _, err := bot.Send(errMsg); err != nil {
			slog.Error("Failed to send OVH error message",
				"error", err, "chat_id", message.Chat.ID)
		}
		return
	}

	// Step 4: Format and send results
	messageText := formatOVHResults(offers)

	msg := tgbotapi.NewMessage(message.Chat.ID, messageText)
	msg.ParseMode = "MarkdownV2"
	msg.DisableWebPagePreview = true

	if _, err := bot.Send(msg); err != nil {
		slog.Error("Failed to send OVH results",
			"error", err,
			"chat_id", message.Chat.ID,
			"offers_count", len(offers))
		return
	}

	slog.Info("OVH results sent successfully",
		"user_id", message.From.ID,
		"chat_id", message.Chat.ID,
		"offers_count", len(offers))
}

// formatOVHResults formats OVH offers for display in Telegram.
// Creates a nicely formatted message with header, server list, and footer.
//
// Parameters:
//   - offers: List of OVH Offer structs with pricing and availability
//
// Returns:
//   - string: Formatted message with MarkdownV2 escaping
func formatOVHResults(offers []ovh.Offer) string {
	// Handle empty results
	if len(offers) == 0 {
		return "No available servers found in London datacenter\\."
	}

	// Build message
	message := "🖥️ *Available OVH Servers*\n"
	message += "_Top 5 cheapest in London \\(GB\\)_\n\n"

	for i, offer := range offers {
		message += ovh.FormatOfferForTelegram(offer, i+1) + "\n"
	}

	message += "\n_Use /start to return to main menu_"

	return message
}
