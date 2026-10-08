package handler

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xuri/excelize/v2"
)

type ReportHandler struct {
	pool *pgxpool.Pool
}

func NewReportHandler(pool *pgxpool.Pool) *ReportHandler {
	return &ReportHandler{pool: pool}
}

// GET /api/export/items.xlsx
func (h *ReportHandler) ItemsXLSX(c *gin.Context) {
	f := excelize.NewFile()
	sheet := "Master_Barang"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"SKU", "Nama", "Kategori", "Lokasi", "Stok", "Satuan", "Kondisi"}
	for i, hd := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, hd)
	}
	rows, _ := h.pool.Query(c, `SELECT sku, name, category, location, current_stock, unit, condition_status FROM inventory_items ORDER BY name`)
	defer rows.Close()
	r := 2
	for rows.Next() {
		var sku, name, cat, loc, unit, cond string
		var stock int32
		rows.Scan(&sku, &name, &cat, &loc, &stock, &unit, &cond)
		vals := []any{sku, name, cat, loc, stock, unit, cond}
		for i, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(i+1, r)
			f.SetCellValue(sheet, cell, v)
		}
		r++
	}
	style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	f.SetCellStyle(sheet, "A1", "G1", style)
	var buf bytes.Buffer
	f.Write(&buf)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

// GET /api/report.pdf
func (h *ReportHandler) PDF(c *gin.Context) {
	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Laporan Inventaris Kantor — Master Barang")
	pdf.Ln(12)
	pdf.SetFont("Arial", "", 9)
	cols := []float64{20, 40, 50, 30, 20, 20}
	hds := []string{"SKU", "Nama", "Kategori", "Lokasi", "Stok", "Satuan"}
	for i, hd := range hds {
		pdf.CellFormat(cols[i], 7, hd, "1", 0, "", true, 0, "")
	}
	pdf.Ln(-1)
	pdf.SetFont("Arial", "", 8)
	rows, _ := h.pool.Query(c, `SELECT sku, name, category, location, current_stock, unit FROM inventory_items ORDER BY name`)
	defer rows.Close()
	for rows.Next() {
		var sku, name, cat, loc, unit string
		var stock int32
		rows.Scan(&sku, &name, &cat, &loc, &stock, &unit)
		vals := []string{sku, truncStr(name, 30), truncStr(cat, 25), truncStr(loc, 20), strconv.Itoa(int(stock)), unit}
		for i, v := range vals {
			pdf.CellFormat(cols[i], 6, v, "1", 0, "", false, 0, "")
		}
		pdf.Ln(-1)
	}
	var buf bytes.Buffer
	pdf.Output(&buf)
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}

func truncStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
