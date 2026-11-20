package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeadersMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SecurityHeadersMiddleware())
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.NotEmpty(t, w.Header().Get("Content-Security-Policy"))
}

func TestIsPublicRoute(t *testing.T) {
	assert.True(t, isPublicRoute("/api/barcode/abc-123.png"))
	assert.True(t, isPublicRoute("/api/barcode/sheet"))
	assert.False(t, isPublicRoute("/api/barcode/templates"))
	assert.False(t, isPublicRoute("/api/barcode/units"))
}

func TestLoginRateLimiterAndSecurityStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	SetRedisClient(nil)
	r := gin.New()
	r.Use(LoginRateLimiter(5, time.Minute))
	r.POST("/api/auth/login", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
		req.RemoteAddr = "198.51.100.10:12345"
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	req.RemoteAddr = "198.51.100.10:12345"
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	_, blocked, _ := GetSecurityStatus()
	found := false
	for _, b := range blocked {
		if b.IP == "198.51.100.10" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected blocked IP in GetSecurityStatus")
	bucketMu.Lock()
	delete(loginBuckets, "198.51.100.10")
	bucketMu.Unlock()
}
