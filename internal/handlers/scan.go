package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"inventariskantor/internal/views"
)

// ScanDetail is the PUBLIC barcode landing page: /scan/{itemID}
// Shows item info + stock without requiring login (read-only, no mutating data).
func (h *Handlers) ScanDetail(w http.ResponseWriter, r *http.Request) {
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
	cats, _ := h.svc.CategoryNames(r.Context())
	h.show(w, r, "Detail Barang", views.ScanDetail(views.ScanDetailData{
		Cats: cats,
		Row:  views.ScanRow{
			ID:           fmtUUID(it.ID),
			SKU:          it.Sku,
			Name:         it.Name,
			Category:     it.Category,
			Location:     it.Location,
			CurrentStock: int(it.CurrentStock),
			MinStock:     int(it.MinStock),
			Unit:         it.Unit,
			Condition:    optStrOr(it.ConditionStatus, "Berfungsi"),
			Merk:         tstr(it.Merk),
			TypeModel:    tstr(it.TypeModel),
			SerialNumber: tstr(it.SerialNumber),
			Year:         tstr(it.ProcurementYear),
			Funding:      tstr(it.FundingSource),
			Distributor:  tstr(it.Distributor),
			Description:  tstr(it.Description),
			PhotoURL:     optStr(it.PhotoUrl),
			IsAvailable:  it.IsAvailable,
		},
	}))
}

// ScanBySKU is a public SKU lookup: /scan/sku/{sku}
func (h *Handlers) ScanBySKU(w http.ResponseWriter, r *http.Request) {
	sku := chi.URLParam(r, "sku")
	var id string
	err := h.pool.QueryRow(r.Context(), `SELECT id FROM inventory_items WHERE sku=$1`, sku).Scan(&id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/scan/"+id, http.StatusSeeOther)
}
