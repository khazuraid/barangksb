package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"inventariskantor/internal/models"
)

// ReportPDF streams a landscape A4 PDF: stock master + stock-in log.
func (h *Handlers) ReportPDF(w http.ResponseWriter, r *http.Request) {
	items := h.allItems(r)
	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 10)
	pdf.SetFont("Arial", "", 9)

	// --- Page 1: Master Barang ---
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Laporan Inventaris Kantor — Master Barang")
	pdf.Ln(12)
	pdf.SetFont("Arial", "", 9)
	cols := []float64{18, 30, 52, 28, 24, 20, 16, 22, 14, 20}
	headers := []string{"SKU", "Nama", "Kategori", "Lokasi", "Merk", "Kondisi", "Stok", "Satuan", "Harga", "Pendanaan"}
	for i, hd := range headers {
		pdf.CellFormat(cols[i], 7, hd, "1", 0, "", true, 0, "")
	}
	pdf.Ln(-1)
	pdf.SetFont("Arial", "", 8)
	for _, it := range items {
		row := []string{
			it.Sku, trunc(it.Name, 40), trunc(it.Category, 22), trunc(it.Location, 20),
			trunc(dash(it.Merk), 14), optStrOr(it.ConditionStatus, "Berfungsi"),
			fmt.Sprintf("%d", it.CurrentStock), it.Unit,
			fmt.Sprintf("%d", it.PricePerUnit.Int64), trunc(dash(it.FundingSource), 14),
		}
		for i, cell := range row {
			pdf.CellFormat(cols[i], 6, cell, "1", 0, "", false, 0, "")
		}
		pdf.Ln(-1)
	}

	// --- Page 2: Log Barang Masuk ---
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Log Barang Masuk")
	pdf.Ln(12)
	pdf.SetFont("Arial", "", 9)
	cols2 := []float64{26, 30, 55, 16, 14, 30, 26, 22, 50}
	headers2 := []string{"Waktu", "SKU", "Nama", "Qty", "Satuan", "Petugas", "Distributor", "No PO", "Kondisi"}
	for i, hd := range headers2 {
		pdf.CellFormat(cols2[i], 7, hd, "1", 0, "", true, 0, "")
	}
	pdf.Ln(-1)
	pdf.SetFont("Arial", "", 8)
	rows, err := h.pool.Query(r.Context(), `SELECT * FROM stock_transactions ORDER BY timestamp DESC LIMIT 300`)
	if err == nil {
		txs, _ := pgx.CollectRows(rows, pgx.RowToStructByPos[models.StockTransaction])
		for _, t := range txs {
			row := []string{
				fmtWIB(t.Timestamp), t.ItemSku, trunc(t.ItemName, 40),
				fmt.Sprintf("+%d", t.Quantity), t.Unit,
				trunc(dash(t.ReceivedBy), 22), trunc(dash2(t.Distributor, t.SupplierOrSource), 22),
				trunc(dash(t.InvoiceOrPoNum), 16), optStrOr(t.ConditionStatus, "Berfungsi"),
			}
			for i, cell := range row {
				pdf.CellFormat(cols2[i], 6, cell, "1", 0, "", false, 0, "")
			}
			pdf.Ln(-1)
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="laporan-%s.pdf"`, time.Now().Format("20060102")))
	w.Write(buf.Bytes())
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

var _ = pgtype.Text{}
