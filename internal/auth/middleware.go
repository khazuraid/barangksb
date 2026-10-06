package auth

import (
	"context"
	"encoding/gob"
	"net/http"

	"github.com/alexedwards/scs/v2"
)

func init() { gob.Register(User{}) }

// RequireAuth loads the user from session; redirects to /login when absent.
func RequireAuth(sess *scs.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uid := sess.GetString(r.Context(), "user_id")
			if uid == "" {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			if u, ok := sess.Get(r.Context(), "user").(User); ok {
				next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), &u)))
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		})
	}
}

func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := FromContext(r.Context())
			if u == nil || u.Role != role {
				http.Error(w, "403 — hanya admin", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func FromContext(ctx context.Context) *User {
	u, _ := ctx.Value(UserKey).(*User)
	return u
}
