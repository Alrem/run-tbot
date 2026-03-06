package handlers

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/Alrem/run-tbot/nbp"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// nbpCurrencies lists the currencies available for selection in the inline keyboard.
var nbpCurrencies = []string{"EUR", "USD", "GBP", "CHF"}

// HandleNBPRate handles the "💱 Kurs NBP" button click.
// This is step 1 of the multi-step NBP conversion flow.
//
// Flow:
//  1. User clicks button → bot asks for date (this function)
//  2. User types date → HandleNBPConvInput routes to handleNBPAmountStep
//  3. User types amount → handleNBPAmountStep shows currency keyboard
//  4. User clicks currency → HandleNBPCurrencyCallback computes result
//
// Parameters:
//   - bot: Telegram Bot API instance
//   - message: Message from button click
func HandleNBPRate(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	// Start (or restart) the conversation
	setConvState(message.Chat.ID, &convState{Step: convStepNBPDate})

	slog.Info("NBP conversion started",
		"user_id", message.From.ID,
		"chat_id", message.Chat.ID)

	today := time.Now().Format("2006-01-02")

	msg := tgbotapi.NewMessage(message.Chat.ID,
		fmt.Sprintf("📅 Enter the income/expense date \\(YYYY\\-MM\\-DD\\):\n_e\\.g\\. %s_",
			escapeMarkdownV2NBP(today)))
	msg.ParseMode = "MarkdownV2"
	// ForceReply prompts the user to reply directly to this message
	msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true, Selective: true}

	if _, err := bot.Send(msg); err != nil {
		slog.Error("Failed to send NBP date prompt",
			"error", err, "chat_id", message.Chat.ID)
		clearConvState(message.Chat.ID)
	}
}

// HandleNBPConvInput routes text input during an active NBP conversation.
// Called by the router when a message arrives and there is an active convState.
//
// Parameters:
//   - bot: Telegram Bot API instance
//   - message: User's text message
//   - state: Current conversation state
func HandleNBPConvInput(bot *tgbotapi.BotAPI, message *tgbotapi.Message, state *convState) {
	switch state.Step {
	case convStepNBPDate:
		handleNBPAmountStep(bot, message, state)
	case convStepNBPAmount:
		handleNBPCurrencyStep(bot, message, state)
	default:
		clearConvState(message.Chat.ID)
	}
}

// handleNBPAmountStep validates the date input and asks for the amount.
// This is step 2 of the flow.
func handleNBPAmountStep(bot *tgbotapi.BotAPI, message *tgbotapi.Message, state *convState) {
	dateStr := strings.TrimSpace(message.Text)

	// Validate date format
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID,
			"❌ Invalid date format\\. Please use *YYYY\\-MM\\-DD*:")
		msg.ParseMode = "MarkdownV2"
		msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true, Selective: true}

		if _, err := bot.Send(msg); err != nil {
			slog.Error("Failed to send date validation error", "error", err, "chat_id", message.Chat.ID)
		}
		return
	}

	// Save date and move to next step
	state.Date = dateStr
	state.Step = convStepNBPAmount
	setConvState(message.Chat.ID, state)

	msg := tgbotapi.NewMessage(message.Chat.ID, "💰 Enter the amount:")
	msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true, Selective: true}

	if _, err := bot.Send(msg); err != nil {
		slog.Error("Failed to send NBP amount prompt", "error", err, "chat_id", message.Chat.ID)
		clearConvState(message.Chat.ID)
	}
}

// handleNBPCurrencyStep validates the amount input and shows the currency keyboard.
// This is step 3 of the flow.
func handleNBPCurrencyStep(bot *tgbotapi.BotAPI, message *tgbotapi.Message, state *convState) {
	// Accept both dot and comma as decimal separator
	amountStr := strings.ReplaceAll(strings.TrimSpace(message.Text), ",", ".")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID,
			"❌ Invalid amount\\. Please enter a positive number:")
		msg.ParseMode = "MarkdownV2"
		msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true, Selective: true}

		if _, err := bot.Send(msg); err != nil {
			slog.Error("Failed to send amount validation error", "error", err, "chat_id", message.Chat.ID)
		}
		return
	}

	// Conversation state no longer needed - currency comes via callback_data
	clearConvState(message.Chat.ID)

	// Build inline keyboard with currency buttons.
	// Date and amount are encoded in callback_data so no server-side state is needed
	// after this point. Format: "nbp:{date}:{amount}:{currency}"
	var buttons []tgbotapi.InlineKeyboardButton
	for _, code := range nbpCurrencies {
		callbackData := fmt.Sprintf("nbp:%s:%.2f:%s", state.Date, amount, code)
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(code, callbackData))
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(buttons...),
	)

	amountFormatted := formatAmount(amount)
	msg := tgbotapi.NewMessage(message.Chat.ID,
		fmt.Sprintf("💱 Select currency for *%s* \\(date: %s\\):",
			escapeMarkdownV2NBP(amountFormatted),
			escapeMarkdownV2NBP(state.Date)))
	msg.ParseMode = "MarkdownV2"
	msg.ReplyMarkup = keyboard

	if _, err := bot.Send(msg); err != nil {
		slog.Error("Failed to send NBP currency keyboard", "error", err, "chat_id", message.Chat.ID)
	}
}

// HandleNBPCurrencyCallback handles the inline keyboard currency selection.
// This is step 4 (final step) of the flow.
//
// callback.Data format: "nbp:{date}:{amount}:{currency}"
// Example: "nbp:2026-03-06:1000.00:EUR"
//
// Parameters:
//   - bot: Telegram Bot API instance
//   - callback: CallbackQuery from inline keyboard button click
func HandleNBPCurrencyCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	// Acknowledge callback immediately to remove loading state on button
	ack := tgbotapi.NewCallback(callback.ID, "")
	if _, err := bot.Request(ack); err != nil {
		slog.Error("Failed to answer callback query", "error", err)
	}

	// Parse callback data: "nbp:{date}:{amount}:{currency}"
	parts := strings.SplitN(callback.Data, ":", 4)
	if len(parts) != 4 {
		slog.Error("Invalid NBP callback data", "data", callback.Data)
		return
	}

	date := parts[1]
	amount, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		slog.Error("Invalid amount in NBP callback data", "data", callback.Data)
		return
	}
	currency := parts[3]

	chatID := callback.Message.Chat.ID

	slog.Info("NBP currency selected",
		"user_id", callback.From.ID,
		"chat_id", chatID,
		"date", date,
		"amount", amount,
		"currency", currency)

	// Fetch NBP rate for the last business day before the given date
	rate, err := nbp.GetRateForPreviousBizDay(currency, date)
	if err != nil {
		slog.Error("Failed to fetch NBP rate for previous biz day",
			"error", err,
			"date", date,
			"currency", currency,
			"chat_id", chatID)

		errMsg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("❌ No NBP rate found for *%s* before %s\\.\nThe date may be too old or there is a network error\\.",
				currency, escapeMarkdownV2NBP(date)))
		errMsg.ParseMode = "MarkdownV2"
		bot.Send(errMsg) //nolint:errcheck
		return
	}

	// Compute converted amount
	result := amount * rate.Mid

	// Format result message
	text := formatNBPResult(date, amount, currency, rate, result)

	// Remove inline keyboard from the original message to keep chat clean
	editMarkup := tgbotapi.NewEditMessageReplyMarkup(
		chatID,
		callback.Message.MessageID,
		tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{}},
	)
	bot.Request(editMarkup) //nolint:errcheck

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "MarkdownV2"

	if _, err := bot.Send(msg); err != nil {
		slog.Error("Failed to send NBP result", "error", err, "chat_id", chatID)
	}
}

// formatNBPResult builds the final MarkdownV2 result message.
func formatNBPResult(incomeDate string, amount float64, currency string, rate nbp.Rate, result float64) string {
	var b strings.Builder

	b.WriteString("💱 *NBP Currency Conversion*\n\n")

	// Input
	b.WriteString(fmt.Sprintf("Amount: *%s %s*\n",
		escapeMarkdownV2NBP(formatAmount(amount)),
		escapeMarkdownV2NBP(currency)))
	b.WriteString(fmt.Sprintf("Income/expense date: *%s*\n\n",
		escapeMarkdownV2NBP(incomeDate)))

	// Applicable rate
	b.WriteString(fmt.Sprintf("NBP rate \\(Table %s, %s\\): *%s PLN*\n\n",
		escapeMarkdownV2NBP(rate.TableNo),
		escapeMarkdownV2NBP(rate.EffectiveDate),
		escapeMarkdownV2NBP(fmt.Sprintf("%.4f", rate.Mid))))

	// Result
	b.WriteString(fmt.Sprintf("*Result: %s PLN*\n\n",
		escapeMarkdownV2NBP(formatAmount(result))))

	// Legal note
	b.WriteString("_Kurs sredni NBP z ostatniego dnia roboczego_\n")
	b.WriteString("_poprzedzajacego dzien uzyskania przychodu\\._")

	return b.String()
}

// formatAmount formats a float with up to 2 decimal places, trimming trailing zeros.
// Examples: 1000.00 -> "1000", 100.50 -> "100.50", 99.99 -> "99.99"
func formatAmount(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

// escapeMarkdownV2NBP escapes special characters for Telegram MarkdownV2.
// Kept local to avoid importing Telegram packages into the nbp package.
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
