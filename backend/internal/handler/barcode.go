package handler

import (
	"bytes"
	"fmt"
	"html"
	"image/png"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/qr"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BarcodeHandler struct {
	pool *pgxpool.Pool
}

func NewBarcodeHandler(pool *pgxpool.Pool) *BarcodeHandler {
	return &BarcodeHandler{pool: pool}
}

// GET /api/barcode/:id.png?fmt=qr — public (no auth), renders barcode image
func (h *BarcodeHandler) PNG(c *gin.Context) {
	id := c.Param("id")
	// strip .png suffix if gin didn't parse it
	id = strings.TrimSuffix(id, ".png")
	slog.Info("barcode request", "raw_param", c.Param("id"), "cleaned_id", id)
	format := c.DefaultQuery("fmt", "qr")

	var sku string
	err := h.pool.QueryRow(c, `SELECT sku FROM inventory_items WHERE id::text=$1`, id).Scan(&sku)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	content := sku
	if format == "qr" || format == "" {
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		content = fmt.Sprintf("%s://%s/scan/%s", scheme, c.Request.Host, id)
	}

	var pngBytes []byte
	switch format {
	case "qr", "":
		pngBytes = renderQR(content)
	case "code128":
		pngBytes = renderCode128(content)
	default:
		pngBytes = renderQR(content)
	}

	c.Data(http.StatusOK, "image/png", pngBytes)
}

// GET /api/scan/:id — public (no auth), returns item detail for QR scan landing page
func (h *BarcodeHandler) ScanItem(c *gin.Context) {
	id := c.Param("id")

	var resp struct {
		Sku             string `json:"sku"`
		Name            string `json:"name"`
		Category        string `json:"category"`
		Location        string `json:"location"`
		CurrentStock    int32  `json:"current_stock"`
		MinStock        int32  `json:"min_stock"`
		Unit            string `json:"unit"`
		ConditionStatus string `json:"condition_status"`
		IsAvailable     bool   `json:"is_available"`
		Merk            string `json:"merk"`
		TypeModel       string `json:"type_model"`
		SerialNumber    string `json:"serial_number"`
		ProcurementYear string `json:"procurement_year"`
		FundingSource   string `json:"funding_source"`
		AklAkd          string `json:"akl_akd"`
		PhotoURL        string   `json:"photo_url"`
		GeoLat          *float64 `json:"geo_lat"`
		GeoLng          *float64 `json:"geo_lng"`
		Maintenances    []gin.H  `json:"maintenances"`
	}

	err := h.pool.QueryRow(c,
		`SELECT sku, name, category, location, current_stock, min_stock, unit,
		        COALESCE(condition_status, 'Berfungsi'), is_available, COALESCE(merk, ''), COALESCE(type_model, ''), COALESCE(serial_number, ''),
		        COALESCE(procurement_year, ''), COALESCE(funding_source, ''), COALESCE(akl_akd, ''), COALESCE(photo_url, ''),
		        geo_lat, geo_lng
		 FROM inventory_items WHERE id::text=$1`, id).
		Scan(&resp.Sku, &resp.Name, &resp.Category, &resp.Location,
			&resp.CurrentStock, &resp.MinStock, &resp.Unit,
			&resp.ConditionStatus, &resp.IsAvailable, &resp.Merk,
			&resp.TypeModel, &resp.SerialNumber, &resp.ProcurementYear,
			&resp.FundingSource, &resp.AklAkd, &resp.PhotoURL,
			&resp.GeoLat, &resp.GeoLng)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item tidak ditemukan"})
		return
	}

	// Fetch maintenance history for public view
	mRows, err := h.pool.Query(c, `
		SELECT service_type, service_date::text,
		       COALESCE(vendor_or_technician, ''), COALESCE(description, ''),
		       next_service_date::text, COALESCE(photo_url, '')
		FROM item_maintenances
		WHERE item_id::text = $1
		ORDER BY service_date DESC, created_at DESC
	`, id)
	resp.Maintenances = []gin.H{}
	if err == nil {
		defer mRows.Close()
		for mRows.Next() {
			var sType, sDate, vendor, desc, photo string
			var nextDate *string
			mRows.Scan(&sType, &sDate, &vendor, &desc, &nextDate, &photo)
			mItem := gin.H{
				"service_type":          sType,
				"service_date":          sDate,
				"vendor_or_technician": vendor,
				"description":           desc,
				"next_service_date":     nextDate,
				"photo_url":             photo,
			}
			resp.Maintenances = append(resp.Maintenances, mItem)
		}
	}

	c.JSON(http.StatusOK, resp)
}

type LabelItem struct {
	ID       string
	Name     string
	SKU      string
	Location string
	Category string
}

// GET /api/barcode/sheet?ids=a,b,c&layout=...&fmt=...&copies=...&offset=...&show_border=...
func (h *BarcodeHandler) Sheet(c *gin.Context) {
	ids := c.Query("ids")
	layout := c.DefaultQuery("layout", "tj_103") // default to Tom & Jerry 103 (popular Indonesian sticker sheet)
	fmtType := c.DefaultQuery("fmt", "qr")
	title := c.DefaultQuery("title", "INVENTARIS KANTOR")
	showName := c.DefaultQuery("show_name", "1") != "0"
	showSKU := c.DefaultQuery("show_sku", "1") != "0"
	showLoc := c.DefaultQuery("show_loc", "1") != "0"
	showBorder := c.DefaultQuery("show_border", "1") != "0"

	copies, _ := strconv.Atoi(c.DefaultQuery("copies", "1"))
	if copies < 1 {
		copies = 1
	}
	if copies > 50 {
		copies = 50
	}

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if offset < 0 {
		offset = 0
	}
	if offset > 100 {
		offset = 100
	}

	idList := splitComma(ids)
	var items []LabelItem
	for _, id := range idList {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		var it LabelItem
		it.ID = id
		err := h.pool.QueryRow(c, `
			SELECT name, COALESCE(sku, ''), COALESCE(location, ''), COALESCE(category, '')
			FROM inventory_items WHERE id::text=$1
		`, id).Scan(&it.Name, &it.SKU, &it.Location, &it.Category)
		if err == nil {
			items = append(items, it)
		}
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, h.renderSheetHTML(items, layout, fmtType, title, showName, showSKU, showLoc, copies, offset, showBorder))
}

func (h *BarcodeHandler) renderSheetHTML(rawItems []LabelItem, layout, fmtType, title string, showName, showSKU, showLoc bool, copies, offset int, showBorder bool) string {
	layoutNames := map[string]string{
		"tj_103":      "Label Tom & Jerry No. 103 (12 Label / Lembar · 3x4 · 64x32 mm)",
		"tj_107":      "Label Tom & Jerry No. 107 (24 Label / Lembar · 3x8 · 50x18 mm)",
		"tj_108":      "Label Tom & Jerry No. 108 (40 Label / Lembar · 5x8 · 38x18 mm)",
		"tj_121":      "Label Tom & Jerry No. 121 (10 Label / Lembar · 2x5 · 75x38 mm)",
		"a4_3x7":      "Kertas Stiker A4 (21 Label / Lembar · 3x7 · 65x38 mm)",
		"a4_4x10":     "Kertas Stiker A4 (40 Label / Lembar · 4x10 · 48x25.5 mm)",
		"thermal_58":  "Printer Thermal Roll 58 mm",
		"thermal_80":  "Printer Thermal Roll 80 mm",
		"label_50x30": "Stiker Rol 50 x 30 mm",
		"label_40x30": "Stiker Rol 40 x 30 mm",
		"label_60x40": "Stiker Rol 60 x 40 mm",
		"label_70x50": "Stiker Rol 70 x 50 mm",
		"label_100x50": "Stiker Rol 100 x 50 mm",
	}
	layoutLabel := layoutNames[layout]
	if layoutLabel == "" {
		layoutLabel = layout
	}

	// Expand items with copies
	var expandedItems []LabelItem
	for _, it := range rawItems {
		for cIdx := 0; cIdx < copies; cIdx++ {
			expandedItems = append(expandedItems, it)
		}
	}

	// Prepend offset (blank stickers)
	var fullItems []LabelItem
	for o := 0; o < offset; o++ {
		fullItems = append(fullItems, LabelItem{ID: ""})
	}
	fullItems = append(fullItems, expandedItems...)

	isSheet := layout == "tj_103" || layout == "tj_107" || layout == "tj_108" || layout == "tj_121" || layout == "a4_3x7" || layout == "a4_4x10"

	perSheet := 12
	switch layout {
	case "tj_103":
		perSheet = 12
	case "tj_107":
		perSheet = 24
	case "tj_108":
		perSheet = 40
	case "tj_121":
		perSheet = 10
	case "a4_3x7":
		perSheet = 21
	case "a4_4x10":
		perSheet = 40
	}

	borderStyle := "border: 1px dashed #cbd5e1;"
	if !showBorder {
		borderStyle = "border: 1px solid transparent;"
	}

	var cssPage, cssContainer, cssCard string
	switch layout {
	case "tj_103":
		// Tom & Jerry No. 103 (3 cols x 4 rows, 64x32 mm per sticker)
		cssPage = `@page { size: auto; margin: 6mm 5mm; }`
		cssContainer = `width: 200mm; margin: 0 auto; display: grid; grid-template-columns: repeat(3, 64mm); gap: 3.5mm 4mm; justify-content: center;`
		cssCard = fmt.Sprintf(`width: 64mm; height: 32mm; %s border-radius: 4px; box-sizing: border-box; padding: 2mm 2.5mm; display: flex; flex-direction: row; align-items: center; justify-content: space-between; overflow: hidden; background: #fff;`, borderStyle)

	case "tj_107":
		// Tom & Jerry No. 107 (3 cols x 8 rows, 50x18 mm per sticker)
		cssPage = `@page { size: auto; margin: 5mm; }`
		cssContainer = `width: 160mm; margin: 0 auto; display: grid; grid-template-columns: repeat(3, 50mm); gap: 2mm 3mm; justify-content: center;`
		cssCard = fmt.Sprintf(`width: 50mm; height: 18mm; %s border-radius: 3px; box-sizing: border-box; padding: 1mm 1.5mm; display: flex; flex-direction: row; align-items: center; justify-content: space-between; overflow: hidden; background: #fff;`, borderStyle)

	case "tj_108":
		// Tom & Jerry No. 108 (5 cols x 8 rows, 38x18 mm per sticker)
		cssPage = `@page { size: auto; margin: 5mm; }`
		cssContainer = `width: 198mm; margin: 0 auto; display: grid; grid-template-columns: repeat(5, 38mm); gap: 2mm 2mm; justify-content: center;`
		cssCard = fmt.Sprintf(`width: 38mm; height: 18mm; %s border-radius: 3px; box-sizing: border-box; padding: 1mm 1.5mm; display: flex; flex-direction: row; align-items: center; justify-content: space-between; overflow: hidden; background: #fff;`, borderStyle)

	case "tj_121":
		// Tom & Jerry No. 121 (2 cols x 5 rows, 75x38 mm per sticker)
		cssPage = `@page { size: auto; margin: 6mm 5mm; }`
		cssContainer = `width: 160mm; margin: 0 auto; display: grid; grid-template-columns: repeat(2, 75mm); gap: 3mm 4mm; justify-content: center;`
		cssCard = fmt.Sprintf(`width: 75mm; height: 38mm; %s border-radius: 4px; box-sizing: border-box; padding: 2.5mm 3mm; display: flex; flex-direction: row; align-items: center; justify-content: space-between; overflow: hidden; background: #fff;`, borderStyle)

	case "a4_3x7":
		cssPage = `@page { size: A4 portrait; margin: 8mm; }`
		cssContainer = `display: grid; grid-template-columns: repeat(3, 1fr); gap: 4mm; padding: 0; margin: 0;`
		cssCard = fmt.Sprintf(`height: 38mm; %s border-radius: 4px; box-sizing: border-box; padding: 2.5mm; display: flex; flex-direction: row; align-items: center; justify-content: space-between; overflow: hidden; background: #fff;`, borderStyle)

	case "a4_4x10":
		cssPage = `@page { size: A4 portrait; margin: 6mm; }`
		cssContainer = `display: grid; grid-template-columns: repeat(4, 1fr); gap: 2.5mm; padding: 0; margin: 0;`
		cssCard = fmt.Sprintf(`height: 25.5mm; %s border-radius: 3px; box-sizing: border-box; padding: 1.5mm; display: flex; flex-direction: row; align-items: center; justify-content: space-between; overflow: hidden; background: #fff;`, borderStyle)

	case "thermal_58":
		cssPage = `@page { size: 58mm auto; margin: 0; }`
		cssContainer = `width: 54mm; margin: 0 auto; padding: 2mm 0;`
		cssCard = `width: 50mm; margin: 0 auto 3mm auto; padding: 2mm; text-align: center; border-bottom: 1px dashed #666; page-break-after: always; break-after: page; box-sizing: border-box;`

	case "thermal_80":
		cssPage = `@page { size: 80mm auto; margin: 0; }`
		cssContainer = `width: 76mm; margin: 0 auto; padding: 2mm 0;`
		cssCard = `width: 72mm; margin: 0 auto 4mm auto; padding: 3mm; text-align: center; border-bottom: 1px dashed #666; page-break-after: always; break-after: page; box-sizing: border-box;`

	case "label_40x30":
		cssPage = `@page { size: 40mm 30mm; margin: 0; }`
		cssContainer = `margin: 0; padding: 0;`
		cssCard = fmt.Sprintf(`width: 40mm; height: 30mm; margin: 0; padding: 1.5mm; box-sizing: border-box; display: flex; flex-direction: row; align-items: center; justify-content: space-between; page-break-after: always; break-after: page; overflow: hidden; %s`, borderStyle)

	case "label_50x30":
		cssPage = `@page { size: 50mm 30mm; margin: 0; }`
		cssContainer = `margin: 0; padding: 0;`
		cssCard = fmt.Sprintf(`width: 50mm; height: 30mm; margin: 0; padding: 1.5mm 2.5mm; box-sizing: border-box; display: flex; flex-direction: row; align-items: center; justify-content: space-between; page-break-after: always; break-after: page; overflow: hidden; %s`, borderStyle)

	case "label_60x40":
		cssPage = `@page { size: 60mm 40mm; margin: 0; }`
		cssContainer = `margin: 0; padding: 0;`
		cssCard = fmt.Sprintf(`width: 60mm; height: 40mm; margin: 0; padding: 2.5mm 3mm; box-sizing: border-box; display: flex; flex-direction: row; align-items: center; justify-content: space-between; page-break-after: always; break-after: page; overflow: hidden; %s`, borderStyle)

	case "label_70x50":
		cssPage = `@page { size: 70mm 50mm; margin: 0; }`
		cssContainer = `margin: 0; padding: 0;`
		cssCard = fmt.Sprintf(`width: 70mm; height: 50mm; margin: 0; padding: 3mm 4mm; box-sizing: border-box; display: flex; flex-direction: row; align-items: center; justify-content: space-between; page-break-after: always; break-after: page; overflow: hidden; %s`, borderStyle)

	case "label_100x50":
		cssPage = `@page { size: 100mm 50mm; margin: 0; }`
		cssContainer = `margin: 0; padding: 0;`
		cssCard = fmt.Sprintf(`width: 100mm; height: 50mm; margin: 0; padding: 3mm 5mm; box-sizing: border-box; display: flex; flex-direction: row; align-items: center; justify-content: space-between; page-break-after: always; break-after: page; overflow: hidden; %s`, borderStyle)

	default:
		cssPage = `@page { size: auto; margin: 6mm; }`
		cssContainer = `width: 200mm; margin: 0 auto; display: grid; grid-template-columns: repeat(3, 64mm); gap: 3.5mm 4mm; justify-content: center;`
		cssCard = fmt.Sprintf(`width: 64mm; height: 32mm; %s border-radius: 4px; box-sizing: border-box; padding: 2mm 2.5mm; display: flex; flex-direction: row; align-items: center; justify-content: space-between; overflow: hidden; background: #fff;`, borderStyle)
	}

	var contentHTML strings.Builder

	if isSheet {
		// Paginate into sheet pages
		totalLabels := len(fullItems)
		sheetCount := (totalLabels + perSheet - 1) / perSheet
		if sheetCount < 1 {
			sheetCount = 1
		}

		for sIdx := 0; sIdx < sheetCount; sIdx++ {
			start := sIdx * perSheet
			end := start + perSheet
			if end > totalLabels {
				end = totalLabels
			}
			slice := fullItems[start:end]
			// Pad remainder of sheet with blank items so layout grid doesn't distort
			for len(slice) < perSheet {
				slice = append(slice, LabelItem{ID: ""})
			}

			contentHTML.WriteString(`<div class="sheet-page">`)
			for _, it := range slice {
				contentHTML.WriteString(h.renderSingleLabel(it, layout, fmtType, title, showName, showSKU, showLoc))
			}
			contentHTML.WriteString(`</div>`)
		}
	} else {
		contentHTML.WriteString(`<div class="print-page-wrapper">`)
		for _, it := range fullItems {
			contentHTML.WriteString(h.renderSingleLabel(it, layout, fmtType, title, showName, showSKU, showLoc))
		}
		contentHTML.WriteString(`</div>`)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="id">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>%s - %s (%d Label)</title>
<style>
* { box-sizing: border-box; -webkit-print-color-adjust: exact; print-color-adjust: exact; }
body {
  margin: 0;
  padding: 0;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  color: #000;
  background: #f1f5f9;
}
@media screen {
  body { padding-top: 56px; }
  .screen-toolbar {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    height: 52px;
    background: #0f172a;
    color: #fff;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 16px;
    z-index: 99999;
    box-shadow: 0 2px 8px rgba(0,0,0,0.25);
    font-size: 13px;
  }
  .screen-toolbar button {
    background: #6366f1;
    color: #fff;
    border: none;
    border-radius: 6px;
    padding: 8px 16px;
    font-weight: 600;
    font-size: 13px;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    transition: background 0.15s;
  }
  .screen-toolbar button:hover { background: #4f46e5; }
  .sheet-page {
    background: #fefce8; /* Characteristic Tom & Jerry backing paper yellow */
    border: 1px solid #fef08a;
    box-shadow: 0 4px 14px rgba(0,0,0,0.08);
    margin: 20px auto;
    border-radius: 6px;
    padding: 8mm 6mm;
  }
  .print-page-wrapper {
    background: #fff;
    margin: 20px auto;
    box-shadow: 0 4px 16px rgba(0,0,0,0.1);
  }
}
@media print {
  .screen-toolbar { display: none !important; }
  body { background: #fff !important; padding: 0 !important; }
  .sheet-page {
    background: transparent !important;
    border: none !important;
    box-shadow: none !important;
    margin: 0 !important;
    padding: 0 !important;
    page-break-after: always;
    break-after: page;
  }
  .print-page-wrapper { box-shadow: none !important; margin: 0 !important; }
}
%s
.sheet-page {
  %s
}
.print-page-wrapper {
  %s
}
.label-unit {
  %s
}
.label-unit.is-blank {
  visibility: hidden !important;
  border: none !important;
  background: transparent !important;
}
</style>
</head>
<body onload="if(window.location.search.indexOf('noprint=1')===-1){window.print();}">
<div class="screen-toolbar">
  <div style="display:flex;align-items:center;gap:12px">
    <strong>🏷️ Lembar Cetak Label Stiker</strong>
    <span style="opacity:0.8;font-size:12px">%s · %d Stiker Dicetak</span>
  </div>
  <div style="display:flex;align-items:center;gap:8px">
    <button onclick="window.print()">🖨️ Cetak Sekarang</button>
  </div>
</div>
%s
</body>
</html>`,
		html.EscapeString(title),
		html.EscapeString(layoutLabel),
		len(expandedItems),
		cssPage,
		cssContainer,
		cssContainer,
		cssCard,
		html.EscapeString(layoutLabel),
		len(expandedItems),
		contentHTML.String(),
	)
}

func (h *BarcodeHandler) renderSingleLabel(it LabelItem, layout, fmtType, title string, showName, showSKU, showLoc bool) string {
	if it.ID == "" {
		return `<div class="label-unit is-blank"></div>`
	}

	codeImgURL := fmt.Sprintf("/api/barcode/%s.png?fmt=%s", it.ID, fmtType)

	// Distinct rendering styles per layout geometry
	switch layout {
	case "tj_103":
		// Tom & Jerry No. 103 (64x32 mm, 3 cols x 4 rows) - Indonesian standard sticker
		if fmtType == "code128" {
			return fmt.Sprintf(`
<div class="label-unit" style="flex-direction:column;align-items:center;justify-content:center;text-align:center">
  <div style="font-size:7.5px;font-weight:800;text-transform:uppercase;letter-spacing:0.4px">%s</div>
  <img src="%s" style="height:12mm;max-width:56mm;margin:1mm 0" alt="%s"/>
  %s
  %s
  %s
</div>`,
				html.EscapeString(title), codeImgURL, html.EscapeString(it.SKU),
				ternaryStr(showName, fmt.Sprintf(`<div style="font-size:8.5px;font-weight:700;line-height:1.1;max-width:58mm;white-space:nowrap;overflow:hidden;text-overflow:ellipsis">%s</div>`, html.EscapeString(it.Name)), ""),
				ternaryStr(showSKU, fmt.Sprintf(`<div style="font-family:monospace;font-size:8px;font-weight:700">%s</div>`, html.EscapeString(it.SKU)), ""),
				ternaryStr(showLoc && it.Location != "", fmt.Sprintf(`<div style="font-size:7px;opacity:0.85">📍 %s</div>`, html.EscapeString(it.Location)), ""),
			)
		}
		return fmt.Sprintf(`
<div class="label-unit">
  <div style="flex:0 0 26mm;display:flex;align-items:center;justify-content:center">
    <img src="%s" style="height:26mm;width:26mm;object-fit:contain" alt="%s"/>
  </div>
  <div style="flex:1;min-width:0;padding-left:2.5mm;display:flex;flex-direction:column;justify-content:center">
    <div style="font-size:7px;font-weight:800;text-transform:uppercase;letter-spacing:0.4px;border-bottom:1px solid #000;padding-bottom:0.5mm">%s</div>
    %s
    %s
    %s
  </div>
</div>`,
			codeImgURL, html.EscapeString(it.SKU),
			html.EscapeString(title),
			ternaryStr(showName, fmt.Sprintf(`<div style="font-size:8.5px;font-weight:700;line-height:1.15;margin-top:1mm;display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden">%s</div>`, html.EscapeString(it.Name)), ""),
			ternaryStr(showSKU, fmt.Sprintf(`<div style="font-family:monospace;font-size:8px;font-weight:700;margin-top:0.8mm">%s</div>`, html.EscapeString(it.SKU)), ""),
			ternaryStr(showLoc && it.Location != "", fmt.Sprintf(`<div style="font-size:7px;opacity:0.85;margin-top:0.5mm;white-space:nowrap;overflow:hidden;text-overflow:ellipsis">📍 %s</div>`, html.EscapeString(it.Location)), ""),
		)

	case "tj_107":
		// Tom & Jerry No. 107 (50x18 mm, 3 cols x 8 rows)
		return fmt.Sprintf(`
<div class="label-unit">
  <div style="flex:0 0 15mm;display:flex;align-items:center;justify-content:center">
    <img src="%s" style="height:15mm;width:15mm;object-fit:contain" alt="%s"/>
  </div>
  <div style="flex:1;min-width:0;padding-left:1.5mm;display:flex;flex-direction:column;justify-content:center">
    %s
    %s
    %s
  </div>
</div>`,
			codeImgURL, html.EscapeString(it.SKU),
			ternaryStr(showName, fmt.Sprintf(`<div style="font-size:7.5px;font-weight:700;line-height:1;white-space:nowrap;overflow:hidden;text-overflow:ellipsis">%s</div>`, html.EscapeString(it.Name)), ""),
			ternaryStr(showSKU, fmt.Sprintf(`<div style="font-family:monospace;font-size:6.5px;font-weight:700">%s</div>`, html.EscapeString(it.SKU)), ""),
			ternaryStr(showLoc && it.Location != "", fmt.Sprintf(`<div style="font-size:6px;opacity:0.8;white-space:nowrap;overflow:hidden;text-overflow:ellipsis">%s</div>`, html.EscapeString(it.Location)), ""),
		)

	case "tj_108":
		// Tom & Jerry No. 108 (38x18 mm, 5 cols x 8 rows)
		return fmt.Sprintf(`
<div class="label-unit">
  <div style="flex:0 0 14mm;display:flex;align-items:center;justify-content:center">
    <img src="%s" style="height:14mm;width:14mm;object-fit:contain" alt="%s"/>
  </div>
  <div style="flex:1;min-width:0;padding-left:1mm;display:flex;flex-direction:column;justify-content:center">
    %s
    %s
  </div>
</div>`,
			codeImgURL, html.EscapeString(it.SKU),
			ternaryStr(showName, fmt.Sprintf(`<div style="font-size:6.5px;font-weight:700;line-height:1;white-space:nowrap;overflow:hidden;text-overflow:ellipsis">%s</div>`, html.EscapeString(it.Name)), ""),
			ternaryStr(showSKU, fmt.Sprintf(`<div style="font-family:monospace;font-size:6px;font-weight:700">%s</div>`, html.EscapeString(it.SKU)), ""),
		)

	case "tj_121":
		// Tom & Jerry No. 121 (75x38 mm, 2 cols x 5 rows)
		return fmt.Sprintf(`
<div class="label-unit">
  <div style="flex:0 0 30mm;display:flex;align-items:center;justify-content:center">
    <img src="%s" style="height:30mm;width:30mm;object-fit:contain" alt="%s"/>
  </div>
  <div style="flex:1;min-width:0;padding-left:3mm;display:flex;flex-direction:column;justify-content:center">
    <div style="font-size:8px;font-weight:800;text-transform:uppercase;letter-spacing:0.5px;border-bottom:1px solid #000;padding-bottom:0.5mm">%s</div>
    %s
    %s
    %s
  </div>
</div>`,
			codeImgURL, html.EscapeString(it.SKU),
			html.EscapeString(title),
			ternaryStr(showName, fmt.Sprintf(`<div style="font-size:9.5px;font-weight:700;line-height:1.2;margin-top:1mm;display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden">%s</div>`, html.EscapeString(it.Name)), ""),
			ternaryStr(showSKU, fmt.Sprintf(`<div style="font-family:monospace;font-size:8.5px;font-weight:700;margin-top:0.8mm">%s</div>`, html.EscapeString(it.SKU)), ""),
			ternaryStr(showLoc && it.Location != "", fmt.Sprintf(`<div style="font-size:7.5px;opacity:0.85;margin-top:0.5mm;white-space:nowrap;overflow:hidden;text-overflow:ellipsis">📍 %s</div>`, html.EscapeString(it.Location)), ""),
		)

	case "thermal_58":
		imgHeight := "46px"
		if fmtType == "code128" {
			imgHeight = "34px"
		}
		return fmt.Sprintf(`
<div class="label-unit">
  <div style="font-size:9px;font-weight:700;letter-spacing:0.5px;text-transform:uppercase;margin-bottom:2px;border-bottom:1px solid #000;padding-bottom:1px">%s</div>
  <div style="margin:2px 0"><img src="%s" style="height:%s;max-width:100%%;object-fit:contain" alt="%s"/></div>
  %s
  %s
  %s
</div>`,
			html.EscapeString(title),
			codeImgURL,
			imgHeight,
			html.EscapeString(it.SKU),
			ternaryStr(showName, fmt.Sprintf(`<div style="font-size:10px;font-weight:700;line-height:1.2;margin-top:2px;word-break:break-word">%s</div>`, html.EscapeString(it.Name)), ""),
			ternaryStr(showSKU, fmt.Sprintf(`<div style="font-family:monospace;font-size:9px;font-weight:600;margin-top:1px">%s</div>`, html.EscapeString(it.SKU)), ""),
			ternaryStr(showLoc && it.Location != "", fmt.Sprintf(`<div style="font-size:8.5px;opacity:0.85;margin-top:1px">%s</div>`, html.EscapeString(it.Location)), ""),
		)

	case "thermal_80":
		imgHeight := "64px"
		if fmtType == "code128" {
			imgHeight = "44px"
		}
		return fmt.Sprintf(`
<div class="label-unit">
  <div style="font-size:10px;font-weight:800;letter-spacing:0.8px;text-transform:uppercase;margin-bottom:3px;border-bottom:1.5px solid #000;padding-bottom:2px">%s</div>
  <div style="margin:3px 0"><img src="%s" style="height:%s;max-width:100%%;object-fit:contain" alt="%s"/></div>
  %s
  %s
  %s
</div>`,
			html.EscapeString(title),
			codeImgURL,
			imgHeight,
			html.EscapeString(it.SKU),
			ternaryStr(showName, fmt.Sprintf(`<div style="font-size:11.5px;font-weight:700;line-height:1.25;margin-top:2px;word-break:break-word">%s</div>`, html.EscapeString(it.Name)), ""),
			ternaryStr(showSKU, fmt.Sprintf(`<div style="font-family:monospace;font-size:10px;font-weight:700;margin-top:2px">%s</div>`, html.EscapeString(it.SKU)), ""),
			ternaryStr(showLoc && it.Location != "", fmt.Sprintf(`<div style="font-size:9.5px;opacity:0.85;margin-top:1px">📍 %s</div>`, html.EscapeString(it.Location)), ""),
		)

	case "label_40x30":
		// Ultra-compact 40x30 sticker (Side by side)
		imgHeight := "24mm"
		imgWidth := "20mm"
		if fmtType == "code128" {
			return fmt.Sprintf(`
<div class="label-unit" style="flex-direction:column;align-items:center;justify-content:center;text-align:center">
  <div style="font-size:7px;font-weight:700;text-transform:uppercase">%s</div>
  <img src="%s" style="height:11mm;max-width:36mm;margin:1mm 0" alt="%s"/>
  <div style="font-size:7.5px;font-weight:700;line-height:1;max-width:38mm;white-space:nowrap;overflow:hidden;text-overflow:ellipsis">%s</div>
  <div style="font-family:monospace;font-size:7px;font-weight:600">%s</div>
</div>`,
				html.EscapeString(title), codeImgURL, html.EscapeString(it.SKU),
				html.EscapeString(it.Name), html.EscapeString(it.SKU))
		}
		return fmt.Sprintf(`
<div class="label-unit">
  <div style="flex:0 0 %s;display:flex;align-items:center;justify-content:center">
    <img src="%s" style="height:%s;width:%s;object-fit:contain" alt="%s"/>
  </div>
  <div style="flex:1;min-width:0;padding-left:1.5mm;display:flex;flex-direction:column;justify-content:center">
    <div style="font-size:6.5px;font-weight:700;text-transform:uppercase;letter-spacing:0.3px">%s</div>
    %s
    %s
    %s
  </div>
</div>`,
			imgWidth, codeImgURL, imgHeight, imgWidth, html.EscapeString(it.SKU),
			html.EscapeString(title),
			ternaryStr(showName, fmt.Sprintf(`<div style="font-size:7.5px;font-weight:700;line-height:1.1;margin-top:0.5mm;display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden">%s</div>`, html.EscapeString(it.Name)), ""),
			ternaryStr(showSKU, fmt.Sprintf(`<div style="font-family:monospace;font-size:7px;font-weight:600;margin-top:0.5mm">%s</div>`, html.EscapeString(it.SKU)), ""),
			ternaryStr(showLoc && it.Location != "", fmt.Sprintf(`<div style="font-size:6.5px;opacity:0.8;margin-top:0.3mm;white-space:nowrap;overflow:hidden;text-overflow:ellipsis">%s</div>`, html.EscapeString(it.Location)), ""),
		)

	case "label_50x30":
		// Standard 50x30 Barcode Sticker
		if fmtType == "code128" {
			return fmt.Sprintf(`
<div class="label-unit" style="flex-direction:column;align-items:center;justify-content:center;text-align:center">
  <div style="font-size:7.5px;font-weight:700;text-transform:uppercase;letter-spacing:0.4px">%s</div>
  <img src="%s" style="height:12mm;max-width:44mm;margin:1mm 0" alt="%s"/>
  <div style="font-size:8px;font-weight:700;line-height:1.1;max-width:46mm;white-space:nowrap;overflow:hidden;text-overflow:ellipsis">%s</div>
  <div style="font-family:monospace;font-size:7.5px;font-weight:700">%s %s</div>
</div>`,
				html.EscapeString(title), codeImgURL, html.EscapeString(it.SKU),
				html.EscapeString(it.Name), html.EscapeString(it.SKU),
				ternaryStr(showLoc && it.Location != "", "· "+html.EscapeString(it.Location), ""))
		}
		return fmt.Sprintf(`
<div class="label-unit">
  <div style="flex:0 0 23mm;display:flex;align-items:center;justify-content:center">
    <img src="%s" style="height:24mm;width:23mm;object-fit:contain" alt="%s"/>
  </div>
  <div style="flex:1;min-width:0;padding-left:2mm;display:flex;flex-direction:column;justify-content:center">
    <div style="font-size:7px;font-weight:800;text-transform:uppercase;letter-spacing:0.5px;border-bottom:1px solid #000;padding-bottom:0.5mm">%s</div>
    %s
    %s
    %s
  </div>
</div>`,
			codeImgURL, html.EscapeString(it.SKU),
			html.EscapeString(title),
			ternaryStr(showName, fmt.Sprintf(`<div style="font-size:8.5px;font-weight:700;line-height:1.15;margin-top:1mm;display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden">%s</div>`, html.EscapeString(it.Name)), ""),
			ternaryStr(showSKU, fmt.Sprintf(`<div style="font-family:monospace;font-size:8px;font-weight:700;margin-top:0.8mm">%s</div>`, html.EscapeString(it.SKU)), ""),
			ternaryStr(showLoc && it.Location != "", fmt.Sprintf(`<div style="font-size:7px;opacity:0.85;margin-top:0.5mm;white-space:nowrap;overflow:hidden;text-overflow:ellipsis">📍 %s</div>`, html.EscapeString(it.Location)), ""),
		)

	case "label_60x40", "label_70x50", "label_100x50":
		// Medium and large stickers
		qrSize := "30mm"
		if layout == "label_70x50" {
			qrSize = "38mm"
		} else if layout == "label_100x50" {
			qrSize = "42mm"
		}
		if fmtType == "code128" {
			return fmt.Sprintf(`
<div class="label-unit" style="flex-direction:column;align-items:center;justify-content:center;text-align:center">
  <div style="font-size:9px;font-weight:800;text-transform:uppercase;letter-spacing:0.8px;margin-bottom:1.5mm">%s</div>
  <img src="%s" style="height:17mm;max-width:85%%;margin:1mm 0" alt="%s"/>
  <div style="font-size:10px;font-weight:700;line-height:1.2;margin-top:1.5mm">%s</div>
  <div style="font-family:monospace;font-size:9.5px;font-weight:700;margin-top:1mm">%s</div>
  %s
</div>`,
				html.EscapeString(title), codeImgURL, html.EscapeString(it.SKU),
				html.EscapeString(it.Name), html.EscapeString(it.SKU),
				ternaryStr(showLoc && it.Location != "", fmt.Sprintf(`<div style="font-size:8.5px;margin-top:0.5mm">📍 %s</div>`, html.EscapeString(it.Location)), ""))
		}
		return fmt.Sprintf(`
<div class="label-unit">
  <div style="flex:0 0 %s;display:flex;align-items:center;justify-content:center">
    <img src="%s" style="height:%s;width:%s;object-fit:contain" alt="%s"/>
  </div>
  <div style="flex:1;min-width:0;padding-left:3mm;display:flex;flex-direction:column;justify-content:center">
    <div style="font-size:8px;font-weight:800;text-transform:uppercase;letter-spacing:0.6px;border-bottom:1px solid #000;padding-bottom:1mm">%s</div>
    %s
    %s
    %s
  </div>
</div>`,
			qrSize, codeImgURL, qrSize, qrSize, html.EscapeString(it.SKU),
			html.EscapeString(title),
			ternaryStr(showName, fmt.Sprintf(`<div style="font-size:10px;font-weight:700;line-height:1.2;margin-top:1.5mm">%s</div>`, html.EscapeString(it.Name)), ""),
			ternaryStr(showSKU, fmt.Sprintf(`<div style="font-family:monospace;font-size:9px;font-weight:700;margin-top:1mm">%s</div>`, html.EscapeString(it.SKU)), ""),
			ternaryStr(showLoc && it.Location != "", fmt.Sprintf(`<div style="font-size:8px;opacity:0.85;margin-top:0.8mm">📍 %s</div>`, html.EscapeString(it.Location)), ""),
		)

	default:
		// A4 grids (a4_3x7, a4_4x10)
		qrSize := "28mm"
		fontSizeName := "9.5px"
		if layout == "a4_4x10" {
			qrSize = "19mm"
			fontSizeName = "8px"
		}
		return fmt.Sprintf(`
<div class="label-unit">
  <div style="flex:0 0 %s;display:flex;align-items:center;justify-content:center">
    <img src="%s" style="height:%s;width:%s;object-fit:contain" alt="%s"/>
  </div>
  <div style="flex:1;min-width:0;padding-left:2mm;display:flex;flex-direction:column;justify-content:center">
    <div style="font-size:7px;font-weight:800;text-transform:uppercase;letter-spacing:0.4px">%s</div>
    %s
    %s
    %s
  </div>
</div>`,
			qrSize, codeImgURL, qrSize, qrSize, html.EscapeString(it.SKU),
			html.EscapeString(title),
			ternaryStr(showName, fmt.Sprintf(`<div style="font-size:%s;font-weight:700;line-height:1.15;margin-top:0.5mm;display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden">%s</div>`, fontSizeName, html.EscapeString(it.Name)), ""),
			ternaryStr(showSKU, fmt.Sprintf(`<div style="font-family:monospace;font-size:7.5px;font-weight:700;margin-top:0.5mm">%s</div>`, html.EscapeString(it.SKU)), ""),
			ternaryStr(showLoc && it.Location != "", fmt.Sprintf(`<div style="font-size:7px;opacity:0.85;margin-top:0.5mm;white-space:nowrap;overflow:hidden;text-overflow:ellipsis">📍 %s</div>`, html.EscapeString(it.Location)), ""),
		)
	}
}

func ternaryStr(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}

func splitComma(s string) []string {
	return strings.Split(s, ",")
}

func renderQR(content string) []byte {
	code, err := qr.Encode(content, qr.M, qr.Auto)
	if err != nil {
		return renderCode128(content)
	}
	scaled, err := barcode.Scale(code, 200, 200)
	if err != nil {
		return []byte{}
	}
	var buf bytes.Buffer
	png.Encode(&buf, scaled)
	return buf.Bytes()
}

func renderCode128(content string) []byte {
	code, err := code128.Encode(content)
	if err != nil {
		return []byte{}
	}
	scaled, err := barcode.Scale(code, 300, 80)
	if err != nil {
		return []byte{}
	}
	var buf bytes.Buffer
	png.Encode(&buf, scaled)
	return buf.Bytes()
}