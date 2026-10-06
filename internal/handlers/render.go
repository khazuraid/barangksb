package handlers

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"inventariskantor/internal/auth"
	"inventariskantor/internal/models"
	"inventariskantor/internal/views"
)

type modelsTX = models.StockTransaction

func tstr(t pgtype.Text) string {
	if t.Valid && t.String != "" {
		return t.String
	}
	return "—"
}

func fmtUUID(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return uuid.UUID(u.Bytes).String()
}

// show renders a templ component inside the layout.
func (h *Handlers) show(w http.ResponseWriter, r *http.Request, title string, content templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	info := views.UserInfo{}
	if u := auth.FromContext(r.Context()); u != nil {
		info = views.UserInfo{Name: u.Name, Role: u.Role}
	}
	if err := views.Layout(title, info, content).Render(r.Context(), w); err != nil {
		slogErr("render failed", "page", title, "err", err)
	}
}

// render is the login-only shim.
func (h *Handlers) render(w http.ResponseWriter, r *http.Request, page string, data any) {
	if page == "login" {
		var err string
		if d, ok := data.(map[string]any); ok {
			err, _ = d["Error"].(string)
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = views.Login(err).Render(r.Context(), w)
		return
	}
	http.Error(w, "unknown page", http.StatusInternalServerError)
}

var _ = fmt.Sprintf
