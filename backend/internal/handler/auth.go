package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"inventariskantor/internal/auth"
	"inventariskantor/internal/config"
)

type AuthHandler struct {
	pool   *pgxpool.Pool
	cfg    *config.Config
}

func NewAuthHandler(pool *pgxpool.Pool, cfg *config.Config) *AuthHandler {
	return &AuthHandler{pool: pool, cfg: cfg}
}

type loginReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email dan password wajib"})
		return
	}

	var id, name, role, hash string
	err := h.pool.QueryRow(c, `SELECT id, name, role, password_hash FROM users WHERE email=$1`,
		strings.ToLower(req.Email)).Scan(&id, &name, &role, &hash)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "email atau password salah"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "email atau password salah"})
		return
	}

	token, err := auth.GenerateToken(h.cfg.JWTSecret, id, req.Email, role, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id": id, "name": name, "email": req.Email, "role": role,
		},
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id": c.GetString("uid"),
		"name": c.GetString("name"),
		"email": c.GetString("email"),
		"role": c.GetString("role"),
	})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req struct {
		Old string `json:"old" binding:"required"`
		New string `json:"new" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	email := c.GetString("email")

	var hash string
	err := h.pool.QueryRow(c, `SELECT password_hash FROM users WHERE email=$1`, email).Scan(&hash)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user tidak ditemukan"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Old)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "password lama salah"})
		return
	}
	newHash, _ := bcrypt.GenerateFromPassword([]byte(req.New), bcrypt.DefaultCost)
	h.pool.Exec(c, `UPDATE users SET password_hash=$2 WHERE email=$1`, email, string(newHash))
	c.JSON(http.StatusOK, gin.H{"message": "password berhasil diganti"})
}

func (h *AuthHandler) CreateUserCLI(c *gin.Context) {
	// not via API — CLI subcommand in main.go
	c.JSON(http.StatusOK, gin.H{"message": "use CLI: go run ./cmd/server -create-user"})
}
