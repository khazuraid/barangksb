package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"inventariskantor/internal/auth"
	"inventariskantor/internal/config"
	"inventariskantor/internal/middleware"
)

var dummyBcryptHash = func() []byte {
	h, _ := bcrypt.GenerateFromPassword([]byte("dummy-timing-equalizer-password"), bcrypt.DefaultCost)
	return h
}()

type AuthHandler struct {
	pool *pgxpool.Pool
	cfg  *config.Config
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

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if !middleware.CheckAccountLoginAllowed(email, 5, 5*time.Minute) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "Terlalu banyak percobaan login gagal untuk akun ini. Silakan tunggu beberapa menit.",
		})
		return
	}

	var id, name, role, hash string
	var tokenVersion int
	err := h.pool.QueryRow(c,
		`SELECT id::text, name, role, password_hash, COALESCE(token_version, 1) FROM users WHERE email=$1`,
		email,
	).Scan(&id, &name, &role, &hash, &tokenVersion)
	if err != nil {
		_ = bcrypt.CompareHashAndPassword(dummyBcryptHash, []byte(req.Password))
		middleware.RecordAccountLoginFailure(c, email, 5, 5*time.Minute)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "email atau password salah"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		middleware.RecordAccountLoginFailure(c, email, 5, 5*time.Minute)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "email atau password salah"})
		return
	}

	ttl := time.Duration(h.cfg.JWTTTLHours) * time.Hour
	token, err := auth.GenerateToken(h.cfg.JWTSecret, id, email, role, name, tokenVersion, ttl)
	if err != nil {
		internalError(c, "gagal membuat token autentikasi", err)
		return
	}

	middleware.ResetLoginAttempts(c.ClientIP(), email)

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id": id, "name": name, "email": email, "role": role,
		},
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id":    c.GetString("uid"),
		"name":  c.GetString("name"),
		"email": c.GetString("email"),
		"role":  c.GetString("role"),
	})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req struct {
		Old string `json:"old" binding:"required"`
		New string `json:"new" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password lama wajib diisi dan password baru minimal 8 karakter"})
		return
	}
	uid := c.GetString("uid")

	var hash string
	err := h.pool.QueryRow(c, `SELECT password_hash FROM users WHERE id::text=$1`, uid).Scan(&hash)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user tidak ditemukan"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Old)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "password lama salah"})
		return
	}
	if len(req.New) > 72 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password baru maksimal 72 karakter"})
		return
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.New), bcrypt.DefaultCost)
	if err != nil {
		internalError(c, "gagal memproses password baru", err)
		return
	}
	if _, err := h.pool.Exec(c,
		`UPDATE users SET password_hash=$2, token_version=COALESCE(token_version, 1)+1 WHERE id::text=$1`,
		uid, string(newHash),
	); err != nil {
		internalError(c, "gagal mengganti password", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password berhasil diganti, silakan login ulang"})
}
