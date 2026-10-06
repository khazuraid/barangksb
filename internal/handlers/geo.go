package handlers

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"inventariskantor/internal/auth"
	"inventariskantor/internal/views"
)

type geoData struct {
	lat, lng, acc pgtype.Float8
	at            pgtype.Timestamptz
	name          pgtype.Text
}

func userInfo(r *http.Request) views.UserInfo {
	info := views.UserInfo{}
	if u := auth.FromContext(r.Context()); u != nil {
		info = views.UserInfo{Name: u.Name, Role: u.Role}
	}
	return info
}

func authUserOf(r *http.Request) *auth.User { return auth.FromContext(r.Context()) }

func authUserName(r *http.Request) string {
	if u := authUserOf(r); u != nil {
		return u.Name
	}
	return ""
}

func geoFromForm(f formValues) geoData {
	lat := atof(f.Get("geo_lat"))
	lng := atof(f.Get("geo_lng"))
	acc := atof(f.Get("geo_acc"))
	if lat == 0 && lng == 0 {
		return geoData{}
	}
	g := geoData{
		lat: pgtype.Float8{Float64: lat, Valid: true},
		lng: pgtype.Float8{Float64: lng, Valid: true},
		acc: pgtype.Float8{Float64: acc, Valid: acc > 0},
		at:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	if n := f.Get("geo_name"); n != "" {
		g.name = pgtype.Text{String: n, Valid: true}
	}
	return g
}
