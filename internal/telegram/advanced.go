package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"inventariskantor/internal/handlers"
)

var _ = json.Marshal
var _ = os.Getenv
// Advanced menambahkan fitur bot lanjutan:
// 1. /stok — rekap barang menipis on-demand + inline refresh
// 2. /foto SKU — reply pesan foto dengan /foto SKU → foto disimpan & dikaitkan
// 3. /bantu — menu bantuan
type Advanced struct {
	Core    *handlers.Core
	Token   string
	ChatIDs map[int64]bool
}

func NewAdvanced(core *handlers.Core, token, chatIDsCSV string) *Advanced {
	ids := map[int64]bool{}
	for _, s := range strings.Split(chatIDsCSV, ",") {
		if v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64); err == nil {
			ids[v] = true
		}
	}
	return &Advanced{Core: core, Token: token, ChatIDs: ids}
}

func (a *Advanced) Register(b *bot.Bot) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "/stok", bot.MatchTypePrefix, a.stok)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/foto", bot.MatchTypePrefix, a.foto)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/bantu", bot.MatchTypeExact, a.bantu)
}

func (a *Advanced) allowed(chatID int64) bool { return a.ChatIDs[chatID] }

func (a *Advanced) bantu(ctx context.Context, b *bot.Bot, u *models.Update) {
	if u.Message == nil || !a.allowed(u.Message.Chat.ID) {
		return
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: u.Message.Chat.ID,
		Text: "�� Perintah tersedia:\n" +
			"/masuk SKU jumlah [ket] — catat barang masuk\n" +
			"/cari kata-kunci — cari barang\n" +
			"/stok — rekap stok menipis\n" +
			"/foto SKU (reply foto) — lampirkan foto bukti",
	})
}

func (a *Advanced) stok(ctx context.Context, b *bot.Bot, u *models.Update) {
	if u.Message == nil || !a.allowed(u.Message.Chat.ID) {
		return
	}
	chat := u.Message.Chat.ID
	rows := a.Core.LowStock(ctx)
	if len(rows) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chat, Text: "✅ Semua stok aman."})
		return
	}
	var sb strings.Builder
	sb.WriteString("⚠️ <b>Stok Menipis</b>\n\n")
	for _, r := range rows {
		fmt.Fprintf(&sb, "• %s (<code>%s</code>): <b>%d %s</b> — %s\n", r.Name, r.SKU, r.Current, r.Unit, r.Location)
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chat,
		Text:      sb.String(),
		ParseMode: models.ParseModeHTML,
		ReplyMarkup: models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{{
				{Text: "🔄 Refresh", CallbackData: "refresh_stok"},
			}},
		},
	})
}

// foto: reply pesan foto dengan /foto SKU → foto dikaitkan ke transaksi.
func (a *Advanced) foto(ctx context.Context, b *bot.Bot, u *models.Update) {
	if u.Message == nil || !a.allowed(u.Message.Chat.ID) {
		return
	}
	chat := u.Message.Chat.ID
	if u.Message.ReplyToMessage == nil || len(u.Message.ReplyToMessage.Photo) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chat,
			Text: "Format: <b>reply</b> pesan foto dengan /foto SKU", ParseMode: models.ParseModeHTML})
		return
	}
	parts := strings.Fields(u.Message.Text)
	if len(parts) < 2 {
		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chat, Text: "SKU tidak ada: /foto SKU"})
		return
	}
	sku := parts[1]
	photos := u.Message.ReplyToMessage.Photo
	fileID := photos[len(photos)-1].FileID

	f, err := b.GetFile(ctx, &bot.GetFileParams{FileID: fileID})
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chat, Text: "❌ ambil foto gagal"})
		return
	}
	data, err := a.download(f.FilePath)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chat, Text: "❌ unduh foto gagal"})
		return
	}
	photoURL := a.storeLocal(data, f.FilePath)
	a.Core.AttachPhoto(ctx, sku, photoURL)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chat,
		Text:   "✅ foto dikaitkan ke " + sku + "\n" + photoURL,
	})
}

func (a *Advanced) download(filePath string) ([]byte, error) {
	resp, err := http.Get("https://api.telegram.org/file/bot" + a.Token + "/" + filePath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// storeLocal simpan ke web/uploads (fallback lokal, tanpa Drive agar bot ringan).
func (a *Advanced) storeLocal(data []byte, filePath string) string {
	name := fmt.Sprintf("tg-%x%s", time.Now().UnixNano(), strings.ToLower(filepath.Ext(filePath)))
	os.MkdirAll("web/uploads", 0o755)
	path := "web/uploads/" + name
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return ""
	}
	return "/uploads/" + name
}

var (
	_ = json.Marshal
	_ = os.Getenv
)
