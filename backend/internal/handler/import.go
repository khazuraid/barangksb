package handler

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xuri/excelize/v2"
)

type ImportHandler struct {
	pool    *pgxpool.Pool
	itemH   *ItemHandler
}

func NewImportHandler(pool *pgxpool.Pool, itemH *ItemHandler) *ImportHandler {
	return &ImportHandler{pool: pool, itemH: itemH}
}

var importHeaders = []string{
	"Nama Barang", "Kategori", "Lokasi", "Stok", "Satuan",
	"Batas Minimum", "Harga Satuan", "Kondisi", "Merk", "Model Tipe",
	"Nomor Seri", "Tahun Pengadaan", "Sumber Dana", "Distributor",
	"AKL/AKD", "Keterangan",
}

// GET /api/items/template
func (h *ImportHandler) DownloadTemplate(c *gin.Context) {
	fmtParam := c.DefaultQuery("fmt", "xlsx")

	if fmtParam == "csv" {
		w := c.Writer
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="template_import_inventaris.csv"`)
		w.Write([]byte("\xEF\xBB\xBF")) // UTF-8 BOM for Excel
		cw := csv.NewWriter(w)
		cw.Write(importHeaders)
		cw.Write([]string{
			"Tensimeter Digital", "Peralatan Medis & Alkes (AKL/AKD)", "Ruang Rawat Inap",
			"5", "unit", "2", "850000", "Berfungsi", "Omron", "HEM-7120",
			"SN-123456", "2024", "APBD", "PT Medika Nusantara", "AKL 20501812345", "Kondisi prima",
		})
		cw.Write([]string{
			"Laptop Admin Kantor", "Elektronik & IT Perkantoran", "Ruang Tata Usaha",
			"3", "unit", "1", "12500000", "Berfungsi", "Lenovo", "ThinkPad E14",
			"LR-987654", "2025", "BOS / Operasional", "PT Multi Sarana", "", "Lengkap dengan adaptor",
		})
		cw.Flush()
		return
	}

	// Excel XLSX Template
	f := excelize.NewFile()
	sheet := "Template_Barang"
	f.SetSheetName("Sheet1", sheet)

	for i, hd := range importHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, hd)
	}

	// Sample Row 1
	sample1 := []any{
		"Tensimeter Digital", "Peralatan Medis & Alkes (AKL/AKD)", "Ruang Rawat Inap",
		5, "unit", 2, 850000, "Berfungsi", "Omron", "HEM-7120",
		"SN-123456", "2024", "APBD", "PT Medika Nusantara", "AKL 20501812345", "Kondisi prima",
	}
	for i, val := range sample1 {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		f.SetCellValue(sheet, cell, val)
	}

	// Sample Row 2
	sample2 := []any{
		"Laptop Admin Kantor", "Elektronik & IT Perkantoran", "Ruang Tata Usaha",
		3, "unit", 1, 12500000, "Berfungsi", "Lenovo", "ThinkPad E14",
		"LR-987654", "2025", "BOS / Operasional", "PT Multi Sarana", "", "Lengkap dengan adaptor",
	}
	for i, val := range sample2 {
		cell, _ := excelize.CoordinatesToCellName(i+1, 3)
		f.SetCellValue(sheet, cell, val)
	}

	// Style header
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E293B"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	endCell, _ := excelize.CoordinatesToCellName(len(importHeaders), 1)
	f.SetCellStyle(sheet, "A1", endCell, headerStyle)
	f.SetRowHeight(sheet, 1, 25)

	// Set column widths
	for i := 1; i <= len(importHeaders); i++ {
		colName, _ := excelize.ColumnNumberToName(i)
		f.SetColWidth(sheet, colName, colName, 22)
	}

	var buf bytes.Buffer
	f.Write(&buf)
	c.Header("Content-Disposition", `attachment; filename="template_import_inventaris.xlsx"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

// POST /api/items/import
func (h *ImportHandler) Import(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20) // Maks 10 MB

	fileHdr, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file Excel / CSV wajib diunggah (maksimal 10 MB)"})
		return
	}

	file, err := fileHdr.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "gagal membuka berkas"})
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(fileHdr.Filename))
	var rows [][]string

	if ext == ".csv" {
		rd := csv.NewReader(file)
		rd.TrimLeadingSpace = true
		records, err := rd.ReadAll()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "berkas CSV tidak terbaca: " + err.Error()})
			return
		}
		rows = records
	} else {
		// Read XLSX
		buf, err := io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "gagal membaca file: " + err.Error()})
			return
		}
		xlFile, err := excelize.OpenReader(bytes.NewReader(buf))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "format file Excel tidak valid: " + err.Error()})
			return
		}
		defer xlFile.Close()

		sheets := xlFile.GetSheetList()
		if len(sheets) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file Excel tidak memiliki sheet"})
			return
		}
		records, err := xlFile.GetRows(sheets[0])
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "gagal membaca baris sheet: " + err.Error()})
			return
		}
		rows = records
	}

	if len(rows) <= 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file tidak berisi baris data"})
		return
	}

	successCount := 0
	failedCount := 0
	var errorList []string

	for idx, row := range rows {
		// Skip header row
		if idx == 0 {
			continue
		}
		rowNum := idx + 1
		if len(row) == 0 {
			continue
		}

		name := strings.TrimSpace(getCol(row, 0))
		category := strings.TrimSpace(getCol(row, 1))
		location := strings.TrimSpace(getCol(row, 2))

		if name == "" {
			failedCount++
			errorList = append(errorList, fmt.Sprintf("Baris %d: Nama barang wajib diisi", rowNum))
			continue
		}
		if category == "" {
			category = "Lainnya"
		}
		if location == "" {
			location = "Gudang Utama"
		}

		// Ensure category exists
		h.ensureCategory(c, category)
		// Ensure location exists
		h.ensureLocation(c, location)

		stock := parseInt32(getCol(row, 3), 0)
		unit := strings.TrimSpace(getCol(row, 4))
		if unit == "" {
			unit = "buah"
		}
		minStock := parseInt32(getCol(row, 5), 0)
		price := parseInt64(getCol(row, 6), 0)
		condition := strings.TrimSpace(getCol(row, 7))
		if condition == "" {
			condition = "Berfungsi"
		}
		merk := strings.TrimSpace(getCol(row, 8))
		typeModel := strings.TrimSpace(getCol(row, 9))
		serialNumber := strings.TrimSpace(getCol(row, 10))
		procurementYear := strings.TrimSpace(getCol(row, 11))
		fundingSource := strings.TrimSpace(getCol(row, 12))
		distributor := strings.TrimSpace(getCol(row, 13))
		aklAkd := strings.TrimSpace(getCol(row, 14))
		description := strings.TrimSpace(getCol(row, 15))

		// Auto-generate SKU
		sku := generateSKU(h.itemH, c, category)

		_, err := h.pool.Exec(c, `
			INSERT INTO inventory_items (
				sku, name, category, location, current_stock, min_stock, unit,
				price_per_unit, condition_status, merk, type_model, serial_number,
				procurement_year, funding_source, distributor, akl_akd, description
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		`, sku, name, category, location, stock, minStock, unit,
			price, condition, nullable(merk), nullable(typeModel), nullable(serialNumber),
			nullable(procurementYear), nullable(fundingSource), nullable(distributor),
			nullable(aklAkd), nullable(description))

		if err != nil {
			failedCount++
			errorList = append(errorList, fmt.Sprintf("Baris %d (%s): %s", rowNum, name, err.Error()))
		} else {
			successCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": successCount > 0,
		"count":   successCount,
		"failed":  failedCount,
		"total":   successCount + failedCount,
		"errors":  errorList,
	})
}

func (h *ImportHandler) ensureCategory(c *gin.Context, catName string) {
	var count int
	h.pool.QueryRow(c, `SELECT count(*) FROM categories WHERE name=$1`, catName).Scan(&count)
	if count == 0 {
		id := strings.ToLower(strings.ReplaceAll(catName, " ", "-"))
		h.pool.Exec(c, `INSERT INTO categories (id, name) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING`, id, catName)
	}
}

func (h *ImportHandler) ensureLocation(c *gin.Context, locName string) {
	var count int
	h.pool.QueryRow(c, `SELECT count(*) FROM locations WHERE name=$1`, locName).Scan(&count)
	if count == 0 {
		id := strings.ToLower(strings.ReplaceAll(locName, " ", "-"))
		h.pool.Exec(c, `INSERT INTO locations (id, name) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING`, id, locName)
	}
}

func getCol(row []string, idx int) string {
	if idx < len(row) {
		return row[idx]
	}
	return ""
}

func parseInt32(s string, def int32) int32 {
	clean := strings.ReplaceAll(strings.TrimSpace(s), ",", "")
	clean = strings.ReplaceAll(clean, ".", "")
	if v, err := strconv.ParseInt(clean, 10, 32); err == nil {
		return int32(v)
	}
	return def
}

func parseInt64(s string, def int64) int64 {
	clean := strings.ReplaceAll(strings.TrimSpace(s), ",", "")
	clean = strings.ReplaceAll(clean, ".", "")
	if v, err := strconv.ParseInt(clean, 10, 64); err == nil {
		return v
	}
	return def
}
