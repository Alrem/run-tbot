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

// nbpCurrencies lists the currencies available for selection in step 4.
var nbpCurrencies = []string{"EUR", "USD", "GBP", "CHF", "UAH"}

var monthNames = [13]string{
	"", // 1-based index
	"Jan", "Feb", "Mar", "Apr", "May", "Jun",
	"Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
}

// HandleNBPRate handles the "💱 Kurs NBP" button click.
// Step 1 of the flow: show the year selection keyboard.
//
// Full flow:
//  1. Button click → year inline keyboard  (this function)
//  2. Year selected → month inline keyboard (handleNBPYearSelected)
//  3. Month selected → day inline keyboard  (handleNBPMonthSelected)
//  4. Day selected → ask for amount         (handleNBPDaySelected)
//  5. Amount typed → currency keyboard      (HandleNBPConvInput)
//  6. Currency selected → result            (handleNBPCurrencySelected)
func HandleNBPRate(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	// Clear any previous in-progress conversation
	clearConvState(message.Chat.ID)

	slog.Info("NBP conversion started",
		"user_id", message.From.ID,
		"chat_id", message.Chat.ID)

	msg := tgbotapi.NewMessage(message.Chat.ID, "📅 Select the income/expense *year*:")
	msg.ParseMode = "MarkdownV2"
	msg.ReplyMarkup = buildYearKeyboard()

	if _, err := bot.Send(msg); err != nil {
		slog.Error("Failed to send NBP year keyboard", "error", err, "chat_id", message.Chat.ID)
	}
}

// HandleNBPCallback is the single entry point for all NBP-related callback queries.
// Dispatches to the appropriate handler based on callback_data prefix.
func HandleNBPCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	data := callback.Data
	switch {
	case strings.HasPrefix(data, "nbp_y:"):
		handleNBPYearSelected(bot, callback)
	case strings.HasPrefix(data, "nbp_m:"):
		handleNBPMonthSelected(bot, callback)
	case strings.HasPrefix(data, "nbp_d:"):
		handleNBPDaySelected(bot, callback)
	case data == "nbp_by":
		handleNBPBackToYear(bot, callback)
	case strings.HasPrefix(data, "nbp_bm:"):
		handleNBPBackToMonth(bot, callback)
	case data == "nbp_x":
		handleNBPCancel(bot, callback)
	case strings.HasPrefix(data, "nbp:"):
		handleNBPCurrencySelected(bot, callback)
	default:
		slog.Warn("Unhandled NBP callback", "data", data)
	}
}

// handleNBPYearSelected handles year button click.
// Edits the message to show month selection for the chosen year.
// callback.Data format: "nbp_y:{year}"
func handleNBPYearSelected(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	ackCallback(bot, callback)

	year, err := strconv.Atoi(strings.TrimPrefix(callback.Data, "nbp_y:"))
	if err != nil {
		return
	}

	editText(bot, callback,
		fmt.Sprintf("📅 Year: *%d* — select month:", year),
		buildMonthKeyboard(year))
}

// handleNBPMonthSelected handles month button click.
// Edits the message to show day selection for the chosen year+month.
// callback.Data format: "nbp_m:{year}:{month}"
func handleNBPMonthSelected(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	ackCallback(bot, callback)

	parts := strings.SplitN(strings.TrimPrefix(callback.Data, "nbp_m:"), ":", 2)
	if len(parts) != 2 {
		return
	}
	year, err1 := strconv.Atoi(parts[0])
	month, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || month < 1 || month > 12 {
		return
	}

	editText(bot, callback,
		fmt.Sprintf("📅 %s %d — select day:", monthNames[month], year),
		buildDayKeyboard(year, month))
}

// handleNBPDaySelected handles day button click.
// Stores the selected date in conversation state and asks for the amount.
// callback.Data format: "nbp_d:{year}:{month}:{day}"
func handleNBPDaySelected(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	ackCallback(bot, callback)

	parts := strings.SplitN(strings.TrimPrefix(callback.Data, "nbp_d:"), ":", 3)
	if len(parts) != 3 {
		return
	}
	year, err1 := strconv.Atoi(parts[0])
	month, err2 := strconv.Atoi(parts[1])
	day, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return
	}

	date := fmt.Sprintf("%04d-%02d-%02d", year, month, day)
	chatID := callback.Message.Chat.ID

	// Save date in conversation state; next text message = amount
	setConvState(chatID, &convState{Step: convStepNBPAmount, Date: date})

	// Update the date picker message to show the confirmed date
	editTextOnly(bot, callback,
		fmt.Sprintf("📅 Date selected: *%s*", escapeMarkdownV2NBP(date)))

	// Ask for amount
	msg := tgbotapi.NewMessage(chatID, "💰 Enter the amount:")
	msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true, Selective: true}

	if _, err := bot.Send(msg); err != nil {
		slog.Error("Failed to send NBP amount prompt", "error", err, "chat_id", chatID)
		clearConvState(chatID)
	}
}

// handleNBPBackToYear handles the "◀ Back" button from month selection.
// Edits the message back to the year keyboard.
// callback.Data: "nbp_by"
func handleNBPBackToYear(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	ackCallback(bot, callback)
	editText(bot, callback, "📅 Select the income/expense *year*:", buildYearKeyboard())
}

// handleNBPBackToMonth handles the "◀ Back" button from day selection.
// Edits the message back to the month keyboard for the given year.
// callback.Data format: "nbp_bm:{year}"
func handleNBPBackToMonth(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	ackCallback(bot, callback)

	year, err := strconv.Atoi(strings.TrimPrefix(callback.Data, "nbp_bm:"))
	if err != nil {
		return
	}

	editText(bot, callback,
		fmt.Sprintf("📅 Year: *%d* — select month:", year),
		buildMonthKeyboard(year))
}

// handleNBPCancel handles the "❌ Cancel" button.
// Removes the keyboard and clears any active conversation state.
// callback.Data: "nbp_x"
func handleNBPCancel(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	ackCallback(bot, callback)
	clearConvState(callback.Message.Chat.ID)
	editTextOnly(bot, callback, "❌ Cancelled\\.")
}

// HandleNBPConvInput routes text input during the NBP amount step.
// Called by the router when a message arrives and there is an active convState.
func HandleNBPConvInput(bot *tgbotapi.BotAPI, message *tgbotapi.Message, state *convState) {
	if state.Step == convStepNBPAmount {
		handleNBPCurrencyStep(bot, message, state)
		return
	}
	clearConvState(message.Chat.ID)
}

// handleNBPCurrencyStep validates the amount and shows the currency keyboard.
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

	// Conversation state no longer needed - date+amount are encoded in callback_data
	clearConvState(message.Chat.ID)

	// Build inline keyboard: each button encodes date, amount, and currency
	// Format: "nbp:{date}:{amount}:{currency}"
	var buttons []tgbotapi.InlineKeyboardButton
	for _, code := range nbpCurrencies {
		callbackData := fmt.Sprintf("nbp:%s:%.2f:%s", state.Date, amount, code)
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(code, callbackData))
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(buttons...),
	)

	msg := tgbotapi.NewMessage(message.Chat.ID,
		fmt.Sprintf("💱 Select currency for *%s* \\(date: %s\\):",
			escapeMarkdownV2NBP(formatAmount(amount)),
			escapeMarkdownV2NBP(state.Date)))
	msg.ParseMode = "MarkdownV2"
	msg.ReplyMarkup = keyboard

	if _, err := bot.Send(msg); err != nil {
		slog.Error("Failed to send NBP currency keyboard", "error", err, "chat_id", message.Chat.ID)
	}
}

// handleNBPCurrencySelected fetches the NBP rate and shows the conversion result.
// callback.Data format: "nbp:{date}:{amount}:{currency}"
func handleNBPCurrencySelected(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	ackCallback(bot, callback)

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
		slog.Error("Failed to fetch NBP rate",
			"error", err, "date", date, "currency", currency, "chat_id", chatID)

		errMsg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("❌ No NBP rate found for *%s* before %s\\.\n_The date may be too old or there is a network error\\._",
				currency, escapeMarkdownV2NBP(date)))
		errMsg.ParseMode = "MarkdownV2"
		bot.Send(errMsg) //nolint:errcheck
		return
	}

	// Remove inline keyboard from the currency selection message
	editTextOnly(bot, callback,
		fmt.Sprintf("💱 Converting *%s %s* for date %s\\.\\.\\.",
			escapeMarkdownV2NBP(formatAmount(amount)),
			currency,
			escapeMarkdownV2NBP(date)))

	result := amount * rate.Mid

	msg := tgbotapi.NewMessage(chatID, formatNBPResult(date, amount, currency, rate, result))
	msg.ParseMode = "MarkdownV2"

	if _, err := bot.Send(msg); err != nil {
		slog.Error("Failed to send NBP result", "error", err, "chat_id", chatID)
	}
}

// --- Keyboard builders ---

// buildYearKeyboard returns an inline keyboard with the current year and 3 previous years.
func buildYearKeyboard() tgbotapi.InlineKeyboardMarkup {
	currentYear := time.Now().Year()

	var buttons []tgbotapi.InlineKeyboardButton
	for i := 0; i < 4; i++ {
		year := currentYear - i
		buttons = append(buttons,
			tgbotapi.NewInlineKeyboardButtonData(
				strconv.Itoa(year),
				fmt.Sprintf("nbp_y:%d", year),
			))
	}

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(buttons...),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", "nbp_x"),
		),
	)
}

// buildMonthKeyboard returns an inline keyboard with all 12 months for the given year.
func buildMonthKeyboard(year int) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// 3 months per row → 4 rows
	for startMonth := 1; startMonth <= 12; startMonth += 3 {
		var row []tgbotapi.InlineKeyboardButton
		for m := startMonth; m < startMonth+3 && m <= 12; m++ {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(
				monthNames[m],
				fmt.Sprintf("nbp_m:%d:%d", year, m),
			))
		}
		rows = append(rows, row)
	}

	// Navigation row
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("◀ Back", "nbp_by"),
		tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", "nbp_x"),
	))

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// buildDayKeyboard returns an inline keyboard with all days of the given month.
func buildDayKeyboard(year, month int) tgbotapi.InlineKeyboardMarkup {
	// time.Date with day=0 of next month gives last day of current month
	daysInMonth := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC).Day()

	var rows [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton

	for day := 1; day <= daysInMonth; day++ {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(
			strconv.Itoa(day),
			fmt.Sprintf("nbp_d:%d:%d:%d", year, month, day),
		))
		// 7 days per row (like a calendar week)
		if len(row) == 7 || day == daysInMonth {
			rows = append(rows, row)
			row = nil
		}
	}

	// Navigation row
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("◀ Back", fmt.Sprintf("nbp_bm:%d", year)),
		tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", "nbp_x"),
	))

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// --- Formatting helpers ---

// formatNBPResult builds the final MarkdownV2 result message.
func formatNBPResult(incomeDate string, amount float64, currency string, rate nbp.Rate, result float64) string {
	var b strings.Builder

	b.WriteString("💱 *NBP Currency Conversion*\n\n")

	b.WriteString(fmt.Sprintf("Amount: *%s %s*\n",
		escapeMarkdownV2NBP(formatAmount(amount)),
		escapeMarkdownV2NBP(currency)))
	b.WriteString(fmt.Sprintf("Income/expense date: *%s*\n\n",
		escapeMarkdownV2NBP(incomeDate)))

	b.WriteString(fmt.Sprintf("NBP rate \\(Table %s, %s\\): *%s PLN*\n\n",
		escapeMarkdownV2NBP(rate.TableNo),
		escapeMarkdownV2NBP(rate.EffectiveDate),
		escapeMarkdownV2NBP(fmt.Sprintf("%.4f", rate.Mid))))

	b.WriteString(fmt.Sprintf("*Result: %s PLN*\n\n",
		escapeMarkdownV2NBP(formatAmount(result))))

	b.WriteString("_Kurs sredni NBP z ostatniego dnia roboczego_\n")
	b.WriteString("_poprzedzajacego dzien uzyskania przychodu\\._")

	return b.String()
}

// formatAmount formats a float with up to 2 decimal places, trimming trailing zeros.
func formatAmount(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

// escapeMarkdownV2NBP escapes special characters for Telegram MarkdownV2.
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

// --- Telegram API helpers ---

// ackCallback answers a callback query to remove the loading indicator on the button.
func ackCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	if _, err := bot.Request(tgbotapi.NewCallback(callback.ID, "")); err != nil {
		slog.Error("Failed to answer callback query", "error", err)
	}
}

// editText edits a callback message's text and inline keyboard.
func editText(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	edit := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)
	edit.ParseMode = "MarkdownV2"
	if _, err := bot.Request(edit); err != nil {
		slog.Error("Failed to edit message", "error", err, "chat_id", callback.Message.Chat.ID)
	}
}

// editTextOnly edits a callback message's text and removes the inline keyboard.
func editTextOnly(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery, text string) {
	edit := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
	)
	edit.ParseMode = "MarkdownV2"
	if _, err := bot.Request(edit); err != nil {
		slog.Error("Failed to edit message text", "error", err, "chat_id", callback.Message.Chat.ID)
	}
}
