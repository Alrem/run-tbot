package handlers

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/Alrem/run-tbot/nbp"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// defaultCurrencies lists the currencies shown by default when user clicks the button.
// These are the most common currencies used in Polish business and tax calculations.
var defaultCurrencies = []string{"EUR", "USD", "GBP", "CHF"}

// HandleNBPRate handles the "💱 Kurs NBP" button click from reply keyboard.
// Shows NBP average (sredni) exchange rates for common currencies to PLN.
//
// This is a PUBLIC feature - all users can access it.
//
// The rates shown comply with Polish tax law requirements:
// "kurs sredni NBP z ostatniego dnia roboczego poprzedzajacego dzien
// uzyskania przychodu, poniesienia kosztu, wydatku lub zaplaty podatku"
//
// NBP publishes Table A on every business day - fetching the last table
// automatically gives the most recent published rate (last business day).
//
// Parameters:
//   - bot: Telegram Bot API instance for sending messages
//   - message: Message from Telegram containing button click
func HandleNBPRate(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	slog.Info("NBP rate requested",
		"user_id", message.From.ID,
		"username", message.From.UserName,
		"chat_id", message.Chat.ID)

	// Fetch last Table A from NBP API
	table, err := nbp.GetLastTableA()
	if err != nil {
		slog.Error("Failed to fetch NBP rates",
			"error", err,
			"user_id", message.From.ID,
			"chat_id", message.Chat.ID)

		errMsg := tgbotapi.NewMessage(message.Chat.ID,
			"❌ Failed to fetch NBP exchange rates\\. Please try again later\\.")
		errMsg.ParseMode = "MarkdownV2"

		if _, err := bot.Send(errMsg); err != nil {
			slog.Error("Failed to send NBP error message",
				"error", err, "chat_id", message.Chat.ID)
		}
		return
	}

	// Format and send the result
	text := formatNBPRates(table)

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "MarkdownV2"

	if _, err := bot.Send(msg); err != nil {
		slog.Error("Failed to send NBP rates message",
			"error", err,
			"chat_id", message.Chat.ID)
		return
	}

	slog.Info("NBP rates sent successfully",
		"user_id", message.From.ID,
		"chat_id", message.Chat.ID,
		"effective_date", table.EffectiveDate)
}

// formatNBPRates formats an NBP RateTable into a MarkdownV2 Telegram message.
//
// Displays:
//   - Table number and effective date (last business day)
//   - Rates for defaultCurrencies to PLN
//   - Legal note about applicability for tax/income purposes
//
// Parameters:
//   - table: NBP RateTable with rates and metadata
//
// Returns:
//   - string: Formatted MarkdownV2 message
func formatNBPRates(table *nbp.RateTable) string {
	var b strings.Builder

	// Header with table info and effective date
	b.WriteString("💱 *NBP Average Exchange Rates \\(Table A\\)*\n")
	b.WriteString(fmt.Sprintf("_Table %s, effective: %s_\n\n",
		escapeMarkdownV2NBP(table.No),
		escapeMarkdownV2NBP(table.EffectiveDate)))

	// Rates for default currencies
	for _, code := range defaultCurrencies {
		rate, ok := nbp.FindRate(table, code)
		if !ok {
			continue
		}

		// Format: EUR: 4.2345 PLN
		midStr := fmt.Sprintf("%.4f", rate.Mid)
		b.WriteString(fmt.Sprintf("*%s*: %s PLN\n",
			escapeMarkdownV2NBP(code),
			escapeMarkdownV2NBP(midStr)))
	}

	// Legal note about applicability
	b.WriteString("\n")
	b.WriteString("_Kurs sredni NBP z ostatniego dnia roboczego_\n")
	b.WriteString("_poprzedzajacego dzien uzyskania przychodu,_\n")
	b.WriteString("_poniesienia kosztu lub zaplaty podatku\\._")

	return b.String()
}

// escapeMarkdownV2NBP escapes special characters for Telegram MarkdownV2.
// Duplicated locally to keep nbp package free of Telegram dependencies.
//
// Parameters:
//   - text: Text to escape
//
// Returns:
//   - string: Escaped text safe for MarkdownV2
func escapeMarkdownV2NBP(text string) string {
	specialChars := []string{
		"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!",
	}
	result := text
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}
