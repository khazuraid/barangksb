package telegram

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/go-telegram/bot"

	"inventariskantor/internal/handlers"
)

var wibLoc = time.FixedZone("WIB", 7*3600)

// Notifier sends daily low-stock alerts.
type Notifier struct {
	Bot     *bot.Bot
	Core    *handlers.Core
	ChatIDs []int64
}

// StartDailyAlert runs the low-stock check every day at 07:00 WIB.
func (n *Notifier) StartDailyAlert(ctx context.Context) {
	if n == nil || n.Bot == nil {
		return
	}
	go func() {
		for {
			now := time.Now().In(wibLoc)
			next := time.Date(now.Year(), now.Month(), now.Day(), 7, 0, 0, 0, wibLoc)
			if next.Before(now) {
				next = next.Add(24 * time.Hour)
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(next.Sub(now)):
			}
			n.SendLowStock(ctx)
		}
	}()
}

// SendLowStock checks and alerts once.
func (n *Notifier) SendLowStock(ctx context.Context) {
	rows := n.Core.LowStock(ctx)
	if len(rows) == 0 {
		return
	}
	var sb strings.Builder
	sb.WriteString("⚠️ *Stok Menipis*\n\n")
	for _, r := range rows {
		sb.WriteString("• " + r.Name + " (`" + r.SKU + "`): " + itoa(int(r.Current)) + " " + r.Unit + "\n")
	}
	for _, id := range n.ChatIDs {
		c, cancel := context.WithTimeout(ctx, 10*time.Second)
		n.Bot.SendMessage(c, &bot.SendMessageParams{
			ChatID: id, Text: sb.String(), ParseMode: models_ParseModeMarkdown,
		})
		cancel()
	}
	slog.Info("telegram: low stock alert sent", "items", len(rows))
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
