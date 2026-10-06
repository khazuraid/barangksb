package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"inventariskantor/internal/service"
	"inventariskantor/internal/views"
)

func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var total, totalStock int
	_ = h.pool.QueryRow(ctx, `SELECT count(*), COALESCE(sum(current_stock),0) FROM inventory_items`).Scan(&total, &totalStock)

	lowRows, _ := h.pool.Query(ctx, `SELECT name, sku, current_stock, min_stock, unit, location
		FROM inventory_items WHERE current_stock <= min_stock ORDER BY current_stock LIMIT 10`)
	type lowRow struct {
		Name    string
		SKU     string
		Current int32
		Min     int32
		Unit    string
		Loc     string
	}
	low, _ := pgx.CollectRows(lowRows, pgx.RowToStructByPos[lowRow])

	txRows, _ := h.pool.Query(ctx, `SELECT timestamp, item_name, item_sku, quantity, unit, type
		FROM stock_transactions ORDER BY timestamp DESC LIMIT 8`)
	type txRow struct {
		Time     time.Time
		ItemName string
		SKU      string
		Quantity int32
		Unit     string
		Type     string
	}
	recent, _ := pgx.CollectRows(txRows, pgx.RowToStructByPos[txRow])

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

	cats, _ := h.svc.CategoryNames(ctx)

	lowViews := make([]views.LowStockRow, 0, len(low))
	for _, l := range low {
		lowViews = append(lowViews, views.LowStockRow{
			Name: l.Name, SKU: l.SKU, Current: int(l.Current), Min: int(l.Min),
			Unit: l.Unit, Location: l.Loc,
		})
	}
	txViews := make([]views.RecentTXRow, 0, len(recent))
	for _, t := range recent {
		txViews = append(txViews, views.RecentTXRow{
			Time: fmtWIBTime(t.Time), ItemName: t.ItemName, SKU: t.SKU,
			Quantity: int(t.Quantity), Unit: t.Unit, Type: t.Type,
		})
	}

	u := authUserOf(r)
	h.show(w, r, "Dashboard", views.Dashboard(views.DashboardData{
		User:       userInfo(r),
		TotalItems: total,
		TotalStock: totalStock,
		LowStock:   lowViews,
		RecentTX:   txViews,
		StockByCat: catMap,
		Categories: cats,
	}))
	_ = u
}

var (
	_ = service.ItemFilter{}
	_ = pgtype.Text{}
	_ = fmt.Sprintf
)
