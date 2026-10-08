package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware adds defensive HTTP headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// In-memory rate limiter for login brute-force prevention
type ipRateBucket struct {
	attempts int
	lastSeen time.Time
}

var (
	loginBuckets = make(map[string]*ipRateBucket)
	bucketMu     sync.Mutex
)

// LoginRateLimiter limits failed login attempts per client IP
func LoginRateLimiter(maxAttempts int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path != "/api/auth/login" || c.Request.Method != http.MethodPost {
			c.Next()
			return
		}

		ip := c.ClientIP()
		bucketMu.Lock()
		b, exists := loginBuckets[ip]
		now := time.Now()
		if !exists || now.Sub(b.lastSeen) > window {
			loginBuckets[ip] = &ipRateBucket{attempts: 1, lastSeen: now}
			bucketMu.Unlock()
			c.Next()
			return
		}

		b.lastSeen = now
		if b.attempts >= maxAttempts {
			bucketMu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Terlalu banyak percobaan login gagal. Silakan tunggu 1 menit sebelum mencoba lagi.",
			})
			c.Abort()
			return
		}

		b.attempts++
		bucketMu.Unlock()
		c.Next()
	}
}

// Gzip response writer
type gzipResponseWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipResponseWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

func (g *gzipResponseWriter) WriteString(s string) (int, error) {
	return g.writer.Write([]byte(s))
}

// GzipMiddleware compresses HTML, JS, CSS, JSON responses to drastically speed up network transfers
func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") ||
			strings.Contains(c.Request.URL.Path, "/uploads/") ||
			strings.Contains(c.Request.URL.Path, "/api/uploads/") {
			c.Next()
			return
		}

		gz, err := gzip.NewWriterLevel(c.Writer, gzip.BestSpeed)
		if err != nil {
			c.Next()
			return
		}
		defer gz.Close()

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")
		c.Writer.Header().Del("Content-Length")

		gw := &gzipResponseWriter{ResponseWriter: c.Writer, writer: gz}
		c.Writer = gw

		c.Next()

		// Flush writer
		_ = gz.Flush()
		_ = io.Discard
	}
}
