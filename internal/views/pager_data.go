package views

// Pager untuk pagination semua tabel.
type Pager struct {
	Page    int
	PerPage int
	Total   int
	BaseURL string
	MaxShow int
}
