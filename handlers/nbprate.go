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
// Each currency is fetched via GET /api/exchangerates/rates/A/{code}/
// which returns the most recently published rate (last business day).
//
// Parameters:
//   - bot: Telegram Bot API instance for sending messages
//   - message: Message from Telegram containing button click
func HandleNBPRate(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	slog.Info("NBP rate requested",
		"user_id", message.From.ID,
		"username", message.From.UserName,
		"chat_id", message.Chat.ID)

	// Fetch each currency individually
	var rates []nbp.Rate
	for _, code := range defaultCurrencies {
		rate, err := nbp.GetRate(code)
		if err != nil {
			slog.Error("Failed to fetch NBP rate",
				"error", err,
				"code", code,
				"user_id", message.From.ID,
				"chat_id", message.Chat.ID)
			continue
		}
		rates = append(rates, rate)
	}

	if len(rates) == 0 {
		errMsg := tgbotapi.NewMessage(message.Chat.ID,
			"❌ Failed to fetch NBP exchange rates\\. Please try again later\\.")
		errMsg.ParseMode = "MarkdownV2"

		if _, err := bot.Send(errMsg); err != nil {
			slog.Error("Failed to send NBP error message",
				"error", err, "chat_id", message.Chat.ID)
		}
		return
	}

	text := formatNBPRates(rates)

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
		"count", len(rates))
}

// formatNBPRates formats a list of NBP rates into a MarkdownV2 Telegram message.
//
// Displays:
//   - Effective date from the first rate (all Table A rates share the same date)
//   - Rates for each currency to PLN
//   - Legal note about applicability for tax/income purposes
//
// Parameters:
//   - rates: List of NBP Rate values
//
// Returns:
//   - string: Formatted MarkdownV2 message
func formatNBPRates(rates []nbp.Rate) string {
	var b strings.Builder

	// Header with effective date (use first rate's date)
	effectiveDate := ""
	tableNo := ""
	if len(rates) > 0 {
		effectiveDate = rates[0].EffectiveDate
		tableNo = rates[0].TableNo
	}

	b.WriteString("💱 *NBP Average Exchange Rates \\(Table A\\)*\n")
	b.WriteString(fmt.Sprintf("_Table %s, effective: %s_\n\n",
		escapeMarkdownV2NBP(tableNo),
		escapeMarkdownV2NBP(effectiveDate)))

	// One line per currency: EUR: 4.2345 PLN
	for _, rate := range rates {
		midStr := fmt.Sprintf("%.4f", rate.Mid)
		b.WriteString(fmt.Sprintf("*%s*: %s PLN\n",
			escapeMarkdownV2NBP(rate.Code),
			escapeMarkdownV2NBP(midStr)))
	}

	// Legal note
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
