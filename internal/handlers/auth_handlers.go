package handlers

import (
	"net/http"

	"inventariskantor/internal/auth"
)

func (h *Handlers) LoginForm(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "login", nil)
}

func (h *Handlers) LoginPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")

	u, hash, err := auth.GetUserByEmail(r.Context(), h.pool, email)
	if err != nil || !auth.CheckPassword(hash, password) {
		h.render(w, r, "login", map[string]any{"Error": "Email atau password salah"})
		return
	}
	h.sess.Put(r.Context(), "user_id", u.ID)
	h.sess.Put(r.Context(), "user", *u)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) LogoutPost(w http.ResponseWriter, r *http.Request) {
	h.sess.Destroy(r.Context())
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

var _ = http.StatusOK
