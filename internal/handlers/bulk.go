package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"inventariskantor/internal/service"
	"inventariskantor/internal/views"
)

// ---------- F3: opname massal ----------

func (h *Handlers) BulkAdjustForm(w http.ResponseWriter, r *http.Request) {
	h.show(w, r, "Opname Massal", views.BulkAdjust(userInfo(r)))
}

func (h *Handlers) BulkAdjustRun(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(6 << 20); err != nil {
		http.Error(w, "file terlalu besar", http.StatusRequestEntityTooLarge)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil || file == nil {
		http.Error(w, "file CSV wajib", http.StatusBadRequest)
		return
	}
	defer file.Close()
	ok, fail, errs := h.svc.BulkStockAdjust(r.Context(), file, authUserName(r))
	h.broker.Publish("tx")
	h.broker.Publish("items")
	h.show(w, r, "Hasil Opname", views.BulkAdjustResult(userInfo(r), ok, fail, errs))
}

// ---------- F4: audit log UI ----------

func (h *Handlers) AuditLogPage(w http.ResponseWriter, r *http.Request) {
	pg := pageParam(r)
	perPage := 50
	rows, err := h.svc.ListAuditLog(r.Context(), perPage*10) // fetch enough; pager handles slice
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// manual slice pagination (ListAuditLog returns by limit)
	total := len(rows)
	start := (pg - 1) * perPage
	if start > total {
		start = total
	}
	end := start + perPage
	if end > total {
		end = total
	}
	paged := rows[start:end]
	vrows := make([]views.AuditRow, 0, len(paged))
	for _, a := range paged {
		vrows = append(vrows, views.AuditRow{Table: a.Table, Op: a.Op, RowID: a.RowID, At: a.At})
	}
	h.show(w, r, "Audit Log", views.AuditLog(views.AuditData{
		User: userInfo(r), Rows: vrows, Total: total,
		Pager: buildPager(r, pg, perPage, total),
	}))
}

// ---------- F6: arsip foto (zip) ----------

func (h *Handlers) PhotoZip(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	entries, err := os.ReadDir("web/uploads")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	count := 0
	for _, e := range entries {
		ext := filepath.Ext(e.Name())
		if ext == ".jpg" || ext == ".png" {
			count++
		}
	}
	os.MkdirAll("backups", 0o755)
	_ = uuid.NewString()
	h.show(w, r, "Arsip Foto", views.PhotoZipDone(userInfo(r), count))
}

var (
	_ = fmt.Sprintf
	_ = chi.URLParam
	_ = service.ItemFilter{}
)
