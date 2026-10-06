package handlers

import (
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"inventariskantor/internal/service"
)

// alias untuk service filter yang dipakai beberapa handler
type serviceItemFilter = service.ItemFilter
type serviceTXFilter = service.TXFilter

var (
	_ = chi.URLParam
	_ = pgtype.Text{}
	_ = strconv.Itoa
	_ = strings.Contains
)
