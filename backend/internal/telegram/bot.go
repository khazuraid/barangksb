package telegram

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"image/png"
	"log/slog"
	"math"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/oned"
	qrcode "github.com/makiuchi-d/gozxing/qrcode"
)

type wizardState struct {
	Step      string // "name", "category", "location", "stock"
	Name      string
	Category  string
	Location  string
	Stock     int32
	Unit      string
	UpdatedAt time.Time
}

type Bot struct {
	token       string
	chatIDs     map[int64]bool
	pool        *pgxpool.Pool
	botInst     *bot.Bot
	wizards     map[int64]*wizardState
	wizardMutex sync.RWMutex
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
	return &Bot{
		token:   token,
		chatIDs: ids,
		pool:    pool,
		wizards: make(map[int64]*wizardState),
	}
}

func (b *Bot) Start(ctx context.Context) {
	if b.token == "" {
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
		bot.WithDefaultHandler(b.handleDefault),
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

	// Register specific slash command handlers
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, b.handleStart)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/bantu", bot.MatchTypeExact, b.handleBantu)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/batal", bot.MatchTypeExact, b.handleBatal)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/tambah", bot.MatchTypePrefix, b.handleTambah)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/masuk", bot.MatchTypePrefix, b.handleMasuk)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/keluar", bot.MatchTypePrefix, b.handleKeluar)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/batch_masuk", bot.MatchTypePrefix, b.handleBatchMasuk)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/batch_keluar", bot.MatchTypePrefix, b.handleBatchKeluar)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/pindah", bot.MatchTypePrefix, b.handlePindah)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/opname", bot.MatchTypePrefix, b.handleOpname)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/label", bot.MatchTypePrefix, b.handleLabel)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/cari", bot.MatchTypePrefix, b.handleCari)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/stok", bot.MatchTypeExact, b.handleStok)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/rusak", bot.MatchTypeExact, b.handleRusak)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/prediksi", bot.MatchTypeExact, b.handlePrediksi)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/rekap", bot.MatchTypeExact, b.handleRekap)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/rekap_mutasi", bot.MatchTypeExact, b.handleRekapMutasi)

	// Register callback query handler for inline buttons
	botInst.RegisterHandler(bot.HandlerTypeCallbackQueryData, "", bot.MatchTypePrefix, b.handleCallback)

	// Register general text handler for conversational wizards, chat mutations, and search
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "", bot.MatchTypePrefix, b.handleDefault)

	go b.dailyAlert(ctx)
	go func() {
		_ = b.SyncBotCommands(ctx)
	}()

	slog.Info("telegram bot started")
	go botInst.Start(ctx)
}

func (b *Bot) getWebAppURL(ctx context.Context) string {
	var u string
	_ = b.pool.QueryRow(ctx, `SELECT value FROM app_settings WHERE key='telegram_webapp_url' AND value != '' LIMIT 1`).Scan(&u)
	if u == "" {
		_ = b.pool.QueryRow(ctx, `SELECT value FROM app_settings WHERE key='app_url' AND value != '' LIMIT 1`).Scan(&u)
	}
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
	if chatID == 0 {
		return false
	}
	if b.chatIDs[chatID] {
		return true
	}
	var exists, active bool
	err := b.pool.QueryRow(context.Background(),
		`SELECT TRUE, active FROM telegram_subscribers WHERE chat_id=$1`,
		chatID).Scan(&exists, &active)
	if err == nil && exists {
		return active
	}
	if len(b.chatIDs) == 0 {
		return true
	}
	return false
}

func (b *Bot) autoRegisterSubscriber(ctx context.Context, chat *models.Chat, from *models.User) {
	var chatID int64
	chatType := "private"
	chatTitle := ""
	chatUsername := ""

	if chat != nil {
		chatID = chat.ID
		chatType = string(chat.Type)
		if chatType == "" {
			chatType = "private"
		}
		chatTitle = chat.Title
		if chatTitle == "" {
			parts := []string{}
			if chat.FirstName != "" {
				parts = append(parts, chat.FirstName)
			}
			if chat.LastName != "" {
				parts = append(parts, chat.LastName)
			}
			chatTitle = strings.TrimSpace(strings.Join(parts, " "))
		}
		chatUsername = chat.Username
	}

	if from != nil {
		if chatID == 0 {
			chatID = from.ID
		}
		if chatUsername == "" {
			chatUsername = from.Username
		}
		if chatTitle == "" {
			parts := []string{}
			if from.FirstName != "" {
				parts = append(parts, from.FirstName)
			}
			if from.LastName != "" {
				parts = append(parts, from.LastName)
			}
			chatTitle = strings.TrimSpace(strings.Join(parts, " "))
		}
	}

	if chatTitle == "" {
		if chatUsername != "" {
			chatTitle = "@" + chatUsername
		} else if chatID != 0 {
			chatTitle = fmt.Sprintf("Chat %d", chatID)
		}
	}

	if chatID == 0 {
		return
	}

	bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _ = b.pool.Exec(bgCtx, `
		INSERT INTO telegram_subscribers (chat_id, type, title, username, notify_in, notify_out, notify_adjust, notify_low_stock, active)
		VALUES ($1, $2, $3, $4, TRUE, TRUE, FALSE, TRUE, TRUE)
		ON CONFLICT (chat_id) DO UPDATE SET
			title = CASE WHEN EXCLUDED.title != '' AND telegram_subscribers.title = '' THEN EXCLUDED.title ELSE telegram_subscribers.title END,
			username = CASE WHEN EXCLUDED.username != '' THEN EXCLUDED.username ELSE telegram_subscribers.username END,
			type = EXCLUDED.type,
			updated_at = now()`,
		chatID, chatType, chatTitle, chatUsername)

	// Also register individual user if message originated from a group
	if from != nil && from.ID != 0 && from.ID != chatID {
		fromName := strings.TrimSpace(from.FirstName + " " + from.LastName)
		if fromName == "" {
			fromName = from.Username
		}
		if fromName == "" {
			fromName = fmt.Sprintf("User %d", from.ID)
		}
		_, _ = b.pool.Exec(bgCtx, `
			INSERT INTO telegram_subscribers (chat_id, type, title, username, notify_in, notify_out, notify_adjust, notify_low_stock, active)
			VALUES ($1, 'private', $2, $3, TRUE, TRUE, FALSE, TRUE, TRUE)
			ON CONFLICT (chat_id) DO UPDATE SET
				title = CASE WHEN EXCLUDED.title != '' AND telegram_subscribers.title = '' THEN EXCLUDED.title ELSE telegram_subscribers.title END,
				username = CASE WHEN EXCLUDED.username != '' THEN EXCLUDED.username ELSE telegram_subscribers.username END,
				updated_at = now()`,
			from.ID, fromName, from.Username)
	}
}

// ---- State Machine Helpers ----

func (b *Bot) setWizard(chatID int64, w *wizardState, fromID ...int64) {
	b.wizardMutex.Lock()
	defer b.wizardMutex.Unlock()
	w.UpdatedAt = time.Now()
	b.wizards[chatID] = w
	for _, f := range fromID {
		if f != 0 && f != chatID {
			b.wizards[f] = w
		}
	}
}

func (b *Bot) getWizard(chatID int64, fromID ...int64) *wizardState {
	b.wizardMutex.RLock()
	w, exists := b.wizards[chatID]
	if !exists {
		for _, f := range fromID {
			if f != 0 {
				w, exists = b.wizards[f]
				if exists {
					break
				}
			}
		}
	}
	b.wizardMutex.RUnlock()

	if !exists || w == nil {
		return nil
	}
	if time.Since(w.UpdatedAt) > 15*time.Minute {
		b.clearWizard(chatID, fromID...)
		return nil
	}
	return w
}

func (b *Bot) clearWizard(chatID int64, fromID ...int64) {
	b.wizardMutex.Lock()
	defer b.wizardMutex.Unlock()
	delete(b.wizards, chatID)
	for _, f := range fromID {
		if f != 0 {
			delete(b.wizards, f)
		}
	}
}

func escapeMarkdown(s string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"`", "\\`",
		"[", "\\[",
	)
	return replacer.Replace(s)
}

func (b *Bot) sendMessage(ctx context.Context, bt *bot.Bot, chatID int64, text string, replyMarkup models.ReplyMarkup) (*models.Message, error) {
	params := &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: replyMarkup,
	}
	msg, err := bt.SendMessage(ctx, params)
	if err != nil {
		slog.Warn("telegram send markdown failed, retrying plain text", "err", err, "chatID", chatID)
		params.ParseMode = ""
		msg, err = bt.SendMessage(ctx, params)
		if err != nil {
			slog.Error("telegram plain text send also failed", "err", err, "chatID", chatID)
			return nil, err
		}
	}
	return msg, nil
}

// ---- Keyboards ----

func (b *Bot) buildMenuKeyboard(ctx context.Context) models.InlineKeyboardMarkup {
	webURL := b.getWebAppURL(ctx)
	kbRows := [][]models.InlineKeyboardButton{}

	if webURL != "" {
		kbRows = append(kbRows, []models.InlineKeyboardButton{
			{
				Text:   "🌐 Buka Aplikasi Web",
				WebApp: &models.WebAppInfo{URL: webURL},
			},
		})
	}

	kbRows = append(kbRows, []models.InlineKeyboardButton{
		{Text: "➕ Tambah Barang Baru", CallbackData: "cmd:tambah_wizard"},
		{Text: "📦 Cek Stok Kritis", CallbackData: "cmd:stok"},
	})

	kbRows = append(kbRows, []models.InlineKeyboardButton{
		{Text: "📂 Jelajah Kategori", CallbackData: "cmd:kategori_list"},
		{Text: "📍 Jelajah Ruangan", CallbackData: "cmd:ruangan_list"},
	})

	kbRows = append(kbRows, []models.InlineKeyboardButton{
		{Text: "🔮 Prediksi Kehabisan", CallbackData: "cmd:prediksi"},
		{Text: "🛠️ Daftar Alat Rusak", CallbackData: "cmd:rusak"},
	})

	kbRows = append(kbRows, []models.InlineKeyboardButton{
		{Text: "📑 Unduh Rekap Barang", CallbackData: "cmd:rekap"},
		{Text: "📊 Rekap Mutasi Masuk/Keluar", CallbackData: "cmd:rekap_mutasi"},
	})

	kbRows = append(kbRows, []models.InlineKeyboardButton{
		{Text: "❓ Panduan Lengkap Bot", CallbackData: "cmd:bantu"},
	})

	// Custom dynamic buttons from telegram_commands where is_menu = true
	cmdRows, err := b.pool.Query(ctx, `SELECT command, label FROM telegram_commands WHERE is_menu=TRUE AND active=TRUE ORDER BY sort_order`)
	if err == nil {
		defer cmdRows.Close()
		var customRow []models.InlineKeyboardButton
		for cmdRows.Next() {
			var cmd, lbl string
			if err := cmdRows.Scan(&cmd, &lbl); err == nil {
				cmd = strings.TrimPrefix(strings.TrimSpace(cmd), "/")
				if lbl == "" {
					lbl = "/" + cmd
				}
				customRow = append(customRow, models.InlineKeyboardButton{
					Text:         "📌 " + lbl,
					CallbackData: "cmd:custom:" + cmd,
				})
				if len(customRow) == 2 {
					kbRows = append(kbRows, customRow)
					customRow = nil
				}
			}
		}
		if len(customRow) > 0 {
			kbRows = append(kbRows, customRow)
		}
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

func (b *Bot) editMessage(ctx context.Context, bt *bot.Bot, chatID int64, messageID int, text string, keyboard models.InlineKeyboardMarkup) {
	if messageID == 0 {
		_, _ = bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        text,
			ParseMode:   models.ParseModeMarkdown,
			ReplyMarkup: keyboard,
		})
		return
	}
	_, err := bt.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   messageID,
		Text:        text,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: keyboard,
	})
	if err != nil {
		_, err2 := bt.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:      chatID,
			MessageID:   messageID,
			Text:        text,
			ReplyMarkup: keyboard,
		})
		if err2 != nil {
			_, _ = bt.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        text,
				ReplyMarkup: keyboard,
			})
		}
	}
}

// ---- Handlers ----

func (b *Bot) handleStart(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil {
		return
	}
	chatID := u.Message.Chat.ID
	var fromID int64
	if u.Message.From != nil {
		fromID = u.Message.From.ID
	}
	b.clearWizard(chatID, fromID)
	b.autoRegisterSubscriber(ctx, &u.Message.Chat, u.Message.From)

	if !b.allowed(chatID) {
		name := u.Message.Chat.FirstName
		if name == "" {
			name = u.Message.Chat.Title
		}
		if name == "" {
			name = "Pengguna"
		}
		text := fmt.Sprintf("👋 *Halo, %s!*\n\n🆔 *Chat ID Anda:* `%d`\n\nAkun Anda dinonaktifkan di sistem Inventaris Puskesmas.\nSilakan hubungi administrator.", name, chatID)
		_, _ = bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      text,
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}
	b.showMenu(ctx, bt, chatID, u.Message.ID)
}

func (b *Bot) showMenu(ctx context.Context, bt *bot.Bot, chatID int64, replyToID int) {
	intro := "📋 *Menu Pintar Inventaris Puskesmas*\n\n" +
		"Pilih menu di bawah, ketik nama barang langsung untuk cari, atau kirim foto barcode fisik:"
	kb := b.buildMenuKeyboard(ctx)
	_, err := bt.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        intro,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: kb,
	})
	if err != nil {
		_, _ = bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        intro,
			ReplyMarkup: kb,
		})
	}
}

func (b *Bot) handleBatal(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil {
		return
	}
	chatID := u.Message.Chat.ID
	var fromID int64
	if u.Message.From != nil {
		fromID = u.Message.From.ID
	}
	b.clearWizard(chatID, fromID)
	bt.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        "❌ *Proses dibatalkan.*\n\nSilakan pilih menu di bawah:",
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: b.buildMenuKeyboard(ctx),
	})
}

func (b *Bot) getBantuText() string {
	return "💡 *Daftar Perintah & Fitur Cerdas Bot:* \n\n" +
		"📦 *Manajemen Barang:*\n" +
		"• `/tambah` — Wizard interaktif pendaftaran barang baru\n" +
		"• `/tambah Nama | Kat | Ruang | Stok | Satuan` — Input cepat\n" +
		"• `/masuk <SKU> <jumlah>` — Tambah stok barang masuk\n" +
		"• `/keluar <SKU> <jumlah>` — Catat pengeluaran barang\n" +
		"• `/pindah <SKU> ke <Ruangan>` — Pindahkan lokasi aset\n" +
		"• `/opname <SKU> <stok_riil>` — Audit penyesuaian stok fisik\n" +
		"• `/label <SKU>` — Kirim stiker QR Code siap print\n\n" +
		"⚡ *Batch Multi-Baris:*\n" +
		"• `/batch_masuk` atau `/batch_keluar` — Catat mutasi banyak barang sekaligus\n\n" +
		"🔍 *Pencarian & Audit Cerdas:*\n" +
		"• *Pencarian Langsung:* Ketik nama barang tanpa garis miring (cth: `tensi`, `kursi roda`)\n" +
		"• *Scan Barcode / QR:* Kirim foto barcode barang fisik langsung ke bot\n" +
		"• `/stok` — Daftar stok menipis di bawah batas minimum\n" +
		"• `/rusak` — Daftar aset yang rusak atau butuh servis\n" +
		"• `/prediksi` — Estimasi kehabisan stok berdasarkan konsumsi riil\n\n" +
		"📑 *Laporan & Export:*\n" +
		"• `/rekap` — Unduh file CSV seluruh data barang\n" +
		"• `/rekap_mutasi` — Unduh file CSV mutasi 30 hari terakhir\n" +
		"• `/batal` — Batalkan proses wizard aktif"
}

func (b *Bot) handleBantu(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil || !b.allowed(u.Message.Chat.ID) {
		return
	}
	text := b.getBantuText()
	bt.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      u.Message.Chat.ID,
		Text:        text,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: b.buildBackKeyboard(),
	})
}

// handleDefault handles interactive wizard inputs, photos, and natural language
func (b *Bot) handleDefault(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil {
		return
	}
	b.autoRegisterSubscriber(ctx, &u.Message.Chat, u.Message.From)
	if !b.allowed(u.Message.Chat.ID) {
		return
	}
	chatID := u.Message.Chat.ID

	// 1. Photo: scan for Barcode / QR Code
	if len(u.Message.Photo) > 0 {
		b.handlePhoto(ctx, bt, u)
		return
	}

	text := strings.TrimSpace(u.Message.Text)
	if text == "" {
		return
	}

	// 2. Check active interactive wizard
	var fromID int64
	if u.Message.From != nil {
		fromID = u.Message.From.ID
	}
	wiz := b.getWizard(chatID, fromID)
	if wiz != nil && !strings.HasPrefix(text, "/") {
		b.handleWizardInput(ctx, bt, u, wiz, text)
		return
	}

	if strings.HasPrefix(text, "/") {
		cmdName := strings.TrimPrefix(strings.Fields(text)[0], "/")
		cmdName = strings.ToLower(cmdName)
		var resp, label string
		err := b.pool.QueryRow(ctx, `SELECT response, label FROM telegram_commands WHERE (LOWER(command)=$1 OR LOWER(command)=$2) AND active=TRUE LIMIT 1`, cmdName, "/"+cmdName).Scan(&resp, &label)
		if err == nil && resp != "" {
			bt.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        fmt.Sprintf("📌 *%s*\n\n%s", label, resp),
				ParseMode:   models.ParseModeMarkdown,
				ReplyMarkup: b.buildBackKeyboard(),
			})
			return
		}
		bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        fmt.Sprintf("❓ Perintah `/%s` tidak dikenali.\n\nKetik /start atau /bantu untuk melihat semua menu yang tersedia.", cmdName),
			ParseMode:   models.ParseModeMarkdown,
		})
		return
	}

	// 3. Smart Natural Language Understanding
	b.handleSmartNLP(ctx, bt, u, text)
}

// handleWizardInput processes conversational wizard responses
func (b *Bot) handleWizardInput(ctx context.Context, bt *bot.Bot, u *models.Update, wiz *wizardState, input string) {
	chatID := u.Message.Chat.ID
	var fromID int64
	if u.Message.From != nil {
		fromID = u.Message.From.ID
	}

	switch wiz.Step {
	case "name":
		wiz.Name = input
		wiz.Step = "category"
		b.setWizard(chatID, wiz, fromID)

		// Ask Category via inline keyboard
		rows, err := b.pool.Query(ctx, `SELECT name FROM categories ORDER BY name LIMIT 12`)
		if err != nil {
			bt.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "⚠️ Gagal memuat daftar kategori."})
			return
		}
		defer rows.Close()

		var kbRows [][]models.InlineKeyboardButton
		for rows.Next() {
			var catName string
			rows.Scan(&catName)
			kbRows = append(kbRows, []models.InlineKeyboardButton{
				{Text: "📂 " + catName, CallbackData: "wizcat:" + catName},
			})
		}
		kbRows = append(kbRows, []models.InlineKeyboardButton{
			{Text: "❌ Batalkan", CallbackData: "cmd:batal"},
		})

		msgText := fmt.Sprintf("📝 *Tambah Barang (Langkah 2/4)*\n\nNama Barang: *%s*\n\nSilakan pilih *Kategori Barang* di bawah:", wiz.Name)
		bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        msgText,
			ParseMode:   models.ParseModeMarkdown,
			ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kbRows},
		})

	case "stock":
		// Parse stock & unit, e.g. "10 unit" or "50 strip" or "10"
		parts := strings.Fields(input)
		var stock int32 = 0
		unit := "unit"

		if len(parts) >= 1 {
			if s, err := strconv.ParseInt(parts[0], 10, 32); err == nil && s >= 0 {
				stock = int32(s)
			}
		}
		if len(parts) >= 2 {
			unit = parts[1]
		}

		wiz.Stock = stock
		wiz.Unit = unit
		b.finalizeWizardItem(ctx, bt, chatID, wiz, fromID)
	}
}

func (b *Bot) finalizeWizardItem(ctx context.Context, bt *bot.Bot, chatID int64, wiz *wizardState, fromID ...int64) {
	sku := b.generateSKU(ctx, wiz.Category)

	var id string
	err := b.pool.QueryRow(ctx,
		`INSERT INTO inventory_items (sku, name, category, location, current_stock, min_stock, unit, condition_status, is_available, track_stock)
		 VALUES ($1, $2, $3, $4, $5, 1, $6, 'Berfungsi', TRUE, TRUE)
		 RETURNING id::text`,
		sku, wiz.Name, wiz.Category, wiz.Location, wiz.Stock, wiz.Unit).Scan(&id)

	b.clearWizard(chatID, fromID...)

	if err != nil {
		bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "⚠️ Gagal menyimpan barang ke inventaris: " + err.Error(),
		})
		return
	}

	if wiz.Stock > 0 {
		b.pool.Exec(ctx,
			`INSERT INTO stock_transactions (type, item_id, item_sku, item_name, quantity, unit, previous_stock, new_stock, notes)
			 VALUES ('IN', $1, $2, $3, $4, $5, 0, $4, 'Pendaftaran barang baru via Bot Telegram')`,
			id, sku, wiz.Name, wiz.Stock, wiz.Unit)
	}

	// Generate QR Code Sticker Image
	qrBytes, err := generateQRImage(sku)
	card := fmt.Sprintf("✅ *Barang Berhasil Didaftarkan!*\n\n📦 *%s*\n• SKU: `%s`\n• Kategori: %s\n• Ruangan: %s\n• Stok Awal: *%d %s*\n• Kondisi: Berfungsi",
		wiz.Name, sku, wiz.Category, wiz.Location, wiz.Stock, wiz.Unit)

	kb := models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "➕ +1 Stok", CallbackData: fmt.Sprintf("cmd:add:%s:1", sku)},
				{Text: "➕ +5 Stok", CallbackData: fmt.Sprintf("cmd:add:%s:5", sku)},
				{Text: "➖ -1 Stok", CallbackData: fmt.Sprintf("cmd:sub:%s:1", sku)},
			},
			{
				{Text: "📍 Pindah Ruangan", CallbackData: fmt.Sprintf("cmd:pindah:%s", sku)},
				{Text: "🛠️ Ubah Kondisi", CallbackData: fmt.Sprintf("cmd:kondisi:%s", sku)},
			},
			{
				{Text: "‹ Kembali ke Menu", CallbackData: "cmd:start"},
			},
		},
	}

	if err == nil && len(qrBytes) > 0 {
		_, _ = bt.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID: chatID,
			Photo: &models.InputFileUpload{
				Filename: fmt.Sprintf("%s.png", sku),
				Data:     bytes.NewReader(qrBytes),
			},
			Caption:     card + "\n\n🏷️ *Stiker QR Code siap dicetak / ditempelkan.*",
			ParseMode:   models.ParseModeMarkdown,
			ReplyMarkup: kb,
		})
	} else {
		bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        card,
			ParseMode:   models.ParseModeMarkdown,
			ReplyMarkup: kb,
		})
	}
}

// handleSmartNLP parses natural language intents
func (b *Bot) handleSmartNLP(ctx context.Context, bt *bot.Bot, u *models.Update, text string) {
	chatID := u.Message.Chat.ID
	lower := strings.ToLower(text)

	// Quick greetings or menu
	if lower == "menu" || lower == "start" || lower == "halo" || lower == "hai" || lower == "help" {
		b.showMenu(ctx, bt, chatID, u.Message.ID)
		return
	}

	// Stock Alerts
	if lower == "stok" || lower == "kritis" || lower == "stok menipis" || lower == "habis" {
		b.handleStok(ctx, bt, u)
		return
	}

	// Damaged equipment
	if lower == "rusak" || lower == "alat rusak" || lower == "servis" || lower == "perbaikan" {
		b.handleRusak(ctx, bt, u)
		return
	}

	// Predictions
	if lower == "prediksi" || lower == "analisis" || lower == "ramalan stok" {
		b.handlePrediksi(ctx, bt, u)
		return
	}

	// Reports
	if lower == "rekap" || lower == "laporan" || lower == "export" {
		b.handleRekap(ctx, bt, u)
		return
	}
	if lower == "rekap mutasi" || lower == "mutasi" || lower == "transaksi" {
		b.handleRekapMutasi(ctx, bt, u)
		return
	}

	// "masuk 10 <SKU>"
	if strings.HasPrefix(lower, "masuk ") {
		fields := strings.Fields(text)
		if len(fields) >= 3 {
			var sku string
			var qty int64
			if q, err := strconv.ParseInt(fields[1], 10, 32); err == nil {
				qty = q
				sku = fields[2]
			} else if q, err := strconv.ParseInt(fields[2], 10, 32); err == nil {
				sku = fields[1]
				qty = q
			}
			if sku != "" && qty > 0 {
				b.executeStockMutation(ctx, bt, chatID, sku, int32(qty), "IN")
				return
			}
		}
	}

	// "keluar 5 <SKU>"
	if strings.HasPrefix(lower, "keluar ") {
		fields := strings.Fields(text)
		if len(fields) >= 3 {
			var sku string
			var qty int64
			if q, err := strconv.ParseInt(fields[1], 10, 32); err == nil {
				qty = q
				sku = fields[2]
			} else if q, err := strconv.ParseInt(fields[2], 10, 32); err == nil {
				sku = fields[1]
				qty = q
			}
			if sku != "" && qty > 0 {
				b.executeStockMutation(ctx, bt, chatID, sku, int32(qty), "OUT")
				return
			}
		}
	}

	// "pindah <SKU> ke <Ruangan>"
	if strings.HasPrefix(lower, "pindah ") && strings.Contains(lower, " ke ") {
		parts := strings.Split(text, " ke ")
		if len(parts) == 2 {
			skuParts := strings.Fields(parts[0])
			if len(skuParts) >= 2 {
				sku := skuParts[1]
				loc := strings.TrimSpace(parts[1])
				b.executeRelocate(ctx, bt, chatID, sku, loc)
				return
			}
		}
	}

	// "opname <SKU> <stok>"
	if strings.HasPrefix(lower, "opname ") {
		fields := strings.Fields(text)
		if len(fields) >= 3 {
			sku := fields[1]
			if actual, err := strconv.ParseInt(fields[2], 10, 32); err == nil {
				b.executeOpname(ctx, bt, chatID, sku, int32(actual))
				return
			}
		}
	}

	// Default: Smart search items
	b.searchItems(ctx, bt, chatID, text)
}

func (b *Bot) searchItems(ctx context.Context, bt *bot.Bot, chatID int64, query string) {
	rows, err := b.pool.Query(ctx,
		`SELECT id::text, sku, name, current_stock, min_stock, unit, location, category, condition_status
		 FROM inventory_items
		 WHERE name ILIKE $1 OR sku ILIKE $1 OR location ILIKE $1 OR category ILIKE $1
		 ORDER BY name LIMIT 5`, "%"+query+"%")
	if err != nil {
		bt.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "⚠️ Gagal mencari data barang."})
		return
	}
	defer rows.Close()

	type itemFound struct {
		ID        string
		SKU       string
		Name      string
		Stock     int32
		MinStock  int32
		Unit      string
		Location  string
		Category  string
		Condition string
	}
	var list []itemFound
	for rows.Next() {
		var it itemFound
		rows.Scan(&it.ID, &it.SKU, &it.Name, &it.Stock, &it.MinStock, &it.Unit, &it.Location, &it.Category, &it.Condition)
		list = append(list, it)
	}

	if len(list) == 0 {
		text := fmt.Sprintf("🔍 Barang dengan kata kunci *\"%s\"* tidak ditemukan.\n\nCoba ketik kata kunci lain atau pilih menu di bawah:", query)
		bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        text,
			ParseMode:   models.ParseModeMarkdown,
			ReplyMarkup: b.buildMenuKeyboard(ctx),
		})
		return
	}

	for _, it := range list {
		b.sendItemCard(ctx, bt, chatID, it.SKU, it.Name, it.Stock, it.MinStock, it.Unit, it.Location, it.Category, it.Condition)
	}
}

func (b *Bot) sendItemCard(ctx context.Context, bt *bot.Bot, chatID int64, sku, name string, stock, minStock int32, unit, location, category, condition string) {
	statusIcon := "🟢"
	if stock <= minStock {
		statusIcon = "🔴"
	}
	condIcon := "✅"
	if condition != "Berfungsi" {
		condIcon = "⚠️"
	}

	card := fmt.Sprintf("%s *%s*\n• SKU: `%s`\n• Stok: *%d %s* (min %d)\n• Ruangan: %s\n• Kategori: %s\n• Kondisi: %s %s",
		statusIcon, name, sku, stock, unit, minStock, location, category, condIcon, condition)

	kb := models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "➕ +1", CallbackData: fmt.Sprintf("cmd:add:%s:1", sku)},
				{Text: "➕ +5", CallbackData: fmt.Sprintf("cmd:add:%s:5", sku)},
				{Text: "➖ -1", CallbackData: fmt.Sprintf("cmd:sub:%s:1", sku)},
			},
			{
				{Text: "📍 Pindah Ruangan", CallbackData: fmt.Sprintf("cmd:pindah:%s", sku)},
				{Text: "🛠️ Ubah Kondisi", CallbackData: fmt.Sprintf("cmd:kondisi:%s", sku)},
			},
			{
				{Text: "🏷️ Stiker Label QR", CallbackData: fmt.Sprintf("cmd:label:%s", sku)},
				{Text: "🗑️ Hapus", CallbackData: fmt.Sprintf("cmd:del:%s", sku)},
			},
			{
				{Text: "‹ Kembali ke Menu", CallbackData: "cmd:start"},
			},
		},
	}

	bt.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        card,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: kb,
	})
}

// handlePhoto scans Barcode & QR Code from uploaded photo
func (b *Bot) handlePhoto(ctx context.Context, bt *bot.Bot, u *models.Update) {
	chatID := u.Message.Chat.ID
	photos := u.Message.Photo
	if len(photos) == 0 {
		return
	}
	largest := photos[len(photos)-1]

	statusMsg, _ := bt.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "🔍 Memindai barcode / QR code dari foto...",
	})

	fileInfo, err := bt.GetFile(ctx, &bot.GetFileParams{FileID: largest.FileID})
	if err != nil {
		if statusMsg != nil {
			_, _ = bt.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:    chatID,
				MessageID: statusMsg.ID,
				Text:      "⚠️ Gagal mengunduh file foto dari Telegram.",
			})
		}
		return
	}

	downloadURL := bt.FileDownloadLink(fileInfo)
	resp, err := http.Get(downloadURL)
	if err != nil {
		if statusMsg != nil {
			_, _ = bt.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:    chatID,
				MessageID: statusMsg.ID,
				Text:      "⚠️ Gagal mengambil file foto.",
			})
		}
		return
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		if statusMsg != nil {
			_, _ = bt.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:    chatID,
				MessageID: statusMsg.ID,
				Text:      "⚠️ Format gambar tidak dapat dibaca.",
			})
		}
		return
	}

	code, format, err := decodeBarcode(img)
	if err != nil || code == "" {
		if statusMsg != nil {
			_, _ = bt.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:    chatID,
				MessageID: statusMsg.ID,
				Text:      "📸 *Foto Diterima*\n\nBarcode atau QR code tidak terdeteksi pada gambar.\n\n💡 *Tips:* Pastikan barcode fokus, pencahayaan terang, dan posisi tegak lurus dengan kamera.",
				ParseMode: models.ParseModeMarkdown,
			})
		}
		return
	}

	// Code detected! Look up item in database
	var it struct {
		ID        string
		SKU       string
		Name      string
		Category  string
		Location  string
		Stock     int32
		MinStock  int32
		Unit      string
		Condition string
	}
	err = b.pool.QueryRow(ctx,
		`SELECT id::text, sku, name, category, location, current_stock, min_stock, unit, condition_status
		 FROM inventory_items
		 WHERE sku = $1 OR serial_number = $1 OR sku ILIKE $2 OR name ILIKE $2
		 LIMIT 1`, code, "%"+code+"%").
		Scan(&it.ID, &it.SKU, &it.Name, &it.Category, &it.Location, &it.Stock, &it.MinStock, &it.Unit, &it.Condition)

	if err != nil {
		text := fmt.Sprintf("📸 *Barcode Terdeteksi!* (%s)\n\nKode: `%s`\n\n⚠️ Barang belum terdaftar di inventaris.\nIngin mendaftarkan sekarang?", format, code)
		kb := models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "➕ Daftarkan Barang Baru", CallbackData: "cmd:tambah_wizard"},
				},
				{
					{Text: "‹ Kembali ke Menu", CallbackData: "cmd:start"},
				},
			},
		}
		if statusMsg != nil {
			_, _ = bt.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:      chatID,
				MessageID:   statusMsg.ID,
				Text:        text,
				ParseMode:   models.ParseModeMarkdown,
				ReplyMarkup: kb,
			})
		}
		return
	}

	if statusMsg != nil {
		_, _ = bt.DeleteMessage(ctx, &bot.DeleteMessageParams{
			ChatID:    chatID,
			MessageID: statusMsg.ID,
		})
	}
	b.sendItemCard(ctx, bt, chatID, it.SKU, it.Name, it.Stock, it.MinStock, it.Unit, it.Location, it.Category, it.Condition)
}

func decodeBarcode(img image.Image) (string, string, error) {
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", "", err
	}
	// Try QR Code
	qrReader := qrcode.NewQRCodeReader()
	if res, err := qrReader.Decode(bmp, nil); err == nil && res != nil {
		return res.GetText(), "QR Code", nil
	}
	// Try 1D Barcodes
	oneDReaders := []struct {
		name   string
		reader gozxing.Reader
	}{
		{"Code 128", oned.NewCode128Reader()},
		{"Code 39", oned.NewCode39Reader()},
		{"EAN-13", oned.NewEAN13Reader()},
		{"UPC-A", oned.NewUPCAReader()},
	}
	for _, r := range oneDReaders {
		if res, err := r.reader.Decode(bmp, nil); err == nil && res != nil {
			return res.GetText(), r.name, nil
		}
	}
	return "", "", fmt.Errorf("barcode tidak terdeteksi")
}

// handleTambah initiates conversational wizard or parses quick syntax
func (b *Bot) handleTambah(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil || !b.allowed(u.Message.Chat.ID) {
		return
	}
	chatID := u.Message.Chat.ID
	var fromID int64
	if u.Message.From != nil {
		fromID = u.Message.From.ID
	}
	raw := strings.TrimSpace(strings.TrimPrefix(u.Message.Text, "/tambah"))

	// If no parameters: start Interactive Wizard
	if raw == "" {
		b.setWizard(chatID, &wizardState{Step: "name"}, fromID)
		text := "📝 *Tambah Barang Baru (Langkah 1/4)*\n\n" +
			"Silakan ketik *Nama Barang* yang ingin didaftarkan:\n" +
			"(Ketik `/batal` kapan saja untuk membatalkan)"
		bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      text,
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}

	// Quick syntax: /tambah Nama | Kategori | Ruangan | Stok | Satuan
	var parts []string
	if strings.Contains(raw, "|") {
		parts = strings.Split(raw, "|")
	} else {
		parts = strings.Split(raw, ",")
	}
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	if len(parts) < 3 {
		bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "⚠️ Format kurang lengkap.\nFormat: `/tambah Nama | Kategori | Ruangan | Stok | Satuan`",
		})
		return
	}

	name := parts[0]
	category := parts[1]
	location := parts[2]
	var stock int32 = 0
	unit := "unit"

	if len(parts) >= 4 {
		if s, err := strconv.ParseInt(parts[3], 10, 32); err == nil && s >= 0 {
			stock = int32(s)
		}
	}
	if len(parts) >= 5 && parts[4] != "" {
		unit = parts[4]
	}

	wiz := &wizardState{
		Name:     name,
		Category: category,
		Location: location,
		Stock:    stock,
		Unit:     unit,
	}
	b.finalizeWizardItem(ctx, bt, chatID, wiz)
}

func (b *Bot) generateSKU(ctx context.Context, category string) string {
	var slug string
	_ = b.pool.QueryRow(ctx, `SELECT id FROM categories WHERE name ILIKE $1 LIMIT 1`, "%"+category+"%").Scan(&slug)
	if slug == "" {
		slug = "BRG"
	}
	parts := strings.Split(slug, "-")
	prefix := strings.ToUpper(parts[0])
	if len(prefix) > 4 {
		prefix = prefix[:4]
	}
	if prefix == "" {
		prefix = "BRG"
	}
	year := strconv.Itoa(time.Now().Year())
	pattern := prefix + "-" + year + "-%"
	var last string
	_ = b.pool.QueryRow(ctx, `SELECT sku FROM inventory_items WHERE sku LIKE $1 ORDER BY sku DESC LIMIT 1`, pattern).Scan(&last)
	serial := 1
	if len(last) > len(prefix)+6 {
		if n, e := strconv.Atoi(last[len(prefix)+6:]); e == nil {
			serial = n + 1
		}
	}
	return fmt.Sprintf("%s-%s-%03d", prefix, year, serial)
}

// handleMasuk handles /masuk <SKU> <jumlah>
func (b *Bot) handleMasuk(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil || !b.allowed(u.Message.Chat.ID) {
		return
	}
	parts := strings.Fields(u.Message.Text)
	if len(parts) < 3 {
		bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    u.Message.Chat.ID,
			Text:      "Format: `/masuk <SKU> <jumlah>`\nContoh: `/masuk ALAT-2026-001 10`",
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}
	sku := parts[1]
	qty, err := strconv.ParseInt(parts[2], 10, 32)
	if err != nil || qty <= 0 {
		bt.SendMessage(ctx, &bot.SendMessageParams{ChatID: u.Message.Chat.ID, Text: "Jumlah harus berupa angka > 0"})
		return
	}
	b.executeStockMutation(ctx, bt, u.Message.Chat.ID, sku, int32(qty), "IN")
}

// handleKeluar handles /keluar <SKU> <jumlah>
func (b *Bot) handleKeluar(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil || !b.allowed(u.Message.Chat.ID) {
		return
	}
	parts := strings.Fields(u.Message.Text)
	if len(parts) < 3 {
		bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    u.Message.Chat.ID,
			Text:      "Format: `/keluar <SKU> <jumlah>`\nContoh: `/keluar ALAT-2026-001 2`",
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}
	sku := parts[1]
	qty, err := strconv.ParseInt(parts[2], 10, 32)
	if err != nil || qty <= 0 {
		bt.SendMessage(ctx, &bot.SendMessageParams{ChatID: u.Message.Chat.ID, Text: "Jumlah harus berupa angka > 0"})
		return
	}
	b.executeStockMutation(ctx, bt, u.Message.Chat.ID, sku, int32(qty), "OUT")
}

func (b *Bot) executeStockMutation(ctx context.Context, bt *bot.Bot, chatID int64, sku string, qty int32, mutType string) {
	var id, name, unit string
	var currentStock int32

	err := b.pool.QueryRow(ctx,
		`SELECT id::text, name, unit, current_stock FROM inventory_items WHERE sku ILIKE $1 OR serial_number ILIKE $1 LIMIT 1`, sku).
		Scan(&id, &name, &unit, &currentStock)
	if err != nil {
		bt.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "⚠️ Barang tidak ditemukan dengan SKU: " + sku})
		return
	}

	var newStock int32
	var actionLabel, icon string
	if mutType == "IN" {
		newStock = currentStock + qty
		actionLabel = "Barang Masuk"
		icon = "📥"
	} else {
		newStock = int32(math.Max(0, float64(currentStock-qty)))
		actionLabel = "Barang Keluar"
		icon = "📤"
	}

	_, err = b.pool.Exec(ctx,
		`UPDATE inventory_items SET current_stock = $2, updated_at = now() WHERE id::text = $1`, id, newStock)
	if err != nil {
		bt.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "⚠️ Gagal memperbarui stok."})
		return
	}

	b.pool.Exec(ctx,
		`INSERT INTO stock_transactions (type, item_id, item_sku, item_name, quantity, unit, previous_stock, new_stock, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'Mutasi via Bot Telegram')`,
		mutType, id, sku, name, qty, unit, currentStock, newStock)

	text := fmt.Sprintf("%s *%s Berhasil!*\n\n📦 *%s* (`%s`)\n• Jumlah: *%d %s*\n• Stok Sebelumnya: %d %s\n• Stok Sekarang: *%d %s*",
		icon, actionLabel, name, sku, qty, unit, currentStock, unit, newStock, unit)

	kb := models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "➕ +1 Lagi", CallbackData: fmt.Sprintf("cmd:add:%s:1", sku)},
				{Text: "➕ +5 Lagi", CallbackData: fmt.Sprintf("cmd:add:%s:5", sku)},
				{Text: "➖ -1", CallbackData: fmt.Sprintf("cmd:sub:%s:1", sku)},
			},
			{
				{Text: "‹ Kembali ke Menu", CallbackData: "cmd:start"},
			},
		},
	}

	bt.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: kb,
	})
}

// handleBatchMasuk parses multiple lines of "SKU QTY"
func (b *Bot) handleBatchMasuk(ctx context.Context, bt *bot.Bot, u *models.Update) {
	b.handleBatchMutation(ctx, bt, u, "IN")
}

// handleBatchKeluar parses multiple lines of "SKU QTY"
func (b *Bot) handleBatchKeluar(ctx context.Context, bt *bot.Bot, u *models.Update) {
	b.handleBatchMutation(ctx, bt, u, "OUT")
}

func (b *Bot) handleBatchMutation(ctx context.Context, bt *bot.Bot, u *models.Update, mutType string) {
	if u.Message == nil || !b.allowed(u.Message.Chat.ID) {
		return
	}
	chatID := u.Message.Chat.ID
	cmdPrefix := "/batch_masuk"
	actionLabel := "Masuk"
	icon := "📥"
	if mutType == "OUT" {
		cmdPrefix = "/batch_keluar"
		actionLabel = "Keluar"
		icon = "📤"
	}

	raw := strings.TrimSpace(strings.TrimPrefix(u.Message.Text, cmdPrefix))
	if raw == "" {
		text := fmt.Sprintf("⚡ *Batch Mutasi Barang %s*\n\nKirim format multi-baris (1 baris per barang):\n`%s\nSKU_1 JUMLAH\nSKU_2 JUMLAH`\n\n📝 *Contoh:*\n`%s\nALAT-2026-001 5\nBHP-2026-002 20\nELEK-2026-003 1`", actionLabel, cmdPrefix, cmdPrefix)
		bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      text,
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}

	lines := strings.Split(raw, "\n")
	successCount := 0
	var report strings.Builder
	report.WriteString(fmt.Sprintf("%s *Hasil Batch Barang %s:*\n\n", icon, actionLabel))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		sku := fields[0]
		qty, err := strconv.ParseInt(fields[1], 10, 32)
		if err != nil || qty <= 0 {
			continue
		}

		var id, name, unit string
		var currentStock int32
		err = b.pool.QueryRow(ctx,
			`SELECT id::text, name, unit, current_stock FROM inventory_items WHERE sku ILIKE $1 LIMIT 1`, sku).
			Scan(&id, &name, &unit, &currentStock)
		if err != nil {
			report.WriteString(fmt.Sprintf("❌ `%s`: SKU tidak ditemukan\n", sku))
			continue
		}

		var newStock int32
		if mutType == "IN" {
			newStock = currentStock + int32(qty)
		} else {
			newStock = int32(math.Max(0, float64(currentStock-int32(qty))))
		}

		b.pool.Exec(ctx, `UPDATE inventory_items SET current_stock=$2, updated_at=now() WHERE id::text=$1`, id, newStock)
		b.pool.Exec(ctx,
			`INSERT INTO stock_transactions (type, item_id, item_sku, item_name, quantity, unit, previous_stock, new_stock, notes)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'Batch mutation via Bot')`,
			mutType, id, sku, name, qty, unit, currentStock, newStock)

		report.WriteString(fmt.Sprintf("✅ *%s* (`%s`): %d -> *%d %s*\n", name, sku, currentStock, newStock, unit))
		successCount++
	}

	report.WriteString(fmt.Sprintf("\n📊 *Total Berhasil:* %d item.", successCount))
	bt.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        report.String(),
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: b.buildBackKeyboard(),
	})
}

// handlePindah handles /pindah <SKU> ke <Ruangan>
func (b *Bot) handlePindah(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil || !b.allowed(u.Message.Chat.ID) {
		return
	}
	chatID := u.Message.Chat.ID
	raw := strings.TrimSpace(strings.TrimPrefix(u.Message.Text, "/pindah"))

	if !strings.Contains(raw, " ke ") {
		text := "📍 *Pindah Lokasi / Ruangan Aset*\n\n" +
			"Format: `/pindah <SKU> ke <Ruangan Tujuan>`\n" +
			"Contoh: `/pindah ALAT-2026-001 ke Poli Gigi`"
		bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      text,
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}

	parts := strings.Split(raw, " ke ")
	sku := strings.TrimSpace(parts[0])
	newLoc := strings.TrimSpace(parts[1])
	b.executeRelocate(ctx, bt, chatID, sku, newLoc)
}

func (b *Bot) executeRelocate(ctx context.Context, bt *bot.Bot, chatID int64, sku, newLoc string) {
	var id, name, oldLoc string
	err := b.pool.QueryRow(ctx,
		`SELECT id::text, name, location FROM inventory_items WHERE sku ILIKE $1 LIMIT 1`, sku).
		Scan(&id, &name, &oldLoc)
	if err != nil {
		bt.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "⚠️ Barang tidak ditemukan dengan SKU: " + sku})
		return
	}

	_, err = b.pool.Exec(ctx, `UPDATE inventory_items SET location=$2, updated_at=now() WHERE id::text=$1`, id, newLoc)
	if err != nil {
		bt.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "⚠️ Gagal memperbarui lokasi aset."})
		return
	}

	note := fmt.Sprintf("Relokasi aset dari %s ke %s via Bot", oldLoc, newLoc)
	b.pool.Exec(ctx,
		`INSERT INTO stock_transactions (type, item_id, item_sku, item_name, quantity, unit, previous_stock, new_stock, notes)
		 VALUES ('ADJUST', $1, $2, $3, 1, 'unit', 0, 0, $4)`,
		id, sku, name, note)

	text := fmt.Sprintf("📍 *Relokasi Aset Berhasil!*\n\n📦 *%s* (`%s`)\n• Ruangan Lama: %s\n• Ruangan Baru: *%s*", name, sku, oldLoc, newLoc)
	bt.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: b.buildBackKeyboard(),
	})
}

// handleOpname handles /opname <SKU> <stok_riil>
func (b *Bot) handleOpname(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil || !b.allowed(u.Message.Chat.ID) {
		return
	}
	chatID := u.Message.Chat.ID
	fields := strings.Fields(u.Message.Text)
	if len(fields) < 3 {
		text := "📋 *Stock Opname (Audit Fisik)*\n\n" +
			"Format: `/opname <SKU> <stok_riil>`\n" +
			"Contoh: `/opname ALAT-2026-001 8`"
		bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      text,
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}

	sku := fields[1]
	actual, err := strconv.ParseInt(fields[2], 10, 32)
	if err != nil || actual < 0 {
		bt.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "Stok fisik harus berupa angka >= 0"})
		return
	}
	b.executeOpname(ctx, bt, chatID, sku, int32(actual))
}

func (b *Bot) executeOpname(ctx context.Context, bt *bot.Bot, chatID int64, sku string, actual int32) {
	var id, name, unit string
	var prev int32
	err := b.pool.QueryRow(ctx,
		`SELECT id::text, name, unit, current_stock FROM inventory_items WHERE sku ILIKE $1 LIMIT 1`, sku).
		Scan(&id, &name, &unit, &prev)
	if err != nil {
		bt.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "⚠️ Barang tidak ditemukan dengan SKU: " + sku})
		return
	}

	diff := actual - prev
	diffStr := fmt.Sprintf("%+d", diff)

	b.pool.Exec(ctx, `UPDATE inventory_items SET current_stock=$2, updated_at=now() WHERE id::text=$1`, id, actual)
	note := fmt.Sprintf("Stock Opname: fisik %d vs sistem %d (selisih %s) via Bot", actual, prev, diffStr)

	qty := int32(math.Abs(float64(diff)))
	if qty == 0 {
		qty = 1
	}

	b.pool.Exec(ctx,
		`INSERT INTO stock_transactions (type, item_id, item_sku, item_name, quantity, unit, previous_stock, new_stock, notes)
		 VALUES ('ADJUST', $1, $2, $3, $4, $5, $6, $7, $8)`,
		id, sku, name, qty, unit, prev, actual, note)

	text := fmt.Sprintf("📋 *Audit Opname Selesai!*\n\n📦 *%s* (`%s`)\n• Stok Sistem: %d %s\n• Stok Fisik Riil: *%d %s*\n• Selisih: *%s %s*",
		name, sku, prev, unit, actual, unit, diffStr, unit)

	bt.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: b.buildBackKeyboard(),
	})
}

// handleLabel generates and sends QR sticker photo
func (b *Bot) handleLabel(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil || !b.allowed(u.Message.Chat.ID) {
		return
	}
	chatID := u.Message.Chat.ID
	sku := strings.TrimSpace(strings.TrimPrefix(u.Message.Text, "/label"))
	if sku == "" {
		bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Format: `/label <SKU>`\nContoh: `/label ALAT-2026-001`",
		})
		return
	}
	b.sendQRLabel(ctx, bt, chatID, sku)
}

func (b *Bot) sendQRLabel(ctx context.Context, bt *bot.Bot, chatID int64, sku string) {
	var it struct {
		Name     string
		Category string
		Location string
	}
	err := b.pool.QueryRow(ctx,
		`SELECT name, category, location FROM inventory_items WHERE sku ILIKE $1 LIMIT 1`, sku).
		Scan(&it.Name, &it.Category, &it.Location)
	if err != nil {
		bt.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "⚠️ Barang tidak ditemukan dengan SKU: " + sku})
		return
	}

	qrBytes, err := generateQRImage(sku)
	if err != nil {
		bt.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "⚠️ Gagal membuat stiker barcode."})
		return
	}

	caption := fmt.Sprintf("🏷️ *STIKER LABEL QR CODE*\n\n📦 *%s*\n• SKU: `%s`\n• Ruangan: %s\n• Kategori: %s\n\n_Bisa langsung dicetak ke printer label portable._",
		it.Name, sku, it.Location, it.Category)

	kb := models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "‹ Kembali ke Menu", CallbackData: "cmd:start"},
			},
		},
	}

	_, _ = bt.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID: chatID,
		Photo: &models.InputFileUpload{
			Filename: fmt.Sprintf("%s.png", sku),
			Data:     bytes.NewReader(qrBytes),
		},
		Caption:     caption,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: kb,
	})
}

func generateQRImage(data string) ([]byte, error) {
	code, err := qr.Encode(data, qr.M, qr.Auto)
	if err != nil {
		return nil, err
	}
	scaled, err := barcode.Scale(code, 350, 350)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, scaled); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (b *Bot) getPrediksiText(ctx context.Context) string {
	rows, err := b.pool.Query(ctx, `
		SELECT
			i.sku,
			i.name,
			i.current_stock,
			i.unit,
			COALESCE(SUM(t.quantity), 0) AS total_out
		FROM inventory_items i
		LEFT JOIN stock_transactions t
			ON t.item_id = i.id
		   AND t.type = 'OUT'
		   AND t.timestamp >= now() - INTERVAL '30 days'
		WHERE i.track_stock = TRUE AND i.current_stock > 0
		GROUP BY i.id, i.sku, i.name, i.current_stock, i.unit
		HAVING COALESCE(SUM(t.quantity), 0) > 0
		ORDER BY (i.current_stock::float / (COALESCE(SUM(t.quantity), 1)::float / 30.0)) ASC
		LIMIT 10
	`)
	if err != nil {
		return "⚠️ Gagal menganalisis riwayat mutasi stok."
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("🔮 *Prediksi Kehabisan Stok (Burn-Rate 30 Hari):*\n\n")

	count := 0
	for rows.Next() {
		var sku, name, unit string
		var curStock, totalOut int32
		rows.Scan(&sku, &name, &curStock, &unit, &totalOut)

		burnRatePerDay := float64(totalOut) / 30.0
		daysLeft := int(float64(curStock) / burnRatePerDay)

		badge := "🟢"
		if daysLeft <= 7 {
			badge = "🔴 *KRITIS*"
		} else if daysLeft <= 14 {
			badge = "🟡 *WASPADA*"
		}

		sb.WriteString(fmt.Sprintf("%s *%s* (`%s`)\n  Stok: %d %s | Konsumsi: %.1f %s/hari\n  Estimasi Habis: *~%d hari lagi*\n\n",
			badge, name, sku, curStock, unit, burnRatePerDay, unit, daysLeft))
		count++
	}

	if count == 0 {
		sb.WriteString("✅ Belum ada pengeluaran signifikan dalam 30 hari terakhir. Semua stok aman.")
	}
	return sb.String()
}

// handlePrediksi predicts stock run-out based on 30-day burn rate
func (b *Bot) handlePrediksi(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil || !b.allowed(u.Message.Chat.ID) {
		return
	}
	chatID := u.Message.Chat.ID
	text := b.getPrediksiText(ctx)

	bt.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: b.buildBackKeyboard(),
	})
}

func (b *Bot) getRusakText(ctx context.Context) string {
	rows, err := b.pool.Query(ctx,
		`SELECT sku, name, location, condition_status FROM inventory_items WHERE condition_status != 'Berfungsi' ORDER BY location LIMIT 20`)
	if err != nil {
		return "⚠️ Gagal memuat data barang rusak."
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("🛠️ *Daftar Aset Butuh Servis / Rusak:*\n\n")
	count := 0
	for rows.Next() {
		var sku, name, loc, cond string
		rows.Scan(&sku, &name, &loc, &cond)
		icon := "⚠️"
		if cond == "Rusak Berat" {
			icon = "❌"
		}
		sb.WriteString(fmt.Sprintf("%s *%s* (`%s`)\n  Lokasi: %s | Kondisi: *%s*\n\n", icon, name, sku, loc, cond))
		count++
	}

	if count == 0 {
		sb.WriteString("✅ Seluruh aset dalam kondisi *Berfungsi* dengan baik.")
	}
	return sb.String()
}

// handleRusak displays damaged items
func (b *Bot) handleRusak(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil || !b.allowed(u.Message.Chat.ID) {
		return
	}
	chatID := u.Message.Chat.ID
	text := b.getRusakText(ctx)

	bt.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: b.buildBackKeyboard(),
	})
}

func (b *Bot) handleCari(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil || !b.allowed(u.Message.Chat.ID) {
		return
	}
	kw := strings.TrimSpace(strings.TrimPrefix(u.Message.Text, "/cari"))
	if kw == "" {
		bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    u.Message.Chat.ID,
			Text:      "Ketik: `/cari <nama/ruangan>`\nContoh: `/cari tensi` atau `/cari poli`",
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}
	b.searchItems(ctx, bt, u.Message.Chat.ID, kw)
}

func (b *Bot) handleStok(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil || !b.allowed(u.Message.Chat.ID) {
		return
	}
	text := b.getLowStockText(ctx)
	bt.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      u.Message.Chat.ID,
		Text:        text,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: b.buildBackKeyboard(),
	})
}

// handleRekap generates and sends a CSV file
func (b *Bot) handleRekap(ctx context.Context, bt *bot.Bot, u *models.Update) {
	chatID := u.Message.Chat.ID
	rows, err := b.pool.Query(ctx,
		`SELECT sku, name, category, location, current_stock, min_stock, unit, condition_status
		 FROM inventory_items ORDER BY category, name`)
	if err != nil {
		bt.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "⚠️ Gagal membuat rekap."})
		return
	}
	defer rows.Close()

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"SKU", "Nama Barang", "Kategori", "Lokasi Ruangan", "Stok", "Min Stok", "Satuan", "Kondisi"})

	count := 0
	for rows.Next() {
		var sku, name, cat, loc, unit, cond string
		var cur, min int32
		rows.Scan(&sku, &name, &cat, &loc, &cur, &min, &unit, &cond)
		_ = writer.Write([]string{sku, name, cat, loc, strconv.Itoa(int(cur)), strconv.Itoa(int(min)), unit, cond})
		count++
	}
	writer.Flush()

	filename := fmt.Sprintf("Rekap_Inventaris_%s.csv", time.Now().Format("2006-01-02"))
	_, err = bt.SendDocument(ctx, &bot.SendDocumentParams{
		ChatID: chatID,
		Document: &models.InputFileUpload{
			Filename: filename,
			Data:     bytes.NewReader(buf.Bytes()),
		},
		Caption: fmt.Sprintf("📑 Rekap data inventaris puskesmas (%d barang).", count),
	})
	if err != nil {
		bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf("📑 Total: %d barang terdaftar. (Gagal kirim file: %v)", count, err),
		})
	}
}

// handleRekapMutasi generates and sends transaction history CSV
func (b *Bot) handleRekapMutasi(ctx context.Context, bt *bot.Bot, u *models.Update) {
	chatID := u.Message.Chat.ID
	rows, err := b.pool.Query(ctx,
		`SELECT timestamp::text, type, item_sku, item_name, quantity, unit, previous_stock, new_stock, COALESCE(notes, '')
		 FROM stock_transactions
		 WHERE timestamp >= now() - INTERVAL '30 days'
		 ORDER BY timestamp DESC LIMIT 200`)
	if err != nil {
		bt.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "⚠️ Gagal membuat rekap mutasi."})
		return
	}
	defer rows.Close()

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"Waktu", "Jenis Mutasi", "SKU", "Nama Barang", "Jumlah", "Satuan", "Stok Sebelum", "Stok Sesudah", "Catatan"})

	count := 0
	for rows.Next() {
		var ts, mType, sku, name, unit, notes string
		var qty, prev, next int32
		rows.Scan(&ts, &mType, &sku, &name, &qty, &unit, &prev, &next, &notes)
		_ = writer.Write([]string{ts, mType, sku, name, strconv.Itoa(int(qty)), unit, strconv.Itoa(int(prev)), strconv.Itoa(int(next)), notes})
		count++
	}
	writer.Flush()

	filename := fmt.Sprintf("Rekap_Mutasi_30Hari_%s.csv", time.Now().Format("2006-01-02"))
	_, err = bt.SendDocument(ctx, &bot.SendDocumentParams{
		ChatID: chatID,
		Document: &models.InputFileUpload{
			Filename: filename,
			Data:     bytes.NewReader(buf.Bytes()),
		},
		Caption: fmt.Sprintf("📊 Rekap transaksi mutasi masuk/keluar 30 hari terakhir (%d mutasi).", count),
	})
	if err != nil {
		bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf("📊 Total: %d mutasi tercatat. (Gagal kirim file: %v)", count, err),
		})
	}
}

// handleCallback processes inline keyboard button clicks
func (b *Bot) handleCallback(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.CallbackQuery == nil {
		return
	}
	cb := u.CallbackQuery

	// Always answer callback query immediately to prevent client UI freeze / loading state
	_, _ = bt.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: cb.ID,
	})

	var chatID int64
	var msgID int
	if cb.Message.Message != nil {
		chatID = cb.Message.Message.Chat.ID
		msgID = cb.Message.Message.ID
	}
	if chatID == 0 && cb.From.ID != 0 {
		chatID = cb.From.ID
	}
	if chatID == 0 {
		return
	}

	var chatObj *models.Chat
	if cb.Message.Message != nil {
		chatObj = &cb.Message.Message.Chat
	}
	var fromObj *models.User
	if cb.From.ID != 0 {
		fromObj = &cb.From
	}
	b.autoRegisterSubscriber(ctx, chatObj, fromObj)

	if !b.allowed(chatID) {
		b.editMessage(ctx, bt, chatID, msgID, "⚠️ Akun Anda belum terdaftar sebagai subscriber aktif.", b.buildBackKeyboard())
		return
	}

	data := cb.Data

	// 1. Wizard category selection
	if strings.HasPrefix(data, "wizcat:") {
		catName := strings.TrimPrefix(data, "wizcat:")
		wiz := b.getWizard(chatID, cb.From.ID)
		if wiz != nil {
			wiz.Category = catName
			wiz.Step = "location"
			b.setWizard(chatID, wiz, cb.From.ID)

			rows, err := b.pool.Query(ctx, `SELECT name FROM locations ORDER BY name LIMIT 12`)
			if err != nil {
				b.editMessage(ctx, bt, chatID, msgID, "⚠️ Gagal memuat ruangan.", b.buildBackKeyboard())
				return
			}
			defer rows.Close()

			var kbRows [][]models.InlineKeyboardButton
			for rows.Next() {
				var locName string
				rows.Scan(&locName)
				kbRows = append(kbRows, []models.InlineKeyboardButton{
					{Text: "📍 " + locName, CallbackData: "wizloc:" + locName},
				})
			}
			kbRows = append(kbRows, []models.InlineKeyboardButton{
				{Text: "❌ Batalkan", CallbackData: "cmd:batal"},
			})

			msgText := fmt.Sprintf("📝 *Tambah Barang (Langkah 3/4)*\n\nNama: *%s*\nKategori: *%s*\n\nSilakan pilih *Ruangan / Lokasi* di bawah:", wiz.Name, wiz.Category)
			b.editMessage(ctx, bt, chatID, msgID, msgText, models.InlineKeyboardMarkup{InlineKeyboard: kbRows})
			return
		}
	}

	// 2. Wizard location selection
	if strings.HasPrefix(data, "wizloc:") {
		locName := strings.TrimPrefix(data, "wizloc:")
		wiz := b.getWizard(chatID, cb.From.ID)
		if wiz != nil {
			wiz.Location = locName
			wiz.Step = "stock"
			b.setWizard(chatID, wiz, cb.From.ID)

			msgText := fmt.Sprintf("🔢 *Tambah Barang (Langkah 4/4)*\n\nNama: *%s*\nKategori: *%s*\nRuangan: *%s*\n\nSilakan ketik *Jumlah Stok Awal & Satuan* pada chat.\nContoh: `10 unit` atau `5 box` atau `1`", wiz.Name, wiz.Category, wiz.Location)
			b.editMessage(ctx, bt, chatID, msgID, msgText, b.buildBackKeyboard())
			return
		}
	}

	if !strings.HasPrefix(data, "cmd:") {
		return
	}
	action := strings.TrimPrefix(data, "cmd:")

	switch {
	case action == "start":
		intro := "📋 *Menu Pintar Inventaris Puskesmas*\n\nPilih menu di bawah, ketik nama barang langsung untuk cari, atau kirim foto barcode fisik:"
		b.editMessage(ctx, bt, chatID, msgID, intro, b.buildMenuKeyboard(ctx))

	case action == "batal":
		b.clearWizard(chatID, cb.From.ID)
		b.editMessage(ctx, bt, chatID, msgID, "❌ Proses dibatalkan.", b.buildMenuKeyboard(ctx))

	case action == "tambah_wizard":
		b.setWizard(chatID, &wizardState{Step: "name"}, cb.From.ID)
		text := "📝 *Tambah Barang Baru (Langkah 1/4)*\n\n" +
			"Silakan ketik *Nama Barang* yang ingin didaftarkan pada chat:\n" +
			"(Atau klik Batalkan di bawah)"
		kb := models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{{Text: "❌ Batalkan", CallbackData: "cmd:batal"}},
			},
		}
		b.editMessage(ctx, bt, chatID, msgID, text, kb)

	case action == "stok":
		text := b.getLowStockText(ctx)
		b.editMessage(ctx, bt, chatID, msgID, text, b.buildBackKeyboard())

	case action == "rusak":
		text := b.getRusakText(ctx)
		b.editMessage(ctx, bt, chatID, msgID, text, b.buildBackKeyboard())

	case action == "prediksi":
		text := b.getPrediksiText(ctx)
		b.editMessage(ctx, bt, chatID, msgID, text, b.buildBackKeyboard())

	case action == "rekap":
		b.editMessage(ctx, bt, chatID, msgID, "⏳ Menyiapkan dan mengirimkan berkas Rekap CSV...", b.buildBackKeyboard())
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()
			b.handleRekap(bgCtx, bt, &models.Update{Message: &models.Message{Chat: models.Chat{ID: chatID}}})
		}()

	case action == "rekap_mutasi":
		b.editMessage(ctx, bt, chatID, msgID, "⏳ Menyiapkan dan mengirimkan berkas Rekap Mutasi CSV...", b.buildBackKeyboard())
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()
			b.handleRekapMutasi(bgCtx, bt, &models.Update{Message: &models.Message{Chat: models.Chat{ID: chatID}}})
		}()

	case action == "bantu":
		text := b.getBantuText()
		b.editMessage(ctx, bt, chatID, msgID, text, b.buildBackKeyboard())

	case action == "kategori_list":
		rows, err := b.pool.Query(ctx, `SELECT category, count(*) FROM inventory_items GROUP BY category ORDER BY category LIMIT 14`)
		if err != nil {
			b.editMessage(ctx, bt, chatID, msgID, "⚠️ Gagal memuat kategori.", b.buildBackKeyboard())
			return
		}
		defer rows.Close()

		var kbRows [][]models.InlineKeyboardButton
		for rows.Next() {
			var cat string
			var count int
			rows.Scan(&cat, &count)
			kbRows = append(kbRows, []models.InlineKeyboardButton{
				{Text: fmt.Sprintf("📂 %s (%d barang)", cat, count), CallbackData: fmt.Sprintf("cmd:cat:%s", cat)},
			})
		}
		kbRows = append(kbRows, []models.InlineKeyboardButton{
			{Text: "‹ Kembali ke Menu", CallbackData: "cmd:start"},
		})
		b.editMessage(ctx, bt, chatID, msgID, "📂 *Pilih Kategori Barang:*", models.InlineKeyboardMarkup{InlineKeyboard: kbRows})

	case strings.HasPrefix(action, "cat:"):
		cat := strings.TrimPrefix(action, "cat:")
		rows, err := b.pool.Query(ctx,
			`SELECT sku, name, current_stock, unit FROM inventory_items WHERE category=$1 ORDER BY name LIMIT 10`, cat)
		if err != nil {
			b.editMessage(ctx, bt, chatID, msgID, "⚠️ Gagal memuat barang.", b.buildBackKeyboard())
			return
		}
		defer rows.Close()

		var kbRows [][]models.InlineKeyboardButton
		for rows.Next() {
			var sku, name, unit string
			var stock int32
			rows.Scan(&sku, &name, &stock, &unit)
			kbRows = append(kbRows, []models.InlineKeyboardButton{
				{Text: fmt.Sprintf("%s (%d %s)", name, stock, unit), CallbackData: fmt.Sprintf("cmd:item:%s", sku)},
			})
		}
		kbRows = append(kbRows, []models.InlineKeyboardButton{
			{Text: "‹ Kembali ke Kategori", CallbackData: "cmd:kategori_list"},
			{Text: "‹ Menu Utama", CallbackData: "cmd:start"},
		})
		b.editMessage(ctx, bt, chatID, msgID, fmt.Sprintf("📂 *Barang Kategori:* %s", cat), models.InlineKeyboardMarkup{InlineKeyboard: kbRows})

	case action == "ruangan_list":
		rows, err := b.pool.Query(ctx, `SELECT location, count(*) FROM inventory_items GROUP BY location ORDER BY location LIMIT 14`)
		if err != nil {
			b.editMessage(ctx, bt, chatID, msgID, "⚠️ Gagal memuat ruangan.", b.buildBackKeyboard())
			return
		}
		defer rows.Close()

		var kbRows [][]models.InlineKeyboardButton
		for rows.Next() {
			var loc string
			var count int
			rows.Scan(&loc, &count)
			kbRows = append(kbRows, []models.InlineKeyboardButton{
				{Text: fmt.Sprintf("📍 %s (%d barang)", loc, count), CallbackData: fmt.Sprintf("cmd:loc:%s", loc)},
			})
		}
		kbRows = append(kbRows, []models.InlineKeyboardButton{
			{Text: "‹ Kembali ke Menu", CallbackData: "cmd:start"},
		})
		b.editMessage(ctx, bt, chatID, msgID, "📍 *Pilih Ruangan / Lokasi:*", models.InlineKeyboardMarkup{InlineKeyboard: kbRows})

	case strings.HasPrefix(action, "loc:"):
		loc := strings.TrimPrefix(action, "loc:")
		rows, err := b.pool.Query(ctx,
			`SELECT sku, name, current_stock, unit FROM inventory_items WHERE location=$1 ORDER BY name LIMIT 10`, loc)
		if err != nil {
			b.editMessage(ctx, bt, chatID, msgID, "⚠️ Gagal memuat barang.", b.buildBackKeyboard())
			return
		}
		defer rows.Close()

		var kbRows [][]models.InlineKeyboardButton
		for rows.Next() {
			var sku, name, unit string
			var stock int32
			rows.Scan(&sku, &name, &stock, &unit)
			kbRows = append(kbRows, []models.InlineKeyboardButton{
				{Text: fmt.Sprintf("%s (%d %s)", name, stock, unit), CallbackData: fmt.Sprintf("cmd:item:%s", sku)},
			})
		}
		kbRows = append(kbRows, []models.InlineKeyboardButton{
			{Text: "‹ Kembali ke Ruangan", CallbackData: "cmd:ruangan_list"},
			{Text: "‹ Menu Utama", CallbackData: "cmd:start"},
		})
		b.editMessage(ctx, bt, chatID, msgID, fmt.Sprintf("📍 *Barang di Ruangan:* %s", loc), models.InlineKeyboardMarkup{InlineKeyboard: kbRows})

	case strings.HasPrefix(action, "item:"):
		sku := strings.TrimPrefix(action, "item:")
		var it struct {
			SKU       string
			Name      string
			Category  string
			Location  string
			Stock     int32
			MinStock  int32
			Unit      string
			Condition string
		}
		err := b.pool.QueryRow(ctx,
			`SELECT sku, name, category, location, current_stock, min_stock, unit, condition_status
			 FROM inventory_items WHERE sku ILIKE $1 LIMIT 1`, sku).
			Scan(&it.SKU, &it.Name, &it.Category, &it.Location, &it.Stock, &it.MinStock, &it.Unit, &it.Condition)
		if err != nil {
			b.editMessage(ctx, bt, chatID, msgID, "⚠️ Barang tidak ditemukan.", b.buildBackKeyboard())
			return
		}
		b.sendItemCard(ctx, bt, chatID, it.SKU, it.Name, it.Stock, it.MinStock, it.Unit, it.Location, it.Category, it.Condition)

	case strings.HasPrefix(action, "label:") || strings.HasPrefix(action, "qr:"):
		sku := strings.TrimPrefix(strings.TrimPrefix(action, "label:"), "qr:")
		b.sendQRLabel(ctx, bt, chatID, sku)

	case strings.HasPrefix(action, "add:"):
		parts := strings.Split(strings.TrimPrefix(action, "add:"), ":")
		if len(parts) >= 2 {
			sku := parts[0]
			qty, _ := strconv.Atoi(parts[1])
			if qty > 0 {
				b.executeStockMutation(ctx, bt, chatID, sku, int32(qty), "IN")
			}
		}

	case strings.HasPrefix(action, "sub:"):
		parts := strings.Split(strings.TrimPrefix(action, "sub:"), ":")
		if len(parts) >= 2 {
			sku := parts[0]
			qty, _ := strconv.Atoi(parts[1])
			if qty > 0 {
				b.executeStockMutation(ctx, bt, chatID, sku, int32(qty), "OUT")
			}
		}

	case strings.HasPrefix(action, "pindah:"):
		sku := strings.TrimPrefix(action, "pindah:")
		rows, err := b.pool.Query(ctx, `SELECT name FROM locations ORDER BY name LIMIT 12`)
		if err != nil {
			b.editMessage(ctx, bt, chatID, msgID, "⚠️ Gagal memuat ruangan.", b.buildBackKeyboard())
			return
		}
		defer rows.Close()

		var kbRows [][]models.InlineKeyboardButton
		for rows.Next() {
			var locName string
			rows.Scan(&locName)
			kbRows = append(kbRows, []models.InlineKeyboardButton{
				{Text: "📍 " + locName, CallbackData: fmt.Sprintf("cmd:setloc:%s:%s", sku, locName)},
			})
		}
		kbRows = append(kbRows, []models.InlineKeyboardButton{
			{Text: "‹ Batal", CallbackData: fmt.Sprintf("cmd:item:%s", sku)},
		})
		b.editMessage(ctx, bt, chatID, msgID, fmt.Sprintf("📍 *Pilih Ruangan Tujuan Baru untuk Barang:* `%s`", sku), models.InlineKeyboardMarkup{InlineKeyboard: kbRows})

	case strings.HasPrefix(action, "setloc:"):
		parts := strings.Split(strings.TrimPrefix(action, "setloc:"), ":")
		if len(parts) >= 2 {
			sku := parts[0]
			newLoc := parts[1]
			b.executeRelocate(ctx, bt, chatID, sku, newLoc)
		}

	case strings.HasPrefix(action, "kondisi:"):
		sku := strings.TrimPrefix(action, "kondisi:")
		kbRows := [][]models.InlineKeyboardButton{
			{
				{Text: "✅ Berfungsi", CallbackData: fmt.Sprintf("cmd:setcond:%s:Berfungsi", sku)},
				{Text: "⚠️ Rusak Ringan", CallbackData: fmt.Sprintf("cmd:setcond:%s:Rusak Ringan", sku)},
			},
			{
				{Text: "❌ Rusak Berat", CallbackData: fmt.Sprintf("cmd:setcond:%s:Rusak Berat", sku)},
				{Text: "🔧 Sedang Servis", CallbackData: fmt.Sprintf("cmd:setcond:%s:Sedang Servis", sku)},
			},
			{
				{Text: "‹ Batal", CallbackData: fmt.Sprintf("cmd:item:%s", sku)},
			},
		}
		b.editMessage(ctx, bt, chatID, msgID, fmt.Sprintf("🛠️ *Pilih Status Kondisi Aset:* `%s`", sku), models.InlineKeyboardMarkup{InlineKeyboard: kbRows})

	case strings.HasPrefix(action, "setcond:"):
		parts := strings.Split(strings.TrimPrefix(action, "setcond:"), ":")
		if len(parts) >= 2 {
			sku := parts[0]
			newCond := parts[1]
			_, err := b.pool.Exec(ctx, `UPDATE inventory_items SET condition_status=$2, updated_at=now() WHERE sku ILIKE $1`, sku, newCond)
			if err == nil {
				text := fmt.Sprintf("🛠️ *Kondisi Aset Diperbarui!*\n\n• SKU: `%s`\n• Status Baru: *%s*", sku, newCond)
				b.editMessage(ctx, bt, chatID, msgID, text, b.buildBackKeyboard())
			}
		}

	case strings.HasPrefix(action, "del:"):
		sku := strings.TrimPrefix(action, "del:")
		text := fmt.Sprintf("⚠️ *Konfirmasi Hapus Barang*\n\nApakah Anda yakin ingin menghapus barang `%s` secara permanen dari inventaris?", sku)
		kb := models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "🗑️ Ya, Hapus Permanen", CallbackData: fmt.Sprintf("cmd:delyes:%s", sku)},
				},
				{
					{Text: "‹ Batal", CallbackData: fmt.Sprintf("cmd:item:%s", sku)},
				},
			},
		}
		b.editMessage(ctx, bt, chatID, msgID, text, kb)

	case strings.HasPrefix(action, "delyes:"):
		sku := strings.TrimPrefix(action, "delyes:")
		_, err := b.pool.Exec(ctx, `DELETE FROM inventory_items WHERE sku ILIKE $1`, sku)
		if err == nil {
			text := fmt.Sprintf("✅ *Barang `%s` berhasil dihapus permanen.*", sku)
			b.editMessage(ctx, bt, chatID, msgID, text, b.buildBackKeyboard())
		} else {
			b.editMessage(ctx, bt, chatID, msgID, "⚠️ Gagal menghapus barang.", b.buildBackKeyboard())
		}

	case strings.HasPrefix(action, "custom:"):
		cmdName := strings.TrimPrefix(action, "custom:")
		cleanCmd := strings.TrimPrefix(strings.ToLower(cmdName), "/")
		var resp, label string
		err := b.pool.QueryRow(ctx, `SELECT response, label FROM telegram_commands WHERE (LOWER(command)=$1 OR LOWER(command)=$2) AND active=TRUE LIMIT 1`, cleanCmd, "/"+cleanCmd).Scan(&resp, &label)
		if err == nil && resp != "" {
			b.editMessage(ctx, bt, chatID, msgID, fmt.Sprintf("📌 *%s*\n\n%s", label, resp), b.buildBackKeyboard())
		} else {
			b.editMessage(ctx, bt, chatID, msgID, fmt.Sprintf("Perintah `/%s` belum memiliki pesan respon.", cmdName), b.buildBackKeyboard())
		}

	case strings.HasPrefix(action, "quick_masuk:"):
		sku := strings.TrimPrefix(action, "quick_masuk:")
		text := fmt.Sprintf("📥 *Catat Barang Masuk*\n\nKirim format:\n`/masuk %s <jumlah>`\n\nContoh: `/masuk %s 10`", sku, sku)
		b.editMessage(ctx, bt, chatID, msgID, text, b.buildBackKeyboard())

	case strings.HasPrefix(action, "cari_sku:"):
		sku := strings.TrimPrefix(action, "cari_sku:")
		var name, unit, loc, cat, cond string
		var cur, min int32
		err := b.pool.QueryRow(ctx, `
			SELECT sku, name, current_stock, min_stock, unit,
			       COALESCE(location, '-'), COALESCE(category, '-'), COALESCE(condition_status, 'Berfungsi')
			FROM inventory_items
			WHERE sku ILIKE $1 LIMIT 1`, sku).Scan(&sku, &name, &cur, &min, &unit, &loc, &cat, &cond)
		if err == nil {
			b.sendItemCard(ctx, bt, chatID, sku, name, cur, min, unit, loc, cat, cond)
		} else {
			b.editMessage(ctx, bt, chatID, msgID, "Barang tidak ditemukan.", b.buildBackKeyboard())
		}

	default:
		b.editMessage(ctx, bt, chatID, msgID, "Perintah tidak ditemukan.", b.buildBackKeyboard())
	}
}

func (b *Bot) getLowStockText(ctx context.Context) string {
	rows, _ := b.pool.Query(ctx,
		`SELECT name, sku, current_stock, min_stock, unit FROM inventory_items WHERE current_stock <= min_stock ORDER BY current_stock LIMIT 20`)
	defer rows.Close()
	var sb strings.Builder
	sb.WriteString("⚠️ *Daftar Barang Stok Menipis:*\n\n")
	count := 0
	for rows.Next() {
		var name, sku, unit string
		var cur, min int32
		rows.Scan(&name, &sku, &cur, &min, &unit)
		sb.WriteString(fmt.Sprintf("• *%s* (`%s`)\n  Stok: *%d %s* (Batas min: %d)\n", name, sku, cur, unit, min))
		count++
	}
	if count == 0 {
		sb.WriteString("✅ Semua stok aman di atas batas minimum.")
	}
	return sb.String()
}

// dailyAlert runs periodically and checks telegram_alert_daily_time (WIB)
func (b *Bot) dailyAlert(ctx context.Context) {
	loc := time.FixedZone("WIB", 7*3600)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	var lastRunDate string

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			nowWIB := now.In(loc)
			today := nowWIB.Format("2006-01-02")
			if lastRunDate == today {
				continue
			}

			var enabled string
			var scheduledTime string
			_ = b.pool.QueryRow(ctx, `SELECT value FROM app_settings WHERE key='telegram_alert_low_stock'`).Scan(&enabled)
			_ = b.pool.QueryRow(ctx, `SELECT value FROM app_settings WHERE key='telegram_alert_daily_time'`).Scan(&scheduledTime)

			if strings.ToLower(strings.TrimSpace(enabled)) != "true" {
				continue
			}

			scheduledTime = strings.TrimSpace(scheduledTime)
			if scheduledTime == "" {
				scheduledTime = "07:00"
			}

			currentHM := nowWIB.Format("15:04")
			if currentHM == scheduledTime {
				lastRunDate = today
				b.sendLowStockAlert(ctx)
			}
		}
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

	kb := models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "📦 Cek Detail Stok", CallbackData: "cmd:stok"},
				{Text: "🔮 Prediksi Kehabisan", CallbackData: "cmd:prediksi"},
			},
		},
	}

	for _, t := range targets {
		ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
		_, err := b.botInst.SendMessage(ctx2, &bot.SendMessageParams{
			ChatID:      t.chatID,
			Text:        text,
			ParseMode:   models.ParseModeMarkdown,
			ReplyMarkup: kb,
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

// SyncBotCommands registers bot commands with Telegram API and sets chat menu button
func (b *Bot) SyncBotCommands(ctx context.Context) error {
	if b.botInst == nil {
		b.Start(ctx)
		if b.botInst == nil {
			return fmt.Errorf("bot belum diinisialisasi atau token belum aktif")
		}
	}

	// 1. Default built-in commands
	cmds := []models.BotCommand{
		{Command: "start", Description: "Menu Utama & Dashboard"},
		{Command: "cari", Description: "Cari Barang (SKU / Nama)"},
		{Command: "tambah", Description: "Tambah Barang Baru (Wizard)"},
		{Command: "masuk", Description: "Catat Barang Masuk (/masuk SKU Jumlah)"},
		{Command: "keluar", Description: "Catat Barang Keluar (/keluar SKU Jumlah)"},
		{Command: "opname", Description: "Sesuaikan Stok Fisik (/opname SKU Stok)"},
		{Command: "label", Description: "Cetak Label QR & Barcode (/label SKU)"},
		{Command: "pindah", Description: "Pindah Lokasi (/pindah SKU ke Ruangan)"},
		{Command: "prediksi", Description: "Prediksi Hari Stok Habis (Burn-Rate)"},
		{Command: "rusak", Description: "Daftar Barang Rusak & Servis"},
		{Command: "rekap", Description: "Unduh Data Semua Barang (CSV)"},
		{Command: "rekap_mutasi", Description: "Unduh Rekap Mutasi 30 Hari (CSV)"},
		{Command: "bantu", Description: "Bantuan & Panduan Lengkap"},
	}

	// 2. Fetch custom commands from telegram_commands
	rows, err := b.pool.Query(ctx, `SELECT command, description FROM telegram_commands WHERE active=TRUE ORDER BY sort_order`)
	if err == nil {
		defer rows.Close()
		existing := map[string]bool{}
		for _, c := range cmds {
			existing[c.Command] = true
		}
		for rows.Next() {
			var c, d string
			if err := rows.Scan(&c, &d); err == nil {
				c = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(c), "/"))
				if c != "" && !existing[c] {
					existing[c] = true
					if d == "" {
						d = "Menu " + c
					}
					cmds = append(cmds, models.BotCommand{
						Command:     c,
						Description: d,
					})
				}
			}
		}
	}

	ctxReq, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	_, err = b.botInst.SetMyCommands(ctxReq, &bot.SetMyCommandsParams{
		Commands: cmds,
	})
	if err != nil {
		slog.Error("SetMyCommands failed", "err", err)
	}

	webURL := b.getWebAppURL(ctx)
	if webURL != "" {
		ctxMB, cancelMB := context.WithTimeout(ctx, 15*time.Second)
		defer cancelMB()
		_, errMB := b.botInst.SetChatMenuButton(ctxMB, &bot.SetChatMenuButtonParams{
			MenuButton: &models.MenuButtonWebApp{
				Type: models.MenuButtonTypeWebApp,
				Text: "Buka Inventaris",
				WebApp: models.WebAppInfo{
					URL: webURL,
				},
			},
		})
		if errMB != nil {
			slog.Error("SetChatMenuButton failed", "err", errMB)
		}
	}

	return err
}

// NotifyMovement broadcasts real-time stock mutation notifications to matching active subscribers
func (b *Bot) NotifyMovement(ctx context.Context, txType, sku, name string, qty int32, unit string, prev, next int32, person, notes string) {
	if b.botInst == nil {
		return
	}

	var notifColumn string
	var icon string
	var header string

	switch txType {
	case "IN":
		notifColumn = "notify_in"
		icon = "📥"
		header = "*MUTASI BARANG MASUK*"
	case "OUT":
		notifColumn = "notify_out"
		icon = "📤"
		header = "*MUTASI BARANG KELUAR*"
	case "ADJUST+", "ADJUST-":
		notifColumn = "notify_adjust"
		icon = "⚖️"
		header = "*PENYESUAIAN STOK (OPNAME)*"
	default:
		return
	}

	query := fmt.Sprintf(`SELECT chat_id, title FROM telegram_subscribers WHERE active=TRUE AND %s=TRUE`, notifColumn)
	rows, err := b.pool.Query(ctx, query)
	if err != nil {
		return
	}
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

	if len(targets) == 0 {
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s %s\n\n", icon, header))
	sb.WriteString(fmt.Sprintf("🏷️ *Barang:* %s\n", name))
	sb.WriteString(fmt.Sprintf("🔖 *SKU:* `%s`\n", sku))
	sb.WriteString(fmt.Sprintf("🔢 *Jumlah:* %d %s\n", qty, unit))
	sb.WriteString(fmt.Sprintf("📊 *Stok:* %d ➔ *%d %s*\n", prev, next, unit))
	if person != "" {
		sb.WriteString(fmt.Sprintf("👤 *Petugas/Penerima:* %s\n", person))
	}
	if notes != "" {
		sb.WriteString(fmt.Sprintf("📝 *Catatan:* %s\n", notes))
	}
	sb.WriteString(fmt.Sprintf("⏰ *Waktu:* %s WIB", time.Now().In(time.FixedZone("WIB", 7*3600)).Format("02-01-2006 15:04")))

	msg := sb.String()

	kb := models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "📦 Cek Barang Ini", CallbackData: fmt.Sprintf("cmd:cari_sku:%s", sku)},
				{Text: "🏷️ Cetak Label QR", CallbackData: fmt.Sprintf("cmd:qr:%s", sku)},
			},
		},
	}

	for _, t := range targets {
		go func(chatID int64, title string) {
			ctxReq, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_, err := b.botInst.SendMessage(ctxReq, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        msg,
				ParseMode:   models.ParseModeMarkdown,
				ReplyMarkup: kb,
			})
			status := "ok"
			errMsg := ""
			if err != nil {
				status = "fail"
				errMsg = err.Error()
			}
			b.pool.Exec(context.Background(),
				`INSERT INTO telegram_message_log (chat_id, chat_title, message, status, error) VALUES ($1,$2,$3,$4,$5)`,
				chatID, title, msg, status, errMsg)
		}(t.chatID, t.title)
	}

	// Check low stock condition
	var minStock int32
	var trackStock bool
	err = b.pool.QueryRow(ctx, `SELECT min_stock, track_stock FROM inventory_items WHERE sku=$1`, sku).Scan(&minStock, &trackStock)
	if err == nil && trackStock && next <= minStock {
		b.NotifyLowStockItem(ctx, sku, name, next, minStock, unit)
	}
}

// NotifyLowStockItem sends an immediate alert when an item hits or drops below min_stock
func (b *Bot) NotifyLowStockItem(ctx context.Context, sku, name string, current, minStock int32, unit string) {
	if b.botInst == nil {
		return
	}
	rows, err := b.pool.Query(ctx, `SELECT chat_id, title FROM telegram_subscribers WHERE active=TRUE AND notify_low_stock=TRUE`)
	if err != nil {
		return
	}
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
	if len(targets) == 0 {
		return
	}

	text := fmt.Sprintf("⚠️ *PERINGATAN STOK MENIPIS!*\n\n🏷️ *Barang:* %s\n🔖 *SKU:* `%s`\n🚨 *Sisa Stok:* *%d %s* (Batas Minimum: %d %s)\n\nSegera lakukan pengadaan atau mutasi masuk.",
		name, sku, current, unit, minStock, unit)

	kb := models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "📥 Catat Masuk", CallbackData: fmt.Sprintf("cmd:quick_masuk:%s", sku)},
				{Text: "🔮 Prediksi Kehabisan", CallbackData: "cmd:prediksi"},
			},
		},
	}

	for _, t := range targets {
		go func(chatID int64, title string) {
			ctxReq, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_, err := b.botInst.SendMessage(ctxReq, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        text,
				ParseMode:   models.ParseModeMarkdown,
				ReplyMarkup: kb,
			})
			status := "ok"
			errMsg := ""
			if err != nil {
				status = "fail"
				errMsg = err.Error()
			}
			b.pool.Exec(context.Background(),
				`INSERT INTO telegram_message_log (chat_id, chat_title, message, status, error) VALUES ($1,$2,$3,$4,$5)`,
				chatID, title, text, status, errMsg)
		}(t.chatID, t.title)
	}
}
