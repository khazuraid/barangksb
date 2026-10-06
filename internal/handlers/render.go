package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/jackc/pgx/v5/pgtype"

	"inventariskantor/internal/auth"
	"inventariskantor/internal/models"
	"inventariskantor/internal/service"
	"inventariskantor/internal/views"
)

type modelsTX = models.StockTransaction

func tstr(t pgtype.Text) string {
	if t.Valid && t.String != "" {
		return t.String
	}
	return "—"
}

// show renders a templ component inside the layout.
func (h *Handlers) show(w http.ResponseWriter, r *http.Request, title string, content templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	info := views.UserInfo{}
	if u := auth.FromContext(r.Context()); u != nil {
		info = views.UserInfo{Name: u.Name, Role: u.Role}
	}
	if err := views.Layout(title, info, content).Render(r.Context(), w); err != nil {
		slog.Error("render failed", "page", title, "err", err)
	}
}

// render renders public pages (login) with layout.
func (h *Handlers) render(w http.ResponseWriter, r *http.Request, page string, data any) {
	if page == "login" {
		var err string
		if d, ok := data.(map[string]any); ok {
			err, _ = d["Error"].(string)
		}
		h.show(w, r, "Masuk", views.LoginContent(err))
		return
	}
	http.Error(w, "unknown page", http.StatusInternalServerError)
}

// itemInputFromForm maps a POST form to service.ItemInput.
func itemInputFromForm(form formValues, photoURL string) service.ItemInput {
	return service.ItemInput{
		SKU:             form.Get("sku"),
		Name:            form.Get("name"),
		Category:        form.Get("category"),
		Location:        form.Get("location"),
		Unit:            form.Get("unit"),
		CurrentStock:    atoi0(form.Get("current_stock")),
		MinStock:        atoi0(form.Get("min_stock")),
		PricePerUnit:    atol(form.Get("price_per_unit")),
		Description:     form.Get("description"),
		PhotoURL:        photoURL,
		Merk:            form.Get("merk"),
		TypeModel:       form.Get("type_model"),
		SerialNumber:    form.Get("serial_number"),
		ProcurementYear: form.Get("procurement_year"),
		ConditionStatus: form.Get("condition_status"),
		FundingSource:   form.Get("funding_source"),
		Distributor:     form.Get("distributor"),
		AklAkd:          form.Get("akl_akd"),
	}
}

// pageParam parses ?page=N
func pageParam(r *http.Request) int {
	p, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if p < 1 {
		p = 1
	}
	return p
}

// buildPager constructs the Pager for a list page, preserving other query params.
func buildPager(r *http.Request, page, perPage, total int) views.Pager {
	q := r.URL.Query()
	q.Del("page")
	base := r.URL.Path
	if enc := q.Encode(); enc != "" {
		base += "?" + enc
	}
	return views.Pager{Page: page, PerPage: perPage, Total: total, BaseURL: base, MaxShow: 7}
}

// itemToRow / itemRows mapping helpers.
func toItemRow(it *models.InventoryItem) views.ItemRow {
	return views.ItemRow{
		ID:           fmtUUID(it.ID),
		SKU:          it.Sku,
		Name:         it.Name,
		Category:     it.Category,
		Location:     it.Location,
		CurrentStock: int(it.CurrentStock),
		MinStock:     int(it.MinStock),
		Unit:         it.Unit,
		Condition:    optStrOr(it.ConditionStatus, "Berfungsi"),
	}
}

func itemRows(items []models.InventoryItem) []views.ItemRow {
	rows := make([]views.ItemRow, 0, len(items))
	for i := range items {
		rows = append(rows, toItemRow(&items[i]))
	}
	return rows
}

var _ = context.Background
var _ = strings.TrimSpace
