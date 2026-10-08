package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ItemHandler struct {
	pool *pgxpool.Pool
}

func NewItemHandler(pool *pgxpool.Pool) *ItemHandler {
	return &ItemHandler{pool: pool}
}

type itemReq struct {
	SKU             string `json:"sku"`
	Name            string `json:"name" binding:"required"`
	Category        string `json:"category" binding:"required"`
	Location        string `json:"location" binding:"required"`
	CurrentStock    int32  `json:"current_stock"`
	MinStock        int32  `json:"min_stock"`
	Unit            string `json:"unit"`
	PricePerUnit    int64  `json:"price_per_unit"`
	Description     string `json:"description"`
	PhotoURL        string `json:"photo_url"`
	Merk            string `json:"merk"`
	TypeModel       string `json:"type_model"`
	SerialNumber    string `json:"serial_number"`
	ProcurementYear string `json:"procurement_year"`
	ConditionStatus string `json:"condition_status"`
	FundingSource   string `json:"funding_source"`
	Distributor     string   `json:"distributor"`
	AklAkd          string   `json:"akl_akd"`
	GeoLat          *float64 `json:"geo_lat"`
	GeoLng          *float64 `json:"geo_lng"`
	GeoAcc          *float64 `json:"geo_acc"`
	GeoName         string   `json:"geo_name"`
	IsAvailable     bool     `json:"is_available"`
}

func (h *ItemHandler) List(c *gin.Context) {
	q := c.Query("q")
	cat := c.Query("cat")
	loc := c.Query("loc")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 { page = 1 }
	perPage := configPerPage(c)
	offset := (page - 1) * perPage

	where := " WHERE TRUE"
	args := []any{}
	if q != "" {
		where += fmt.Sprintf(` AND (name ILIKE $%d OR sku ILIKE $%d OR location ILIKE $%d)`, len(args)+1, len(args)+1, len(args)+1)
		args = append(args, "%"+q+"%")
	}
	if cat != "" { where += fmt.Sprintf(` AND category = $%d`, len(args)+1); args = append(args, cat) }
	if loc != "" { where += fmt.Sprintf(` AND location = $%d`, len(args)+1); args = append(args, loc) }

	var total int
	h.pool.QueryRow(c, `SELECT count(*) FROM inventory_items`+where, args...).Scan(&total)

	args = append(args, perPage, offset)
	rows, err := h.pool.Query(c, fmt.Sprintf(`SELECT id, sku, name, category, location, current_stock, min_stock, unit, price_per_unit, condition_status, is_available, photo_url, geo_lat, geo_lng, COALESCE(geo_name, '') FROM inventory_items%s ORDER BY name LIMIT $%d OFFSET $%d`, where, len(args)-1, len(args)), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type itemRow struct {
		ID           string   `json:"id"`
		SKU          string   `json:"sku"`
		Name         string   `json:"name"`
		Category     string   `json:"category"`
		Location     string   `json:"location"`
		CurrentStock int32    `json:"current_stock"`
		MinStock     int32    `json:"min_stock"`
		Unit         string   `json:"unit"`
		PricePerUnit int64    `json:"price_per_unit"`
		Condition    string   `json:"condition_status"`
		IsAvailable  bool     `json:"is_available"`
		PhotoURL     string   `json:"photo_url"`
		GeoLat       *float64 `json:"geo_lat"`
		GeoLng       *float64 `json:"geo_lng"`
		GeoName      string   `json:"geo_name"`
	}
	items := []itemRow{}
	for rows.Next() {
		var it itemRow
		rows.Scan(&it.ID, &it.SKU, &it.Name, &it.Category, &it.Location, &it.CurrentStock, &it.MinStock, &it.Unit, &it.PricePerUnit, &it.Condition, &it.IsAvailable, &it.PhotoURL, &it.GeoLat, &it.GeoLng, &it.GeoName)
		items = append(items, it)
	}

	pages := total / perPage
	if total%perPage > 0 { pages++ }
	c.JSON(http.StatusOK, paginatedResp{Data: items, Total: total, Page: page, PerPage: perPage, Pages: pages})
}

func (h *ItemHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var it itemReq
	err := h.pool.QueryRow(c, `SELECT sku, name, category, location, current_stock, min_stock, unit, COALESCE(price_per_unit, 0), COALESCE(description, ''), COALESCE(photo_url, ''), COALESCE(merk, ''), COALESCE(type_model, ''), COALESCE(serial_number, ''), COALESCE(procurement_year, ''), COALESCE(condition_status, 'Berfungsi'), COALESCE(funding_source, ''), COALESCE(distributor, ''), COALESCE(akl_akd, ''), is_available, geo_lat, geo_lng, geo_acc, COALESCE(geo_name, '') FROM inventory_items WHERE id::text=$1`, id).
		Scan(&it.SKU, &it.Name, &it.Category, &it.Location, &it.CurrentStock, &it.MinStock, &it.Unit, &it.PricePerUnit, &it.Description, &it.PhotoURL, &it.Merk, &it.TypeModel, &it.SerialNumber, &it.ProcurementYear, &it.ConditionStatus, &it.FundingSource, &it.Distributor, &it.AklAkd, &it.IsAvailable, &it.GeoLat, &it.GeoLng, &it.GeoAcc, &it.GeoName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, it)
}

func (h *ItemHandler) Create(c *gin.Context) {
	var req itemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sku := strings.TrimSpace(req.SKU)
	if sku == "" {
		sku = generateSKU(h, c, req.Category)
	}
	var id string
	err := h.pool.QueryRow(c, `INSERT INTO inventory_items (sku, name, category, location, current_stock, min_stock, unit, price_per_unit, description, photo_url, merk, type_model, serial_number, procurement_year, condition_status, funding_source, distributor, akl_akd, geo_lat, geo_lng, geo_acc, geo_name) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22) RETURNING id::text`,
		sku, req.Name, req.Category, req.Location, req.CurrentStock, req.MinStock, req.Unit, req.PricePerUnit,
		nullable(req.Description), nullable(req.PhotoURL), nullable(req.Merk), nullable(req.TypeModel), nullable(req.SerialNumber),
		nullable(req.ProcurementYear), nullable(req.ConditionStatus), nullable(req.FundingSource), nullable(req.Distributor), nullable(req.AklAkd),
		req.GeoLat, req.GeoLng, req.GeoAcc, nullable(req.GeoName)).
		Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "sku": sku})
}

func (h *ItemHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req itemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.pool.Exec(c, `UPDATE inventory_items SET sku=$2, name=$3, category=$4, location=$5, current_stock=$6, min_stock=$7, unit=$8, price_per_unit=$9, description=$10, photo_url=$11, merk=$12, type_model=$13, serial_number=$14, procurement_year=$15, condition_status=$16, funding_source=$17, distributor=$18, akl_akd=$19, geo_lat=$20, geo_lng=$21, geo_acc=$22, geo_name=$23, updated_at=now() WHERE id=$1`,
		id, req.SKU, req.Name, req.Category, req.Location, req.CurrentStock, req.MinStock, req.Unit, req.PricePerUnit,
		nullable(req.Description), nullable(req.PhotoURL), nullable(req.Merk), nullable(req.TypeModel), nullable(req.SerialNumber),
		nullable(req.ProcurementYear), nullable(req.ConditionStatus), nullable(req.FundingSource), nullable(req.Distributor), nullable(req.AklAkd),
		req.GeoLat, req.GeoLng, req.GeoAcc, nullable(req.GeoName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *ItemHandler) Delete(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "hanya admin"})
		return
	}
	id := c.Param("id")

	// Check if item has transactions — refuse to cascade-delete history
	var txCount int
	h.pool.QueryRow(c, `SELECT count(*) FROM stock_transactions WHERE item_id::text=$1`, id).Scan(&txCount)
	if txCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error": "barang memiliki " + strconv.Itoa(txCount) + " transaksi terkait. Hapus/Arsipkan transaksi terlebih dahulu, atau set is_available=false untuk menonaktifkan.",
		})
		return
	}

	h.pool.Exec(c, `DELETE FROM inventory_items WHERE id=$1`, id)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *ItemHandler) Dashboard(c *gin.Context) {
	var totalItems, totalStock int
	h.pool.QueryRow(c, `SELECT count(*), COALESCE(sum(current_stock),0) FROM inventory_items`).Scan(&totalItems, &totalStock)

	lowRows, _ := h.pool.Query(c, `SELECT name, sku, current_stock, min_stock, unit, location FROM inventory_items WHERE current_stock <= min_stock ORDER BY current_stock LIMIT 10`)
	defer lowRows.Close()
	type lowRow struct {
		Name string `json:"name"`
		SKU  string `json:"sku"`
		Current int `json:"current"`
		Min  int `json:"min"`
		Unit string `json:"unit"`
		Location string `json:"location"`
	}
	lowStock := []lowRow{}
	for lowRows.Next() {
		var l lowRow
		lowRows.Scan(&l.Name, &l.SKU, &l.Current, &l.Min, &l.Unit, &l.Location)
		lowStock = append(lowStock, l)
	}

	txRows, _ := h.pool.Query(c, `SELECT timestamp, item_name, item_sku, quantity, unit, type FROM stock_transactions ORDER BY timestamp DESC LIMIT 8`)
	defer txRows.Close()
	type txRow struct {
		Time string `json:"time"`
		Name string `json:"item_name"`
		SKU  string `json:"item_sku"`
		Qty  int    `json:"quantity"`
		Unit string `json:"unit"`
		Type string `json:"type"`
	}
	recent := []txRow{}
	for txRows.Next() {
		var t txRow
		var ts time.Time
		txRows.Scan(&ts, &t.Name, &t.SKU, &t.Qty, &t.Unit, &t.Type)
		t.Time = ts.In(time.FixedZone("WIB", 7*3600)).Format("02-01-2006 15:04")
		recent = append(recent, t)
	}

	catRows, _ := h.pool.Query(c, `SELECT category, COALESCE(sum(current_stock),0) FROM inventory_items GROUP BY category`)
	defer catRows.Close()
	stockByCat := map[string]int{}
	for catRows.Next() {
		var cat string
		var n int
		catRows.Scan(&cat, &n)
		stockByCat[cat] = n
	}

	c.JSON(http.StatusOK, gin.H{
		"total_items": totalItems,
		"total_stock": totalStock,
		"low_stock":   lowStock,
		"recent_tx":   recent,
		"stock_by_cat": stockByCat,
	})
}

// helpers

func generateSKU(h *ItemHandler, c *gin.Context, category string) string {
	var slug string
	h.pool.QueryRow(c, `SELECT id FROM categories WHERE name=$1`, category).Scan(&slug)
	parts := strings.Split(slug, "-")
	prefix := strings.ToUpper(parts[0])
	if len(prefix) > 4 {
		prefix = prefix[:4]
	}
	year := strconv.Itoa(getCurrentYear())
	pattern := prefix + "-" + year + "-%"
	var last string
	h.pool.QueryRow(c, `SELECT sku FROM inventory_items WHERE sku LIKE $1 ORDER BY sku DESC LIMIT 1`, pattern).Scan(&last)
	serial := 1
	if len(last) > len(prefix)+6 {
		if n, e := strconv.Atoi(last[len(prefix)+6:]); e == nil {
			serial = n + 1
		}
	}
	return fmt.Sprintf("%s-%s-%03d", prefix, year, serial)
}

func getCurrentYear() int {
	return time.Now().Year()
}
