package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"inventariskantor/internal/views"
)

func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var total int
	_ = h.pool.QueryRow(ctx, `SELECT count(*) FROM inventory_items`).Scan(&total)

	lowRows, _ := h.pool.Query(ctx, `SELECT name, sku, current_stock, min_stock, unit
		FROM inventory_items WHERE current_stock <= min_stock ORDER BY current_stock`)
	low, _ := pgx.CollectRows(lowRows, pgx.RowToStructByPos[lowStockRow])

	txRows, _ := h.pool.Query(ctx, `SELECT timestamp, item_name, item_sku, quantity, unit
		FROM stock_transactions ORDER BY timestamp DESC LIMIT 10`)
	recent, _ := pgx.CollectRows(txRows, pgx.RowToStructByPos[recentTXRow])

	catRows, _ := h.pool.Query(ctx, `SELECT category, COALESCE(sum(current_stock),0) FROM inventory_items GROUP BY category`)
	catMap := map[string]int{}
	for catRows.Next() {
		var c string
		var n int
		if catRows.Scan(&c, &n) == nil {
			catMap[c] = n
		}
	}
	catRows.Close()

	var cats []string
	catList, _ := h.pool.Query(ctx, `SELECT name FROM categories ORDER BY name`)
	cats, _ = pgx.CollectRows(catList, pgx.RowTo[string])

	lowViews := make([]views.LowStockRow, 0, len(low))
	for _, l := range low {
		lowViews = append(lowViews, views.LowStockRow{Name: l.Name, SKU: l.SKU, Current: int(l.Current), Min: int(l.Min), Unit: l.Unit})
	}
	txViews := make([]views.RecentTXRow, 0, len(recent))
	for _, t := range recent {
		txViews = append(txViews, views.RecentTXRow{
			Time: t.Time.In(wibLoc()).Format("02-01 15:04"), ItemName: t.ItemName,
			SKU: t.SKU, Quantity: int(t.Quantity), Unit: t.Unit,
		})
	}

	h.show(w, r, "Dashboard", views.Dashboard(views.DashboardData{
		User: userInfo(r), TotalItems: total,
		LowStock: lowViews, RecentTX: txViews, StockByCat: catMap, Categories: cats,
	}))
}

type lowStockRow struct {
	Name    string
	SKU     string
	Current int32
	Min     int32
	Unit    string
}

type recentTXRow struct {
	Time     time.Time
	ItemName string
	SKU      string
	Quantity int32
	Unit     string
}

var _ = fmt.Sprintf
