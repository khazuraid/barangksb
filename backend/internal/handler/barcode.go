package handler

import (
	"bytes"
	"fmt"
	"image/png"
	"log/slog"
	"net/http"
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

// GET /api/barcode/sheet?ids=a,b,c — public, printable HTML label sheet
func (h *BarcodeHandler) Sheet(c *gin.Context) {
	ids := c.Query("ids")
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, `<!DOCTYPE html><html><body onload="window.print()">
	<div style="display:grid;grid-template-columns:repeat(3,1fr);gap:8mm;padding:12mm">
	`+h.generateLabels(c, ids)+`
	</div></body></html>`)
}

func (h *BarcodeHandler) generateLabels(c *gin.Context, ids string) string {
	html := ""
	for _, id := range splitComma(ids) {
		var name, sku string
		h.pool.QueryRow(c, `SELECT name, sku FROM inventory_items WHERE id::text=$1`, id).Scan(&name, &sku)
		html += fmt.Sprintf(`<div style="text-align:center;border:1px dashed #ccc;padding:6px">
			<div style="font-size:11px;font-weight:700">%s</div>
			<img src="/api/barcode/%s.png?fmt=qr" height="60"/>
			<div style="font-size:10px;font-family:monospace">%s</div>
		</div>`, name, id, sku)
	}
	return html
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