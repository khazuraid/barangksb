package views

import (
	"fmt"
	"strings"
)

// PaginationData: view-safe pager (tanpa import cycle).
type PaginationData struct {
	Page    int
	PerPage int
	Total   int
	BaseURL string
	MaxShow int
}

func (p PaginationData) Pages() int {
	if p.PerPage < 1 {
		return 1
	}
	n := p.Total / p.PerPage
	if p.Total%p.PerPage > 0 {
		n++
	}
	if n < 1 {
		n = 1
	}
	return n
}

func (p PaginationData) HasPrev() bool { return p.Page > 1 }
func (p PaginationData) HasNext() bool { return p.Page < p.Pages() }

func (p PaginationData) Window() []int {
	if p.MaxShow < 1 {
		p.MaxShow = 7
	}
	total := p.Pages()
	start := p.Page - p.MaxShow/2
	if start < 1 {
		start = 1
	}
	end := start + p.MaxShow - 1
	if end > total {
		end = total
		start = end - p.MaxShow + 1
		if start < 1 {
			start = 1
		}
	}
	out := make([]int, 0, end-start+1)
	for i := start; i <= end; i++ {
		out = append(out, i)
	}
	return out
}

func (p PaginationData) URL(page int) string {
	sep := "?"
	if strings.Contains(p.BaseURL, "?") {
		sep = "&"
	}
	return fmt.Sprintf("%s%spage=%d", p.BaseURL, sep, page)
}

// Pager = alias PaginationData (dipakai di data structs).
type Pager = PaginationData
