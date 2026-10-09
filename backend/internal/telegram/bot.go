package telegram

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
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
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/oned"
	"github.com/makiuchi-d/gozxing/qrcode"
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
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/masuk", bot.MatchTypePrefix, b.handleMasuk)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/keluar", bot.MatchTypePrefix, b.handleKeluar)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/cari", bot.MatchTypePrefix, b.handleCari)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/stok", bot.MatchTypeExact, b.handleStok)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/tambah", bot.MatchTypePrefix, b.handleTambah)
	botInst.RegisterHandler(bot.HandlerTypeMessageText, "/rekap", bot.MatchTypeExact, b.handleRekap)

	// Register callback query handler for inline buttons
	botInst.RegisterHandler(bot.HandlerTypeCallbackQueryData, "", bot.MatchTypePrefix, b.handleCallback)

	go b.dailyAlert(ctx)

	// Ensure standard commands show in menu if DB has old seed
	_, _ = b.pool.Exec(ctx, `UPDATE telegram_commands SET is_menu=TRUE WHERE command IN ('stok', 'cari', 'masuk', 'bantu') AND (SELECT count(*) FROM telegram_commands WHERE is_menu=TRUE AND command != 'start') = 0`)

	// Configure Telegram Mini App Menu Button
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
		{Text: "📦 Cek Stok Menipis", CallbackData: "cmd:stok"},
		{Text: "➕ Tambah Barang", CallbackData: "cmd:tambah_menu"},
	})

	kbRows = append(kbRows, []models.InlineKeyboardButton{
		{Text: "📂 Jelajah Kategori", CallbackData: "cmd:kategori_list"},
		{Text: "📍 Jelajah Ruangan", CallbackData: "cmd:ruangan_list"},
	})

	kbRows = append(kbRows, []models.InlineKeyboardButton{
		{Text: "🔍 Cari Barang", CallbackData: "cmd:cari"},
		{Text: "📑 Unduh Rekap CSV", CallbackData: "cmd:rekap"},
	})

	kbRows = append(kbRows, []models.InlineKeyboardButton{
		{Text: "❓ Bantuan Perintah", CallbackData: "cmd:bantu"},
	})

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

// ---- Handlers ----

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
		text := fmt.Sprintf("👋 *Halo, %s!*\n\n🆔 *Chat ID Anda:* `%d`\n\nAkun Anda belum terdaftar sebagai subscriber aktif di sistem Inventaris Puskesmas.\n\nSilakan salin Chat ID di atas dan tambahkan pada menu *Administrasi > Telegram* di web panel untuk mendapatkan akses penuh.", name, chatID)
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
	intro := "📋 *Menu Pintar Inventaris Kantor*\n\nSilakan pilih opsi menu di bawah atau kirim foto barcode/ketik nama barang langsung:"
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

func (b *Bot) handleBantu(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil || !b.allowed(u.Message.Chat.ID) {
		return
	}
	text := "💡 *Daftar Perintah Bot Pintar:*\n\n" +
		"• *Pencarian Cerdas:* Ketik nama barang langsung di chat (tanpa garis miring), cth: `tensimeter` atau `kursi roda`.\n" +
		"• *Scan Barcode:* Kirim foto barcode/QR barang fisik ke chat bot.\n" +
		"• `/tambah <Nama> | <Kategori> | <Ruangan> | <Stok> | <Satuan>` — Tambah barang baru cepat.\n" +
		"• `/masuk <SKU> <jumlah>` — Tambah stok barang masuk.\n" +
		"• `/keluar <SKU> <jumlah>` — Catat pengeluaran barang.\n" +
		"• `/cari <kata kunci>` — Cari barang berdasarkan nama/ruangan.\n" +
		"• `/stok` — Tampilkan daftar barang yang menipis.\n" +
		"• `/rekap` — Unduh file spreadsheet CSV rekap inventaris.\n" +
		"• `/start` — Buka menu interaktif utama."
	bt.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    u.Message.Chat.ID,
		Text:      text,
		ParseMode: models.ParseModeMarkdown,
	})
}

// handleDefault catches plain text searches and photos
func (b *Bot) handleDefault(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil || !b.allowed(u.Message.Chat.ID) {
		return
	}

	// 1. Photo: scan for Barcode / QR Code
	if len(u.Message.Photo) > 0 {
		b.handlePhoto(ctx, bt, u)
		return
	}

	// 2. Plain Text: Smart Natural Search
	text := strings.TrimSpace(u.Message.Text)
	if text == "" || strings.HasPrefix(text, "/") {
		return
	}
	b.handleSmartText(ctx, bt, u, text)
}

// handleSmartText searches items when user sends a word without slash
func (b *Bot) handleSmartText(ctx context.Context, bt *bot.Bot, u *models.Update, query string) {
	chatID := u.Message.Chat.ID
	lower := strings.ToLower(query)

	if lower == "menu" || lower == "start" || lower == "halo" || lower == "hai" {
		b.showMenu(ctx, bt, chatID, u.Message.ID)
		return
	}
	if lower == "stok" || lower == "stok menipis" || lower == "kritis" || lower == "habis" {
		text := b.getLowStockText(ctx)
		bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      text,
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}
	if lower == "rekap" || lower == "laporan" {
		b.handleRekap(ctx, bt, u)
		return
	}

	// Query database
	rows, err := b.pool.Query(ctx,
		`SELECT id::text, sku, name, current_stock, min_stock, unit, location, category, condition_status
		 FROM inventory_items
		 WHERE name ILIKE $1 OR sku ILIKE $1 OR location ILIKE $1 OR category ILIKE $1
		 ORDER BY name LIMIT 5`, "%"+query+"%")
	if err != nil {
		bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "⚠️ Gagal mencari data barang.",
		})
		return
	}
	defer rows.Close()

	type foundItem struct {
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
	var list []foundItem
	for rows.Next() {
		var it foundItem
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

	webURL := b.getWebAppURL(ctx)
	for _, it := range list {
		statusIcon := "🟢"
		if it.Stock <= it.MinStock {
			statusIcon = "🔴"
		}
		card := fmt.Sprintf("%s *%s*\n• SKU: `%s`\n• Stok: *%d %s* (min %d)\n• Ruangan: %s\n• Kategori: %s\n• Kondisi: %s",
			statusIcon, it.Name, it.SKU, it.Stock, it.Unit, it.MinStock, it.Location, it.Category, it.Condition)

		kb := models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "➕ +1", CallbackData: fmt.Sprintf("cmd:add:%s:1", it.SKU)},
					{Text: "➕ +5", CallbackData: fmt.Sprintf("cmd:add:%s:5", it.SKU)},
					{Text: "➖ -1", CallbackData: fmt.Sprintf("cmd:sub:%s:1", it.SKU)},
				},
				{
					{
						Text:   "🌐 Buka di Web",
						WebApp: &models.WebAppInfo{URL: fmt.Sprintf("%s/items/%s", webURL, it.ID)},
					},
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
		Text:   "🔍 Memindai barcode / QR dari foto...",
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
				Text:      "📸 *Foto Diterima*\n\nBarcode atau QR code tidak terdeteksi pada gambar.\n\n💡 *Tips:* Pastikan barcode tidak blur, pencahayaan cukup terang, dan posisi tegak lurus dengan kamera.",
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
		Unit      string
		Condition string
	}
	err = b.pool.QueryRow(ctx,
		`SELECT id::text, sku, name, category, location, current_stock, unit, condition_status
		 FROM inventory_items
		 WHERE sku = $1 OR serial_number = $1 OR sku ILIKE $2 OR name ILIKE $2
		 LIMIT 1`, code, "%"+code+"%").
		Scan(&it.ID, &it.SKU, &it.Name, &it.Category, &it.Location, &it.Stock, &it.Unit, &it.Condition)

	webURL := b.getWebAppURL(ctx)

	if err != nil {
		text := fmt.Sprintf("📸 *Barcode Terdeteksi!* (%s)\n\nKode: `%s`\n\n⚠️ Barang dengan kode ini belum terdaftar di inventaris.", format, code)
		kb := models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{
						Text:   "➕ Daftarkan via Form Web",
						WebApp: &models.WebAppInfo{URL: fmt.Sprintf("%s/items/new", webURL)},
					},
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

	text := fmt.Sprintf("📸 *Barcode Cocok!* (%s)\n\n📦 *%s*\n• SKU: `%s`\n• Stok: *%d %s*\n• Lokasi: %s\n• Kategori: %s\n• Kondisi: %s",
		format, it.Name, it.SKU, it.Stock, it.Unit, it.Location, it.Category, it.Condition)

	kb := models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "➕ +1 Stok", CallbackData: fmt.Sprintf("cmd:add:%s:1", it.SKU)},
				{Text: "➕ +5 Stok", CallbackData: fmt.Sprintf("cmd:add:%s:5", it.SKU)},
				{Text: "➖ -1 Stok", CallbackData: fmt.Sprintf("cmd:sub:%s:1", it.SKU)},
			},
			{
				{
					Text:   "🌐 Buka Detail di Web",
					WebApp: &models.WebAppInfo{URL: fmt.Sprintf("%s/items/%s", webURL, it.ID)},
				},
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
}

func decodeBarcode(img image.Image) (string, string, error) {
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", "", err
	}
	// Try QR Code
	qr := qrcode.NewQRCodeReader()
	if res, err := qr.Decode(bmp, nil); err == nil && res != nil {
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

// handleTambah adds a new item via quick syntax: /tambah Nama | Kategori | Ruangan | Stok | Satuan
func (b *Bot) handleTambah(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil || !b.allowed(u.Message.Chat.ID) {
		return
	}
	chatID := u.Message.Chat.ID
	raw := strings.TrimSpace(strings.TrimPrefix(u.Message.Text, "/tambah"))
	webURL := b.getWebAppURL(ctx)

	if raw == "" {
		text := "➕ *Tambah Barang Baru*\n\n" +
			"Gunakan format pemisah tegak `|` atau koma `,`:\n" +
			"`/tambah <Nama> | <Kategori> | <Ruangan> | <Stok> | <Satuan>`\n\n" +
			"📝 *Contoh:*\n" +
			"`/tambah Tensimeter Digital | Alat Medis | Poli Umum | 5 | unit`\n" +
			"`/tambah Paracetamol 500mg | Obat | Farmasi | 100 | strip`\n\n" +
			"Atau buka form lengkap melalui tombol di bawah:"
		kb := models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{
						Text:   "➕ Buka Form Lengkap di Web",
						WebApp: &models.WebAppInfo{URL: fmt.Sprintf("%s/items/new", webURL)},
					},
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
		return
	}

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
			Text:   "⚠️ Format kurang lengkap. Minimal: `/tambah Nama | Kategori | Ruangan`",
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

	sku := b.generateSKU(ctx, category)

	var id string
	err := b.pool.QueryRow(ctx,
		`INSERT INTO inventory_items (sku, name, category, location, current_stock, min_stock, unit, condition_status, is_available, track_stock)
		 VALUES ($1, $2, $3, $4, $5, 1, $6, 'Berfungsi', TRUE, TRUE)
		 RETURNING id::text`,
		sku, name, category, location, stock, unit).Scan(&id)
	if err != nil {
		bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "⚠️ Gagal menyimpan barang: " + err.Error(),
		})
		return
	}

	// Record initial transaction
	if stock > 0 {
		b.pool.Exec(ctx,
			`INSERT INTO stock_transactions (type, item_id, item_sku, item_name, quantity, unit, previous_stock, new_stock)
			 VALUES ('IN', $1, $2, $3, $4, $5, 0, $4)`,
			id, sku, name, stock, unit)
	}

	text := fmt.Sprintf("✅ *Barang Berhasil Didaftarkan!*\n\n📦 *%s*\n• SKU: `%s`\n• Kategori: %s\n• Ruangan: %s\n• Stok Awal: *%d %s*",
		name, sku, category, location, stock, unit)

	kb := models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "➕ +1 Stok", CallbackData: fmt.Sprintf("cmd:add:%s:1", sku)},
				{Text: "➕ +5 Stok", CallbackData: fmt.Sprintf("cmd:add:%s:5", sku)},
			},
			{
				{
					Text:   "🌐 Lihat di Web App",
					WebApp: &models.WebAppInfo{URL: fmt.Sprintf("%s/items/%s", webURL, id)},
				},
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

func (b *Bot) generateSKU(ctx context.Context, category string) string {
	var slug string
	_ = b.pool.QueryRow(ctx, `SELECT id FROM categories WHERE name=$1`, category).Scan(&slug)
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

func (b *Bot) handleMasuk(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil || !b.allowed(u.Message.Chat.ID) {
		return
	}
	parts := strings.Fields(u.Message.Text)
	if len(parts) < 3 {
		bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: u.Message.Chat.ID,
			Text:   "Format: `/masuk <SKU> <jumlah>`\nContoh: `/masuk ALAT-2026-001 10`",
			ParseMode: models.ParseModeMarkdown,
		})
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
		Text:   fmt.Sprintf("✅ *Barang Masuk Berhasil!*\n\n📦 *%s* (%s)\n• Tambah: +%d %s\n• Stok Sekarang: *%d %s*", name, sku, qty, unit, next, unit),
		ParseMode: models.ParseModeMarkdown,
	})
}

func (b *Bot) handleKeluar(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil || !b.allowed(u.Message.Chat.ID) {
		return
	}
	parts := strings.Fields(u.Message.Text)
	if len(parts) < 3 {
		bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: u.Message.Chat.ID,
			Text:   "Format: `/keluar <SKU> <jumlah>`\nContoh: `/keluar ALAT-2026-001 2`",
			ParseMode: models.ParseModeMarkdown,
		})
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
		`UPDATE inventory_items SET current_stock = GREATEST(0, current_stock - $2), updated_at = now()
		 WHERE sku = $1 RETURNING name, unit, current_stock + $2, current_stock`,
		sku, qty).Scan(&name, &unit, &prev, &next)
	if err != nil {
		bt.SendMessage(ctx, &bot.SendMessageParams{ChatID: u.Message.Chat.ID, Text: "SKU tidak ditemukan: " + sku})
		return
	}

	b.pool.Exec(ctx, `INSERT INTO stock_transactions (type, item_id, item_sku, item_name, quantity, unit, previous_stock, new_stock)
		VALUES ('OUT', (SELECT id FROM inventory_items WHERE sku=$1), $1, $2, $3, $4, $5, $6)`,
		sku, name, qty, unit, prev, next)

	bt.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: u.Message.Chat.ID,
		Text:   fmt.Sprintf("📤 *Pengeluaran Barang Tercatat!*\n\n📦 *%s* (%s)\n• Keluar: -%d %s\n• Sisa Stok: *%d %s*", name, sku, qty, unit, next, unit),
		ParseMode: models.ParseModeMarkdown,
	})
}

func (b *Bot) handleCari(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil || !b.allowed(u.Message.Chat.ID) {
		return
	}
	kw := strings.TrimSpace(strings.TrimPrefix(u.Message.Text, "/cari"))
	if kw == "" {
		bt.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: u.Message.Chat.ID,
			Text:   "Ketik: `/cari <nama/ruangan>`\nContoh: `/cari tensi` atau `/cari poli`",
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}
	b.handleSmartText(ctx, bt, u, kw)
}

func (b *Bot) handleStok(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.Message == nil || !b.allowed(u.Message.Chat.ID) {
		return
	}
	text := b.getLowStockText(ctx)
	bt.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    u.Message.Chat.ID,
		Text:      text,
		ParseMode: models.ParseModeMarkdown,
		ReplyMarkup: b.buildBackKeyboard(),
	})
}

// handleRekap generates and sends a CSV file to Telegram chat
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
			Text:   fmt.Sprintf("📑 Total: %d barang terdaftar. (Gagal upload dokumen: %v)", count, err),
		})
	}
}

// handleCallback handles all inline keyboard clicks (drilldown, actions, pagination)
func (b *Bot) handleCallback(ctx context.Context, bt *bot.Bot, u *models.Update) {
	if u.CallbackQuery == nil {
		return
	}
	cb := u.CallbackQuery
	chatID := cb.Message.Message.Chat.ID
	if !b.allowed(chatID) {
		return
	}

	bt.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: cb.ID,
	})

	data := cb.Data
	if !strings.HasPrefix(data, "cmd:") {
		return
	}
	action := strings.TrimPrefix(data, "cmd:")
	msgID := cb.Message.Message.ID

	switch {
	case action == "start":
		intro := "📋 *Menu Pintar Inventaris Kantor*\n\nSilakan pilih opsi menu di bawah atau kirim foto barcode/ketik nama barang langsung:"
		b.editMessage(ctx, bt, chatID, msgID, intro, b.buildMenuKeyboard(ctx))

	case action == "stok":
		text := b.getLowStockText(ctx)
		b.editMessage(ctx, bt, chatID, msgID, text, b.buildBackKeyboard())

	case action == "cari":
		text := "🔍 *Pencarian Barang*\n\nKetik langsung nama barang di chat ini tanpa tanda apapun.\nContoh:\n• `tensimeter`\n• `kursi roda`\n• `paracetamol`\n• `poli umum`"
		b.editMessage(ctx, bt, chatID, msgID, text, b.buildBackKeyboard())

	case action == "bantu":
		text := "💡 *Daftar Perintah & Fitur Bot Pintar:*\n\n" +
			"• *Cari Cepat:* Ketik nama barang langsung di chat.\n" +
			"• *Scan Barcode:* Kirim foto barcode/QR barang fisik.\n" +
			"• *Jelajah Kategori & Ruangan:* Gunakan tombol menu interaktif.\n" +
			"• *Tambah Stok Langsung:* Klik tombol `[+1]` atau `[+5]` pada hasil pencarian.\n" +
			"• `/tambah Nama | Kategori | Ruangan | Stok | Satuan` — Input barang baru cepat.\n" +
			"• `/masuk <SKU> <jumlah>` — Tambah stok barang.\n" +
			"• `/keluar <SKU> <jumlah>` — Kurang stok barang.\n" +
			"• `/rekap` — Unduh file CSV rekap seluruh inventaris."
		b.editMessage(ctx, bt, chatID, msgID, text, b.buildBackKeyboard())

	case action == "tambah_menu":
		webURL := b.getWebAppURL(ctx)
		text := "➕ *Tambah Barang Baru*\n\n" +
			"Ketik pesan di chat dengan format:\n" +
			"`/tambah <Nama> | <Kategori> | <Ruangan> | <Stok> | <Satuan>`\n\n" +
			"📝 *Contoh:*\n" +
			"`/tambah Termometer Gun | Alat Medis | Poli Umum | 10 | unit`\n\n" +
			"Atau buka form lengkap di Web App:"
		kb := models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{
						Text:   "➕ Buka Form Lengkap di Web",
						WebApp: &models.WebAppInfo{URL: fmt.Sprintf("%s/items/new", webURL)},
					},
				},
				{
					{Text: "‹ Kembali ke Menu", CallbackData: "cmd:start"},
				},
			},
		}
		b.editMessage(ctx, bt, chatID, msgID, text, kb)

	case action == "rekap":
		// Call handleRekap in a new goroutine
		go b.handleRekap(ctx, bt, &models.Update{Message: &models.Message{Chat: models.Chat{ID: chatID}}})

	case action == "kategori_list":
		// Drilldown: Category list
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
		// Drilldown: Items in category
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
		b.editMessage(ctx, bt, chatID, msgID, fmt.Sprintf("📂 *Barang dalam Kategori:* %s", cat), models.InlineKeyboardMarkup{InlineKeyboard: kbRows})

	case action == "ruangan_list":
		// Drilldown: Location list
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
		// Drilldown: Items in location
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
		// Item Detail
		sku := strings.TrimPrefix(action, "item:")
		var it struct {
			ID        string
			SKU       string
			Name      string
			Category  string
			Location  string
			Stock     int32
			Unit      string
			Condition string
		}
		err := b.pool.QueryRow(ctx,
			`SELECT id::text, sku, name, category, location, current_stock, unit, condition_status
			 FROM inventory_items WHERE sku=$1`, sku).
			Scan(&it.ID, &it.SKU, &it.Name, &it.Category, &it.Location, &it.Stock, &it.Unit, &it.Condition)
		if err != nil {
			b.editMessage(ctx, bt, chatID, msgID, "⚠️ Barang tidak ditemukan.", b.buildBackKeyboard())
			return
		}

		webURL := b.getWebAppURL(ctx)
		text := fmt.Sprintf("📦 *%s*\n• SKU: `%s`\n• Stok: *%d %s*\n• Ruangan: %s\n• Kategori: %s\n• Kondisi: %s",
			it.Name, it.SKU, it.Stock, it.Unit, it.Location, it.Category, it.Condition)

		kb := models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "➕ +1 Stok", CallbackData: fmt.Sprintf("cmd:add:%s:1", it.SKU)},
					{Text: "➕ +5 Stok", CallbackData: fmt.Sprintf("cmd:add:%s:5", it.SKU)},
					{Text: "➖ -1 Stok", CallbackData: fmt.Sprintf("cmd:sub:%s:1", it.SKU)},
				},
				{
					{
						Text:   "🌐 Buka Detail di Web",
						WebApp: &models.WebAppInfo{URL: fmt.Sprintf("%s/items/%s", webURL, it.ID)},
					},
				},
				{
					{Text: "‹ Menu Utama", CallbackData: "cmd:start"},
				},
			},
		}
		b.editMessage(ctx, bt, chatID, msgID, text, kb)

	case strings.HasPrefix(action, "add:"):
		// Quick Action: Add Stock
		parts := strings.Split(strings.TrimPrefix(action, "add:"), ":")
		if len(parts) >= 2 {
			sku := parts[0]
			qty, _ := strconv.Atoi(parts[1])
			if qty > 0 {
				var name, unit, id string
				var next int32
				err := b.pool.QueryRow(ctx,
					`UPDATE inventory_items SET current_stock = current_stock + $2, updated_at = now()
					 WHERE sku = $1 RETURNING id::text, name, unit, current_stock`,
					sku, qty).Scan(&id, &name, &unit, &next)
				if err == nil {
					b.pool.Exec(ctx,
						`INSERT INTO stock_transactions (type, item_id, item_sku, item_name, quantity, unit, previous_stock, new_stock)
						 VALUES ('IN', $1, $2, $3, $4, $5, $6, $7)`,
						id, sku, name, qty, unit, next-int32(qty), next)

					webURL := b.getWebAppURL(ctx)
					text := fmt.Sprintf("✅ *Stok Berhasil Ditambah!*\n\n📦 *%s* (`%s`)\n• Ditambah: +%d %s\n• Stok Baru: *%d %s*",
						name, sku, qty, unit, next, unit)
					kb := models.InlineKeyboardMarkup{
						InlineKeyboard: [][]models.InlineKeyboardButton{
							{
								{Text: "➕ +1 Lagi", CallbackData: fmt.Sprintf("cmd:add:%s:1", sku)},
								{Text: "➕ +5 Lagi", CallbackData: fmt.Sprintf("cmd:add:%s:5", sku)},
								{Text: "➖ -1", CallbackData: fmt.Sprintf("cmd:sub:%s:1", sku)},
							},
							{
								{
									Text:   "🌐 Buka Detail di Web",
									WebApp: &models.WebAppInfo{URL: fmt.Sprintf("%s/items/%s", webURL, id)},
								},
							},
							{
								{Text: "‹ Kembali ke Menu", CallbackData: "cmd:start"},
							},
						},
					}
					b.editMessage(ctx, bt, chatID, msgID, text, kb)
				}
			}
		}

	case strings.HasPrefix(action, "sub:"):
		// Quick Action: Deduct Stock
		parts := strings.Split(strings.TrimPrefix(action, "sub:"), ":")
		if len(parts) >= 2 {
			sku := parts[0]
			qty, _ := strconv.Atoi(parts[1])
			if qty > 0 {
				var name, unit, id string
				var next int32
				err := b.pool.QueryRow(ctx,
					`UPDATE inventory_items SET current_stock = GREATEST(0, current_stock - $2), updated_at = now()
					 WHERE sku = $1 RETURNING id::text, name, unit, current_stock`,
					sku, qty).Scan(&id, &name, &unit, &next)
				if err == nil {
					b.pool.Exec(ctx,
						`INSERT INTO stock_transactions (type, item_id, item_sku, item_name, quantity, unit, previous_stock, new_stock)
						 VALUES ('OUT', $1, $2, $3, $4, $5, $6, $7)`,
						id, sku, name, qty, unit, next+int32(qty), next)

					webURL := b.getWebAppURL(ctx)
					text := fmt.Sprintf("📤 *Pengeluaran Berhasil Dicatat!*\n\n📦 *%s* (`%s`)\n• Dikurang: -%d %s\n• Sisa Stok: *%d %s*",
						name, sku, qty, unit, next, unit)
					kb := models.InlineKeyboardMarkup{
						InlineKeyboard: [][]models.InlineKeyboardButton{
							{
								{Text: "➕ +1", CallbackData: fmt.Sprintf("cmd:add:%s:1", sku)},
								{Text: "➖ -1 Lagi", CallbackData: fmt.Sprintf("cmd:sub:%s:1", sku)},
							},
							{
								{
									Text:   "🌐 Buka Detail di Web",
									WebApp: &models.WebAppInfo{URL: fmt.Sprintf("%s/items/%s", webURL, id)},
								},
							},
							{
								{Text: "‹ Kembali ke Menu", CallbackData: "cmd:start"},
							},
						},
					}
					b.editMessage(ctx, bt, chatID, msgID, text, kb)
				}
			}
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

// dailyAlert runs daily at 07:00 WIB
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

	webURL := b.getWebAppURL(ctx)
	kb := models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{
					Text:   "🌐 Buka Inventaris Web",
					WebApp: &models.WebAppInfo{URL: webURL},
				},
			},
			{
				{Text: "📦 Cek Detail Stok", CallbackData: "cmd:stok"},
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
