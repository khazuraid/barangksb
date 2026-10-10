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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xuri/excelize/v2"
)

type ImportHandler struct {
	pool  *pgxpool.Pool
	itemH *ItemHandler
}

func NewImportHandler(pool *pgxpool.Pool, itemH *ItemHandler) *ImportHandler {
	return &ImportHandler{pool: pool, itemH: itemH}
}

// Header standar yang sepenuhnya selaras dengan formulir barang
var importHeaders = []string{
	"Jenis Barang / Nama Barang", "Kategori", "Ruangan / Lokasi", "No. Kode Lokasi",
	"No. Kode Barang", "No. Register", "Merk", "Model / Tipe", "No. Seri Pabrik",
	"Ukuran", "Bahan", "Tahun Pembuatan / Pembelian", "Keadaan Barang",
	"Jumlah / Stok", "Satuan", "Batas Minimum Stok", "Estimasi Harga Satuan",
	"Sumber Dana", "Distributor / Rekanan", "Izin Edar AKL / AKD", "Keterangan", "SKU",
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

		// Baris Contoh 1: Aset Mebel / KIR KIB
		cw.Write([]string{
			"Kursi Kerja Putar Ergonomis", "Mebel & Furniture Kantor", "Ruang Tata Usaha", "11.02.01",
			"02.06.01.02.05", "0001", "Chitose", "Ergo-X", "CHT-2024-889",
			"60 x 60 x 95 cm", "Besi & Busa Fabric", "2024", "Baik (B)",
			"1", "unit", "0", "1450000",
			"APBD", "PT Sarana Mandiri", "", "Pengadaan Meubelair Ruang TU TA 2024", "",
		})
		// Baris Contoh 2: Peralatan Medis & Alkes (AKL/AKD)
		cw.Write([]string{
			"Tensimeter Digital", "Peralatan Medis & Alkes (AKL/AKD)", "Poli Umum", "11.03.02",
			"02.07.01.01.12", "0012", "Omron", "HEM-7120", "OMR-SN-99120",
			"12 x 10 x 8 cm", "Plastik ABS", "2024", "Baik (B)",
			"1", "unit", "0", "850000",
			"DAK Non Fisik", "PT Medika Jaya", "KEMENKES RI AKL 20501812345", "Lengkap manset M & adaptor", "",
		})
		// Baris Contoh 3: Barang Habis Pakai / Konsumabel
		cw.Write([]string{
			"Kertas HVS A4 80gr", "Alat Tulis Kantor & Kertas", "Gudang Utama", "11.01.01",
			"", "", "PaperOne", "A4 80 GSM", "",
			"210 x 297 mm", "Kertas", "2025", "Baik (B)",
			"50", "rim", "10", "55000",
			"BOS / Operasional", "CV Prima Stationery", "", "Dus isi 5 rim", "",
		})
		cw.Flush()
		return
	}

	// Ambil kategori dari database
	catRows, err := h.pool.Query(c, `SELECT name FROM categories ORDER BY name`)
	var categories []string
	if err == nil {
		for catRows.Next() {
			var catName string
			if catRows.Scan(&catName) == nil && catName != "" {
				categories = append(categories, catName)
			}
		}
		catRows.Close()
	}
	if len(categories) == 0 {
		categories = []string{
			"Peralatan Medis & Alkes (AKL/AKD)",
			"Elektronik & IT Perkantoran",
			"Furnitur & Perlengkapan Ruangan",
			"ATK (Alat Tulis Kantor)",
			"Alat Laboratorium & Diagnostik",
			"Pantri & Fasilitas Umum",
			"Kebersihan & Sanitasi",
			"Keamanan & K3",
			"Lainnya",
		}
	}

	// Ambil lokasi dari database
	locRows, err := h.pool.Query(c, `SELECT name, COALESCE(code, '') FROM locations ORDER BY name`)
	type locRef struct {
		Name string
		Code string
	}
	var locations []locRef
	if err == nil {
		for locRows.Next() {
			var l locRef
			if locRows.Scan(&l.Name, &l.Code) == nil && l.Name != "" {
				locations = append(locations, l)
			}
		}
		locRows.Close()
	}
	if len(locations) == 0 {
		locations = []locRef{
			{Name: "Gudang Utama", Code: "11.01.01"},
			{Name: "Ruang Tata Usaha", Code: "11.02.01"},
			{Name: "Poli Umum", Code: "11.03.02"},
			{Name: "Ruang Rawat Inap", Code: "11.04.01"},
		}
	}

	// Excel XLSX Template
	f := excelize.NewFile()
	sheet := "Template_Barang"
	f.SetSheetName("Sheet1", sheet)

	for i, hd := range importHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, hd)
	}

	// Baris Contoh 1: Aset Mebel / KIR KIB
	sample1 := []any{
		"Kursi Kerja Putar Ergonomis", "Mebel & Furniture Kantor", "Ruang Tata Usaha", "11.02.01",
		"02.06.01.02.05", "0001", "Chitose", "Ergo-X", "CHT-2024-889",
		"60 x 60 x 95 cm", "Besi & Busa Fabric", "2024", "Baik (B)",
		1, "unit", 0, 1450000,
		"APBD", "PT Sarana Mandiri", "", "Pengadaan Meubelair Ruang TU TA 2024", "",
	}
	for i, val := range sample1 {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		f.SetCellValue(sheet, cell, val)
	}

	// Baris Contoh 2: Peralatan Medis & Alkes (AKL/AKD)
	sample2 := []any{
		"Tensimeter Digital", "Peralatan Medis & Alkes (AKL/AKD)", "Poli Umum", "11.03.02",
		"02.07.01.01.12", "0012", "Omron", "HEM-7120", "OMR-SN-99120",
		"12 x 10 x 8 cm", "Plastik ABS", "2024", "Baik (B)",
		1, "unit", 0, 850000,
		"DAK Non Fisik", "PT Medika Jaya", "KEMENKES RI AKL 20501812345", "Lengkap manset M & adaptor", "",
	}
	for i, val := range sample2 {
		cell, _ := excelize.CoordinatesToCellName(i+1, 3)
		f.SetCellValue(sheet, cell, val)
	}

	// Baris Contoh 3: Barang Habis Pakai / Konsumabel
	sample3 := []any{
		"Kertas HVS A4 80gr", "Alat Tulis Kantor & Kertas", "Gudang Utama", "11.01.01",
		"", "", "PaperOne", "A4 80 GSM", "",
		"210 x 297 mm", "Kertas", "2025", "Baik (B)",
		50, "rim", 10, 55000,
		"BOS / Operasional", "CV Prima Stationery", "", "Dus isi 5 rim", "",
	}
	for i, val := range sample3 {
		cell, _ := excelize.CoordinatesToCellName(i+1, 4)
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
	f.SetRowHeight(sheet, 1, 26)

	// Set column widths
	for i := 1; i <= len(importHeaders); i++ {
		colName, _ := excelize.ColumnNumberToName(i)
		f.SetColWidth(sheet, colName, colName, 22)
	}

	// Buat Sheet Referensi untuk Drop Down Kategori & Ruangan
	refSheet := "Data_Referensi"
	f.NewSheet(refSheet)

	f.SetCellValue(refSheet, "A1", "Daftar Kategori")
	f.SetCellValue(refSheet, "B1", "Daftar Ruangan / Lokasi")
	f.SetCellValue(refSheet, "C1", "No. Kode Lokasi")
	f.SetCellValue(refSheet, "D1", "Keadaan Barang")
	f.SetCellValue(refSheet, "E1", "Satuan Standar")

	refHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#334155"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	f.SetCellStyle(refSheet, "A1", "E1", refHeaderStyle)
	f.SetRowHeight(refSheet, 1, 24)

	for i, cat := range categories {
		cell, _ := excelize.CoordinatesToCellName(1, i+2)
		f.SetCellValue(refSheet, cell, cat)
	}

	for i, loc := range locations {
		cellName, _ := excelize.CoordinatesToCellName(2, i+2)
		f.SetCellValue(refSheet, cellName, loc.Name)
		cellCode, _ := excelize.CoordinatesToCellName(3, i+2)
		f.SetCellValue(refSheet, cellCode, loc.Code)
	}

	conditions := []string{
		"Baik (B)",
		"Kurang Baik (KB)",
		"Rusak Berat (RB)",
		"Berfungsi",
		"Rusak Ringan",
		"Perlu Kalibrasi",
	}
	for i, cond := range conditions {
		cell, _ := excelize.CoordinatesToCellName(4, i+2)
		f.SetCellValue(refSheet, cell, cond)
	}

	units := []string{
		"buah", "unit", "set", "box", "rim", "pack", "dus", "botol", "roll", "lembar",
	}
	for i, u := range units {
		cell, _ := excelize.CoordinatesToCellName(5, i+2)
		f.SetCellValue(refSheet, cell, u)
	}

	f.SetColWidth(refSheet, "A", "A", 35)
	f.SetColWidth(refSheet, "B", "B", 30)
	f.SetColWidth(refSheet, "C", "C", 16)
	f.SetColWidth(refSheet, "D", "D", 22)
	f.SetColWidth(refSheet, "E", "E", 18)

	// Tambahkan Dropdown Validasi Data pada Template_Barang
	// 1. Dropdown Kategori di kolom B (baris 2 s/d 500)
	dvCat := excelize.NewDataValidation(true)
	dvCat.SetSqref("B2:B500")
	dvCat.SetSqrefDropList(fmt.Sprintf("Data_Referensi!$A$2:$A$%d", len(categories)+1))
	f.AddDataValidation(sheet, dvCat)

	// 2. Dropdown Ruangan / Lokasi di kolom C (baris 2 s/d 500)
	dvLoc := excelize.NewDataValidation(true)
	dvLoc.SetSqref("C2:C500")
	dvLoc.SetSqrefDropList(fmt.Sprintf("Data_Referensi!$B$2:$B$%d", len(locations)+1))
	f.AddDataValidation(sheet, dvLoc)

	// 3. Dropdown Keadaan Barang di kolom M (baris 2 s/d 500)
	dvCond := excelize.NewDataValidation(true)
	dvCond.SetSqref("M2:M500")
	dvCond.SetSqrefDropList(fmt.Sprintf("Data_Referensi!$D$2:$D$%d", len(conditions)+1))
	f.AddDataValidation(sheet, dvCond)

	// 4. Dropdown Satuan di kolom O (baris 2 s/d 500)
	dvUnit := excelize.NewDataValidation(true)
	dvUnit.SetSqref("O2:O500")
	dvUnit.SetSqrefDropList(fmt.Sprintf("Data_Referensi!$E$2:$E$%d", len(units)+1))
	f.AddDataValidation(sheet, dvUnit)

	// Set sheet aktif kembali ke Template_Barang
	templateIdx, _ := f.GetSheetIndex(sheet)
	if templateIdx >= 0 {
		f.SetActiveSheet(templateIdx)
	}

	var buf bytes.Buffer
	f.Write(&buf)
	c.Header("Content-Disposition", `attachment; filename="template_import_inventaris.xlsx"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

// Pemetaan alias header fleksibel (tidak peka huruf besar/kecil atau variasi nama)
var headerAliases = map[string]string{
	// Nama
	"namabarang":            "name",
	"jenisbarangnamabarang": "name",
	"jenisbarang":           "name",
	"nama":                  "name",
	"itemname":              "name",

	// Kategori
	"kategori":       "category",
	"kategoribarang": "category",
	"category":       "category",

	// Lokasi
	"ruanganlokasi": "location",
	"lokasi":        "location",
	"ruangan":       "location",
	"location":      "location",

	// Kode Lokasi
	"nokodelokasi": "location_code",
	"kodelokasi":   "location_code",
	"locationcode": "location_code",

	// Kode Barang
	"nokodebarang": "item_code",
	"kodebarang":   "item_code",
	"itemcode":     "item_code",

	// Register
	"noregister":           "register_number",
	"register":             "register_number",
	"jumlahbarangregister": "register_number",
	"nomorregister":        "register_number",
	"registernumber":       "register_number",

	// Merk
	"merk":  "merk",
	"brand": "merk",

	// Model / Tipe
	"modeltipe": "type_model",
	"model":     "type_model",
	"tipe":      "type_model",
	"typemodel": "type_model",

	// Serial Number
	"noseripabrik": "serial_number",
	"nomorseri":    "serial_number",
	"seri":         "serial_number",
	"serialnumber": "serial_number",
	"sn":           "serial_number",

	// Ukuran
	"ukuran":  "size",
	"dimensi": "size",
	"size":    "size",

	// Bahan
	"bahan":    "material",
	"material": "material",

	// Tahun
	"tahunpembuatanpembelian": "procurement_year",
	"tahunpengadaan":          "procurement_year",
	"tahunpembuatan":          "procurement_year",
	"tahunpembelian":          "procurement_year",
	"tahun":                   "procurement_year",
	"procurementyear":         "procurement_year",

	// Kondisi
	"keadaanbarang":   "condition",
	"kondisi":         "condition",
	"kondisibarang":   "condition",
	"condition":       "condition",
	"conditionstatus": "condition",

	// Stok
	"jumlahstok":   "stock",
	"stok":         "stock",
	"jumlah":       "stock",
	"stock":        "stock",
	"currentstock": "stock",
	"qty":          "stock",

	// Satuan
	"satuan":     "unit",
	"satuanunit": "unit",
	"unit":       "unit",

	// Min Stok
	"batasminimumstok": "min_stock",
	"batasminimum":     "min_stock",
	"minimumstok":      "min_stock",
	"minstock":         "min_stock",

	// Harga
	"estimasihargasatuan": "price",
	"hargasatuan":         "price",
	"harga":               "price",
	"price":               "price",
	"priceperunit":        "price",

	// Sumber Dana
	"sumberdana":    "funding_source",
	"fundingsource": "funding_source",

	// Distributor
	"distributorrekanan": "distributor",
	"distributor":        "distributor",
	"rekanan":            "distributor",
	"vendor":             "distributor",

	// AKL/AKD
	"izinedaraklakd": "akl_akd",
	"aklakd":         "akl_akd",
	"izinedar":       "akl_akd",

	// Keterangan
	"keterangan":  "description",
	"catatan":     "description",
	"deskripsi":   "description",
	"description": "description",

	// SKU
	"sku":       "sku",
	"skukustom": "sku",

	// Model Pengelolaan
	"jenispengelolaan": "management_type",
	"modelpengelolaan": "management_type",
	"trackstock":       "management_type",
}

func normalizeHeader(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func normalizeCondition(c string) string {
	c = strings.TrimSpace(c)
	if c == "" {
		return "Baik (B)"
	}
	low := strings.ToLower(c)
	switch {
	case strings.Contains(low, "rusak berat") || low == "rb":
		return "Rusak Berat (RB)"
	case strings.Contains(low, "kurang baik") || low == "kb":
		return "Kurang Baik (KB)"
	case strings.Contains(low, "rusak ringan") || low == "rr":
		return "Rusak Ringan"
	case strings.Contains(low, "kalibrasi"):
		return "Perlu Kalibrasi"
	case strings.Contains(low, "berfungsi"):
		return "Berfungsi"
	case strings.Contains(low, "baik") || low == "b":
		return "Baik (B)"
	default:
		return c
	}
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
		// Prioritaskan sheet data item (abaikan sheet referensi dropdown)
		dataSheet := sheets[0]
		for _, s := range sheets {
			if s != "Data_Referensi" {
				dataSheet = s
				break
			}
		}
		records, err := xlFile.GetRows(dataSheet)
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

	// Buat peta kolom dari baris header pertama
	colMap := make(map[string]int)
	for i, col := range rows[0] {
		norm := normalizeHeader(col)
		if field, ok := headerAliases[norm]; ok {
			colMap[field] = i
		}
	}

	// Helper membaca isi kolom berdasarkan alias atau urutan bawaan
	getField := func(row []string, field string, fallbackIdx int) string {
		if idx, ok := colMap[field]; ok && idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
		if fallbackIdx >= 0 && fallbackIdx < len(row) {
			return strings.TrimSpace(row[fallbackIdx])
		}
		return ""
	}

	successCount := 0
	skippedCount := 0
	failedCount := 0
	var errorList []string
	seenKeys := make(map[string]bool)

	for idx, row := range rows {
		// Lewati baris header
		if idx == 0 {
			continue
		}
		rowNum := idx + 1
		if len(row) == 0 {
			continue
		}

		name := getField(row, "name", 0)
		category := getField(row, "category", 1)
		location := getField(row, "location", 2)

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

		// Field KIR / KIB
		itemCode := getField(row, "item_code", 4)
		registerNumber := getField(row, "register_number", 5)
		merk := getField(row, "merk", 6)
		typeModel := getField(row, "type_model", 7)
		serialNumber := getField(row, "serial_number", 8)
		size := getField(row, "size", 9)
		material := getField(row, "material", 10)
		procurementYear := getField(row, "procurement_year", 11)
		condition := normalizeCondition(getField(row, "condition", 12))

		// Stok & Satuan
		stock := parseInt32(getField(row, "stock", 13), 0)
		unit := getField(row, "unit", 14)
		if unit == "" {
			unit = "buah"
		}
		minStock := parseInt32(getField(row, "min_stock", 15), 0)
		price := parseInt64(getField(row, "price", 16), 0)

		// Pengadaan & Keterangan
		fundingSource := getField(row, "funding_source", 17)
		distributor := getField(row, "distributor", 18)
		aklAkd := getField(row, "akl_akd", 19)
		description := getField(row, "description", 20)
		customSKU := getField(row, "sku", 21)

		// Cek apakah barang sudah ada di sistem (jika sudah ada jangan buat baru)
		var exists bool
		var dedupKey string
		if customSKU != "" {
			dedupKey = "sku:" + strings.ToLower(customSKU)
			if seenKeys[dedupKey] {
				skippedCount++
				continue
			}
			_ = h.pool.QueryRow(c, `SELECT EXISTS(SELECT 1 FROM inventory_items WHERE LOWER(sku) = LOWER($1))`, customSKU).Scan(&exists)
		} else if itemCode != "" && registerNumber != "" {
			dedupKey = fmt.Sprintf("reg:%s:%s", itemCode, registerNumber)
			if seenKeys[dedupKey] {
				skippedCount++
				continue
			}
			_ = h.pool.QueryRow(c, `SELECT EXISTS(SELECT 1 FROM inventory_items WHERE item_code = $1 AND register_number = $2)`, itemCode, registerNumber).Scan(&exists)
		} else if serialNumber != "" {
			dedupKey = "sn:" + serialNumber
			if seenKeys[dedupKey] {
				skippedCount++
				continue
			}
			_ = h.pool.QueryRow(c, `SELECT EXISTS(SELECT 1 FROM inventory_items WHERE serial_number = $1)`, serialNumber).Scan(&exists)
		} else {
			dedupKey = fmt.Sprintf("name_loc:%s:%s", strings.ToLower(strings.TrimSpace(name)), strings.ToLower(strings.TrimSpace(location)))
			if seenKeys[dedupKey] {
				skippedCount++
				continue
			}
			_ = h.pool.QueryRow(c, `SELECT EXISTS(SELECT 1 FROM inventory_items WHERE LOWER(TRIM(name)) = LOWER(TRIM($1)) AND LOWER(TRIM(location)) = LOWER(TRIM($2)))`, name, location).Scan(&exists)
		}

		if exists {
			seenKeys[dedupKey] = true
			skippedCount++
			continue
		}

		// Pastikan Kategori & Lokasi tersimpan di master data
		h.ensureCategory(c, category)
		locCode := getField(row, "location_code", 3)
		resolvedLocCode := h.ensureLocation(c, location, locCode)
		if locCode == "" {
			locCode = resolvedLocCode
		}

		// Model pengelolaan stok
		trackStock := true
		if mgmt := strings.ToLower(getField(row, "management_type", -1)); mgmt != "" {
			if strings.Contains(mgmt, "aset") || strings.Contains(mgmt, "tetap") || strings.Contains(mgmt, "mandiri") || strings.Contains(mgmt, "false") {
				trackStock = false
			}
		}

		// Tentukan SKU (gunakan custom jika terisi, atau auto-generate)
		sku := customSKU
		if sku == "" {
			sku = generateSKU(h.itemH, c, category)
		}

		_, err := h.pool.Exec(c, `
			INSERT INTO inventory_items (
				sku, name, category, location, location_code, current_stock, min_stock, unit,
				price_per_unit, condition_status, merk, type_model, serial_number,
				procurement_year, funding_source, distributor, akl_akd, description,
				size, material, item_code, register_number, track_stock
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23)
		`, sku, name, category, location, nullable(locCode), stock, minStock, unit,
			price, condition, nullable(merk), nullable(typeModel), nullable(serialNumber),
			nullable(procurementYear), nullable(fundingSource), nullable(distributor),
			nullable(aklAkd), nullable(description), nullable(size), nullable(material),
			nullable(itemCode), nullable(registerNumber), trackStock)

		if err != nil {
			failedCount++
			errorList = append(errorList, fmt.Sprintf("Baris %d (%s): %s", rowNum, name, err.Error()))
		} else {
			seenKeys[dedupKey] = true
			successCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": successCount > 0 || skippedCount > 0,
		"count":   successCount,
		"skipped": skippedCount,
		"failed":  failedCount,
		"total":   successCount + skippedCount + failedCount,
		"errors":  errorList,
	})
}

func (h *ImportHandler) ensureCategory(c *gin.Context, catName string) {
	var count int
	h.pool.QueryRow(c, `SELECT count(*) FROM categories WHERE name=$1`, catName).Scan(&count)
	if count == 0 {
		id := strings.ToLower(strings.ReplaceAll(catName, " ", "-"))
		var idCount int
		h.pool.QueryRow(c, `SELECT count(*) FROM categories WHERE id=$1`, id).Scan(&idCount)
		if idCount > 0 {
			id = fmt.Sprintf("%s-%d", id, time.Now().UnixNano()%10000)
		}
		h.pool.Exec(c, `INSERT INTO categories (id, name) VALUES ($1, $2) ON CONFLICT (name) DO NOTHING`, id, catName)
	}
}

func (h *ImportHandler) ensureLocation(c *gin.Context, locName string, locCode string) string {
	var codeFromDB string
	err := h.pool.QueryRow(c, `SELECT COALESCE(code, '') FROM locations WHERE name=$1`, locName).Scan(&codeFromDB)
	if err != nil {
		id := strings.ToLower(strings.ReplaceAll(locName, " ", "-"))
		var idCount int
		h.pool.QueryRow(c, `SELECT count(*) FROM locations WHERE id=$1`, id).Scan(&idCount)
		if idCount > 0 {
			id = fmt.Sprintf("%s-%d", id, time.Now().UnixNano()%10000)
		}
		h.pool.Exec(c, `INSERT INTO locations (id, name, code) VALUES ($1, $2, $3) ON CONFLICT (name) DO NOTHING`, id, locName, locCode)
		return locCode
	}
	if locCode != "" && codeFromDB == "" {
		h.pool.Exec(c, `UPDATE locations SET code=$2 WHERE name=$1`, locName, locCode)
		return locCode
	}
	if locCode == "" {
		return codeFromDB
	}
	return locCode
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
