package handlers

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"inventariskantor/internal/models"
	"inventariskantor/internal/views"
)

func (h *Handlers) CategoriesList(w http.ResponseWriter, r *http.Request) {
	cats, err := h.listCategoryRows(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.show(w, r, "Kategori", views.Categories(views.CategoriesData{User: userInfo(r), Cats: cats}))
}

func (h *Handlers) CategoriesCreate(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r) {
		return
	}
	name := strings.TrimSpace(r.PostFormValue("name"))
	if name == "" {
		http.Error(w, "nama kategori wajib", http.StatusBadRequest)
		return
	}
	id := slugify(name)
	_, err := h.pool.Exec(r.Context(),
		`INSERT INTO categories (id, name) VALUES ($1,$2) ON CONFLICT (id) DO UPDATE SET name=$2`,
		id, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.broker.Publish("items")
	http.Redirect(w, r, "/categories", http.StatusSeeOther)
}

func (h *Handlers) CategoriesRename(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !parseForm(w, r) {
		return
	}
	name := strings.TrimSpace(r.PostFormValue("name"))
	if name == "" {
		http.Error(w, "nama kategori wajib", http.StatusBadRequest)
		return
	}
	if _, err := h.pool.Exec(r.Context(), `UPDATE categories SET name=$2 WHERE id=$1`, id, name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// sinkronkan nama di item yang memakai kategori lama
	var oldName string
	_ = h.pool.QueryRow(r.Context(), `SELECT name FROM categories WHERE id=$1`, id).Scan(&oldName)
	if oldName != "" && oldName != name {
		h.pool.Exec(r.Context(), `UPDATE inventory_items SET category=$2 WHERE category=$1`, oldName, name)
	}
	h.broker.Publish("items")
	http.Redirect(w, r, "/categories", http.StatusSeeOther)
}

func (h *Handlers) CategoriesDelete(w http.ResponseWriter, r *http.Request) {
	u := authUserOf(r)
	if u == nil || u.Role != "admin" {
		http.Error(w, "403 — hanya admin boleh menghapus kategori", http.StatusForbidden)
		return
	}
	id := chi.URLParam(r, "id")
	var name string
	if err := h.pool.QueryRow(r.Context(), `SELECT name FROM categories WHERE id=$1`, id).Scan(&name); err != nil {
		http.NotFound(w, r)
		return
	}
	// larang hapus jika masih dipakai barang
	var used int
	_ = h.pool.QueryRow(r.Context(), `SELECT count(*) FROM inventory_items WHERE category=$1`, name).Scan(&used)
	if used > 0 {
		http.Error(w, "Kategori masih dipakai "+itoa(used)+" barang — pindahkan dulu.", http.StatusConflict)
		return
	}
	if _, err := h.pool.Exec(r.Context(), `DELETE FROM categories WHERE id=$1`, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.broker.Publish("items")
	http.Redirect(w, r, "/categories", http.StatusSeeOther)
}

func (h *Handlers) listCategoryRows(ctx ctx2) ([]models.Category, error) {
	rows, err := h.pool.Query(ctx, `SELECT id, name FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	return collectCategories(rows)
}

func slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, c := range s {
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
			b.WriteRune(c)
		case c == ' ' || c == '-' || c == '/' || c == '&':
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

var _ = models.Category{}
