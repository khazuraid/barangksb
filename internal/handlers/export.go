package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"inventariskantor/internal/models"
)

var masterHeaders = []string{
	"Ada (Keberadaan)", "No Seri", "Nama Barang", "Kode Barcode / SKU", "Merk",
	"Type / Model", "Thn Pengadaan", "Kondisi (Berfungsi)", "Kategori",
	"Lokasi Simpan / Ruangan", "Jumlah Stok", "Satuan", "Nilai Satuan (Rp)",
	"Sumber Pendanaan", "Distributor / Vendor", "AKL / AKD", "Foto Barang",
	"Geotagging GPS", "Keterangan", "Terakhir Diperbarui",
}

var logInHeaders = []string{
	"ID Transaksi", "Tanggal & Waktu (WIB)", "Kode Barcode / SKU", "Nama Barang",
	"Jumlah Masuk (+)", "Satuan", "Merk", "Type / Model", "No Seri", "Thn Pengadaan",
	"Kondisi", "Sumber Pendanaan", "Distributor / Vendor", "Petugas Penerima",
	"No PO / Surat Jalan", "Foto Bukti Fisik", "Koordinat Geotag GPS", "Catatan / Keterangan",
}

func (h *Handlers) ExportItemsCSV(w http.ResponseWriter, r *http.Request) {
	items := h.allItems(r)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="master_barang.csv"`)
	w.Write([]byte("\xEF\xBB\xBF")) // BOM for Excel
	cw := csv.NewWriter(w)
	cw.Write(masterHeaders)
	for _, it := range items {
		cw.Write([]string{
			boolAda(it.IsAvailable), optStr(it.SerialNumber), it.Name, it.Sku,
			dash(it.Merk), dash(it.TypeModel), dash(it.ProcurementYear),
			optStrOr(it.ConditionStatus, "Berfungsi"), it.Category, it.Location,
			fmt.Sprintf("%d", it.CurrentStock), it.Unit,
			fmt.Sprintf("%d", it.PricePerUnit.Int64),
			dash(it.FundingSource), dash(it.Distributor), dash(it.AklAkd),
			photoCell(it.PhotoUrl), geoCell(it.GeoLat, it.GeoLng),
			dash(it.Description), fmtWIB(it.UpdatedAt),
		})
	}
	cw.Flush()
}

func (h *Handlers) ExportTXCSV(w http.ResponseWriter, r *http.Request) {
	txs, err := h.allTX(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="log_barang_masuk.csv"`)
	w.Write([]byte("\xEF\xBB\xBF"))
	cw := csv.NewWriter(w)
	cw.Write(logInHeaders)
	for _, t := range txs {
		cw.Write([]string{
			fmtUUID(t.ID), fmtWIB(t.Timestamp), t.ItemSku, t.ItemName,
			fmt.Sprintf("%d", t.Quantity), t.Unit,
			dash(t.Merk), dash(t.TypeModel), dash(t.SerialNumber), dash(t.ProcurementYear),
			optStrOr(t.ConditionStatus, "Berfungsi"), dash(t.FundingSource),
			dash2(t.Distributor, t.SupplierOrSource), dash(t.ReceivedBy), dash(t.InvoiceOrPoNum),
			photoCell(t.PhotoUrl), geoCell(t.GeoLat, t.GeoLng), dash(t.Notes),
		})
	}
	cw.Flush()
}

func (h *Handlers) allItems(r *http.Request) []models.InventoryItem {
	rows, err := h.pool.Query(r.Context(), `SELECT * FROM inventory_items ORDER BY name`)
	if err != nil {
		return nil
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.InventoryItem])
	if err != nil {
		return nil
	}
	return items
}

func (h *Handlers) allTX(r *http.Request) ([]models.StockTransaction, error) {
	rows, err := h.pool.Query(r.Context(), `SELECT * FROM stock_transactions ORDER BY timestamp DESC`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[models.StockTransaction])
}

type pgtypeText = pgtype.Text
type pgtypeFloat8 = pgtype.Float8

func boolAda(b bool) string {
	if b {
		return "ADA"
	}
	return "TIDAK"
}

func dash(t pgtypeText) string {
	if t.Valid && t.String != "" {
		return t.String
	}
	return "-"
}

func dash2(a, b pgtypeText) string {
	if s := dash(a); s != "-" {
		return s
	}
	return dash(b)
}

func photoCell(t pgtypeText) string {
	if t.Valid && t.String != "" {
		return t.String
	}
	return "-"
}

func geoCell(lat, lng pgtypeFloat8) string {
	if !lat.Valid || !lng.Valid {
		return "-"
	}
	return fmt.Sprintf("%.6f, %.6f", lat.Float64, lng.Float64)
}

var _ = strings.TrimSpace
