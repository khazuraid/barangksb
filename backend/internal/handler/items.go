package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

type rowQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func resolveLocationCode(ctx context.Context, q rowQuerier, locationName, fallback string) string {
	loc := strings.TrimSpace(locationName)
	if loc != "" {
		var code string
		if err := q.QueryRow(ctx, `SELECT COALESCE(code, '') FROM locations WHERE name=$1 LIMIT 1`, loc).Scan(&code); err == nil && strings.TrimSpace(code) != "" {
			return strings.TrimSpace(code)
		}
	}
	return strings.TrimSpace(fallback)
}

func resolveFixedAssetUnits(initialUnits, currentStock int32) (int, error) {
	units := int(initialUnits)
	if units <= 0 {
		units = int(currentStock)
	}
	if units < 1 {
		units = 1
	}
	if units > 500 {
		return 0, errors.New("jumlah unit maksimal 500 per permintaan")
	}
	return units, nil
}

func unitRegisterFor(baseReg string, unitNum, totalUnits int) string {
	baseReg = strings.TrimSpace(baseReg)
	if totalUnits == 1 && baseReg != "" {
		return baseReg
	}
	startReg := 1
	if baseReg != "" {
		if parsed, err := strconv.Atoi(strings.TrimLeft(baseReg, "0")); err == nil && parsed > 0 {
			startReg = parsed
		}
	}
	return fmt.Sprintf("%04d", startReg+unitNum-1)
}

func unitBarcodeFor(sku string, unitNum, totalUnits int) string {
	sku = strings.TrimSpace(sku)
	if totalUnits == 1 && sku != "" {
		return sku
	}
	return fmt.Sprintf("%s-U%d", sku, unitNum)
}

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
	InitialUnits    int32    `json:"initial_units,omitempty"`
}

type ItemUnit struct {
	ID              string   `json:"id"`
	ItemID          string   `json:"item_id"`
	UnitNumber      int      `json:"unit_number"`
	RegisterNumber  string   `json:"register_number"`
	Barcode         string   `json:"barcode"`
	SerialNumber    string   `json:"serial_number"`
	ConditionStatus string   `json:"condition_status"`
	Location        string   `json:"location"`
	LocationCode    string   `json:"location_code"`
	PhotoURL        string   `json:"photo_url"`
	GeoLat          *float64 `json:"geo_lat"`
	GeoLng          *float64 `json:"geo_lng"`
	GeoAcc          *float64 `json:"geo_acc"`
	GeoName         string   `json:"geo_name"`
	Notes           string   `json:"notes"`
	CreatedAt       string   `json:"created_at,omitempty"`
	UpdatedAt       string   `json:"updated_at,omitempty"`
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
	if err := h.pool.QueryRow(c, `SELECT count(*) FROM inventory_items`+where, args...).Scan(&total); err != nil {
		internalError(c, "gagal menghitung total barang", err)
		return
	}

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
	queryStr := fmt.Sprintf(`
		SELECT
			id::text,
			COALESCE(sku, ''),
			COALESCE(name, ''),
			COALESCE(category, 'Lainnya'),
			COALESCE(location, 'Gudang Utama'),
			COALESCE(current_stock, 0),
			COALESCE(min_stock, 0),
			COALESCE(unit, 'unit'),
			COALESCE(price_per_unit, 0),
			COALESCE(condition_status, 'Baik (B)'),
			COALESCE(is_available, true),
			COALESCE(photo_url, ''),
			geo_lat,
			geo_lng,
			COALESCE(geo_name, ''),
			COALESCE(track_stock, true),
			COALESCE(item_code, ''),
			COALESCE(register_number, ''),
			COALESCE(size, ''),
			COALESCE(material, ''),
			COALESCE(location_code, ''),
			COALESCE(merk, ''),
			COALESCE(type_model, ''),
			COALESCE(serial_number, ''),
			COALESCE(procurement_year, ''),
			COALESCE(funding_source, ''),
			COALESCE(description, '')
		FROM inventory_items%s %s LIMIT $%d OFFSET $%d`, where, orderClause, len(args)-1, len(args))

	rows, err := h.pool.Query(c, queryStr, args...)
	if err != nil {
		internalError(c, "gagal mengambil daftar barang", err)
		return
	}
	defer rows.Close()

	type itemRow struct {
		ID              string   `json:"id"`
		SKU             string   `json:"sku"`
		Name            string   `json:"name"`
		Category        string   `json:"category"`
		Location        string   `json:"location"`
		LocationCode    string   `json:"location_code"`
		CurrentStock    int32    `json:"current_stock"`
		MinStock        int32    `json:"min_stock"`
		Unit            string   `json:"unit"`
		PricePerUnit    int64    `json:"price_per_unit"`
		Condition       string   `json:"condition_status"`
		IsAvailable     bool     `json:"is_available"`
		PhotoURL        string   `json:"photo_url"`
		ItemCode        string   `json:"item_code"`
		RegisterNumber  string   `json:"register_number"`
		Size            string   `json:"size"`
		Material        string   `json:"material"`
		Merk            string   `json:"merk"`
		TypeModel       string   `json:"type_model"`
		SerialNumber    string   `json:"serial_number"`
		ProcurementYear string   `json:"procurement_year"`
		FundingSource   string   `json:"funding_source"`
		Description     string   `json:"description"`
		GeoLat          *float64 `json:"geo_lat"`
		GeoLng          *float64 `json:"geo_lng"`
		GeoName         string   `json:"geo_name"`
		TrackStock      bool     `json:"track_stock"`
	}
	items := make([]itemRow, 0, perPage)
	for rows.Next() {
		var it itemRow
		if err := rows.Scan(
			&it.ID, &it.SKU, &it.Name, &it.Category, &it.Location,
			&it.CurrentStock, &it.MinStock, &it.Unit, &it.PricePerUnit,
			&it.Condition, &it.IsAvailable, &it.PhotoURL, &it.GeoLat,
			&it.GeoLng, &it.GeoName, &it.TrackStock, &it.ItemCode,
			&it.RegisterNumber, &it.Size, &it.Material, &it.LocationCode,
			&it.Merk, &it.TypeModel, &it.SerialNumber, &it.ProcurementYear,
			&it.FundingSource, &it.Description,
		); err != nil {
			internalError(c, "gagal membaca data barang", err)
			return
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		internalError(c, "gagal membaca daftar barang", err)
		return
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "data barang tidak valid"})
		return
	}
	safePhoto, err := SanitizePhotoURL(req.PhotoURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.PhotoURL = safePhoto
	if req.CurrentStock < 0 || req.MinStock < 0 || req.PricePerUnit < 0 || req.InitialUnits < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stok, stok minimum, harga, dan jumlah unit tidak boleh negatif"})
		return
	}
	customSKU := strings.TrimSpace(req.SKU)
	trackStock := true
	if req.TrackStock != nil {
		trackStock = *req.TrackStock
	}
	initialStock := req.CurrentStock
	unitsCount := 0
	if !trackStock {
		unitsCount, err = resolveFixedAssetUnits(req.InitialUnits, req.CurrentStock)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		initialStock = int32(unitsCount)
	}

	var id, sku string
	maxAttempts := 1
	if customSKU == "" {
		maxAttempts = 3
	}
	for attempt := 0; attempt < maxAttempts; attempt++ {
		tx, txErr := h.pool.Begin(c)
		if txErr != nil {
			internalError(c, "gagal memulai transaksi penyimpanan barang", txErr)
			return
		}
		locCode := resolveLocationCode(c, tx, req.Location, req.LocationCode)
		if customSKU != "" {
			sku = customSKU
		} else {
			sku = generateSKU(tx, c, req.Category)
		}

		err = tx.QueryRow(c, `INSERT INTO inventory_items (sku, name, category, location, location_code, current_stock, min_stock, unit, price_per_unit, description, photo_url, merk, type_model, serial_number, procurement_year, condition_status, funding_source, distributor, akl_akd, size, material, item_code, register_number, geo_lat, geo_lng, geo_acc, geo_name, track_stock) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28) RETURNING id::text`,
			sku, req.Name, req.Category, req.Location, nullable(locCode), initialStock, req.MinStock, req.Unit, req.PricePerUnit,
			nullable(req.Description), nullable(req.PhotoURL), nullable(req.Merk), nullable(req.TypeModel), nullable(req.SerialNumber),
			nullable(req.ProcurementYear), nullable(req.ConditionStatus), nullable(req.FundingSource), nullable(req.Distributor), nullable(req.AklAkd),
			nullable(req.Size), nullable(req.Material), nullable(req.ItemCode), nullable(req.RegisterNumber),
			req.GeoLat, req.GeoLng, req.GeoAcc, nullable(req.GeoName), trackStock).
			Scan(&id)
		if err != nil {
			_ = tx.Rollback(c)
			if customSKU == "" && isUniqueViolation(err) && attempt+1 < maxAttempts {
				continue
			}
			if isUniqueViolation(err) {
				c.JSON(http.StatusConflict, gin.H{"error": "SKU sudah digunakan, silakan gunakan SKU lain"})
				return
			}
			internalError(c, "gagal menyimpan barang baru", err)
			return
		}

		if !trackStock {
			for i := 1; i <= unitsCount; i++ {
				unitNum := i
				regStr := unitRegisterFor(req.RegisterNumber, unitNum, unitsCount)
				unitBarcode := unitBarcodeFor(sku, unitNum, unitsCount)
				var unitPhoto any
				var uLat, uLng, uAcc *float64
				var uGeoName any
				if i == 1 && req.PhotoURL != "" {
					unitPhoto = nullable(req.PhotoURL)
					uLat = req.GeoLat
					uLng = req.GeoLng
					uAcc = req.GeoAcc
					uGeoName = nullable(req.GeoName)
				}
				var uSerial any
				if i == 1 && req.SerialNumber != "" {
					uSerial = nullable(req.SerialNumber)
				}
				if _, err = tx.Exec(c, `INSERT INTO item_units (item_id, unit_number, register_number, barcode, serial_number, condition_status, location, location_code, photo_url, geo_lat, geo_lng, geo_acc, geo_name) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
					id, unitNum, regStr, unitBarcode, uSerial, nullable(req.ConditionStatus), req.Location, nullable(locCode), unitPhoto, uLat, uLng, uAcc, uGeoName); err != nil {
					_ = tx.Rollback(c)
					internalError(c, "gagal menyimpan unit fisik barang", err)
					return
				}
			}
		}

		if err = tx.Commit(c); err != nil {
			internalError(c, "gagal menyimpan barang baru", err)
			return
		}
		break
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "sku": sku})
}

func (h *ItemHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req itemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data barang tidak valid"})
		return
	}
	safePhoto, err := SanitizePhotoURL(req.PhotoURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.PhotoURL = safePhoto
	if req.CurrentStock < 0 || req.MinStock < 0 || req.PricePerUnit < 0 || req.InitialUnits < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stok, stok minimum, harga, dan jumlah unit tidak boleh negatif"})
		return
	}
	trackStock := true
	if req.TrackStock != nil {
		trackStock = *req.TrackStock
	}

	tx, err := h.pool.Begin(c)
	if err != nil {
		internalError(c, "gagal memulai transaksi pembaruan barang", err)
		return
	}
	defer tx.Rollback(c)

	var existingSKU string
	if err := tx.QueryRow(c, `SELECT COALESCE(sku, '') FROM inventory_items WHERE id::text=$1`, id).Scan(&existingSKU); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "barang tidak ditemukan"})
			return
		}
		internalError(c, "gagal memuat data barang", err)
		return
	}
	sku := strings.TrimSpace(req.SKU)
	if sku == "" {
		sku = existingSKU
	}
	req.LocationCode = resolveLocationCode(c, tx, req.Location, req.LocationCode)

	targetStock := req.CurrentStock
	if !trackStock {
		targetUnits, err := resolveFixedAssetUnits(req.InitialUnits, req.CurrentStock)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var existingUnits, maxUnitNum int
		if err := tx.QueryRow(c, `SELECT COUNT(*), COALESCE(MAX(unit_number), 0) FROM item_units WHERE item_id::text=$1`, id).Scan(&existingUnits, &maxUnitNum); err != nil {
			internalError(c, "gagal memeriksa unit fisik barang", err)
			return
		}
		if targetUnits > existingUnits {
			for k := 1; k <= targetUnits-existingUnits; k++ {
				unitNum := maxUnitNum + k
				regStr := unitRegisterFor(req.RegisterNumber, unitNum, targetUnits)
				unitBarcode := unitBarcodeFor(sku, unitNum, targetUnits)
				var unitPhoto any
				var uLat, uLng, uAcc *float64
				var uGeoName, uSerial any
				if existingUnits == 0 && k == 1 {
					if req.PhotoURL != "" {
						unitPhoto = nullable(req.PhotoURL)
						uLat = req.GeoLat
						uLng = req.GeoLng
						uAcc = req.GeoAcc
						uGeoName = nullable(req.GeoName)
					}
					if req.SerialNumber != "" {
						uSerial = nullable(req.SerialNumber)
					}
				}
				if _, err := tx.Exec(c, `INSERT INTO item_units (item_id, unit_number, register_number, barcode, serial_number, condition_status, location, location_code, photo_url, geo_lat, geo_lng, geo_acc, geo_name) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
					id, unitNum, regStr, unitBarcode, uSerial, nullable(req.ConditionStatus), req.Location, nullable(req.LocationCode), unitPhoto, uLat, uLng, uAcc, uGeoName); err != nil {
					internalError(c, "gagal menambah unit fisik barang", err)
					return
				}
			}
		} else if targetUnits < existingUnits {
			toRemove := existingUnits - targetUnits
			tag, err := tx.Exec(c, `WITH victims AS (
				SELECT id FROM item_units
				WHERE item_id::text=$1
				  AND COALESCE(photo_url, '') = ''
				  AND COALESCE(serial_number, '') = ''
				  AND COALESCE(notes, '') = ''
				ORDER BY unit_number DESC, created_at DESC
				LIMIT $2
			) DELETE FROM item_units WHERE id IN (SELECT id FROM victims)`, id, toRemove)
			if err != nil {
				internalError(c, "gagal mengurangi unit fisik barang", err)
				return
			}
			if tag.RowsAffected() < int64(toRemove) {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("tidak dapat mengurangi jumlah barang menjadi %d karena masih ada unit fisik yang memiliki foto, nomor seri, atau catatan; hapus unit spesifik di tab Daftar Unit Fisik", targetUnits)})
				return
			}
		}
		targetStock = int32(targetUnits)
	}

	_, err = tx.Exec(c, `UPDATE inventory_items SET sku=$2, name=$3, category=$4, location=$5, location_code=$6, current_stock=$7, min_stock=$8, unit=$9, price_per_unit=$10, description=$11, photo_url=$12, merk=$13, type_model=$14, serial_number=$15, procurement_year=$16, condition_status=$17, funding_source=$18, distributor=$19, akl_akd=$20, size=$21, material=$22, item_code=$23, register_number=$24, geo_lat=$25, geo_lng=$26, geo_acc=$27, geo_name=$28, track_stock=$29, updated_at=now() WHERE id::text=$1`,
		id, sku, req.Name, req.Category, req.Location, nullable(req.LocationCode), targetStock, req.MinStock, req.Unit, req.PricePerUnit,
		nullable(req.Description), nullable(req.PhotoURL), nullable(req.Merk), nullable(req.TypeModel), nullable(req.SerialNumber),
		nullable(req.ProcurementYear), nullable(req.ConditionStatus), nullable(req.FundingSource), nullable(req.Distributor), nullable(req.AklAkd),
		nullable(req.Size), nullable(req.Material), nullable(req.ItemCode), nullable(req.RegisterNumber),
		req.GeoLat, req.GeoLng, req.GeoAcc, nullable(req.GeoName), trackStock)
	if err != nil {
		internalError(c, "gagal memperbarui barang", err)
		return
	}
	if err := tx.Commit(c); err != nil {
		internalError(c, "gagal menyimpan pembaruan barang", err)
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

	// 1. Gather all photo URLs associated with this item before deleting DB rows
	var photosToDelete []string
	var itemPhoto *string
	_ = h.pool.QueryRow(c, `SELECT photo_url FROM inventory_items WHERE id::text=$1`, id).Scan(&itemPhoto)
	if itemPhoto != nil && *itemPhoto != "" {
		photosToDelete = append(photosToDelete, *itemPhoto)
	}

	txPhotos, errTx := h.pool.Query(c, `SELECT photo_url FROM stock_transactions WHERE (item_id::text=$1 OR item_sku IN (SELECT sku FROM inventory_items WHERE id::text=$1)) AND photo_url IS NOT NULL`, id)
	if errTx == nil {
		for txPhotos.Next() {
			var p string
			if txPhotos.Scan(&p) == nil && p != "" {
				photosToDelete = append(photosToDelete, p)
			}
		}
		txPhotos.Close()
	}

	maintPhotos, errMaint := h.pool.Query(c, `SELECT photo_url FROM item_maintenances WHERE item_id::text=$1 AND photo_url IS NOT NULL`, id)
	if errMaint == nil {
		for maintPhotos.Next() {
			var p string
			if maintPhotos.Scan(&p) == nil && p != "" {
				photosToDelete = append(photosToDelete, p)
			}
		}
		maintPhotos.Close()
	}

	unitPhotos, errUnitPhotos := h.pool.Query(c, `SELECT photo_url FROM item_units WHERE item_id::text=$1 AND photo_url IS NOT NULL AND photo_url != ''`, id)
	if errUnitPhotos == nil {
		for unitPhotos.Next() {
			var up string
			if unitPhotos.Scan(&up) == nil && up != "" {
				photosToDelete = append(photosToDelete, up)
			}
		}
		unitPhotos.Close()
	}

	// 2. Delete related transactions, maintenance records, and the item from database atomically
	tx, err := h.pool.Begin(c)
	if err != nil {
		internalError(c, "gagal memulai transaksi hapus barang", err)
		return
	}
	defer tx.Rollback(c)

	if _, err := tx.Exec(c, `DELETE FROM stock_transactions WHERE item_id::text=$1 OR item_sku IN (SELECT sku FROM inventory_items WHERE id::text=$1)`, id); err != nil {
		internalError(c, "gagal menghapus riwayat transaksi barang", err)
		return
	}
	if _, err := tx.Exec(c, `DELETE FROM item_maintenances WHERE item_id::text=$1`, id); err != nil {
		internalError(c, "gagal menghapus riwayat pemeliharaan barang", err)
		return
	}
	tag, err := tx.Exec(c, `DELETE FROM inventory_items WHERE id::text=$1`, id)
	if err != nil {
		internalError(c, "gagal menghapus barang", err)
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "barang tidak ditemukan"})
		return
	}
	if err := tx.Commit(c); err != nil {
		internalError(c, "gagal menyimpan penghapusan barang", err)
		return
	}

	for _, p := range photosToDelete {
		DeleteUploadedPhoto(c.Request.Context(), h.pool, p)
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *ItemHandler) Dashboard(c *gin.Context) {
	var totalItems, totalStock int
	h.pool.QueryRow(c, `SELECT count(*), COALESCE(sum(current_stock),0) FROM inventory_items`).Scan(&totalItems, &totalStock)

	lowRows, err := h.pool.Query(c, `SELECT name, sku, current_stock, min_stock, unit, COALESCE(location, '') FROM inventory_items WHERE COALESCE(track_stock, true) = true AND current_stock <= min_stock ORDER BY current_stock LIMIT 10`)
	if err != nil {
		internalError(c, "gagal mengambil data stok menipis", err)
		return
	}
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

	txRows, err := h.pool.Query(c, `SELECT timestamp, item_name, item_sku, quantity, unit, type FROM stock_transactions ORDER BY timestamp DESC LIMIT 8`)
	if err != nil {
		internalError(c, "gagal mengambil transaksi terbaru", err)
		return
	}
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

	catRows, err := h.pool.Query(c, `SELECT category, COALESCE(sum(current_stock),0) FROM inventory_items GROUP BY category`)
	if err != nil {
		internalError(c, "gagal mengambil ringkasan kategori", err)
		return
	}
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

func generateSKU(q rowQuerier, ctx context.Context, category string) string {
	var slug string
	_ = q.QueryRow(ctx, `SELECT id FROM categories WHERE name=$1`, category).Scan(&slug)
	parts := strings.Split(slug, "-")
	prefix := strings.ToUpper(parts[0])
	if len(prefix) > 4 {
		prefix = prefix[:4]
	}
	if prefix == "" {
		prefix = "BRG"
	}
	year := strconv.Itoa(getCurrentYear())
	prefixYear := prefix + "-" + year + "-"
	pattern := prefixYear + "%"
	var last string
	_ = q.QueryRow(ctx, `SELECT sku FROM inventory_items WHERE sku LIKE $1 ORDER BY LENGTH(sku) DESC, sku DESC LIMIT 1`, pattern).Scan(&last)
	serial := 1
	if strings.HasPrefix(last, prefixYear) {
		if n, e := strconv.Atoi(last[len(prefixYear):]); e == nil {
			serial = n + 1
		}
	}
	return fmt.Sprintf("%s-%s-%03d", prefix, year, serial)
}

func getCurrentYear() int {
	return time.Now().Year()
}

// Physical Units (Item Units) Handlers

func (h *ItemHandler) ListUnits(c *gin.Context) {
	id := c.Param("id")
	rows, err := h.pool.Query(c, `SELECT id::text, item_id::text, unit_number, COALESCE(register_number, ''), COALESCE(barcode, ''), COALESCE(serial_number, ''), COALESCE(condition_status, 'Baik (B)'), COALESCE(location, ''), COALESCE(location_code, ''), COALESCE(photo_url, ''), geo_lat, geo_lng, geo_acc, COALESCE(geo_name, ''), COALESCE(notes, ''), created_at::text, updated_at::text FROM item_units WHERE item_id::text=$1 ORDER BY unit_number ASC, created_at ASC`, id)
	if err != nil {
		internalError(c, "gagal mengambil daftar unit barang", err)
		return
	}
	defer rows.Close()

	units := []ItemUnit{}
	for rows.Next() {
		var u ItemUnit
		err := rows.Scan(&u.ID, &u.ItemID, &u.UnitNumber, &u.RegisterNumber, &u.Barcode, &u.SerialNumber, &u.ConditionStatus, &u.Location, &u.LocationCode, &u.PhotoURL, &u.GeoLat, &u.GeoLng, &u.GeoAcc, &u.GeoName, &u.Notes, &u.CreatedAt, &u.UpdatedAt)
		if err == nil {
			units = append(units, u)
		}
	}
	c.JSON(http.StatusOK, gin.H{"data": units, "total": len(units)})
}

func (h *ItemHandler) CreateUnit(c *gin.Context) {
	id := c.Param("id")
	var req ItemUnit
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	safePhoto, err := SanitizePhotoURL(req.PhotoURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.PhotoURL = safePhoto

	tx, err := h.pool.Begin(c)
	if err != nil {
		internalError(c, "gagal memulai transaksi tambah unit", err)
		return
	}
	defer tx.Rollback(c)

	// Lock parent item so concurrent unit additions cannot assign duplicate unit_number
	var itemSKU, itemLoc, itemLocCode, itemCond string
	err = tx.QueryRow(c, `SELECT sku, location, COALESCE(location_code, ''), COALESCE(condition_status, 'Baik (B)') FROM inventory_items WHERE id::text=$1 FOR UPDATE`, id).
		Scan(&itemSKU, &itemLoc, &itemLocCode, &itemCond)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "item tidak ditemukan"})
			return
		}
		internalError(c, "gagal memuat data barang", err)
		return
	}

	var maxUnitNum int
	if err := tx.QueryRow(c, `SELECT COALESCE(MAX(unit_number), 0) FROM item_units WHERE item_id::text=$1`, id).Scan(&maxUnitNum); err != nil {
		internalError(c, "gagal menghitung nomor unit", err)
		return
	}
	nextUnitNum := maxUnitNum + 1

	if req.UnitNumber <= 0 {
		req.UnitNumber = nextUnitNum
	}
	if strings.TrimSpace(req.Barcode) == "" {
		req.Barcode = fmt.Sprintf("%s-U%d", itemSKU, req.UnitNumber)
	}
	if strings.TrimSpace(req.RegisterNumber) == "" {
		req.RegisterNumber = fmt.Sprintf("%04d", req.UnitNumber)
	}
	if strings.TrimSpace(req.Location) == "" {
		req.Location = itemLoc
	}
	req.LocationCode = resolveLocationCode(c, tx, req.Location, req.LocationCode)
	if strings.TrimSpace(req.LocationCode) == "" {
		req.LocationCode = itemLocCode
	}
	if strings.TrimSpace(req.ConditionStatus) == "" {
		req.ConditionStatus = itemCond
	}

	var unitID string
	err = tx.QueryRow(c, `INSERT INTO item_units (item_id, unit_number, register_number, barcode, serial_number, condition_status, location, location_code, photo_url, geo_lat, geo_lng, geo_acc, geo_name, notes) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING id::text`,
		id, req.UnitNumber, req.RegisterNumber, req.Barcode, nullable(req.SerialNumber), req.ConditionStatus, req.Location, nullable(req.LocationCode), nullable(req.PhotoURL), req.GeoLat, req.GeoLng, req.GeoAcc, nullable(req.GeoName), nullable(req.Notes)).
		Scan(&unitID)
	if err != nil {
		internalError(c, "gagal menambah unit barang", err)
		return
	}

	// Sync current_stock to total physical units count
	if _, err = tx.Exec(c, `UPDATE inventory_items SET current_stock = (SELECT COUNT(*) FROM item_units WHERE item_id::text=$1), updated_at = now() WHERE id::text=$1`, id); err != nil {
		internalError(c, "gagal menyinkronkan jumlah unit barang", err)
		return
	}
	if err = tx.Commit(c); err != nil {
		internalError(c, "gagal menyimpan unit barang", err)
		return
	}

	req.ID = unitID
	req.ItemID = id
	c.JSON(http.StatusCreated, req)
}

func (h *ItemHandler) UpdateUnit(c *gin.Context) {
	id := c.Param("id")
	unitID := c.Param("unitId")
	var req ItemUnit
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data unit tidak valid"})
		return
	}
	safePhoto, err := SanitizePhotoURL(req.PhotoURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.PhotoURL = safePhoto
	req.LocationCode = resolveLocationCode(c, h.pool, req.Location, req.LocationCode)

	_, err = h.pool.Exec(c, `UPDATE item_units SET register_number=$3, barcode=$4, serial_number=$5, condition_status=$6, location=$7, location_code=$8, photo_url=$9, geo_lat=$10, geo_lng=$11, geo_acc=$12, geo_name=$13, notes=$14, updated_at=now() WHERE id::text=$1 AND item_id::text=$2`,
		unitID, id, req.RegisterNumber, req.Barcode, nullable(req.SerialNumber), req.ConditionStatus, req.Location, nullable(req.LocationCode), nullable(req.PhotoURL), req.GeoLat, req.GeoLng, req.GeoAcc, nullable(req.GeoName), nullable(req.Notes))
	if err != nil {
		internalError(c, "gagal memperbarui unit barang", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *ItemHandler) DeleteUnit(c *gin.Context) {
	id := c.Param("id")
	unitID := c.Param("unitId")

	// Fetch photo before deleting DB row, then delete physical file only after DB delete succeeds
	var photo string
	_ = h.pool.QueryRow(c, `SELECT COALESCE(photo_url, '') FROM item_units WHERE id::text=$1 AND item_id::text=$2`, unitID, id).Scan(&photo)

	tag, err := h.pool.Exec(c, `DELETE FROM item_units WHERE id::text=$1 AND item_id::text=$2`, unitID, id)
	if err != nil {
		internalError(c, "gagal menghapus unit barang", err)
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "unit tidak ditemukan"})
		return
	}

	// Sync current_stock to total physical units count
	_, _ = h.pool.Exec(c, `UPDATE inventory_items SET current_stock = (SELECT COUNT(*) FROM item_units WHERE item_id::text=$1) WHERE id::text=$1`, id)

	if photo != "" {
		DeleteUploadedPhoto(c.Request.Context(), h.pool, photo)
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
