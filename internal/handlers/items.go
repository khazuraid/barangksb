package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"inventariskantor/internal/auth"
	"inventariskantor/internal/models"
	"inventariskantor/internal/views"
)

func (h *Handlers) ItemsList(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	cat := r.URL.Query().Get("cat")

	items, err := h.listItems(r.Context(), q, cat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cats, _ := h.listCats(r.Context())
	rows := make([]views.ItemRow, 0, len(items))
	for i := range items {
		rows = append(rows, toItemRow(&items[i]))
	}
	h.show(w, r, "Barang", views.Items(views.ItemsData{
		User:  userInfo(r), Items: rows, Query: q, Cat: cat, Cats: cats,
	}))
}

func (h *Handlers) ItemNewForm(w http.ResponseWriter, r *http.Request) {
	cats, _ := h.listCats(r.Context())
	h.show(w, r, "Tambah Barang", views.ItemForm(views.ItemFormData{User: userInfo(r), Cats: cats}))
}

func (h *Handlers) ItemEditForm(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chiURLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	it, err := h.getItem(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	cats, _ := h.listCats(r.Context())
	h.show(w, r, "Edit Barang", views.ItemForm(views.ItemFormData{
		User: userInfo(r), Item: toFormValues(&it), Cats: cats,
	}))
}

func (h *Handlers) ItemsCreate(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r) {
		return
	}
	// SKU opsional: jika kosong, generate otomatis CAT-YYYY-XXX
	if strings.TrimSpace(r.PostFormValue("sku")) == "" {
		r.PostForm.Set("sku", "AUTO") // placeholder, digenerate di insertItem setelah tahu id
	}
	photoURL := h.processUpload(r, "")
	it, err := h.insertItem(r.Context(), r.PostForm, photoURL)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
			http.Error(w, "SKU sudah dipakai barang lain", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.broker.Publish("items")
	http.Redirect(w, r, "/items?created="+it.Sku, http.StatusSeeOther)
}

func (h *Handlers) ItemsUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chiURLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if !parseForm(w, r) {
		return
	}
	old, err := h.getItem(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	photoURL := h.processUpload(r, optStr(old.PhotoUrl))
	if err := h.updateItem(r.Context(), id, r.PostForm, photoURL); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.broker.Publish("items")
	http.Redirect(w, r, "/items?updated=1", http.StatusSeeOther)
}

func (h *Handlers) ItemDelete(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	if u == nil || u.Role != "admin" {
		http.Error(w, "403 — hanya admin boleh menghapus barang", http.StatusForbidden)
		return
	}
	id, err := uuid.Parse(chiURLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if _, err := h.pool.Exec(r.Context(), `DELETE FROM inventory_items WHERE id=$1`, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.broker.Publish("items")
	http.Redirect(w, r, "/items?deleted=1", http.StatusSeeOther)
}

func (h *Handlers) ItemDetail(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/items", http.StatusSeeOther)
}

// --- mappers ---

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

// --- queries ---

func (h *Handlers) listItems(ctx context.Context, q, cat string) ([]models.InventoryItem, error) {
	sql := `SELECT * FROM inventory_items WHERE TRUE`
	args := []any{}
	if q != "" {
		n := len(args) + 1
		sql += fmt.Sprintf(` AND (name ILIKE $%d OR sku ILIKE $%d OR location ILIKE $%d)`, n, n, n)
		args = append(args, "%"+q+"%")
	}
	if cat != "" {
		n := len(args) + 1
		sql += fmt.Sprintf(` AND category = $%d`, n)
		args = append(args, cat)
	}
	sql += ` ORDER BY name`
	rows, err := h.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[models.InventoryItem])
}

func (h *Handlers) getItem(ctx context.Context, id uuid.UUID) (models.InventoryItem, error) {
	rows, err := h.pool.Query(ctx, `SELECT * FROM inventory_items WHERE id=$1`, id)
	if err != nil {
		return models.InventoryItem{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByPos[models.InventoryItem])
}

func (h *Handlers) listCats(ctx context.Context) ([]string, error) {
	rows, err := h.pool.Query(ctx, `SELECT name FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[string])
}

func (h *Handlers) insertItem(ctx context.Context, form formValues, photoURL string) (models.InventoryItem, error) {
	// SKU otomatis: jika kosong/"AUTO", generate <KATEGORI>-<TAHUN>-<3digit seri>
	sku := strings.TrimSpace(form.Get("sku"))
	if sku == "" || sku == "AUTO" {
		generated, err := h.nextSKU(ctx, form.Get("category"))
		if err != nil {
			return models.InventoryItem{}, err
		}
		sku = generated
	}
	sql := `INSERT INTO inventory_items
		(sku, name, category, location, current_stock, min_stock, unit, price_per_unit,
		 description, photo_url, merk, type_model, serial_number, procurement_year,
		 condition_status, funding_source, distributor, akl_akd, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18, now())
		RETURNING *`
	rows, err := h.pool.Query(ctx, sql,
		sku, form.Get("name"), form.Get("category"), form.Get("location"),
		atoi0(form.Get("current_stock")), atoi0(form.Get("min_stock")), form.Get("unit"),
		nullInt8(form.Get("price_per_unit")),
		nullStr(form.Get("description")), nullStr(photoURL),
		nullStr(form.Get("merk")), nullStr(form.Get("type_model")), nullStr(form.Get("serial_number")),
		nullStr(form.Get("procurement_year")), nullStr(form.Get("condition_status")),
		nullStr(form.Get("funding_source")), nullStr(form.Get("distributor")), nullStr(form.Get("akl_akd")))
	if err != nil {
		return models.InventoryItem{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByPos[models.InventoryItem])
}

// nextSKU generates <KODEKATEGORI>-<TAHUN>-<3digit> berurutan per kategori per tahun.
// Kode kategori diambil dari 2-4 huruf awal slug kategori.
func (h *Handlers) nextSKU(ctx context.Context, category string) (string, error) {
	var slug string
	err := h.pool.QueryRow(ctx, `SELECT id FROM categories WHERE name=$1`, category).Scan(&slug)
	if err != nil {
		slug = slugify(category)
	}
	parts := strings.Split(strings.Trim(slug, "-"), "-")
	prefix := strings.ToUpper(parts[0])
	if len(prefix) < 2 {
		prefix = strings.ToUpper(slugify(category))[:3]
	}
	prefix = strings.Trim(prefix, "-")[:min2(len(prefix), 4)]

	year := time.Now().Format("2006")
	pattern := prefix + "-" + year + "-%"
	var last string
	err = h.pool.QueryRow(ctx,
		`SELECT sku FROM inventory_items WHERE sku LIKE $1 ORDER BY sku DESC LIMIT 1`, pattern).Scan(&last)
	serial := 1
	if err == nil && len(last) > len(prefix)+5 {
		if n, e := strconv.Atoi(last[len(prefix)+5:]); e == nil {
			serial = n + 1
		}
	}
	return fmt.Sprintf("%s-%s-%03d", prefix, year, serial), nil
}

func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (h *Handlers) updateItem(ctx context.Context, id uuid.UUID, form formValues, photoURL string) error {
	sql := `UPDATE inventory_items SET
		sku=$2, name=$3, category=$4, location=$5, current_stock=$6, min_stock=$7, unit=$8,
		price_per_unit=$9, description=$10, photo_url=$11, merk=$12, type_model=$13,
		serial_number=$14, procurement_year=$15, condition_status=$16, funding_source=$17,
		distributor=$18, akl_akd=$19, updated_at=now()
		WHERE id=$1`
	_, err := h.pool.Exec(ctx, sql,
		id, form.Get("sku"), form.Get("name"), form.Get("category"), form.Get("location"),
		atoi0(form.Get("current_stock")), atoi0(form.Get("min_stock")), form.Get("unit"),
		nullInt8(form.Get("price_per_unit")),
		nullStr(form.Get("description")), nullStr(photoURL),
		nullStr(form.Get("merk")), nullStr(form.Get("type_model")), nullStr(form.Get("serial_number")),
		nullStr(form.Get("procurement_year")), nullStr(form.Get("condition_status")),
		nullStr(form.Get("funding_source")), nullStr(form.Get("distributor")), nullStr(form.Get("akl_akd")))
	return err
}

var _ = pgtype.Text{}
var _ = fmt.Sprintf
