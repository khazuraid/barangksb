package pdf

import (
	"bytes"
	"fmt"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/jackc/pgx/v5/pgtype"

	"inventariskantor/internal/models"
)

// BuildReport generates the full PDF report bytes (dipakai handler & cron).
func BuildReport(items []models.InventoryItem, txs []models.StockTransaction) ([]byte, error) {
	p := fpdf.New("L", "mm", "A4", "")
	p.SetAutoPageBreak(true, 10)

	// --- Master Barang ---
	p.AddPage()
	p.SetFont("Arial", "B", 14)
	p.Cell(0, 10, "Laporan Inventaris Kantor — Master Barang")
	p.Ln(12)
	p.SetFont("Arial", "", 9)
	cols := []float64{18, 30, 52, 28, 24, 20, 16, 22, 14, 20}
	headers := []string{"SKU", "Nama", "Kategori", "Lokasi", "Merk", "Kondisi", "Stok", "Satuan", "Harga", "Pendanaan"}
	for i, hd := range headers {
		p.CellFormat(cols[i], 7, hd, "1", 0, "", true, 0, "")
	}
	p.Ln(-1)
	p.SetFont("Arial", "", 8)
	for _, it := range items {
		row := []string{
			it.Sku, trunc(it.Name, 40), trunc(it.Category, 22), trunc(it.Location, 20),
			trunc(txt(it.Merk), 14), strDef(it.ConditionStatus, "Berfungsi"),
			fmt.Sprintf("%d", it.CurrentStock), it.Unit,
			fmt.Sprintf("%d", it.PricePerUnit.Int64), trunc(txt(it.FundingSource), 14),
		}
		for i, cell := range row {
			p.CellFormat(cols[i], 6, cell, "1", 0, "", false, 0, "")
		}
		p.Ln(-1)
	}

	// --- Log Mutasi ---
	p.AddPage()
	p.SetFont("Arial", "B", 14)
	p.Cell(0, 10, "Log Mutasi Barang (Masuk/Keluar/Opname)")
	p.Ln(12)
	p.SetFont("Arial", "", 9)
	cols2 := []float64{26, 20, 30, 55, 16, 14, 30, 50}
	headers2 := []string{"Waktu", "Jenis", "SKU", "Nama", "Qty", "Satuan", "Person", "Kondisi"}
	for i, hd := range headers2 {
		p.CellFormat(cols2[i], 7, hd, "1", 0, "", true, 0, "")
	}
	p.Ln(-1)
	p.SetFont("Arial", "", 8)
	for _, t := range txs {
		row := []string{
			timeStr(t.Timestamp), t.Type, t.ItemSku, trunc(t.ItemName, 40),
			fmt.Sprintf("%d", t.Quantity), t.Unit,
			trunc(txt(t.ReceivedBy), 22), strDef(t.ConditionStatus, "-"),
		}
		for i, cell := range row {
			p.CellFormat(cols2[i], 6, cell, "1", 0, "", false, 0, "")
		}
		p.Ln(-1)
	}

	var buf bytes.Buffer
	if err := p.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func txt(t pgtype.Text) string {
	if t.Valid && t.String != "" {
		return t.String
	}
	return "-"
}

func strDef(t pgtype.Text, def string) string {
	if t.Valid && t.String != "" {
		return t.String
	}
	return def
}

func timeStr(t pgtype.Timestamptz) string {
	if !t.Valid {
		return "-"
	}
	return t.Time.In(time.FixedZone("WIB", 7*3600)).Format("02-01-2006 15:04")
}
