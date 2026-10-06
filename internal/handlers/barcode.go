package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"image/png"
	"io"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/ean"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"

	"inventariskantor/internal/views"
)

func (h *Handlers) BarcodePage(w http.ResponseWriter, r *http.Request) {
	items, err := h.listItems(r.Context(), "", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows := make([]views.ItemRow, 0, len(items))
	for i := range items {
		rows = append(rows, toItemRow(&items[i]))
	}
	h.show(w, r, "Barcode", views.BarcodePage(views.BarcodeData{User: userInfo(r), Items: rows}))
}

// BarcodeSheet renders a printable label sheet: GET /barcode/sheet?ids=a,b,c
func (h *Handlers) BarcodeSheet(w http.ResponseWriter, r *http.Request) {
	ids := strings.Split(r.URL.Query().Get("ids"), ",")
	var rows []views.ItemRow
	for _, idStr := range ids {
		id, err := uuid.Parse(strings.TrimSpace(idStr))
		if err != nil {
			continue
		}
		it, err := h.getItem(r.Context(), id)
		if err != nil {
			continue
		}
		rows = append(rows, toItemRow(&it))
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = views.BarcodeSheet(rows).Render(r.Context(), w)
}

// BarcodePNG renders one barcode: GET /barcode/{itemID}.png?fmt=code128|ean13|qr
func (h *Handlers) BarcodePNG(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "itemID"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	it, err := h.getItem(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	format := r.URL.Query().Get("fmt")
	if format == "" {
		format = optStrOr(it.BarcodeFormat, "CODE128")
	}
	pngBytes, err := RenderBarcode(it.Sku, format)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(pngBytes)
}

// RenderBarcode encodes SKU into PNG bytes for the requested format.
func RenderBarcode(sku, format string) ([]byte, error) {
	switch strings.ToUpper(format) {
	case "QR":
		q, err := qrcode.New(sku, qrcode.Medium)
		if err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, q.Image(200)); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	case "EAN13":
		if len(sku) != 13 {
			return nil, fmt.Errorf("EAN13 butuh 13 digit, SKU %q punya %d", sku, len(sku))
		}
		code, err := ean.Encode(sku)
		if err != nil {
			return nil, fmt.Errorf("SKU tidak valid untuk EAN13: %w", err)
		}
		return scalePNG(code, 300, 80)
	default: // CODE128
		code, err := code128.Encode(sku)
		if err != nil {
			return nil, err
		}
		return scalePNG(code, 300, 80)
	}
}

func scalePNG(bc barcode.Barcode, w, h int) ([]byte, error) {
	scaled, err := barcode.Scale(bc, w, h)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, scaled); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var (
	_ = templ.URL
	_ = io.EOF
	_ = errors.New
)
