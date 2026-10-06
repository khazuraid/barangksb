package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"inventariskantor/internal/handlers"
)

// Handler routes bot commands: /masuk SKU qty, /cari keyword.
type Handler struct {
	Core    *handlers.Core
	ChatIDs map[int64]bool
}

func NewHandler(core *handlers.Core, chatIDsCSV string) *Handler {
	ids := map[int64]bool{}
	for _, s := range strings.Split(chatIDsCSV, ",") {
		if v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64); err == nil {
			ids[v] = true
		}
	}
	return &Handler{Core: core, ChatIDs: ids}
}

func (h *Handler) Register(b *bot.Bot) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "/masuk", bot.MatchTypePrefix, h.masuk)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/cari", bot.MatchTypePrefix, h.cari)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, h.start)
}

func (h *Handler) allowed(chatID int64) bool { return h.ChatIDs[chatID] }

// ChatIDList returns allowlisted IDs as a slice (for the notifier).
func (h *Handler) ChatIDList() []int64 {
	out := make([]int64, 0, len(h.ChatIDs))
	for id := range h.ChatIDs {
		out = append(out, id)
	}
	return out
}

func (h *Handler) start(ctx context.Context, b *bot.Bot, u *models.Update) {
	if u.Message == nil || !h.allowed(u.Message.Chat.ID) {
		return
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: u.Message.Chat.ID,
		Text:   "🤖 Bot Inventaris Kantor\n/masuk SKU jumlah [keterangan] — catat barang masuk\n/cari kata-kunci — cari barang",
	})
}

// masuk: /masuk SKU qty [keterangan]
func (h *Handler) masuk(ctx context.Context, b *bot.Bot, u *models.Update) {
	if u.Message == nil || !h.allowed(u.Message.Chat.ID) {
		slog.Warn("telegram: unauthorized chat")
		return
	}
	chat := u.Message.Chat.ID
	parts := strings.Fields(u.Message.Text)
	if len(parts) < 3 {
		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chat, Text: "Format: /masuk SKU jumlah [keterangan]"})
		return
	}
	sku := parts[1]
	qty, err := strconv.Atoi(parts[2])
	if err != nil || qty <= 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chat, Text: "Jumlah harus angka > 0"})
		return
	}
	notes := strings.Join(parts[3:], " ")
	receivedBy := u.Message.From.FirstName

	res, err := h.Core.RecordStockIn(ctx, sku, int32(qty), receivedBy, notes)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chat, Text: "❌ " + err.Error()})
		return
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chat,
		Text:   fmt.Sprintf("✅ %s (+%d %s)\nStok: %d → %d", res.Name, qty, res.Unit, res.Previous, res.New),
	})
}

// cari: /cari keyword
func (h *Handler) cari(ctx context.Context, b *bot.Bot, u *models.Update) {
	if u.Message == nil || !h.allowed(u.Message.Chat.ID) {
		return
	}
	chat := u.Message.Chat.ID
	kw := strings.TrimSpace(strings.TrimPrefix(u.Message.Text, "/cari"))
	if kw == "" {
		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chat, Text: "Format: /cari kata-kunci"})
		return
	}
	items := h.Core.SearchItems(ctx, kw, 10)
	if len(items) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chat, Text: "Tidak ditemukan: " + kw})
		return
	}
	var sb strings.Builder
	sb.WriteString("🔍 Hasil pencarian:\n\n")
	for _, it := range items {
		fmt.Fprintf(&sb, "• %s (`%s`)\n  Stok: %d %s — %s\n", it.Name, it.SKU, it.Current, it.Unit, it.Location)
	}
	b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chat, Text: sb.String(), ParseMode: models.ParseModeMarkdown})
}
