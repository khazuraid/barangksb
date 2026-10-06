package handlers
import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"inventariskantor/internal/service"
	"inventariskantor/internal/views"
)

func atoi64(s string) (int64, error) {
	return strconvParseInt(s)
}

func (h *Handlers) MovementForm(w http.ResponseWriter, r *http.Request) {
	tab := r.URL.Query().Get("tab")
	if tab != "out" && tab != "adjust" {
		tab = "in"
	}
	items, _, err := h.svc.ListItems(r.Context(), serviceFilterAll())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cats, _ := h.svc.CategoryNames(r.Context())
	locs, _ := h.svc.LocationNames(r.Context())
	h.show(w, r, "Mutasi Barang", views.Movement(views.MovementData{
		User: userInfo(r), Cats: cats, Locs: locs,
		SKU: r.URL.Query().Get("sku"), Items: itemRows(items), Tab: tab,
	}))
}

func strconvParseInt(s string) (int64, error) {
	var n int64
	neg := false
	i := 0
	if len(s) > 0 && (s[0] == '-' || s[0] == '+') {
		neg = s[0] == '-'
		i = 1
	}
	for ; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, fmt.Errorf("invalid number")
		}
		n = n*10 + int64(s[i]-'0')
	}
	if neg {
		n = -n
	}
	return n, nil
}

func serviceFilterAll() (f serviceItemFilter) { return }

// RecordStockInFull: stock-in dengan foto/geotag (dipakai form web).
func (h *Handlers) RecordStockInFull(ctx context.Context, sku string, qty int32, receivedBy, notes, photoURL string, geo geoData) (*service.StockInResult, error) {
	res, err := h.svc.RecordStockIn(ctx, sku, qty, receivedBy, notes)
	if err == nil && (photoURL != "" || geo.lat.Valid) {
		h.pool.Exec(ctx, `UPDATE stock_transactions SET photo_url=$2, geo_lat=$3, geo_lng=$4, geo_acc=$5, geo_at=$6, geo_name=$7
			WHERE id = (SELECT id FROM stock_transactions WHERE item_sku=$1 ORDER BY timestamp DESC LIMIT 1)`,
			sku, nullStr(photoURL), geo.lat, geo.lng, geo.acc, geo.at, geo.name)
	}
	return res, err
}

// MovementPost: barang MASUK (dengan foto/geotag lengkap).
func (h *Handlers) MovementPost(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r) {
		return
	}
	form := formValues(r.PostForm)
	itemID := form.Get("item_id")
	var sku string
	if _, err := parseUUID(itemID); err == nil {
		if it, err := h.svc.GetItem(r.Context(), itemID); err == nil {
			sku = it.Sku
		}
	} else {
		sku = itemID // bot/API kirim SKU langsung
	}
	if sku == "" {
		http.Error(w, "item tidak valid", http.StatusBadRequest)
		return
	}
	qty := atoi0(form.Get("quantity"))
	if qty <= 0 {
		http.Error(w, "jumlah harus > 0", http.StatusBadRequest)
		return
	}
	photoURL := h.processUpload(r, "")
	geo := geoFromForm(form)

	res, err := h.RecordStockInFull(r.Context(), sku, qty, form.Get("received_by"), form.Get("notes"), photoURL, geo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.broker.Publish("tx")
	h.broker.Publish("items")
	http.Redirect(w, r, "/history?ok="+res.SKU, http.StatusSeeOther)
}

// MovementOut: barang KELUAR.
func (h *Handlers) MovementOut(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r) {
		return
	}
	form := formValues(r.PostForm)
	qty := atoi0(form.Get("quantity"))
	if qty <= 0 {
		http.Error(w, "jumlah harus > 0", http.StatusBadRequest)
		return
	}
	res, err := h.svc.RecordStockOut(r.Context(), form.Get("item_id"), qty, form.Get("received_by"), form.Get("notes"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.broker.Publish("tx")
	h.broker.Publish("items")
	http.Redirect(w, r, "/history?ok="+res.SKU, http.StatusSeeOther)
}

// MovementAdjust: stok opname.
func (h *Handlers) MovementAdjust(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r) {
		return
	}
	form := formValues(r.PostForm)
	actual, err := atoi64(form.Get("quantity"))
	if err != nil || actual < 0 {
		http.Error(w, "stok fisik harus angka ≥ 0", http.StatusBadRequest)
		return
	}
	res, err := h.svc.AdjustStock(r.Context(), form.Get("item_id"), int32(actual), form.Get("received_by"), form.Get("notes"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.broker.Publish("tx")
	h.broker.Publish("items")
	http.Redirect(w, r, "/history?ok="+res.SKU, http.StatusSeeOther)
}

// History: riwayat dengan filter + pagination.
func (h *Handlers) History(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := serviceTXFilter{
		SKU: q.Get("sku"), Type: q.Get("type"),
		From: q.Get("from"), To: q.Get("to"),
		Page: pageParam(r), PerPage: 50,
	}
	txs, total, err := h.svc.ListTransactions(r.Context(), f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	vrows := make([]views.TXRow, 0, len(txs))
	for i := range txs {
		t := &txs[i]
		vrows = append(vrows, views.TXRow{
			Time:        fmtWIB(t.Timestamp),
			Type:        t.Type,
			SKU:         t.ItemSku,
			Name:        t.ItemName,
			Unit:        t.Unit,
			Quantity:    int(t.Quantity),
			ReceivedBy:  tstr(t.ReceivedBy),
			Distributor: tstr(t.Distributor),
			PONumber:    tstr(t.InvoiceOrPoNum),
			Condition:   tstr(t.ConditionStatus),
			Geo:         geoDisplay(t.GeoLat, t.GeoLng),
		})
	}
	h.show(w, r, "Riwayat", views.History(views.HistoryData{
		User: userInfo(r), TX: vrows,
		Pager: buildPager(r, f.Page, f.PerPage, total),
		FSKU: f.SKU, FType: f.Type, FFrom: f.From, FTo: f.To,
		Types: []string{"IN", "OUT", "ADJUST+", "ADJUST-"},
	}))
}

// ---------- formatting ----------

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
	return formatGeo(lat.Float64, lng.Float64)
}

func wibLoc() *time.Location {
	return time.FixedZone("WIB", 7*3600)
}

var _ = time.Now
