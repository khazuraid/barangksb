package handlers

import (
	"net/http"
	"strconv"

	"inventariskantor/internal/views"
	"inventariskantor/internal/service"
)

// ItemsList uses service.ListItems with filter + pagination.
func (h *Handlers) ItemsList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := service.ItemFilter{
		Query:   q.Get("q"),
		Cat:     q.Get("cat"),
		Loc:     q.Get("loc"),
		Page:    pageParam(r),
		PerPage: 25,
	}
	items, total, err := h.svc.ListItems(r.Context(), f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cats, _ := h.svc.CategoryNames(r.Context())
	locs, _ := h.svc.LocationNames(r.Context())
	h.show(w, r, "Barang", views.Items(views.ItemsData{
		User:  userInfo(r),
		Items: itemRows(items),
		Query: f.Query, Cat: f.Cat, Loc: f.Loc,
		Cats: cats, Locs: locs,
		Pager: buildPager(r, f.Page, f.PerPage, total),
	}))
}

func (h *Handlers) ItemNewForm(w http.ResponseWriter, r *http.Request) {
	cats, _ := h.svc.CategoryNames(r.Context())
	locs, _ := h.svc.LocationNames(r.Context())
	h.show(w, r, "Tambah Barang", views.ItemForm(views.ItemFormData{User: userInfo(r), Cats: cats, Locs: locs}))
}

func (h *Handlers) ItemEditForm(w http.ResponseWriter, r *http.Request) {
	it, err := h.svc.GetItem(r.Context(), chiURLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	cats, _ := h.svc.CategoryNames(r.Context())
	locs, _ := h.svc.LocationNames(r.Context())
	h.show(w, r, "Edit Barang", views.ItemForm(views.ItemFormData{
		User: userInfo(r), Item: toFormValues(&it), Cats: cats, Locs: locs,
	}))
}

func (h *Handlers) ItemsCreate(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r) {
		return
	}
	photoURL := h.processUpload(r, "")
	it, err := h.svc.CreateItem(r.Context(), itemInputFromForm(r.PostForm, photoURL))
	if err != nil {
		if isDuplicateErr(err) {
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
	id := chiURLParam(r, "id")
	if !parseForm(w, r) {
		return
	}
	old, err := h.svc.GetItem(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	photoURL := h.processUpload(r, optStr(old.PhotoUrl))
	in := itemInputFromForm(r.PostForm, photoURL)
	if in.SKU == "" {
		in.SKU = old.Sku
	}
	if err := h.svc.UpdateItem(r.Context(), id, in); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.broker.Publish("items")
	http.Redirect(w, r, "/items?updated=1", http.StatusSeeOther)
}

func (h *Handlers) ItemDelete(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	if err := h.svc.DeleteItem(r.Context(), chiURLParam(r, "id")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.broker.Publish("items")
	http.Redirect(w, r, "/items?deleted=1", http.StatusSeeOther)
}

func (h *Handlers) ItemDetail(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/items", http.StatusSeeOther)
}

func requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	u := authUserOf(r)
	if u == nil || u.Role != "admin" {
		http.Error(w, "403 — hanya admin", http.StatusForbidden)
		return false
	}
	return true
}

func isDuplicateErr(err error) bool {
	return containsStr(err.Error(), "duplicate key") || containsStr(err.Error(), "unique")
}

var _ = strconv.Itoa
