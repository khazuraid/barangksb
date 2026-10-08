package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MaintenanceHandler struct {
	pool *pgxpool.Pool
}

func NewMaintenanceHandler(pool *pgxpool.Pool) *MaintenanceHandler {
	return &MaintenanceHandler{pool: pool}
}

type maintenanceRow struct {
	ID                 string     `json:"id"`
	ItemID             string     `json:"item_id"`
	ServiceType        string     `json:"service_type"`
	ServiceDate        string     `json:"service_date"`
	VendorOrTechnician string     `json:"vendor_or_technician"`
	Cost               int64      `json:"cost"`
	Description        string     `json:"description"`
	NextServiceDate    *string    `json:"next_service_date"`
	PhotoURL           string     `json:"photo_url"`
	PerformedBy        string     `json:"performed_by"`
	CreatedAt          string     `json:"created_at"`
	ItemName           string     `json:"item_name,omitempty"`
	ItemSKU            string     `json:"item_sku,omitempty"`
}

type maintenanceReq struct {
	ServiceType        string `json:"service_type" binding:"required"`
	ServiceDate        string `json:"service_date" binding:"required"`
	VendorOrTechnician string `json:"vendor_or_technician"`
	Cost               int64  `json:"cost"`
	Description        string `json:"description"`
	NextServiceDate    string `json:"next_service_date"`
	PhotoURL           string `json:"photo_url"`
	UpdateCondition    string `json:"update_condition"`
}

// GET /api/items/:id/maintenance
func (h *MaintenanceHandler) ListByItem(c *gin.Context) {
	itemID := c.Param("id")

	rows, err := h.pool.Query(c, `
		SELECT id, item_id, service_type, service_date::text,
		       COALESCE(vendor_or_technician, ''), COALESCE(cost, 0),
		       COALESCE(description, ''), next_service_date::text,
		       COALESCE(photo_url, ''), COALESCE(performed_by, ''),
		       created_at::text
		FROM item_maintenances
		WHERE item_id::text = $1
		ORDER BY service_date DESC, created_at DESC
	`, itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	list := []maintenanceRow{}
	for rows.Next() {
		var m maintenanceRow
		var nextDate *string
		rows.Scan(&m.ID, &m.ItemID, &m.ServiceType, &m.ServiceDate,
			&m.VendorOrTechnician, &m.Cost, &m.Description,
			&nextDate, &m.PhotoURL, &m.PerformedBy, &m.CreatedAt)
		m.NextServiceDate = nextDate
		list = append(list, m)
	}

	c.JSON(http.StatusOK, list)
}

// POST /api/items/:id/maintenance
func (h *MaintenanceHandler) Create(c *gin.Context) {
	itemID := c.Param("id")
	var req maintenanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	performedBy := c.GetString("name")
	if performedBy == "" {
		performedBy = "Petugas"
	}

	var nextDate any
	if strings.TrimSpace(req.NextServiceDate) != "" {
		nextDate = req.NextServiceDate
	}

	var id string
	err := h.pool.QueryRow(c, `
		INSERT INTO item_maintenances (
			item_id, service_type, service_date, vendor_or_technician,
			cost, description, next_service_date, photo_url, performed_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id::text
	`, itemID, req.ServiceType, req.ServiceDate, nullable(req.VendorOrTechnician),
		req.Cost, nullable(req.Description), nextDate, nullable(req.PhotoURL), performedBy).
		Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update item condition if specified
	if strings.TrimSpace(req.UpdateCondition) != "" {
		h.pool.Exec(c, `UPDATE inventory_items SET condition_status=$1, updated_at=now() WHERE id::text=$2`,
			req.UpdateCondition, itemID)
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "riwayat servis tersimpan"})
}

// DELETE /api/items/:id/maintenance/:mId
func (h *MaintenanceHandler) Delete(c *gin.Context) {
	mID := c.Param("mId")
	h.pool.Exec(c, `DELETE FROM item_maintenances WHERE id::text=$1`, mID)
	c.JSON(http.StatusOK, gin.H{"message": "riwayat servis dihapus"})
}

// GET /api/maintenance/upcoming — assets with calibration or service due within 30 days or overdue
func (h *MaintenanceHandler) Upcoming(c *gin.Context) {
	rows, err := h.pool.Query(c, `
		SELECT m.id, m.item_id, m.service_type, m.service_date::text,
		       COALESCE(m.vendor_or_technician, ''), COALESCE(m.cost, 0),
		       COALESCE(m.description, ''), m.next_service_date::text,
		       COALESCE(m.photo_url, ''), COALESCE(m.performed_by, ''),
		       m.created_at::text, i.name, i.sku
		FROM item_maintenances m
		JOIN inventory_items i ON i.id = m.item_id
		WHERE m.next_service_date IS NOT NULL
		  AND m.next_service_date <= (CURRENT_DATE + INTERVAL '30 day')
		ORDER BY m.next_service_date ASC
		LIMIT 20
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	list := []maintenanceRow{}
	for rows.Next() {
		var m maintenanceRow
		var nextDate *string
		rows.Scan(&m.ID, &m.ItemID, &m.ServiceType, &m.ServiceDate,
			&m.VendorOrTechnician, &m.Cost, &m.Description,
			&nextDate, &m.PhotoURL, &m.PerformedBy, &m.CreatedAt,
			&m.ItemName, &m.ItemSKU)
		m.NextServiceDate = nextDate
		list = append(list, m)
	}

	c.JSON(http.StatusOK, list)
}
