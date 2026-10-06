package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"inventariskantor/internal/auth"
	"inventariskantor/internal/views"
)

type geoData struct {
	lat, lng, acc pgtype.Float8
	at            pgtype.Timestamptz
	name          pgtype.Text
}

func userInfo(r *http.Request) views.UserInfo {
	info := views.UserInfo{}
	if u := auth.FromContext(r.Context()); u != nil {
		info = views.UserInfo{Name: u.Name, Role: u.Role}
	}
	return info
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
		at:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	if n := f.Get("geo_name"); n != "" {
		g.name = pgtype.Text{String: n, Valid: true}
	}
	return g
}

func (h *Handlers) MovementForm(w http.ResponseWriter, r *http.Request) {
	cats, _ := h.listCats(r.Context())
	items, err := h.listItems(r.Context(), "", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows := make([]views.ItemRow, 0, len(items))
	for i := range items {
		rows = append(rows, toItemRow(&items[i]))
	}
	h.show(w, r, "Barang Masuk", views.Movement(views.MovementData{
		User: userInfo(r), Cats: cats,
		SKU: r.URL.Query().Get("sku"), Items: rows,
	}))
}

// MovementPost records a stock-in inside ONE db transaction: insert tx + bump stock.
func (h *Handlers) MovementPost(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r) {
		return
	}
	form := formValues(r.PostForm)
	itemID, err := parseUUID(form.Get("item_id"))
	if err != nil {
		http.Error(w, "item tidak valid", http.StatusBadRequest)
		return
	}
	qty := atoi0(form.Get("quantity"))
	if qty <= 0 {
		http.Error(w, "jumlah harus > 0", http.StatusBadRequest)
		return
	}
	geo := geoFromForm(form)
	photoURL := h.processUpload(r, "")

	ctx := r.Context()
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	var sku, name, unit string
	var prev, next int32
	err = tx.QueryRow(ctx,
		`UPDATE inventory_items SET current_stock = current_stock + $2, updated_at = now()
		 WHERE id = $1 RETURNING sku, name, unit, current_stock - $2, current_stock`,
		itemID, qty).Scan(&sku, &name, &unit, &prev, &next)
	if err != nil {
		http.Error(w, "barang tidak ditemukan: "+err.Error(), http.StatusBadRequest)
		return
	}

	_, err = tx.Exec(ctx, `INSERT INTO stock_transactions
		(type, item_id, item_sku, item_name, quantity, unit, previous_stock, new_stock,
		 supplier_or_source, received_by, invoice_or_po_number, photo_url,
		 geo_lat, geo_lng, geo_acc, geo_at, geo_name,
		 merk, type_model, serial_number, procurement_year, condition_status,
		 funding_source, distributor, akl_akd, notes)
		VALUES ('IN',$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25)`,
		itemID, sku, name, qty, unit, prev, next,
		nullStr(form.Get("supplier_or_source")), nullStr(form.Get("received_by")),
		nullStr(form.Get("invoice_or_po_number")), nullStr(photoURL),
		geo.lat, geo.lng, geo.acc, geo.at, geo.name,
		nullStr(form.Get("merk")), nullStr(form.Get("type_model")),
		nullStr(form.Get("serial_number")), nullStr(form.Get("procurement_year")),
		nullStr(form.Get("condition_status")), nullStr(form.Get("funding_source")),
		nullStr(form.Get("distributor")), nullStr(form.Get("akl_akd")),
		nullStr(form.Get("notes")))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.broker.Publish("tx")
	h.broker.Publish("items")
	http.Redirect(w, r, "/history?ok="+sku, http.StatusSeeOther)
}

func (h *Handlers) History(w http.ResponseWriter, r *http.Request) {
	rows, err := h.pool.Query(r.Context(), `SELECT * FROM stock_transactions ORDER BY timestamp DESC LIMIT 500`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	txs, err := pgx.CollectRows(rows, pgx.RowToStructByPos[modelsTX])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	vrows := make([]views.TXRow, 0, len(txs))
	for _, t := range txs {
		vrows = append(vrows, views.TXRow{
			Time:       fmtWIB(t.Timestamp),
			SKU:        t.ItemSku,
			Name:       t.ItemName,
			Unit:       t.Unit,
			Quantity:   int(t.Quantity),
			ReceivedBy: tstr(t.ReceivedBy),
			Distributor: tstr(t.Distributor),
			PONumber:   tstr(t.InvoiceOrPoNum),
			Condition:  tstr(t.ConditionStatus),
			Geo:        geoDisplay(t.GeoLat, t.GeoLng),
		})
	}
	h.show(w, r, "Riwayat", views.History(views.HistoryData{User: userInfo(r), TX: vrows}))
}

// --- formatting helpers ---

func fmtWIB(t pgtype.Timestamptz) string {
	if !t.Valid {
		return "—"
	}
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	return t.Time.In(loc).Format("02-01-2006 15:04")
}

func geoDisplay(lat, lng pgtype.Float8) string {
	if !lat.Valid || !lng.Valid {
		return "—"
	}
	return fmt.Sprintf("%.6f, %.6f", lat.Float64, lng.Float64)
}

func parseUUID(s string) (uuid.UUID, error) { return uuid.Parse(s) }

var _ = pgtype.Text{}
