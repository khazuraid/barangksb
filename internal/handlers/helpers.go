package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"inventariskantor/internal/auth"
	"inventariskantor/internal/models"
	"inventariskantor/internal/service"
	"inventariskantor/internal/views"
)

func timeNow() time.Time { return time.Now() }

// ---------- URL/param helpers ----------

func chiURLParam(r *http.Request, key string) string { return chi.URLParam(r, key) }

// ---------- parse helpers ----------

func atoi0(s string) int32 {
	n, _ := strconv.Atoi(s)
	if n < 0 {
		return 0
	}
	return int32(n)
}

func atol(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	if n < 0 {
		return 0
	}
	return n
}

func atoi64(s string) (int64, error) { return strconv.ParseInt(strings.TrimSpace(s), 10, 64) }

func atof(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func nullStr(s string) pgtype.Text {
	if strings.TrimSpace(s) == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func nullInt8(s string) pgtype.Int8 {
	if strings.TrimSpace(s) == "" {
		return pgtype.Int8{}
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: n, Valid: true}
}

func optStr(t pgtype.Text) string {
	if t.Valid {
		return t.String
	}
	return ""
}

func optStrOr(t pgtype.Text, def string) string {
	if t.Valid && t.String != "" {
		return t.String
	}
	return def
}

func tstr(t pgtype.Text) string {
	if t.Valid && t.String != "" {
		return t.String
	}
	return "—"
}

func fmtUUID(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return uuid.UUID(u.Bytes).String()
}

func containsStr(s, sub string) bool { return strings.Contains(s, sub) }

// ---------- auth/context helpers ----------

func authUserOf(r *http.Request) *auth.User { return auth.FromContext(r.Context()) }

func authUserName(r *http.Request) string {
	if u := authUserOf(r); u != nil {
		return u.Name
	}
	return ""
}

func userInfo(r *http.Request) views.UserInfo {
	info := views.UserInfo{}
	if u := authUserOf(r); u != nil {
		info = views.UserInfo{Name: u.Name, Role: u.Role}
	}
	return info
}

func requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	u := authUserOf(r)
	if u == nil || u.Role != "admin" {
		http.Error(w, "403 — hanya admin", http.StatusForbidden)
		return false
	}
	return true
}

// ---------- geo ----------

type geoData struct {
	lat, lng, acc pgtype.Float8
	at            pgtype.Timestamptz
	name          pgtype.Text
}

func geoFromForm(f formValues) geoData {
	lat := atof(f.Get("geo_lat"))
	lng := atof(f.Get("geo_lng"))
	acc := atof(f.Get("geo_acc"))
	if lat == 0 && lng == 0 {
		return geoData{}
	}
	g := geoData{
		lat: pgtype.Float8{Float64: lat, Valid: true},
		lng: pgtype.Float8{Float64: lng, Valid: true},
		acc: pgtype.Float8{Float64: acc, Valid: acc > 0},
		at:  pgtype.Timestamptz{Time: timeNow(), Valid: true},
	}
	if n := f.Get("geo_name"); n != "" {
		g.name = pgtype.Text{String: n, Valid: true}
	}
	return g
}

// ---------- form & formatting ----------

type formValues interface{ Get(key string) string }

func parseForm(w http.ResponseWriter, r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		if err := r.ParseMultipartForm(6 << 20); err != nil {
			http.Error(w, "form terlalu besar (max 6MB)", http.StatusRequestEntityTooLarge)
			return false
		}
		return true
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form tidak valid", http.StatusBadRequest)
		return false
	}
	return true
}

func fmtWIB(t pgtype.Timestamptz) string {
	if !t.Valid {
		return "—"
	}
	return t.Time.In(wibLoc()).Format("02-01-2006 15:04")
}

func geoDisplay(lat, lng pgtype.Float8) string {
	if !lat.Valid || !lng.Valid {
		return "—"
	}
	return fmt.Sprintf("%.6f, %.6f", lat.Float64, lng.Float64)
}

func wibLoc() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.FixedZone("WIB", 7*3600)
	}
	return loc
}

// ---------- render & mapping ----------

// show renders a templ component inside the layout.
func (h *Handlers) show(w http.ResponseWriter, r *http.Request, title string, content templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	info := userInfo(r)
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

func toFormValues(it *models.InventoryItem) *views.ItemFormValues {
	return &views.ItemFormValues{
		ID:              fmtUUID(it.ID),
		SKU:             it.Sku,
		Name:            it.Name,
		Category:        it.Category,
		Location:        it.Location,
		CurrentStock:    int(it.CurrentStock),
		MinStock:        int(it.MinStock),
		Unit:            it.Unit,
		PricePerUnit:    it.PricePerUnit.Int64,
		Description:     optStr(it.Description),
		Merk:            optStr(it.Merk),
		TypeModel:       optStr(it.TypeModel),
		SerialNumber:    optStr(it.SerialNumber),
		ProcurementYear: optStr(it.ProcurementYear),
		ConditionStatus: optStrOr(it.ConditionStatus, "Berfungsi"),
		FundingSource:   optStr(it.FundingSource),
		Distributor:     optStr(it.Distributor),
		AklAkd:          optStr(it.AklAkd),
	}
}

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

// buildPager constructs the Pager preserving other query params.
func buildPager(r *http.Request, page, perPage, total int) views.Pager {
	q := r.URL.Query()
	q.Del("page")
	base := r.URL.Path
	if enc := q.Encode(); enc != "" {
		base += "?" + enc
	}
	return views.Pager{Page: page, PerPage: perPage, Total: total, BaseURL: base, MaxShow: 7}
}

func isDuplicateErr(err error) bool {
	return containsStr(err.Error(), "duplicate key") || containsStr(err.Error(), "unique")
}
