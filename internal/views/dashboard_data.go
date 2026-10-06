package views

// DashboardData v2 — bento + chart. (User dideklarasikan di data.go)
type DashboardData struct {
	User       UserInfo
	TotalItems int
	TotalStock int
	LowStock   []LowStockRow
	RecentTX   []RecentTXRow
	StockByCat map[string]int
	Categories []string
	ChartCats  []ChartPoint
	ChartTypes map[string]int
}

type LowStockRow struct {
	Name, SKU string
	Current   int
	Min       int
	Unit      string
	Location  string
}

type RecentTXRow struct {
	Time, ItemName, SKU, Unit string
	Quantity                  int
	Type                      string
}

// ChartPoint — satu titik data chart.
type ChartPoint struct {
	Label string `json:"label"`
	Value int    `json:"value"`
}
