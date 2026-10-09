package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
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
	if apiURL == "" {
		apiURL = os.Getenv("TELEGRAM_API_URL")
	}

	httpClient := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	opts := []bot.Option{
		bot.WithSkipGetMe(),
		bot.WithHTTPClient(30*time.Second, httpClient),
		bot.WithErrorsHandler(func(err error) {
			if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline") {
				slog.Debug("telegram polling timeout", "err", err)
				return
			}
			slog.Warn("telegram bot error", "err", err)
		}),
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

	// Ensure standard commands show in menu if DB has old seed
	_, _ = b.pool.Exec(ctx, `UPDATE telegram_commands SET is_menu=TRUE WHERE command IN ('stok', 'cari', 'masuk', 'bantu') AND (SELECT count(*) FROM telegram_commands WHERE is_menu=TRUE AND command != 'start') = 0`)

	// Set Telegram Mini App Menu Button
	webURL := b.getWebAppURL(ctx)
	if webURL != "" {
		_, _ = botInst.SetChatMenuButton(ctx, &bot.SetChatMenuButtonParams{
			MenuButton: &models.MenuButtonWebApp{
				Type: models.MenuButtonTypeWebApp,
				Text: "Buka Inventaris",
				WebApp: models.WebAppInfo{
					URL: webURL,
				},
			},
		})
	}

	slog.Info("telegram bot started")
	go botInst.Start(ctx)
}

func (b *Bot) getWebAppURL(ctx context.Context) string {
	var u string
	_ = b.pool.QueryRow(ctx, `SELECT value FROM app_settings WHERE key='telegram_webapp_url' OR key='app_url' LIMIT 1`).Scan(&u)
	u = strings.TrimSpace(u)
	if u == "" {
		u = strings.TrimSpace(os.Getenv("WEBAPP_URL"))
	}
	if u == "" {
		u = strings.TrimSpace(os.Getenv("APP_URL"))
	}
	if u == "" {
		u = "https://barang.kesling.biz.id"
	}
	return u
}

func (b *Bot) allowed(chatID int64) bool {
	if b.chatIDs[chatID] {
		return true
	}
	var exists bool
	_ = b.pool.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM telegram_subscribers WHERE chat_id=$1 AND active=TRUE)`,
		chatID).Scan(&exists)
	if exists {
		return true
	}
	if len(b.chatIDs) == 0 {
		var count int
		_ = b.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM telegram_subscribers WHERE active=TRUE`).Scan(&count)
		if count == 0 {
			return true
		}
	}
	return false
}

// ---- inline keyboard builder ----

func (b *Bot) buildMenuKeyboard(ctx context.Context) models.InlineKeyboardMarkup {
	rows, _ := b.pool.Query(ctx, `SELECT command, label FROM telegram_commands WHERE is_menu=TRUE AND active=TRUE AND command != 'start' ORDER BY sort_order`)
	defer rows.Close()

	var kbRows [][]models.InlineKeyboardButton
	row := []models.InlineKeyboardButton{}
	col := 0
	for rows != nil && rows.Next() {
		var cmd, label string
		if err := rows.Scan(&cmd, &label); err == nil {
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
	}
	if col > 0 {
		kbRows = append(kbRows, row)
	}

	// Fallback menu buttons if DB has none configured
	if len(kbRows) == 0 {
		kbRows = [][]models.InlineKeyboardButton{
			{
				{Text: "📦 Stok Menipis", CallbackData: "cmd:stok"},
				{Text: "🔍 Cari Barang", CallbackData: "cmd:cari"},
			},
			{
				{Text: "📥 Barang Masuk", CallbackData: "cmd:masuk"},
				{Text: "❓ Bantuan", CallbackData: "cmd:bantu"},
			},
		}
	}

	// Add Web App button at top
	webURL := b.getWebAppURL(ctx)
	if webURL != "" {
		kbRows = append([][]models.InlineKeyboardButton{{
			{
				Text:   "🌐 Buka Aplikasi Web",
				WebApp: &models.WebAppInfo{URL: webURL},
			},
		}}, kbRows...)
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
	if u.Message == nil {
		return
	}
	chatID := u.Message.Chat.ID
	if !b.allowed(chatID) {
		name := u.Message.Chat.FirstName
		if name == "" {
			name = u.Message.Chat.Title
		}
		if name == "" {
			name = "Pengguna"
		}
		text := fmt.Sprintf("👋 *Halo, %s!*\n\n🆔 *Chat ID Anda:* `%d`\n\nAkun Anda belum terdaftar sebagai subscriber aktif di sistem Inventaris Puskesmas.\n\nSilakan salin Chat ID di atas dan tambahkan pada menu *Administrasi > Telegram* di web panel untuk mendapatkan notifikasi dan akses menu bot.", name, chatID)
		_, _ = bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      text,
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}
	b.showMenu(ctx, bt, chatID, u.Message.ID)
}

// showMenu sends the interactive menu directly to the chat
func (b *Bot) showMenu(ctx context.Context, bt *bot.Bot, chatID int64, replyToID int) {
	intro := "📋 *Menu Inventaris Kantor*\n\nPilih menu di bawah:"
	if cmd := b.getCommandByCmd(ctx, "start"); cmd != nil && strings.TrimSpace(cmd.Response) != "" {
		intro = cmd.Response
	}

	kb := b.buildMenuKeyboard(ctx)
	_, err := bt.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        intro,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: kb,
	})
	if err != nil {
		// Fallback without Markdown if parsing fails
		_, _ = bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        intro,
			ReplyMarkup: kb,
		})
	}
}

// editMessage updates an existing message with markdown fallback
func (b *Bot) editMessage(ctx context.Context, bt *bot.Bot, chatID int64, messageID int, text string, keyboard models.InlineKeyboardMarkup) {
	_, err := bt.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   messageID,
		Text:        text,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: keyboard,
	})
	if err != nil {
		_, _ = bt.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:      chatID,
			MessageID:   messageID,
			Text:        text,
			ReplyMarkup: keyboard,
		})
	}
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
		if c := b.getCommandByCmd(ctx, "start"); c != nil && strings.TrimSpace(c.Response) != "" {
			intro = c.Response
		}
		b.editMessage(ctx, bt, chatID, msgID, intro, b.buildMenuKeyboard(ctx))

	case "bantu":
		text := b.getCommandResponse(ctx, "bantu")
		b.editMessage(ctx, bt, chatID, msgID, text, b.buildBackKeyboard())

	case "stok":
		text := b.getLowStockText(ctx)
		b.editMessage(ctx, bt, chatID, msgID, text, b.buildBackKeyboard())

	case "cari":
		text := "🔍 *Pencarian Barang*\n\nKetik `/cari <nama barang>` di chat.\nContoh:\n`/cari tensimeter`\n`/cari paracetamol`"
		b.editMessage(ctx, bt, chatID, msgID, text, b.buildBackKeyboard())

	case "masuk":
		text := "📥 *Catat Barang Masuk*\n\nKetik `/masuk <SKU> <jumlah>` di chat.\nContoh:\n`/masuk MED-2026-001 50`"
		b.editMessage(ctx, bt, chatID, msgID, text, b.buildBackKeyboard())

	default:
		// Check if it's a DB command
		if c := b.getCommandByCmd(ctx, cmd); c != nil {
			text := c.Response
			b.editMessage(ctx, bt, chatID, msgID, text, b.buildBackKeyboard())
		} else {
			b.editMessage(ctx, bt, chatID, msgID, "Perintah tidak ditemukan.", b.buildBackKeyboard())
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