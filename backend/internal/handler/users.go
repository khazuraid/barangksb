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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	_, err := h.pool.Exec(c,
		`INSERT INTO users (name, email, password_hash, role) VALUES ($1,$2,$3,$4)
		 ON CONFLICT (email) DO UPDATE SET name=$1, password_hash=$3, role=$4`,
		req.Name, strings.ToLower(req.Email), string(hash), req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "user created"})
}

func (h *UserHandler) UpdateRole(c *gin.Context) {
	email := c.Param("email")
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
	h.pool.Exec(c, `UPDATE users SET role=$2 WHERE email=$1`, email, req.Role)
	c.JSON(http.StatusOK, gin.H{"message": "role updated"})
}

func (h *UserHandler) Delete(c *gin.Context) {
	email := c.Param("email")
	h.pool.Exec(c, `DELETE FROM users WHERE email=$1`, email)
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
	h.pool.Exec(c, `INSERT INTO categories (id, name) VALUES ($1,$2) ON CONFLICT (id) DO UPDATE SET name=$2`, id, req.Name)
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *UserHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	h.pool.Exec(c, `DELETE FROM categories WHERE id=$1`, id)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// Locations CRUD
func (h *UserHandler) CreateLocation(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := slugify(req.Name)
	h.pool.Exec(c, `INSERT INTO locations (id, name) VALUES ($1,$2) ON CONFLICT (id) DO UPDATE SET name=$2`, id, req.Name)
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *UserHandler) DeleteLocation(c *gin.Context) {
	id := c.Param("id")
	h.pool.Exec(c, `DELETE FROM locations WHERE id=$1`, id)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
