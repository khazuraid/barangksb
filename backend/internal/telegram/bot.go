package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Bot struct {
	token   string
	chatIDs map[int64]bool
	pool    *pgxpool.Pool
	botInst *bot.Bot
}

func NewBot(token, chatIDsCSV string, pool *pgxpool.Pool) *Bot {
	ids := map[int64]bool{}
	for _, s := range strings.Split(chatIDsCSV, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil || id <= 0 {
			continue
		}
		ids[id] = true
	}
	return &Bot{token: token, chatIDs: ids, pool: pool}
}

func (b *Bot) Start(ctx context.Context) {
	if b.token == "" {
		// try load from DB
		var token string
		b.pool.QueryRow(ctx, `SELECT value FROM app_settings WHERE key='telegram_bot_token'`).Scan(&token)
		if token == "" {
			return
		}
		b.token = token
	}

	var apiURL string
	b.pool.QueryRow(ctx, `SELECT value FROM app_settings WHERE key='telegram_api_url'`).Scan(&apiURL)

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 15 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	opts := []bot.Option{
		bot.WithSkipGetMe(),
		bot.WithHTTPClient(30*time.Second, httpClient),
	}
	apiURL = strings.TrimRight(strings.TrimSpace(apiURL), "/")
	if apiURL != "" {
		opts = append(opts, bot.WithServerURL(apiURL))
	}

	botInst, err := bot.New(b.token, opts...)
	if err != nil {
		slog.Error("telegram bot init failed", "err", err)
		return
	}
	b.botInst = botInst

	// Register text command handlers
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, b.handleStart)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/bantu", bot.MatchTypeExact, b.handleCommand)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/masuk", bot.MatchTypePrefix, b.handleMasuk)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/cari", bot.MatchTypePrefix, b.handleCari)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/stok", bot.MatchTypeExact, b.handleStok)

	// Register callback query handler for inline keyboard buttons
	botInst.RegisterHandler(bot.HandlerTypeCallbackQueryData, "", bot.MatchTypePrefix, b.handleCallback)

	go b.dailyAlert(ctx)

	slog.Info("telegram bot started")
	go botInst.Start(ctx)
}

func (b *Bot) allowed(chatID int64) bool { return b.chatIDs[chatID] }

// ---- inline keyboard builder ----

func (b *Bot) buildMenuKeyboard(ctx context.Context) models.InlineKeyboardMarkup {
	rows, _ := b.pool.Query(ctx, `SELECT command, label FROM telegram_commands WHERE is_menu=TRUE AND active=TRUE ORDER BY sort_order`)
	defer rows.Close()

	var kbRows [][]models.InlineKeyboardButton
	row := []models.InlineKeyboardButton{}
	col := 0
	for rows.Next() {
		var cmd, label string
		rows.Scan(&cmd, &label)
		btn := models.InlineKeyboardButton{
			Text:         label,
			CallbackData: "cmd:" + cmd,
		}
		row = append(row, btn)
		col++
		if col == 2 {
			kbRows = append(kbRows, row)
			row = []models.InlineKeyboardButton{}
			col = 0
		}
	}
	if col > 0 {
		kbRows = append(kbRows, row)
	}

	return models.InlineKeyboardMarkup{InlineKeyboard: kbRows}
}

func (b *Bot) buildBackKeyboard() models.InlineKeyboardMarkup {
	return models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{{
			{Text: "‹ Kembali ke Menu", CallbackData: "cmd:start"},
		}},
	}
}

// ---- handlers ----

func (b *Bot) handleStart(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil || !b.allowed(u.Message.Chat.ID) {
		return
	}
	b.showMenu(ctx, bt, u.Message.Chat.ID, u.Message.ID)
}

// showMenu sends a new message with loading → edits to menu
func (b *Bot) showMenu(ctx context.Context, bt *bot.Bot, chatID int64, replyToID int) {
	// Send loading message
	msg, err := bt.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "⏳ Memuat menu...",
	})
	if err != nil {
		return
	}

	// Small delay for UX
	time.Sleep(400 * time.Millisecond)

	// Get menu intro text from DB
	intro := "📋 *Menu Inventaris Kantor*\n\nPilih menu di bawah:"
	if cmd := b.getCommandByCmd(ctx, "start"); cmd != nil && cmd.Response != "" {
		intro = cmd.Response
	}

	// Edit message to show menu
	_, _ = bt.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   msg.ID,
		Text:        intro,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: b.buildMenuKeyboard(ctx),
	})
}

// showLoadingAndEdit: edit existing message to loading then to content
func (b *Bot) showLoadingAndEdit(ctx context.Context, bt *bot.Bot, chatID int64, messageID int, text string, keyboard models.InlineKeyboardMarkup) {
	// Edit to loading
	_, _ = bt.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: messageID,
		Text:      "⏳ Memuat...",
	})
	time.Sleep(300 * time.Millisecond)
	// Edit to content
	_, _ = bt.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   messageID,
		Text:        text,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: keyboard,
	})
}

// handleCallback handles inline keyboard button presses
func (b *Bot) handleCallback(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.CallbackQuery == nil {
		return
	}
	cb := u.CallbackQuery
	chatID := cb.Message.Message.Chat.ID
	if !b.allowed(chatID) {
		return
	}

	// Answer callback to remove loading state on button
	bt.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: cb.ID,
	})

	data := cb.Data
	if !strings.HasPrefix(data, "cmd:") {
		return
	}
	cmd := strings.TrimPrefix(data, "cmd:")
	msgID := cb.Message.Message.ID

	switch cmd {
	case "start":
		// Back to menu
		intro := "📋 *Menu Inventaris Kantor*\n\nPilih menu di bawah:"
		if c := b.getCommandByCmd(ctx, "start"); c != nil && c.Response != "" {
			intro = c.Response
		}
		b.showLoadingAndEdit(ctx, bt, chatID, msgID, intro, b.buildMenuKeyboard(ctx))

	case "bantu":
		text := b.getCommandResponse(ctx, "bantu")
		b.showLoadingAndEdit(ctx, bt, chatID, msgID, text, b.buildBackKeyboard())

	case "stok":
		text := b.getLowStockText(ctx)
		b.showLoadingAndEdit(ctx, bt, chatID, msgID, text, b.buildBackKeyboard())

	default:
		// Check if it's a DB command
		if c := b.getCommandByCmd(ctx, cmd); c != nil {
			text := c.Response
			// For dynamic commands like masuk/cari, show format hint
			b.showLoadingAndEdit(ctx, bt, chatID, msgID, text, b.buildBackKeyboard())
		} else {
			b.showLoadingAndEdit(ctx, bt, chatID, msgID, "Perintah tidak ditemukan.", b.buildBackKeyboard())
		}
	}
}

// handleCommand handles text commands that have DB entries (bantu, etc)
func (b *Bot) handleCommand(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil || !b.allowed(u.Message.Chat.ID) {
		return
	}
	cmdText := strings.TrimPrefix(u.Message.Text, "/")
	cmdParts := strings.Fields(cmdText)
	if len(cmdParts) == 0 {
		return
	}
	cmd := b.getCommandByCmd(ctx, cmdParts[0])
	if cmd == nil {
		return
	}
	bt.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    u.Message.Chat.ID,
		Text:      cmd.Response,
		ParseMode: models.ParseModeMarkdown,
	})
}

func (b *Bot) handleMasuk(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil || !b.allowed(u.Message.Chat.ID) {
		return
	}
	parts := strings.Fields(u.Message.Text)
	if len(parts) < 3 {
		text := b.getCommandResponse(ctx, "masuk")
		bt.SendMessage(ctx, &bot.SendMessageParams{ChatID: u.Message.Chat.ID, Text: text})
		return
	}
	sku := parts[1]
	qty, err := strconv.ParseInt(parts[2], 10, 32)
	if err != nil || qty <= 0 {
		bt.SendMessage(ctx, &bot.SendMessageParams{ChatID: u.Message.Chat.ID, Text: "Jumlah harus > 0"})
		return
	}

	var name, unit string
	var prev, next int32
	err = b.pool.QueryRow(ctx,
		`UPDATE inventory_items SET current_stock = current_stock + $2, updated_at = now()
		 WHERE sku = $1 RETURNING name, unit, current_stock - $2, current_stock`,
		sku, qty).Scan(&name, &unit, &prev, &next)
	if err != nil {
		bt.SendMessage(ctx, &bot.SendMessageParams{ChatID: u.Message.Chat.ID, Text: "SKU tidak ditemukan: " + sku})
		return
	}

	b.pool.Exec(ctx, `INSERT INTO stock_transactions (type, item_id, item_sku, item_name, quantity, unit, previous_stock, new_stock)
		VALUES ('IN', (SELECT id FROM inventory_items WHERE sku=$1), $1, $2, $3, $4, $5, $6)`,
		sku, name, qty, unit, prev, next)

	bt.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: u.Message.Chat.ID,
		Text:   fmt.Sprintf("✅ %s +%d %s\nStok: %d", name, qty, unit, next),
	})
}

func (b *Bot) handleCari(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil || !b.allowed(u.Message.Chat.ID) {
		return
	}
	kw := strings.TrimSpace(strings.TrimPrefix(u.Message.Text, "/cari"))
	if kw == "" {
		text := b.getCommandResponse(ctx, "cari")
		bt.SendMessage(ctx, &bot.SendMessageParams{ChatID: u.Message.Chat.ID, Text: text})
		return
	}
	rows, _ := b.pool.Query(ctx,
		`SELECT name, sku, current_stock, unit, location FROM inventory_items
		 WHERE name ILIKE $1 OR sku ILIKE $1 OR location ILIKE $1 LIMIT 10`, "%"+kw+"%")
	defer rows.Close()
	var sb strings.Builder
	sb.WriteString("🔍 Hasil cari:\n")
	count := 0
	for rows.Next() {
		var name, sku, unit, loc string
		var stock int32
		rows.Scan(&name, &sku, &stock, &unit, &loc)
		sb.WriteString(fmt.Sprintf("• %s (%s) %d %s @ %s\n", name, sku, stock, unit, loc))
		count++
	}
	if count == 0 {
		sb.WriteString("Tidak ditemukan: " + kw)
	}
	bt.SendMessage(ctx, &bot.SendMessageParams{ChatID: u.Message.Chat.ID, Text: sb.String()})
}

func (b *Bot) handleStok(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil || !b.allowed(u.Message.Chat.ID) {
		return
	}
	text := b.getLowStockText(ctx)
	bt.SendMessage(ctx, &bot.SendMessageParams{ChatID: u.Message.Chat.ID, Text: text})
}

// ---- helpers ----

func (b *Bot) getCommandByCmd(ctx context.Context, command string) *dbCommand {
	var c dbCommand
	err := b.pool.QueryRow(ctx,
		`SELECT id, command, label, description, response, is_menu, sort_order, active FROM telegram_commands WHERE command=$1 AND active=TRUE`,
		command).Scan(&c.ID, &c.Command, &c.Label, &c.Description, &c.Response, &c.IsMenu, &c.SortOrder, &c.Active)
	if err != nil {
		return nil
	}
	return &c
}

func (b *Bot) getCommandResponse(ctx context.Context, command string) string {
	if c := b.getCommandByCmd(ctx, command); c != nil {
		return c.Response
	}
	return "Perintah tidak ditemukan."
}

func (b *Bot) getLowStockText(ctx context.Context) string {
	rows, _ := b.pool.Query(ctx,
		`SELECT name, sku, current_stock, min_stock, unit FROM inventory_items WHERE current_stock <= min_stock ORDER BY current_stock LIMIT 20`)
	defer rows.Close()
	var sb strings.Builder
	sb.WriteString("⚠️ *Stok Menipis:*\n\n")
	count := 0
	for rows.Next() {
		var name, sku, unit string
		var cur, min int32
		rows.Scan(&name, &sku, &cur, &min, &unit)
		sb.WriteString(fmt.Sprintf("• %s (%s) %d/%d %s\n", name, sku, cur, min, unit))
		count++
	}
	if count == 0 {
		sb.WriteString("✅ Semua stok aman.")
	}
	return sb.String()
}

// ---- daily alert ----

func (b *Bot) dailyAlert(ctx context.Context) {
	loc := time.FixedZone("WIB", 7*3600)
	for {
		now := time.Now().In(loc)
		next := time.Date(now.Year(), now.Month(), now.Day(), 7, 0, 0, 0, loc)
		if next.Before(now) {
			next = next.Add(24 * time.Hour)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(next.Sub(now)):
		}
		b.sendLowStockAlert(ctx)
	}
}

func (b *Bot) sendLowStockAlert(ctx context.Context) {
	text := b.getLowStockText(ctx)
	rows, _ := b.pool.Query(ctx, `SELECT chat_id, title FROM telegram_subscribers WHERE active=TRUE AND notify_low_stock=TRUE`)
	defer rows.Close()
	type target struct {
		chatID int64
		title  string
	}
	var targets []target
	for rows.Next() {
		var t target
		rows.Scan(&t.chatID, &t.title)
		targets = append(targets, t)
	}
	if len(targets) == 0 || b.botInst == nil {
		return
	}
	for _, t := range targets {
		ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
		_, err := b.botInst.SendMessage(ctx2, &bot.SendMessageParams{
			ChatID:    t.chatID,
			Text:      text,
			ParseMode: models.ParseModeMarkdown,
		})
		status := "ok"
		errMsg := ""
		if err != nil {
			status = "fail"
			errMsg = err.Error()
		}
		b.pool.Exec(ctx, `INSERT INTO telegram_message_log (chat_id, chat_title, message, status, error) VALUES ($1,$2,$3,$4,$5)`,
			t.chatID, t.title, text, status, errMsg)
		cancel()
	}
}

// ---- types ----

type dbCommand struct {
	ID          int    `json:"id"`
	Command     string `json:"command"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Response    string `json:"response"`
	IsMenu      bool   `json:"is_menu"`
	SortOrder   int    `json:"sort_order"`
	Active      bool   `json:"active"`
}