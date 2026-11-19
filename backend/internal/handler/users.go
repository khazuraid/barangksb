package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"inventariskantor/internal/middleware"
)

type UserHandler struct {
	pool *pgxpool.Pool
}

func NewUserHandler(pool *pgxpool.Pool) *UserHandler {
	return &UserHandler{pool: pool}
}

func (h *UserHandler) List(c *gin.Context) {
	rows, err := h.pool.Query(c, `SELECT id::text, name, email, role, created_at FROM users ORDER BY created_at`)
	if err != nil {
		internalError(c, "gagal mengambil daftar user", err)
		return
	}
	defer rows.Close()
	type user struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	users := []user{}
	for rows.Next() {
		var u user
		var ts time.Time
		rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &ts)
		users = append(users, u)
	}
	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) Create(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required,min=8"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Role != "admin" && req.Role != "petugas" {
		req.Role = "petugas"
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))

	var exists bool
	_ = h.pool.QueryRow(c, `SELECT TRUE FROM users WHERE email=$1`, email).Scan(&exists)
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Email sudah terdaftar di sistem"})
		return
	}

	if len(req.Password) > 72 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password maksimal 72 karakter"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		internalError(c, "gagal memproses password", err)
		return
	}
	_, err = h.pool.Exec(c,
		`INSERT INTO users (name, email, password_hash, role) VALUES ($1,$2,$3,$4)`,
		req.Name, email, string(hash), req.Role)
	if err != nil {
		internalError(c, "gagal membuat user", err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "user created"})
}

func (h *UserHandler) UpdateRole(c *gin.Context) {
	email := strings.ToLower(c.Param("email"))
	var req struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Role != "admin" && req.Role != "petugas" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role tidak valid"})
		return
	}

	tx, err := h.pool.Begin(c)
	if err != nil {
		internalError(c, "gagal memulai transaksi update role", err)
		return
	}
	defer tx.Rollback(c)

	var currentRole string
	if err := tx.QueryRow(c, `SELECT role FROM users WHERE email=$1 FOR UPDATE`, email).Scan(&currentRole); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user tidak ditemukan"})
		return
	}

	// Cegah demosi admin terakhir secara atomik
	if currentRole == "admin" && req.Role != "admin" {
		rows, err := tx.Query(c, `SELECT id FROM users WHERE role='admin' FOR UPDATE`)
		if err != nil {
			internalError(c, "gagal memeriksa jumlah admin", err)
			return
		}
		adminCount := 0
		for rows.Next() {
			adminCount++
		}
		rows.Close()
		if adminCount <= 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tidak dapat mengubah role admin terakhir"})
			return
		}
	}

	if _, err := tx.Exec(c,
		`UPDATE users SET role=$2, token_version=COALESCE(token_version, 1)+1 WHERE email=$1`,
		email, req.Role,
	); err != nil {
		internalError(c, "gagal mengubah role user", err)
		return
	}
	if err := tx.Commit(c); err != nil {
		internalError(c, "gagal menyimpan perubahan role", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "role updated"})
}

func (h *UserHandler) Delete(c *gin.Context) {
	email := strings.ToLower(c.Param("email"))
	currentEmail := strings.ToLower(c.GetString("email"))
	if email == currentEmail {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tidak dapat menghapus akun Anda sendiri"})
		return
	}

	tx, err := h.pool.Begin(c)
	if err != nil {
		internalError(c, "gagal memulai transaksi hapus user", err)
		return
	}
	defer tx.Rollback(c)

	var role string
	if err := tx.QueryRow(c, `SELECT role FROM users WHERE email=$1 FOR UPDATE`, email).Scan(&role); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user tidak ditemukan"})
		return
	}
	if role == "admin" {
		rows, err := tx.Query(c, `SELECT id FROM users WHERE role='admin' FOR UPDATE`)
		if err != nil {
			internalError(c, "gagal memeriksa jumlah admin", err)
			return
		}
		adminCount := 0
		for rows.Next() {
			adminCount++
		}
		rows.Close()
		if adminCount <= 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tidak dapat menghapus admin terakhir"})
			return
		}
	}

	if _, err := tx.Exec(c, `DELETE FROM users WHERE email=$1`, email); err != nil {
		internalError(c, "gagal menghapus user", err)
		return
	}
	if err := tx.Commit(c); err != nil {
		internalError(c, "gagal menyimpan penghapusan user", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// Categories CRUD
func (h *UserHandler) CreateCategory(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := slugify(req.Name)
	if _, err := h.pool.Exec(c, `INSERT INTO categories (id, name) VALUES ($1,$2) ON CONFLICT (id) DO UPDATE SET name=$2`, id, req.Name); err != nil {
		internalError(c, "gagal menyimpan kategori", err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *UserHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	if _, err := h.pool.Exec(c, `DELETE FROM categories WHERE id=$1`, id); err != nil {
		internalError(c, "gagal menghapus kategori", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// Locations CRUD
func (h *UserHandler) CreateLocation(c *gin.Context) {
	var req struct {
		ID   string `json:"id"`
		Name string `json:"name" binding:"required"`
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := strings.TrimSpace(req.ID)
	if id == "" {
		id = slugify(req.Name)
	}
	code := strings.TrimSpace(req.Code)
	if _, err := h.pool.Exec(c, `INSERT INTO locations (id, name, code) VALUES ($1,$2,$3) ON CONFLICT (id) DO UPDATE SET name=$2, code=$3`, id, req.Name, code); err != nil {
		internalError(c, "gagal menyimpan lokasi", err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "name": req.Name, "code": code})
}

func (h *UserHandler) DeleteLocation(c *gin.Context) {
	id := c.Param("id")
	if _, err := h.pool.Exec(c, `DELETE FROM locations WHERE id=$1`, id); err != nil {
		internalError(c, "gagal menghapus lokasi", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// Audit
func (h *UserHandler) AuditLog(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	perPage := configPerPage(c)
	offset := (page - 1) * perPage

	var total int
	h.pool.QueryRow(c, `SELECT count(*) FROM audit_log`).Scan(&total)

	rows, err := h.pool.Query(c, `SELECT table_name, op, row_id, old_data, new_data, at FROM audit_log ORDER BY at DESC LIMIT $1 OFFSET $2`, perPage, offset)
	if err != nil {
		internalError(c, "gagal mengambil audit log", err)
		return
	}
	defer rows.Close()
	type log struct {
		Table   string          `json:"table_name"`
		Op      string          `json:"op"`
		RowID   string          `json:"row_id"`
		OldData json.RawMessage `json:"old_data"`
		NewData json.RawMessage `json:"new_data"`
		At      string          `json:"at"`
	}
	logs := []log{}
	for rows.Next() {
		var l log
		var ts time.Time
		var rowID *string
		rows.Scan(&l.Table, &l.Op, &rowID, &l.OldData, &l.NewData, &ts)
		if rowID != nil {
			l.RowID = *rowID
		}
		l.At = ts.In(time.FixedZone("WIB", 7*3600)).Format("02-01-2006 15:04:05")
		logs = append(logs, l)
	}
	pages := total / perPage
	if total%perPage > 0 {
		pages++
	}
	c.JSON(http.StatusOK, paginatedResp{Data: logs, Total: total, Page: page, PerPage: perPage, Pages: pages})
}

// SecurityAuditLog returns burst attack and rate limit telemetry
func (h *UserHandler) SecurityAuditLog(c *gin.Context) {
	events, blocked, stats := middleware.GetSecurityStatus()
	c.JSON(http.StatusOK, gin.H{
		"events":      events,
		"blocked_ips": blocked,
		"stats":       stats,
	})
}

func (h *UserHandler) UnblockIP(c *gin.Context) {
	var req struct {
		IP string `json:"ip" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IP wajib diisi"})
		return
	}
	success := middleware.UnblockSecurityIP(req.IP)
	c.JSON(http.StatusOK, gin.H{"success": success, "ip": req.IP})
}

func (h *UserHandler) ClearSecurityEvents(c *gin.Context) {
	middleware.ClearSecurityEvents()
	c.JSON(http.StatusOK, gin.H{"message": "log keamanan dibersihkan"})
}

func slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	prevDash := false
	for _, c := range s {
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
			b.WriteRune(c)
			prevDash = false
		case c == ' ' || c == '-' || c == '/' || c == '&':
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
