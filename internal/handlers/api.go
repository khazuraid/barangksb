package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"inventariskantor/internal/drive"
	"inventariskantor/internal/service"
)

func driveUpload(sa, folder, name string, data []byte, mime string) (string, error) {
	return drive.UploadPhoto(sa, folder, name, data, mime)
}

// ---------- API JSON (F12) ----------

// GET /api/items?q=&cat=&page=
func (h *Handlers) APIItems(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := serviceItemFilter{
		Query: q.Get("q"), Cat: q.Get("cat"), Loc: q.Get("loc"),
		Page: pageParam(r), PerPage: 50,
	}
	items, total, err := h.svc.ListItems(r.Context(), f)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonWrite(w, map[string]any{
		"data": items, "total": total,
		"page": f.Page, "per_page": f.PerPage,
	})
}

// APIItem single item by ID: GET /api/items/{id}
func (h *Handlers) APIItem(w http.ResponseWriter, r *http.Request) {
	it, err := h.svc.GetItem(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	jsonWrite(w, it)
}

// GET /api/transactions?sku=&type=&from=&to=&page=
func (h *Handlers) APITransactions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := serviceTXFilter{
		SKU: q.Get("sku"), Type: q.Get("type"),
		From: q.Get("from"), To: q.Get("to"),
		Page: pageParam(r), PerPage: 100,
	}
	txs, total, err := h.svc.ListTransactions(r.Context(), f)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonWrite(w, map[string]any{"data": txs, "total": total, "page": f.Page})
}

// POST /api/movement  {sku, quantity, type: IN|OUT|ADJUST, person, notes}
func (h *Handlers) APIMovement(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SKU      string `json:"sku"`
		Quantity int32  `json:"quantity"`
		Type     string `json:"type"`
		Person   string `json:"person"`
		Notes    string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "bad json", http.StatusBadRequest)
		return
	}
	if body.SKU == "" || body.Quantity <= 0 {
		jsonError(w, "sku & quantity>0 wajib", http.StatusBadRequest)
		return
	}
	var res *service.StockInResult
	var err error
	switch body.Type {
	case "OUT":
		res, err = h.svc.RecordStockOut(r.Context(), body.SKU, body.Quantity, body.Person, body.Notes)
	case "ADJUST":
		res, err = h.svc.AdjustStock(r.Context(), body.SKU, body.Quantity, body.Person, body.Notes)
	default:
		res, err = h.svc.RecordStockIn(r.Context(), body.SKU, body.Quantity, body.Person, body.Notes)
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.broker.Publish("tx")
	jsonWrite(w, res)
}

func jsonWrite(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

var _ = strconv.Itoa
