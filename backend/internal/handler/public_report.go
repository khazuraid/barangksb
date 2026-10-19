package handler

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PublicReportHandler struct {
	pool *pgxpool.Pool
}

func NewPublicReportHandler(pool *pgxpool.Pool) *PublicReportHandler {
	return &PublicReportHandler{pool: pool}
}

type CreateReportInput struct {
	ItemID          *string `json:"item_id"`
	ItemSKU         string  `json:"item_sku"`
	ItemName        string  `json:"item_name"`
	ReporterName    string  `json:"reporter_name" binding:"required"`
	ReporterContact string  `json:"reporter_contact" binding:"required"`
	ReportType      string  `json:"report_type" binding:"required"`
	Description     string  `json:"description" binding:"required"`
	Location        string  `json:"location" binding:"required"`
	PhotoURL        string  `json:"photo_url"`
}

type UpdateReportStatusInput struct {
	Status              string  `json:"status" binding:"required"`
	AdminNotes          string  `json:"admin_notes"`
	UpdateItemCondition *string `json:"update_item_condition"`
}

// generateReportNumber creates sequential-like daily ticket ID: LAP-YYYYMMDD-XXXX
func generateReportNumber() string {
	b := make([]byte, 2)
	rand.Read(b)
	randHex := strings.ToUpper(hex.EncodeToString(b))
	return fmt.Sprintf("LAP-%s-%s", time.Now().Format("20060102"), randHex)
}

// POST /api/public/reports - Submit report from public website (no auth required)
func (h *PublicReportHandler) PublicCreate(c *gin.Context) {
	var in CreateReportInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mohon lengkapi data wajib: Nama, Kontak, Jenis Laporan, dan Deskripsi"})
		return
	}

	reportNum := generateReportNumber()

	// If item_sku or item_id provided, sync item details
	var itemID *string = in.ItemID
	var itemName = in.ItemName
	var itemSKU = in.ItemSKU

	if itemID != nil && *itemID != "" {
		var name, sku, loc string
		err := h.pool.QueryRow(c, `SELECT name, sku, location FROM inventory_items WHERE id::text=$1`, *itemID).Scan(&name, &sku, &loc)
		if err == nil {
			itemName = name
			itemSKU = sku
			if in.Location == "" {
				in.Location = loc
			}
		}
	} else if itemSKU != "" {
		var id, name, loc string
		err := h.pool.QueryRow(c, `SELECT id::text, name, location FROM inventory_items WHERE sku=$1`, itemSKU).Scan(&id, &name, &loc)
		if err == nil {
			itemID = &id
			itemName = name
			if in.Location == "" {
				in.Location = loc
			}
		}
	}

	var newID string
	err := h.pool.QueryRow(c, `
		INSERT INTO item_reports (
			report_number, item_id, item_sku, item_name,
			reporter_name, reporter_contact, report_type, description,
			location, photo_url, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'PENDING')
		RETURNING id::text
	`, reportNum, itemID, itemSKU, itemName, in.ReporterName, in.ReporterContact, in.ReportType, in.Description, in.Location, in.PhotoURL).Scan(&newID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan laporan: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":       "Laporan berhasil dikirim",
		"id":            newID,
		"report_number": reportNum,
		"status":        "PENDING",
	})
}

// GET /api/public/reports/:num - Track status by report_number or ID (no auth required)
func (h *PublicReportHandler) PublicGet(c *gin.Context) {
	num := strings.TrimSpace(c.Param("num"))

	var r struct {
		ID              string     `json:"id"`
		ReportNumber    string     `json:"report_number"`
		ItemID          *string    `json:"item_id"`
		ItemSKU         *string    `json:"item_sku"`
		ItemName        *string    `json:"item_name"`
		ReporterName    string     `json:"reporter_name"`
		ReporterContact string     `json:"reporter_contact"`
		ReportType      string     `json:"report_type"`
		Description     string     `json:"description"`
		Location        *string    `json:"location"`
		PhotoURL        *string    `json:"photo_url"`
		Status          string     `json:"status"`
		AdminNotes      *string    `json:"admin_notes"`
		HandledBy       *string    `json:"handled_by"`
		HandledAt       *time.Time `json:"handled_at"`
		CreatedAt       time.Time  `json:"created_at"`
		UpdatedAt       time.Time  `json:"updated_at"`
	}

	query := `
		SELECT id, report_number, item_id::text, item_sku, item_name,
		       reporter_name, reporter_contact, report_type, description,
		       location, photo_url, status, admin_notes, handled_by, handled_at,
		       created_at, updated_at
		FROM item_reports
		WHERE report_number ILIKE $1 OR id::text = $1
		LIMIT 1
	`
	err := h.pool.QueryRow(c, query, num).Scan(
		&r.ID, &r.ReportNumber, &r.ItemID, &r.ItemSKU, &r.ItemName,
		&r.ReporterName, &r.ReporterContact, &r.ReportType, &r.Description,
		&r.Location, &r.PhotoURL, &r.Status, &r.AdminNotes, &r.HandledBy, &r.HandledAt,
		&r.CreatedAt, &r.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Laporan dengan nomor atau ID tersebut tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, r)
}

// POST /api/public/upload - Upload public report photo (no auth required)
func (h *PublicReportHandler) PublicUploadPhoto(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20)

	file, hdr, err := c.Request.FormFile("photo")
	if err != nil || file == nil || hdr.Size == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File foto wajib dilampirkan (maksimal 10 MB)"})
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(hdr.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		ext = ".jpg"
	}

	filename := "report_" + uuid.NewString() + ext
	uploadDir := filepath.Join("data", "uploads")
	_ = os.MkdirAll(uploadDir, 0755)
	destPath := filepath.Join(uploadDir, filename)

	out, err := os.Create(destPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file foto"})
		return
	}
	defer out.Close()

	var buf bytes.Buffer
	_, err = io.Copy(out, io.TeeReader(file, &buf))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menulis file foto"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url": fmt.Sprintf("/api/uploads/%s", filename),
	})
}

// GET /api/reports - Admin/Petugas view list of all reports on website
func (h *PublicReportHandler) ListReports(c *gin.Context) {
	status := strings.TrimSpace(c.Query("status"))
	reportType := strings.TrimSpace(c.Query("type"))
	q := strings.TrimSpace(c.Query("q"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var conditions []string
	var args []any
	argIdx := 1

	if status != "" && status != "ALL" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}

	if reportType != "" && reportType != "ALL" {
		conditions = append(conditions, fmt.Sprintf("report_type = $%d", argIdx))
		args = append(args, reportType)
		argIdx++
	}

	if q != "" {
		conditions = append(conditions, fmt.Sprintf("(report_number ILIKE $%d OR item_sku ILIKE $%d OR item_name ILIKE $%d OR reporter_name ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx, argIdx, argIdx, argIdx))
		args = append(args, "%"+q+"%")
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Stats counts
	var statTotal, statPending, statInProgress, statResolved, statRejected int
	_ = h.pool.QueryRow(c, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'PENDING'),
			COUNT(*) FILTER (WHERE status = 'IN_PROGRESS'),
			COUNT(*) FILTER (WHERE status = 'RESOLVED'),
			COUNT(*) FILTER (WHERE status = 'REJECTED')
		FROM item_reports
	`).Scan(&statTotal, &statPending, &statInProgress, &statResolved, &statRejected)

	// Total filtered
	var totalFiltered int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM item_reports %s", whereClause)
	_ = h.pool.QueryRow(c, countQuery, args...).Scan(&totalFiltered)

	// Fetch page
	listQuery := fmt.Sprintf(`
		SELECT id, report_number, item_id::text, item_sku, item_name,
		       reporter_name, reporter_contact, report_type, description,
		       location, photo_url, status, admin_notes, handled_by, handled_at,
		       created_at, updated_at
		FROM item_reports
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, perPage, offset)

	rows, err := h.pool.Query(c, listQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca data laporan: " + err.Error()})
		return
	}
	defer rows.Close()

	type ReportItem struct {
		ID              string     `json:"id"`
		ReportNumber    string     `json:"report_number"`
		ItemID          *string    `json:"item_id"`
		ItemSKU         *string    `json:"item_sku"`
		ItemName        *string    `json:"item_name"`
		ReporterName    string     `json:"reporter_name"`
		ReporterContact string     `json:"reporter_contact"`
		ReportType      string     `json:"report_type"`
		Description     string     `json:"description"`
		Location        *string    `json:"location"`
		PhotoURL        *string    `json:"photo_url"`
		Status          string     `json:"status"`
		AdminNotes      *string    `json:"admin_notes"`
		HandledBy       *string    `json:"handled_by"`
		HandledAt       *time.Time `json:"handled_at"`
		CreatedAt       time.Time  `json:"created_at"`
		UpdatedAt       time.Time  `json:"updated_at"`
	}

	items := []ReportItem{}
	for rows.Next() {
		var it ReportItem
		err := rows.Scan(
			&it.ID, &it.ReportNumber, &it.ItemID, &it.ItemSKU, &it.ItemName,
			&it.ReporterName, &it.ReporterContact, &it.ReportType, &it.Description,
			&it.Location, &it.PhotoURL, &it.Status, &it.AdminNotes, &it.HandledBy, &it.HandledAt,
			&it.CreatedAt, &it.UpdatedAt,
		)
		if err == nil {
			items = append(items, it)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  items,
		"total": totalFiltered,
		"page":  page,
		"stats": gin.H{
			"total":       statTotal,
			"pending":     statPending,
			"in_progress": statInProgress,
			"resolved":    statResolved,
			"rejected":    statRejected,
		},
	})
}

// GET /api/reports/:id - Get single report detail (Admin/Petugas)
func (h *PublicReportHandler) GetReport(c *gin.Context) {
	id := c.Param("id")

	var r struct {
		ID              string     `json:"id"`
		ReportNumber    string     `json:"report_number"`
		ItemID          *string    `json:"item_id"`
		ItemSKU         *string    `json:"item_sku"`
		ItemName        *string    `json:"item_name"`
		ReporterName    string     `json:"reporter_name"`
		ReporterContact string     `json:"reporter_contact"`
		ReportType      string     `json:"report_type"`
		Description     string     `json:"description"`
		Location        *string    `json:"location"`
		PhotoURL        *string    `json:"photo_url"`
		Status          string     `json:"status"`
		AdminNotes      *string    `json:"admin_notes"`
		HandledBy       *string    `json:"handled_by"`
		HandledAt       *time.Time `json:"handled_at"`
		CreatedAt       time.Time  `json:"created_at"`
		UpdatedAt       time.Time  `json:"updated_at"`
	}

	query := `
		SELECT id, report_number, item_id::text, item_sku, item_name,
		       reporter_name, reporter_contact, report_type, description,
		       location, photo_url, status, admin_notes, handled_by, handled_at,
		       created_at, updated_at
		FROM item_reports
		WHERE id::text = $1 OR report_number = $1
		LIMIT 1
	`
	err := h.pool.QueryRow(c, query, id).Scan(
		&r.ID, &r.ReportNumber, &r.ItemID, &r.ItemSKU, &r.ItemName,
		&r.ReporterName, &r.ReporterContact, &r.ReportType, &r.Description,
		&r.Location, &r.PhotoURL, &r.Status, &r.AdminNotes, &r.HandledBy, &r.HandledAt,
		&r.CreatedAt, &r.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Laporan tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, r)
}

// PUT /api/reports/:id/status - Update report status and follow up (Admin/Petugas)
func (h *PublicReportHandler) UpdateReportStatus(c *gin.Context) {
	id := c.Param("id")
	var in UpdateReportStatusInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status wajib diisi"})
		return
	}

	validStatuses := map[string]bool{
		"PENDING":     true,
		"IN_PROGRESS": true,
		"RESOLVED":    true,
		"REJECTED":    true,
	}
	statusUpper := strings.ToUpper(strings.TrimSpace(in.Status))
	if !validStatuses[statusUpper] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status tidak valid. Pilihan: PENDING, IN_PROGRESS, RESOLVED, REJECTED"})
		return
	}

	handledBy := c.GetString("name")
	if handledBy == "" {
		handledBy = "Petugas Inventaris"
	}

	now := time.Now()
	res, err := h.pool.Exec(c, `
		UPDATE item_reports
		SET status = $1, admin_notes = $2, handled_by = $3, handled_at = $4, updated_at = now()
		WHERE id::text = $5
	`, statusUpper, in.AdminNotes, handledBy, now, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui laporan: " + err.Error()})
		return
	}
	if res.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Laporan tidak ditemukan"})
		return
	}

	// If update_item_condition is given (e.g. 'Rusak Ringan', 'Rusak Berat', 'Berfungsi')
	if in.UpdateItemCondition != nil && *in.UpdateItemCondition != "" {
		var itemID *string
		_ = h.pool.QueryRow(c, `SELECT item_id::text FROM item_reports WHERE id::text=$1`, id).Scan(&itemID)
		if itemID != nil && *itemID != "" {
			_, _ = h.pool.Exec(c, `UPDATE inventory_items SET condition_status=$1, updated_at=now() WHERE id::text=$2`, *in.UpdateItemCondition, *itemID)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Status laporan berhasil diperbarui",
		"status":     statusUpper,
		"handled_by": handledBy,
		"handled_at": now,
	})
}

// DELETE /api/reports/:id - Delete report (Admin)
func (h *PublicReportHandler) DeleteReport(c *gin.Context) {
	id := c.Param("id")
	res, err := h.pool.Exec(c, `DELETE FROM item_reports WHERE id::text = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus laporan: " + err.Error()})
		return
	}
	if res.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Laporan tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Laporan berhasil dihapus"})
}
