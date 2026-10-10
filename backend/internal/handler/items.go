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
	ID              string   `json:"id,omitempty"`
	SKU             string   `json:"sku"`
	Name            string   `json:"name" binding:"required"`
	Category        string   `json:"category" binding:"required"`
	Location        string   `json:"location" binding:"required"`
	LocationCode    string   `json:"location_code"`
	CurrentStock    int32    `json:"current_stock"`
	MinStock        int32    `json:"min_stock"`
	Unit            string   `json:"unit"`
	PricePerUnit    int64    `json:"price_per_unit"`
	Description     string   `json:"description"`
	PhotoURL        string   `json:"photo_url"`
	Merk            string   `json:"merk"`
	TypeModel       string   `json:"type_model"`
	SerialNumber    string   `json:"serial_number"`
	ProcurementYear string   `json:"procurement_year"`
	ConditionStatus string   `json:"condition_status"`
	FundingSource   string   `json:"funding_source"`
	Distributor     string   `json:"distributor"`
	AklAkd          string   `json:"akl_akd"`
	Size            string   `json:"size"`
	Material        string   `json:"material"`
	ItemCode        string   `json:"item_code"`
	RegisterNumber  string   `json:"register_number"`
	GeoLat          *float64 `json:"geo_lat"`
	GeoLng          *float64 `json:"geo_lng"`
	GeoAcc          *float64 `json:"geo_acc"`
	GeoName         string   `json:"geo_name"`
	IsAvailable     bool     `json:"is_available"`
	TrackStock      *bool    `json:"track_stock"`
}

func (h *ItemHandler) List(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	cat := strings.TrimSpace(c.Query("cat"))
	loc := strings.TrimSpace(c.Query("loc"))
	sortBy := strings.TrimSpace(c.DefaultQuery("sort", "newest"))
	lowStockOnly := c.Query("low_stock") == "true"
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 { page = 1 }
	perPage := configPerPage(c)
	offset := (page - 1) * perPage

	where := " WHERE TRUE"
	args := []any{}
	if q != "" {
		where += fmt.Sprintf(` AND (name ILIKE $%d OR sku ILIKE $%d OR location ILIKE $%d OR location_code ILIKE $%d OR item_code ILIKE $%d OR register_number ILIKE $%d OR serial_number ILIKE $%d OR COALESCE(merk, '') ILIKE $%d OR COALESCE(type_model, '') ILIKE $%d OR COALESCE(description, '') ILIKE $%d)`, len(args)+1, len(args)+1, len(args)+1, len(args)+1, len(args)+1, len(args)+1, len(args)+1, len(args)+1, len(args)+1, len(args)+1)
		args = append(args, "%"+q+"%")
	}
	if cat != "" { where += fmt.Sprintf(` AND category = $%d`, len(args)+1); args = append(args, cat) }
	if loc != "" { where += fmt.Sprintf(` AND location = $%d`, len(args)+1); args = append(args, loc) }
	if lowStockOnly { where += ` AND current_stock <= min_stock AND COALESCE(track_stock, true) = true` }

	var total int
	h.pool.QueryRow(c, `SELECT count(*) FROM inventory_items`+where, args...).Scan(&total)

	orderClause := "ORDER BY created_at DESC, name ASC"
	switch sortBy {
	case "name_asc", "name":
		orderClause = "ORDER BY name ASC"
	case "name_desc":
		orderClause = "ORDER BY name DESC"
	case "stock_asc":
		orderClause = "ORDER BY current_stock ASC, name ASC"
	case "stock_desc":
		orderClause = "ORDER BY current_stock DESC, name ASC"
	case "oldest":
		orderClause = "ORDER BY created_at ASC"
	case "newest":
		orderClause = "ORDER BY created_at DESC, name ASC"
	}

	args = append(args, perPage, offset)
	rows, err := h.pool.Query(c, fmt.Sprintf(`SELECT id::text, sku, name, category, location, current_stock, min_stock, unit, price_per_unit, condition_status, is_available, photo_url, geo_lat, geo_lng, COALESCE(geo_name, ''), COALESCE(track_stock, true), COALESCE(item_code, ''), COALESCE(register_number, ''), COALESCE(size, ''), COALESCE(material, ''), COALESCE(location_code, ''), COALESCE(merk, ''), COALESCE(type_model, '') FROM inventory_items%s %s LIMIT $%d OFFSET $%d`, where, orderClause, len(args)-1, len(args)), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type itemRow struct {
		ID             string   `json:"id"`
		SKU            string   `json:"sku"`
		Name           string   `json:"name"`
		Category       string   `json:"category"`
		Location       string   `json:"location"`
		LocationCode   string   `json:"location_code"`
		CurrentStock   int32    `json:"current_stock"`
		MinStock       int32    `json:"min_stock"`
		Unit           string   `json:"unit"`
		PricePerUnit   int64    `json:"price_per_unit"`
		Condition      string   `json:"condition_status"`
		IsAvailable    bool     `json:"is_available"`
		PhotoURL       string   `json:"photo_url"`
		ItemCode       string   `json:"item_code"`
		RegisterNumber string   `json:"register_number"`
		Size           string   `json:"size"`
		Material       string   `json:"material"`
		Merk           string   `json:"merk"`
		TypeModel      string   `json:"type_model"`
		GeoLat         *float64 `json:"geo_lat"`
		GeoLng         *float64 `json:"geo_lng"`
		GeoName        string   `json:"geo_name"`
		TrackStock     bool     `json:"track_stock"`
	}
	items := []itemRow{}
	for rows.Next() {
		var it itemRow
		rows.Scan(&it.ID, &it.SKU, &it.Name, &it.Category, &it.Location, &it.CurrentStock, &it.MinStock, &it.Unit, &it.PricePerUnit, &it.Condition, &it.IsAvailable, &it.PhotoURL, &it.GeoLat, &it.GeoLng, &it.GeoName, &it.TrackStock, &it.ItemCode, &it.RegisterNumber, &it.Size, &it.Material, &it.LocationCode, &it.Merk, &it.TypeModel)
		items = append(items, it)
	}

	pages := total / perPage
	if total%perPage > 0 { pages++ }
	c.JSON(http.StatusOK, paginatedResp{Data: items, Total: total, Page: page, PerPage: perPage, Pages: pages})
}

func (h *ItemHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" || id == "undefined" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id barang tidak valid"})
		return
	}
	var it itemReq
	var trackStock bool
	err := h.pool.QueryRow(c, `SELECT id::text, sku, name, category, location, current_stock, min_stock, unit, COALESCE(price_per_unit, 0), COALESCE(description, ''), COALESCE(photo_url, ''), COALESCE(merk, ''), COALESCE(type_model, ''), COALESCE(serial_number, ''), COALESCE(procurement_year, ''), COALESCE(condition_status, 'Berfungsi'), COALESCE(funding_source, ''), COALESCE(distributor, ''), COALESCE(akl_akd, ''), is_available, geo_lat, geo_lng, geo_acc, COALESCE(geo_name, ''), COALESCE(track_stock, true), COALESCE(item_code, ''), COALESCE(register_number, ''), COALESCE(size, ''), COALESCE(material, ''), COALESCE(location_code, '') FROM inventory_items WHERE id::text=$1`, id).
		Scan(&it.ID, &it.SKU, &it.Name, &it.Category, &it.Location, &it.CurrentStock, &it.MinStock, &it.Unit, &it.PricePerUnit, &it.Description, &it.PhotoURL, &it.Merk, &it.TypeModel, &it.SerialNumber, &it.ProcurementYear, &it.ConditionStatus, &it.FundingSource, &it.Distributor, &it.AklAkd, &it.IsAvailable, &it.GeoLat, &it.GeoLng, &it.GeoAcc, &it.GeoName, &trackStock, &it.ItemCode, &it.RegisterNumber, &it.Size, &it.Material, &it.LocationCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item tidak ditemukan"})
		return
	}
	it.TrackStock = &trackStock
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
	trackStock := true
	if req.TrackStock != nil {
		trackStock = *req.TrackStock
	}
	initialStock := req.CurrentStock
	if !trackStock {
		initialStock = 1
	}
	var id string
	err := h.pool.QueryRow(c, `INSERT INTO inventory_items (sku, name, category, location, location_code, current_stock, min_stock, unit, price_per_unit, description, photo_url, merk, type_model, serial_number, procurement_year, condition_status, funding_source, distributor, akl_akd, size, material, item_code, register_number, geo_lat, geo_lng, geo_acc, geo_name, track_stock) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28) RETURNING id::text`,
		sku, req.Name, req.Category, req.Location, nullable(req.LocationCode), initialStock, req.MinStock, req.Unit, req.PricePerUnit,
		nullable(req.Description), nullable(req.PhotoURL), nullable(req.Merk), nullable(req.TypeModel), nullable(req.SerialNumber),
		nullable(req.ProcurementYear), nullable(req.ConditionStatus), nullable(req.FundingSource), nullable(req.Distributor), nullable(req.AklAkd),
		nullable(req.Size), nullable(req.Material), nullable(req.ItemCode), nullable(req.RegisterNumber),
		req.GeoLat, req.GeoLng, req.GeoAcc, nullable(req.GeoName), trackStock).
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
	trackStock := true
	if req.TrackStock != nil {
		trackStock = *req.TrackStock
	}
	currentStock := req.CurrentStock
	if !trackStock {
		currentStock = 1
	}
	_, err := h.pool.Exec(c, `UPDATE inventory_items SET sku=$2, name=$3, category=$4, location=$5, location_code=$6, current_stock=$7, min_stock=$8, unit=$9, price_per_unit=$10, description=$11, photo_url=$12, merk=$13, type_model=$14, serial_number=$15, procurement_year=$16, condition_status=$17, funding_source=$18, distributor=$19, akl_akd=$20, size=$21, material=$22, item_code=$23, register_number=$24, geo_lat=$25, geo_lng=$26, geo_acc=$27, geo_name=$28, track_stock=$29, updated_at=now() WHERE id=$1`,
		id, req.SKU, req.Name, req.Category, req.Location, nullable(req.LocationCode), currentStock, req.MinStock, req.Unit, req.PricePerUnit,
		nullable(req.Description), nullable(req.PhotoURL), nullable(req.Merk), nullable(req.TypeModel), nullable(req.SerialNumber),
		nullable(req.ProcurementYear), nullable(req.ConditionStatus), nullable(req.FundingSource), nullable(req.Distributor), nullable(req.AklAkd),
		nullable(req.Size), nullable(req.Material), nullable(req.ItemCode), nullable(req.RegisterNumber),
		req.GeoLat, req.GeoLng, req.GeoAcc, nullable(req.GeoName), trackStock)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *ItemHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" || id == "undefined" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id barang tidak valid"})
		return
	}

	// 1. Gather all photo URLs associated with this item to delete physical files
	var itemPhoto *string
	_ = h.pool.QueryRow(c, `SELECT photo_url FROM inventory_items WHERE id::text=$1`, id).Scan(&itemPhoto)
	if itemPhoto != nil && *itemPhoto != "" {
		DeleteUploadedPhoto(*itemPhoto)
	}

	txPhotos, errTx := h.pool.Query(c, `SELECT photo_url FROM stock_transactions WHERE (item_id::text=$1 OR item_sku IN (SELECT sku FROM inventory_items WHERE id::text=$1)) AND photo_url IS NOT NULL`, id)
	if errTx == nil {
		for txPhotos.Next() {
			var p string
			txPhotos.Scan(&p)
			DeleteUploadedPhoto(p)
		}
		txPhotos.Close()
	}

	maintPhotos, errMaint := h.pool.Query(c, `SELECT photo_url FROM item_maintenances WHERE item_id::text=$1 AND photo_url IS NOT NULL`, id)
	if errMaint == nil {
		for maintPhotos.Next() {
			var p string
			maintPhotos.Scan(&p)
			DeleteUploadedPhoto(p)
		}
		maintPhotos.Close()
	}

	// 2. Delete related transactions, maintenance records, and the item from database
	_, _ = h.pool.Exec(c, `DELETE FROM stock_transactions WHERE item_id::text=$1 OR item_sku IN (SELECT sku FROM inventory_items WHERE id::text=$1)`, id)
	_, _ = h.pool.Exec(c, `DELETE FROM item_maintenances WHERE item_id::text=$1`, id)
	tag, err := h.pool.Exec(c, `DELETE FROM inventory_items WHERE id::text=$1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "barang tidak ditemukan"})
		return
	}
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
