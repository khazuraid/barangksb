package handlers

import (
	"bytes"
	"fmt"
	"image/png"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/ean"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"

	"inventariskantor/internal/service"
	"inventariskantor/internal/views"
)

func (h *Handlers) BarcodePage(w http.ResponseWriter, r *http.Request) {
	f := service.ItemFilter{Page: pageParam(r), PerPage: 24}
	items, total, err := h.svc.ListItems(r.Context(), f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.show(w, r, "QR Code", views.BarcodePage(views.BarcodeData{
		User:  userInfo(r),
		Items: itemRows(items),
		Pager: buildPager(r, f.Page, f.PerPage, total),
	}))
}

// BarcodeSheet renders a printable label sheet: GET /barcode/sheet?ids=a,b,c
func (h *Handlers) BarcodeSheet(w http.ResponseWriter, r *http.Request) {
	ids := strings.Split(r.URL.Query().Get("ids"), ",")
	var rows []views.ItemRow
	for _, idStr := range ids {
		it, err := h.svc.GetItem(r.Context(), strings.TrimSpace(idStr))
		if err != nil {
			continue
		}
		rows = append(rows, toItemRow(&it))
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = views.BarcodeSheet(rows).Render(r.Context(), w)
}

// BarcodePNG renders one barcode: GET /barcode/{itemID}.png?fmt=qr|code128|ean13
func (h *Handlers) BarcodePNG(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "itemID"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	it, err := h.svc.GetItem(r.Context(), id.String())
	if err != nil {
		http.NotFound(w, r)
		return
	}
	format := r.URL.Query().Get("fmt")
	if format == "" {
		format = "QR"
	}

	content := it.Sku
	if strings.EqualFold(format, "QR") {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		if fwd := r.Header.Get("X-Forwarded-Proto"); fwd != "" {
			scheme = fwd
		}
		content = scheme + "://" + r.Host + "/scan/" + it.ID.String()
	}

	pngBytes, err := RenderBarcode(content, format)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(pngBytes)
}

// RenderBarcode encodes content into PNG bytes for the requested format.
func RenderBarcode(content, format string) ([]byte, error) {
	switch strings.ToUpper(format) {
	case "QR":
		q, err := qrcode.New(content, qrcode.Medium)
		if err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, q.Image(200)); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	case "EAN13":
		if len(content) != 13 {
			return nil, fmt.Errorf("EAN13 butuh 13 digit")
		}
		code, err := ean.Encode(content)
		if err != nil {
			return nil, fmt.Errorf("SKU tidak valid untuk EAN13: %w", err)
		}
		return scalePNG(code, 300, 80)
	default: // CODE128
		code, err := code128.Encode(content)
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

var _ = templ.URL
