package handlers

import (
	"net/url"

	"inventariskantor/internal/views"
)

func viewsPager(t interface{ Fatalf(string, ...interface{}) }, page, perPage, total int) views.Pager {
	_ = t
	return views.Pager{Page: page, PerPage: perPage, Total: total, BaseURL: "/items"}
}

func testFormGet(v url.Values, key string) string { return v.Get(key) }
