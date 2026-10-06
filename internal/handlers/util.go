package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func chiURLParam(r *http.Request, key string) string { return chi.URLParam(r, key) }

func atoi0(s string) int32 {
	n, _ := strconv.Atoi(s)
	if n < 0 {
		return 0
	}
	return int32(n)
}

func atol(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	if n < 0 {
		return 0
	}
	return n
}

// atoi64 ada di movement.go

func atof(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func nullStr(s string) pgtype.Text {
	if strings.TrimSpace(s) == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func nullInt8(s string) pgtype.Int8 {
	if strings.TrimSpace(s) == "" {
		return pgtype.Int8{}
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: n, Valid: true}
}

func optStr(t pgtype.Text) string {
	if t.Valid {
		return t.String
	}
	return ""
}

func optStrOr(t pgtype.Text, def string) string {
	if t.Valid && t.String != "" {
		return t.String
	}
	return def
}
