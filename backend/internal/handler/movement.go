package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MovementHandler struct {
	pool *pgxpool.Pool
	tg   TelegramNotifier
}

func NewMovementHandler(pool *pgxpool.Pool, tg TelegramNotifier) *MovementHandler {
	return &MovementHandler{pool: pool, tg: tg}
}

type movementReq struct {
	ItemID   string `json:"item_id" binding:"required"`
	Quantity int32  `json:"quantity" binding:"required"`
	Person   string `json:"received_by"`
	Notes    string `json:"notes"`
	PhotoURL string `json:"photo_url"`
}

func (h *MovementHandler) StockIn(c *gin.Context)  { h.doMovement(c, "IN") }
func (h *MovementHandler) StockOut(c *gin.Context) { h.doMovement(c, "OUT") }

func (h *MovementHandler) doMovement(c *gin.Context, txType string) {
	var req movementReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "quantity harus > 0"})
		return
	}

	tx, err := h.pool.Begin(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(c)

	itemID := req.ItemID

	var name, unit string
	var prev, next int32

	if txType == "OUT" {
		err = tx.QueryRow(c,
			`UPDATE inventory_items SET current_stock = current_stock - $2, updated_at = now()
			 WHERE id::text = $1 AND current_stock >= $2
			 RETURNING name, unit, current_stock + $2, current_stock`,
			itemID, req.Quantity).Scan(&name, &unit, &prev, &next)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "stok tidak cukup atau barang tidak ditemukan"})
			return
		}
	} else {
		err = tx.QueryRow(c,
			`UPDATE inventory_items SET current_stock = current_stock + $2, updated_at = now()
			 WHERE id::text = $1
			 RETURNING name, unit, current_stock - $2, current_stock`,
			itemID, req.Quantity).Scan(&name, &unit, &prev, &next)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "barang tidak ditemukan"})
			return
		}
	}

	var txSKU string
	tx.QueryRow(c, `SELECT sku FROM inventory_items WHERE id::text=$1`, itemID).Scan(&txSKU)

	tx.Exec(c, `INSERT INTO stock_transactions (type, item_id, item_sku, item_name, quantity, unit, previous_stock, new_stock, received_by, notes, photo_url)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		txType, itemID, txSKU, name, req.Quantity, unit, prev, next,
		nullableS(req.Person), nullableS(req.Notes), nullableS(req.PhotoURL))

	tx.Commit(c)

	if h.tg != nil {
		go h.tg.NotifyMovement(context.Background(), txType, txSKU, name, req.Quantity, unit, prev, next, req.Person, req.Notes)
	}

	c.JSON(http.StatusOK, gin.H{"sku": txSKU, "name": name, "unit": unit, "previous": prev, "new": next})
}

func (h *MovementHandler) Adjust(c *gin.Context) {
	var req struct {
		ItemID   string `json:"item_id" binding:"required"`
		Quantity int32  `json:"quantity"`
		Person   string `json:"received_by"`
		Notes    string `json:"notes"`
		PhotoURL string `json:"photo_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var current int32
	var name, unit, sku string
	err := h.pool.QueryRow(c, `SELECT current_stock, name, unit, sku FROM inventory_items WHERE id::text=$1`, req.ItemID).
		Scan(&current, &name, &unit, &sku)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "barang tidak ditemukan"})
		return
	}

	diff := req.Quantity - current
	if diff == 0 {
		c.JSON(http.StatusOK, gin.H{"sku": sku, "name": name, "previous": current, "new": req.Quantity})
		return
	}

	txType := "ADJUST+"
	if diff < 0 { txType = "ADJUST-" }
	absDiff := diff
	if absDiff < 0 { absDiff = -absDiff }

	h.pool.Exec(c, `UPDATE inventory_items SET current_stock=$2, updated_at=now() WHERE id::text=$1`, req.ItemID, req.Quantity)
	h.pool.Exec(c, `INSERT INTO stock_transactions (type, item_id, item_sku, item_name, quantity, unit, previous_stock, new_stock, received_by, notes, photo_url)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		txType, req.ItemID, sku, name, absDiff, unit, current, req.Quantity,
		nullableS(req.Person), nullableS("Opname: "+req.Notes), nullableS(req.PhotoURL))

	if h.tg != nil {
		go h.tg.NotifyMovement(context.Background(), txType, sku, name, absDiff, unit, current, req.Quantity, req.Person, "Opname: "+req.Notes)
	}

	c.JSON(http.StatusOK, gin.H{"sku": sku, "name": name, "previous": current, "new": req.Quantity})
}

func (h *MovementHandler) ListTransactions(c *gin.Context) {
	sku := c.Query("sku")
	txType := c.Query("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 { page = 1 }
	perPage := configPerPage(c)
	offset := (page - 1) * perPage

	where := " WHERE TRUE"
	args := []any{}
	if sku != "" { where += fmtSprintf(` AND item_sku ILIKE $%d`, len(args)+1); args = append(args, "%"+sku+"%") }
	if txType != "" { where += fmtSprintf(` AND type = $%d`, len(args)+1); args = append(args, txType) }

	var total int
	h.pool.QueryRow(c, `SELECT count(*) FROM stock_transactions`+where, args...).Scan(&total)

	args = append(args, perPage, offset)
	rows, err := h.pool.Query(c, fmtSprintf(`SELECT id::text, timestamp, type, item_sku, item_name, quantity, unit, previous_stock, new_stock, received_by, notes, COALESCE(photo_url, '') FROM stock_transactions%s ORDER BY timestamp DESC LIMIT $%d OFFSET $%d`, where, len(args)-1, len(args)), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type txRow struct {
		ID       string `json:"id"`
		Time     string `json:"timestamp"`
		Type     string `json:"type"`
		SKU      string `json:"item_sku"`
		Name     string `json:"item_name"`
		Quantity int32  `json:"quantity"`
		Unit     string `json:"unit"`
		Prev     int32  `json:"previous_stock"`
		New      int32  `json:"new_stock"`
		By       string `json:"received_by"`
		Notes    string `json:"notes"`
		PhotoURL string `json:"photo_url"`
	}
	txs := []txRow{}
	for rows.Next() {
		var t txRow
		rows.Scan(&t.ID, &t.Time, &t.Type, &t.SKU, &t.Name, &t.Quantity, &t.Unit, &t.Prev, &t.New, &t.By, &t.Notes, &t.PhotoURL)
		txs = append(txs, t)
	}

	pages := total / perPage
	if total%perPage > 0 { pages++ }
	c.JSON(http.StatusOK, paginatedResp{Data: txs, Total: total, Page: page, PerPage: perPage, Pages: pages})
}

func (h *MovementHandler) Categories(c *gin.Context) {
	rows, _ := h.pool.Query(c, `SELECT id, name FROM categories ORDER BY name`)
	defer rows.Close()
	type cat struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	cats := []cat{}
	for rows.Next() {
		var c cat
		rows.Scan(&c.ID, &c.Name)
		cats = append(cats, c)
	}
	c.JSON(http.StatusOK, cats)
}

func (h *MovementHandler) Locations(c *gin.Context) {
	rows, _ := h.pool.Query(c, `SELECT id, name FROM locations ORDER BY name`)
	defer rows.Close()
	type loc struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	locs := []loc{}
	for rows.Next() {
		var l loc
		rows.Scan(&l.ID, &l.Name)
		locs = append(locs, l)
	}
	c.JSON(http.StatusOK, locs)
}

func nullableS(s string) any {
	if strings.TrimSpace(s) == "" { return nil }
	return s
}

var _ = strings.TrimSpace
