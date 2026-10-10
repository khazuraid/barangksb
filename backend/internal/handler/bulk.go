package handler

import (
	"bytes"
	"context"
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

type BulkHandler struct {
	pool *pgxpool.Pool
	tg   TelegramNotifier
}

func NewBulkHandler(pool *pgxpool.Pool, tg TelegramNotifier) *BulkHandler {
	return &BulkHandler{pool: pool, tg: tg}
}

var opnameHeaders = []string{
	"SKU", "Nama Barang", "Kategori", "Lokasi", "Stok Sistem", "Satuan", "Stok Fisik", "Keterangan",
}

// GET /api/adjust/template?fmt=xlsx|csv
func (h *BulkHandler) DownloadTemplate(c *gin.Context) {
	fmtParam := c.DefaultQuery("fmt", "xlsx")

	if fmtParam == "csv" {
		w := c.Writer
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="template_opname_massal.csv"`)
		w.Write([]byte("\xEF\xBB\xBF"))
		cw := csv.NewWriter(w)
		cw.Write(opnameHeaders)
		cw.Write([]string{
			"ALAT-2026-001", "Tensimeter Digital", "Peralatan Medis & Alkes (AKL/AKD)", "Ruang Rawat Inap", "5", "unit", "5", "Sesuai fisik",
		})
		cw.Write([]string{
			"ELEK-2026-002", "Laptop Admin Kantor", "Elektronik & IT Perkantoran", "Ruang Tata Usaha", "3", "unit", "2", "1 unit perbaikan",
		})
		cw.Flush()
		return
	}

	// Excel XLSX
	f := excelize.NewFile()
	sheet := "Template_Opname"
	f.SetSheetName("Sheet1", sheet)

	for i, hd := range opnameHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, hd)
	}

	sample1 := []any{"ALAT-2026-001", "Tensimeter Digital", "Peralatan Medis & Alkes (AKL/AKD)", "Ruang Rawat Inap", 5, "unit", 5, "Sesuai fisik"}
	for i, val := range sample1 {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		f.SetCellValue(sheet, cell, val)
	}
	sample2 := []any{"ELEK-2026-002", "Laptop Admin Kantor", "Elektronik & IT Perkantoran", "Ruang Tata Usaha", 3, "unit", 2, "1 unit perbaikan"}
	for i, val := range sample2 {
		cell, _ := excelize.CoordinatesToCellName(i+1, 3)
		f.SetCellValue(sheet, cell, val)
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#312E81"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	endCell, _ := excelize.CoordinatesToCellName(len(opnameHeaders), 1)
	f.SetCellStyle(sheet, "A1", endCell, headerStyle)
	f.SetRowHeight(sheet, 1, 25)

	for i := 1; i <= len(opnameHeaders); i++ {
		colName, _ := excelize.ColumnNumberToName(i)
		f.SetColWidth(sheet, colName, colName, 20)
	}

	var buf bytes.Buffer
	f.Write(&buf)
	c.Header("Content-Disposition", `attachment; filename="template_opname_massal.xlsx"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

// GET /api/adjust/export?fmt=xlsx|csv
func (h *BulkHandler) ExportOpname(c *gin.Context) {
	fmtParam := c.DefaultQuery("fmt", "xlsx")

	rows, err := h.pool.Query(c, `SELECT sku, name, category, location, current_stock, unit FROM inventory_items WHERE COALESCE(track_stock, true)=TRUE ORDER BY name`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal membaca database barang"})
		return
	}
	defer rows.Close()

	type itemRow struct {
		sku, name, cat, loc, unit string
		stock                     int32
	}
	var items []itemRow
	for rows.Next() {
		var it itemRow
		if err := rows.Scan(&it.sku, &it.name, &it.cat, &it.loc, &it.stock, &it.unit); err == nil {
			items = append(items, it)
		}
	}

	if fmtParam == "csv" {
		w := c.Writer
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="data_opname_barang.csv"`)
		w.Write([]byte("\xEF\xBB\xBF"))
		cw := csv.NewWriter(w)
		cw.Write(opnameHeaders)
		for _, it := range items {
			cw.Write([]string{
				it.sku, it.name, it.cat, it.loc,
				fmt.Sprintf("%d", it.stock), it.unit,
				fmt.Sprintf("%d", it.stock), "",
			})
		}
		cw.Flush()
		return
	}

	// Excel XLSX
	f := excelize.NewFile()
	sheet := "Data_Opname"
	f.SetSheetName("Sheet1", sheet)

	for i, hd := range opnameHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, hd)
	}

	for rIdx, it := range items {
		rowNum := rIdx + 2
		vals := []any{
			it.sku, it.name, it.cat, it.loc,
			it.stock, it.unit, it.stock, "",
		}
		for cIdx, val := range vals {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, rowNum)
			f.SetCellValue(sheet, cell, val)
		}
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E1B4B"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	endCell, _ := excelize.CoordinatesToCellName(len(opnameHeaders), 1)
	f.SetCellStyle(sheet, "A1", endCell, headerStyle)
	f.SetRowHeight(sheet, 1, 25)

	for i := 1; i <= len(opnameHeaders); i++ {
		colName, _ := excelize.ColumnNumberToName(i)
		f.SetColWidth(sheet, colName, colName, 20)
	}

	var buf bytes.Buffer
	f.Write(&buf)
	c.Header("Content-Disposition", `attachment; filename="data_opname_barang.xlsx"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

// POST /api/adjust/bulk (multipart: file CSV or XLSX)
func (h *BulkHandler) BulkAdjust(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20)

	fileHdr, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file Excel / CSV wajib diunggah (maks 10 MB)"})
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

	if ext == ".csv" || ext == ".txt" {
		buf, err := io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "gagal membaca berkas CSV"})
			return
		}
		content := bytes.TrimPrefix(buf, []byte("\xef\xbb\xbf"))

		firstLine := string(content)
		if idx := strings.IndexAny(firstLine, "\r\n"); idx != -1 {
			firstLine = firstLine[:idx]
		}
		delim := ','
		if strings.Count(firstLine, ";") > strings.Count(firstLine, ",") {
			delim = ';'
		}

		rd := csv.NewReader(bytes.NewReader(content))
		rd.Comma = delim
		rd.TrimLeadingSpace = true
		records, err := rd.ReadAll()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "berkas CSV tidak terbaca: " + err.Error()})
			return
		}
		rows = records
	} else {
		buf, err := io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "gagal membaca berkas: " + err.Error()})
			return
		}
		xlFile, err := excelize.OpenReader(bytes.NewReader(buf))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "format berkas Excel tidak valid: " + err.Error()})
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

	if len(rows) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "berkas tidak berisi baris data"})
		return
	}

	skuCol := -1
	physCol := -1
	notesCol := -1
	startRow := 0

	headerCandidate := rows[0]
	isHeader := false
	for i, col := range headerCandidate {
		norm := cleanColName(col)
		if norm == "sku" || norm == "kode" || norm == "kodebarang" || norm == "kodesku" || norm == "id" {
			skuCol = i
			isHeader = true
		} else if isPhysicalStockHeader(norm) {
			physCol = i
			isHeader = true
		} else if norm == "keterangan" || norm == "catatan" || norm == "alasan" || norm == "notes" || norm == "note" {
			notesCol = i
			isHeader = true
		}
	}

	if isHeader && physCol == -1 {
		for i, col := range headerCandidate {
			norm := cleanColName(col)
			if (norm == "stok" || norm == "stock" || norm == "stokbaru" || norm == "jumlah") && i != skuCol {
				physCol = i
				break
			}
		}
	}

	if isHeader {
		startRow = 1
	}

	if skuCol == -1 || physCol == -1 {
		if len(rows[0]) == 2 {
			skuCol = 0
			physCol = 1
			if strings.EqualFold(strings.TrimSpace(rows[0][0]), "sku") {
				startRow = 1
			}
		} else if len(rows[0]) >= 5 && isMasterBarangShape(rows[0]) {
			skuCol = 0
			physCol = 4
			startRow = 1
		} else {
			skuCol = 0
			physCol = 1
			if isHeader {
				startRow = 1
			}
		}
	}

	ok, unchanged, fail := 0, 0, 0
	errs := []string{}
	person := c.GetString("name")
	if person == "" {
		person = "Petugas"
	}

	for i := startRow; i < len(rows); i++ {
		row := rows[i]
		rowNum := i + 1
		if len(row) == 0 {
			continue
		}

		sku := ""
		if skuCol < len(row) {
			sku = strings.TrimSpace(row[skuCol])
		}
		if sku == "" {
			continue
		}

		physStr := ""
		if physCol < len(row) {
			physStr = strings.TrimSpace(row[physCol])
		}
		if physStr == "" || physStr == "-" {
			continue
		}

		physical, err := parseQty(physStr)
		if err != nil {
			fail++
			errs = append(errs, fmt.Sprintf("Baris %d (%s): angka stok fisik '%s' tidak valid", rowNum, sku, physStr))
			continue
		}

		note := "Opname massal"
		if notesCol != -1 && notesCol < len(row) {
			userNote := strings.TrimSpace(row[notesCol])
			if userNote != "" {
				note = "Opname massal: " + userNote
			}
		}

		diff, err := h.adjustSingleItem(c.Request.Context(), sku, physical, person, note)
		if err != nil {
			fail++
			errs = append(errs, fmt.Sprintf("Baris %d (%s): %s", rowNum, sku, err.Error()))
			continue
		}

		if diff == 0 {
			unchanged++
		} else {
			ok++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":        ok,
		"unchanged": unchanged,
		"fail":      fail,
		"total":     ok + unchanged + fail,
		"errors":    errs,
	})
}

func (h *BulkHandler) adjustSingleItem(ctx context.Context, sku string, actual int32, person, note string) (int32, error) {
	var current int32
	var name, unit string
	var trackStock bool
	err := h.pool.QueryRow(ctx, `SELECT current_stock, name, unit, COALESCE(track_stock, true) FROM inventory_items WHERE sku=$1`, sku).
		Scan(&current, &name, &unit, &trackStock)
	if err != nil {
		return 0, fmt.Errorf("SKU tidak terdaftar di sistem")
	}
	if !trackStock {
		return 0, fmt.Errorf("barang adalah aset tetap (tidak mengelola kuantitas stok)")
	}

	diff := actual - current
	if diff == 0 {
		return 0, nil
	}

	txType := "ADJUST+"
	absDiff := diff
	if diff < 0 {
		txType = "ADJUST-"
		absDiff = -diff
	}

	_, err = h.pool.Exec(ctx, `UPDATE inventory_items SET current_stock=$2, updated_at=now() WHERE sku=$1`, sku, actual)
	if err != nil {
		return 0, fmt.Errorf("gagal update stok: %w", err)
	}

	_, err = h.pool.Exec(ctx, `INSERT INTO stock_transactions (type, item_id, item_sku, item_name, quantity, unit, previous_stock, new_stock, received_by, notes)
		VALUES ($1, (SELECT id FROM inventory_items WHERE sku=$2), $2, $3, $4, $5, $6, $7, $8, $9)`,
		txType, sku, name, absDiff, unit, current, actual, nullableS(person), nullableS(note))
	if err != nil {
		return 0, fmt.Errorf("gagal mencatat transaksi: %w", err)
	}

	if h.tg != nil {
		go h.tg.NotifyMovement(context.Background(), txType, sku, name, absDiff, unit, current, actual, person, note)
	}

	return diff, nil
}

func cleanColName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, "-", "")
	return s
}

func isPhysicalStockHeader(norm string) bool {
	switch norm {
	case "stokfisik", "fisik", "stoknyata", "stokaktual", "stokhitung", "actual", "physical", "actualstock", "physicalstock":
		return true
	default:
		return false
	}
}

func isMasterBarangShape(row []string) bool {
	if len(row) >= 5 {
		c0 := cleanColName(row[0])
		c4 := cleanColName(row[4])
		return (c0 == "sku" || c0 == "kode") && (c4 == "stok" || c4 == "stoksistem" || c4 == "stock")
	}
	return false
}

func parseQty(s string) (int32, error) {
	s = strings.TrimSpace(s)
	clean := strings.ReplaceAll(s, ",", "")
	if idx := strings.Index(clean, "."); idx != -1 {
		clean = clean[:idx]
	}
	v, err := strconv.ParseInt(clean, 10, 32)
	if err != nil {
		return 0, err
	}
	if v < 0 {
		return 0, fmt.Errorf("stok tidak boleh negatif")
	}
	return int32(v), nil
}
