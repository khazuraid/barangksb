package handler

import (
	"bytes"
	"fmt"
	"image/png"
	"net/http"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BarcodeHandler struct {
	pool *pgxpool.Pool
}

func NewBarcodeHandler(pool *pgxpool.Pool) *BarcodeHandler {
	return &BarcodeHandler{pool: pool}
}

// GET /api/barcode/:id.png?fmt=qr
func (h *BarcodeHandler) PNG(c *gin.Context) {
	id := c.Param("id")
	format := c.DefaultQuery("fmt", "qr")

	var sku string
	err := h.pool.QueryRow(c, `SELECT sku FROM inventory_items WHERE id::text=$1`, id).Scan(&sku)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	content := sku
	if format == "qr" || format == "" {
		// QR berisi URL scan publik
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

func renderQR(content string) []byte {
	// simple QR via skip2/go-qrcode replacement — use code128 fallback for now
	// TODO: add go-qrcode when module resolves
	return renderCode128(content)
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

// GET /api/barcode/sheet?ids=a,b,c
func (h *BarcodeHandler) Sheet(c *gin.Context) {
	// Return simple HTML printable page
	ids := c.Query("ids")
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, `<!DOCTYPE html><html><body onload="window.print()">
	<div style="display:grid;grid-template-columns:repeat(3,1fr);gap:8mm;padding:12mm">
	`+generateLabels(h, c, ids)+`
	</div></body></html>`)
}

func generateLabels(h *BarcodeHandler, c *gin.Context, ids string) string {
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
	var out []string
	start := 0
	for i, c := range s {
		if c == ',' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}
