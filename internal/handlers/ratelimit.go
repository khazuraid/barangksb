package handlers

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimitedLogin wraps a login handler with per-IP rate limiting (5/min, burst 5).
func (h *Handlers) RateLimitedLogin(next http.HandlerFunc) http.HandlerFunc {
	return RateLimitLogin(next)
}

// limiterMap throttles login attempts per IP.
type limiterMap struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
}

func newLimiterMap() *limiterMap { return &limiterMap{limiters: map[string]*rate.Limiter{}} }

func (l *limiterMap) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	if lim, ok := l.limiters[ip]; ok {
		return lim
	}
	lim := rate.NewLimiter(rate.Every(time.Minute/5), 5) // 5 req/menit, burst 5
	l.limiters[ip] = lim
	return lim
}

var loginLimiters = newLimiterMap()

// RateLimitLogin wraps an http.HandlerFunc with per-IP rate limiting.
func RateLimitLogin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if fwd := r.Header.Get("X-Real-IP"); fwd != "" {
			ip = fwd
		} else if host, _, err := net.SplitHostPort(ip); err == nil {
			ip = host // buang port — kalau tidak, tiap koneksi = "IP" baru
		}
		if !loginLimiters.get(ip).Allow() {
			http.Error(w, "Terlalu banyak percobaan. Coba lagi nanti.", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}
