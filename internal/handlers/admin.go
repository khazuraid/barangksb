package handlers

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"inventariskantor/internal/views"
)

// ---------- Kelola User (admin only) ----------

func (h *Handlers) UsersList(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	users, err := h.svc.ListUsers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.show(w, r, "Pengguna", views.Users(views.UsersData{User: userInfo(r), Users: users}))
}

func (h *Handlers) UsersCreate(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	if !parseForm(w, r) {
		return
	}
	err := h.svc.CreateUser(r.Context(),
		r.PostFormValue("name"), strings.ToLower(r.PostFormValue("email")),
		r.PostFormValue("password"), r.PostFormValue("role"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

func (h *Handlers) UsersRole(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	if err := h.svc.UpdateUserRole(r.Context(), chi.URLParam(r, "email"), r.PostFormValue("role")); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

func (h *Handlers) UsersDelete(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	if err := h.svc.DeleteUser(r.Context(), chi.URLParam(r, "email")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

// ---------- Ganti password sendiri ----------

func (h *Handlers) PasswordForm(w http.ResponseWriter, r *http.Request) {
	h.show(w, r, "Ganti Password", views.Password(userInfo(r), "", false))
}

func (h *Handlers) PasswordPost(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r) {
		return
	}
	u := authUserOf(r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	newPW := r.PostFormValue("new")
	if newPW != r.PostFormValue("confirm") {
		h.show(w, r, "Ganti Password", views.Password(userInfo(r), "Konfirmasi password tidak cocok", true))
		return
	}
	if err := h.svc.ChangePassword(r.Context(), u.Email, r.PostFormValue("old"), newPW); err != nil {
		h.show(w, r, "Ganti Password", views.Password(userInfo(r), err.Error(), true))
		return
	}
	h.show(w, r, "Ganti Password", views.Password(userInfo(r), "Password berhasil diganti", false))
}

// ---------- Master Lokasi (admin only) ----------

func (h *Handlers) LocationsList(w http.ResponseWriter, r *http.Request) {
	locs, err := h.svc.ListLocations(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.show(w, r, "Lokasi", views.Locations(views.LocationsData{User: userInfo(r), Locs: locs}))
}

func (h *Handlers) LocationsCreate(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	if !parseForm(w, r) {
		return
	}
	if err := h.svc.CreateLocation(r.Context(), strings.TrimSpace(r.PostFormValue("name"))); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/locations", http.StatusSeeOther)
}

func (h *Handlers) LocationsRename(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	if !parseForm(w, r) {
		return
	}
	if err := h.svc.RenameLocation(r.Context(), chi.URLParam(r, "id"), strings.TrimSpace(r.PostFormValue("name"))); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/locations", http.StatusSeeOther)
}

func (h *Handlers) LocationsDelete(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	if err := h.svc.DeleteLocation(r.Context(), chi.URLParam(r, "id")); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	http.Redirect(w, r, "/locations", http.StatusSeeOther)
}
