package handler

import (
	"encoding/csv"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExportHandler struct {
	pool *pgxpool.Pool
}

func NewExportHandler(pool *pgxpool.Pool) *ExportHandler {
	return &ExportHandler{pool: pool}
}

var masterHeaders = []string{"SKU", "Nama", "Kategori", "Lokasi", "Stok", "Satuan", "Kondisi", "Merk"}

func (h *ExportHandler) ItemsCSV(c *gin.Context) {
	w := c.Writer
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="master_barang.csv"`)
	w.Write([]byte("\xEF\xBB\xBF"))
	cw := csv.NewWriter(w)
	cw.Write(masterHeaders)
	rows, _ := h.pool.Query(c, `SELECT sku, name, category, location, current_stock, unit, condition_status, merk FROM inventory_items ORDER BY name`)
	defer rows.Close()
	for rows.Next() {
		var sku, name, cat, loc, unit, cond, merk string
		var stock int32
		rows.Scan(&sku, &name, &cat, &loc, &stock, &unit, &cond, &merk)
		cw.Write([]string{sku, name, cat, loc, fmt.Sprintf("%d", stock), unit, cond, merk})
	}
	cw.Flush()
}

func (h *ExportHandler) TxCSV(c *gin.Context) {
	w := c.Writer
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="log_transaksi.csv"`)
	w.Write([]byte("\xEF\xBB\xBF"))
	cw := csv.NewWriter(w)
	cw.Write([]string{"Waktu", "Jenis", "SKU", "Nama", "Qty", "Unit", "Petugas"})
	rows, _ := h.pool.Query(c, `SELECT timestamp, type, item_sku, item_name, quantity, unit, received_by FROM stock_transactions ORDER BY timestamp DESC`)
	defer rows.Close()
	for rows.Next() {
		var ts interface{}
		var tType, sku, name, unit, by string
		var qty int32
		rows.Scan(&ts, &tType, &sku, &name, &qty, &unit, &by)
		cw.Write([]string{fmt.Sprintf("%v", ts), tType, sku, name, fmt.Sprintf("%d", qty), unit, by})
	}
	cw.Flush()
}
