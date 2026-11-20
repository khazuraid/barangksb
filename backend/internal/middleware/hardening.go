package middleware

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"inventariskantor/internal/worker"
)

// SecurityHeadersMiddleware adds defensive HTTP headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("X-XSS-Protection", "0")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(self), microphone=(), geolocation=(self), payment=()")
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			if strings.HasPrefix(c.Request.URL.Path, "/api/barcode/sheet") {
				c.Header("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; frame-ancestors 'self'; base-uri 'self'")
			} else {
				c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
			}
		} else {
			c.Header("Content-Security-Policy", "default-src 'self'; img-src 'self' data: blob: https:; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' data: https://fonts.gstatic.com; script-src 'self' https://telegram.org; connect-src 'self' https://nominatim.openstreetmap.org https://server.arcgisonline.com https://tile.openstreetmap.org; frame-src 'self' https://www.openstreetmap.org; frame-ancestors 'self'; base-uri 'self'; form-action 'self'")
		}
		if c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}

// Security Incident data
type SecurityEvent struct {
	ID           int64     `json:"id"`
	EventType    string    `json:"event_type"` // "BURST_ATTACK", "BRUTE_FORCE_LOGIN"
	IP           string    `json:"ip"`
	Endpoint     string    `json:"endpoint"`
	Method       string    `json:"method"`
	AttemptCount int       `json:"attempt_count"`
	Status       string    `json:"status"` // "BLOCKED (429)", "THROTTLED"
	UserAgent    string    `json:"user_agent"`
	Details      string    `json:"details"`
	CreatedAt    time.Time `json:"created_at"`
	TimeStr      string    `json:"time_str"`
}

type BlockedIPInfo struct {
	IP           string `json:"ip"`
	Type         string `json:"type"`
	Attempts     int    `json:"attempts"`
	LastSeen     string `json:"last_seen"`
	ExpiresInSec int    `json:"expires_in_sec"`
}

var (
	recentEvents      []SecurityEvent
	eventsMu          sync.RWMutex
	eventSeq          int64
	globalDBPool      *pgxpool.Pool
	poolMu            sync.RWMutex
	globalRedisClient *redis.Client
	globalWorkerQueue *worker.Queue
	redisMu           sync.RWMutex

	// In-memory rate limiter for login brute-force prevention
	loginBuckets        = make(map[string]*ipRateBucket)
	accountLoginBuckets = make(map[string]*ipRateBucket)
	bucketMu            sync.Mutex

	// In-memory rate limiter for public endpoints
	publicBuckets = make(map[string]*ipRateBucket)
	publicMu      sync.Mutex

	// In-memory rate limiter for general API burst attacks
	burstBuckets = make(map[string]*burstBucket)
	burstMu      sync.Mutex
)

func init() {
	// Background garbage collector to prevent memory leak from IP tracking maps
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			now := time.Now()
			bucketMu.Lock()
			for ip, b := range loginBuckets {
				if now.Sub(b.lastSeen) > 10*time.Minute {
					delete(loginBuckets, ip)
				}
			}
			for acct, b := range accountLoginBuckets {
				if now.Sub(b.lastSeen) > 15*time.Minute {
					delete(accountLoginBuckets, acct)
				}
			}
			bucketMu.Unlock()

			publicMu.Lock()
			for key, b := range publicBuckets {
				if now.Sub(b.lastSeen) > 10*time.Minute {
					delete(publicBuckets, key)
				}
			}
			publicMu.Unlock()

			burstMu.Lock()
			for ip, b := range burstBuckets {
				if now.After(b.blockedUntil) && now.Sub(b.windowStart) > 10*time.Minute {
					delete(burstBuckets, ip)
				}
			}
			burstMu.Unlock()
		}
	}()
}

// SetRedisClient registers a Redis client for distributed rate limiting & blacklist
func SetRedisClient(r *redis.Client) {
	redisMu.Lock()
	globalRedisClient = r
	redisMu.Unlock()
}

// SetWorkerQueue registers a background worker queue for asynchronous security event logging
func SetWorkerQueue(q *worker.Queue) {
	redisMu.Lock()
	globalWorkerQueue = q
	redisMu.Unlock()
}

// ResetLoginAttempts clears per-account failed login attempts upon successful authentication.
// Note: per-IP rate-limit buckets are intentionally NOT cleared so an attacker on a shared IP
// cannot reset IP throttling by logging into a low-privilege account.
func ResetLoginAttempts(ip string, account ...string) {
	acct := ""
	if len(account) > 0 {
		acct = strings.ToLower(strings.TrimSpace(account[0]))
	}
	if acct != "" {
		bucketMu.Lock()
		delete(accountLoginBuckets, acct)
		bucketMu.Unlock()

		redisMu.RLock()
		rdb := globalRedisClient
		redisMu.RUnlock()
		if rdb != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			_ = rdb.Del(ctx, "rate:login:acct:"+acct, "rate:blocked:login:acct:"+acct).Err()
		}
	}
}

// CheckAccountLoginAllowed returns false if the target email has exceeded failed login limits.
func CheckAccountLoginAllowed(account string, maxAttempts int, window time.Duration) bool {
	acct := strings.ToLower(strings.TrimSpace(account))
	if acct == "" {
		return true
	}
	redisMu.RLock()
	rdb := globalRedisClient
	redisMu.RUnlock()
	if rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if exists, _ := rdb.Exists(ctx, "rate:blocked:login:acct:"+acct).Result(); exists > 0 {
			return false
		}
		return true
	}
	bucketMu.Lock()
	defer bucketMu.Unlock()
	if b, ok := accountLoginBuckets[acct]; ok {
		if time.Since(b.lastSeen) <= window && b.attempts >= maxAttempts {
			return false
		}
	}
	return true
}

// RecordAccountLoginFailure increments failed login counter for a specific email account.
func RecordAccountLoginFailure(c *gin.Context, account string, maxAttempts int, window time.Duration) {
	acct := strings.ToLower(strings.TrimSpace(account))
	if acct == "" {
		return
	}
	redisMu.RLock()
	rdb := globalRedisClient
	redisMu.RUnlock()
	if rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		key := "rate:login:acct:" + acct
		count, err := rdb.Incr(ctx, key).Result()
		if err == nil {
			if count == 1 {
				rdb.Expire(ctx, key, window)
			}
			if count >= int64(maxAttempts) {
				rdb.Set(ctx, "rate:blocked:login:acct:"+acct, "1", window)
				RecordSecurityEvent(c, "BRUTE_FORCE_LOGIN", fmt.Sprintf("Akun %s gagal login %d kali dalam %v", acct, count, window), int(count), "BLOCKED (429)")
			}
			return
		}
	}
	bucketMu.Lock()
	now := time.Now()
	b, exists := accountLoginBuckets[acct]
	if !exists || now.Sub(b.lastSeen) > window {
		accountLoginBuckets[acct] = &ipRateBucket{attempts: 1, lastSeen: now}
		bucketMu.Unlock()
		return
	}
	b.attempts++
	b.lastSeen = now
	attempts := b.attempts
	bucketMu.Unlock()
	if attempts >= maxAttempts {
		RecordSecurityEvent(c, "BRUTE_FORCE_LOGIN", fmt.Sprintf("Akun %s gagal login %d kali dalam %v", acct, attempts, window), attempts, "BLOCKED (429)")
	}
}

// PublicRateLimiter applies a dedicated rate limit to unauthenticated /api/public/* endpoints.
func PublicRateLimiter(maxReq int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := c.Request.URL.Path
		if !strings.HasPrefix(p, "/api/public/") {
			c.Next()
			return
		}
		ip := c.ClientIP()
		key := ip + ":" + c.Request.Method

		redisMu.RLock()
		rdb := globalRedisClient
		redisMu.RUnlock()
		if rdb != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			rKey := "rate:public:" + key
			count, err := rdb.Incr(ctx, rKey).Result()
			if err == nil {
				if count == 1 {
					rdb.Expire(ctx, rKey, window)
				}
				cancel()
				if count > int64(maxReq) {
					c.JSON(http.StatusTooManyRequests, gin.H{
						"error": "Terlalu banyak permintaan publik. Silakan coba lagi beberapa saat.",
					})
					c.Abort()
					return
				}
				c.Next()
				return
			}
			cancel()
		}

		publicMu.Lock()
		now := time.Now()
		b, exists := publicBuckets[key]
		if !exists || now.Sub(b.lastSeen) > window {
			publicBuckets[key] = &ipRateBucket{attempts: 1, lastSeen: now}
			publicMu.Unlock()
			c.Next()
			return
		}
		b.lastSeen = now
		if b.attempts >= maxReq {
			publicMu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Terlalu banyak permintaan publik. Silakan coba lagi beberapa saat.",
			})
			c.Abort()
			return
		}
		b.attempts++
		publicMu.Unlock()
		c.Next()
	}
}

type ipRateBucket struct {
	attempts int
	lastSeen time.Time
}

type burstBucket struct {
	count        int
	windowStart  time.Time
	blockedUntil time.Time
}

// SetDBPool registers pool for asynchronous security event logging
func SetDBPool(p *pgxpool.Pool) {
	poolMu.Lock()
	globalDBPool = p
	poolMu.Unlock()
}

// RecordSecurityEvent saves event to in-memory history and optional postgres table
func RecordSecurityEvent(c *gin.Context, eventType, details string, count int, status string) {
	now := time.Now()
	ip := c.ClientIP()
	ep := c.Request.URL.Path
	meth := c.Request.Method
	ua := c.GetHeader("User-Agent")

	eventsMu.Lock()
	eventSeq++
	ev := SecurityEvent{
		ID:           eventSeq,
		EventType:    eventType,
		IP:           ip,
		Endpoint:     ep,
		Method:       meth,
		AttemptCount: count,
		Status:       status,
		UserAgent:    ua,
		Details:      details,
		CreatedAt:    now,
		TimeStr:      now.In(time.FixedZone("WIB", 7*3600)).Format("02-01-2006 15:04:05"),
	}

	// Keep last 150 events in memory
	recentEvents = append([]SecurityEvent{ev}, recentEvents...)
	if len(recentEvents) > 150 {
		recentEvents = recentEvents[:150]
	}
	eventsMu.Unlock()

	// Asynchronously record into DB table via worker queue or fallback goroutine
	redisMu.RLock()
	q := globalWorkerQueue
	redisMu.RUnlock()

	if q != nil {
		_ = q.Enqueue(context.Background(), worker.TaskTypeSecurityEvent, worker.SecurityEventPayload{
			EventType:    eventType,
			IP:           ip,
			Endpoint:     ep,
			Method:       meth,
			AttemptCount: count,
			Status:       status,
			UserAgent:    ua,
			Details:      details,
			CreatedAt:    now,
		})
	} else {
		poolMu.RLock()
		p := globalDBPool
		poolMu.RUnlock()

		if p != nil {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				_, _ = p.Exec(ctx, `INSERT INTO security_events (event_type, ip, endpoint, method, attempt_count, status, user_agent, details, created_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
					eventType, ip, ep, meth, count, status, ua, details, now)
			}()
		}
	}
}

// GetSecurityStatus retrieves realtime telemetry for the audit page
func GetSecurityStatus() ([]SecurityEvent, []BlockedIPInfo, map[string]any) {
	now := time.Now()

	eventsMu.RLock()
	copiedEvents := make([]SecurityEvent, len(recentEvents))
	copy(copiedEvents, recentEvents)
	eventsMu.RUnlock()

	if len(copiedEvents) == 0 {
		poolMu.RLock()
		p := globalDBPool
		poolMu.RUnlock()
		if p != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			rows, err := p.Query(ctx, `SELECT id, event_type, ip, endpoint, method, attempt_count, status, COALESCE(user_agent, ''), COALESCE(details, ''), created_at
				FROM security_events ORDER BY created_at DESC LIMIT 50`)
			if err == nil {
				for rows.Next() {
					var ev SecurityEvent
					if err := rows.Scan(&ev.ID, &ev.EventType, &ev.IP, &ev.Endpoint, &ev.Method, &ev.AttemptCount, &ev.Status, &ev.UserAgent, &ev.Details, &ev.CreatedAt); err == nil {
						ev.TimeStr = ev.CreatedAt.In(time.FixedZone("WIB", 7*3600)).Format("02-01-2006 15:04:05")
						copiedEvents = append(copiedEvents, ev)
					}
				}
				rows.Close()
			}
			cancel()
		}
	}

	var blocked []BlockedIPInfo

	// Check login rate limit bucket
	bucketMu.Lock()
	for ip, b := range loginBuckets {
		if now.Sub(b.lastSeen) <= time.Minute && b.attempts >= 5 {
			rem := 60 - int(now.Sub(b.lastSeen).Seconds())
			if rem < 0 {
				rem = 0
			}
			blocked = append(blocked, BlockedIPInfo{
				IP:           ip,
				Type:         "Login Brute Force",
				Attempts:     b.attempts,
				LastSeen:     b.lastSeen.In(time.FixedZone("WIB", 7*3600)).Format("15:04:05"),
				ExpiresInSec: rem,
			})
		}
	}
	bucketMu.Unlock()

	// Check burst attack bucket
	burstMu.Lock()
	for ip, b := range burstBuckets {
		if now.Before(b.blockedUntil) {
			rem := int(b.blockedUntil.Sub(now).Seconds())
			blocked = append(blocked, BlockedIPInfo{
				IP:           ip,
				Type:         "Burst Attack Spikes",
				Attempts:     b.count,
				LastSeen:     b.windowStart.In(time.FixedZone("WIB", 7*3600)).Format("15:04:05"),
				ExpiresInSec: rem,
			})
		}
	}
	burstMu.Unlock()

	// Merge Redis-backed blocks using SCAN (KEYS is disabled in Redis config)
	redisMu.RLock()
	rdb := globalRedisClient
	redisMu.RUnlock()
	if rdb != nil {
		seen := make(map[string]bool, len(blocked))
		for _, b := range blocked {
			seen[b.Type+":"+b.IP] = true
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		scanBlocked := func(prefix, blockType string) {
			var cursor uint64
			for {
				keys, next, err := rdb.Scan(ctx, cursor, prefix+"*", 100).Result()
				if err != nil {
					return
				}
				for _, key := range keys {
					ip := strings.TrimPrefix(key, prefix)
					if ip == "" || strings.HasPrefix(ip, "acct:") || seen[blockType+":"+ip] {
						continue
					}
					ttl, err := rdb.TTL(ctx, key).Result()
					if err != nil || ttl <= 0 {
						continue
					}
					seen[blockType+":"+ip] = true
					blocked = append(blocked, BlockedIPInfo{
						IP:           ip,
						Type:         blockType,
						Attempts:     0,
						LastSeen:     now.In(time.FixedZone("WIB", 7*3600)).Format("15:04:05"),
						ExpiresInSec: int(ttl.Seconds()),
					})
				}
				cursor = next
				if cursor == 0 {
					return
				}
			}
		}
		scanBlocked("rate:blocked:login:", "Login Brute Force")
		scanBlocked("rate:blocked:burst:", "Burst Attack Spikes")
		cancel()
	}

	bruteCount := 0
	burstCount := 0
	for _, ev := range copiedEvents {
		if ev.EventType == "BRUTE_FORCE_LOGIN" {
			bruteCount++
		} else if ev.EventType == "BURST_ATTACK" {
			burstCount++
		}
	}

	stats := map[string]any{
		"total_incidents":    len(copiedEvents),
		"blocked_ips_count":  len(blocked),
		"brute_force_count":  bruteCount,
		"burst_attack_count": burstCount,
		"protection_status":  "AKTIF",
		"login_limit":        "5 perc./menit",
		"burst_limit":        "30 req./5 detik",
	}

	return copiedEvents, blocked, stats
}

// UnblockSecurityIP removes restrictions on a given IP
func UnblockSecurityIP(targetIP string) bool {
	unblocked := false

	bucketMu.Lock()
	if _, ok := loginBuckets[targetIP]; ok {
		delete(loginBuckets, targetIP)
		unblocked = true
	}
	bucketMu.Unlock()

	burstMu.Lock()
	if _, ok := burstBuckets[targetIP]; ok {
		delete(burstBuckets, targetIP)
		unblocked = true
	}
	burstMu.Unlock()

	redisMu.RLock()
	rdb := globalRedisClient
	redisMu.RUnlock()
	if rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		_ = rdb.Del(ctx, "rate:login:"+targetIP, "rate:blocked:login:"+targetIP, "rate:burst:"+targetIP, "rate:blocked:burst:"+targetIP).Err()
		unblocked = true
	}

	return unblocked
}

// ClearSecurityEvents empties the recent security event log synchronously.
func ClearSecurityEvents(ctx context.Context) error {
	poolMu.RLock()
	p := globalDBPool
	poolMu.RUnlock()
	if p != nil {
		if ctx == nil {
			ctx = context.Background()
		}
		qCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if _, err := p.Exec(qCtx, `TRUNCATE TABLE security_events`); err != nil {
			return err
		}
	}

	eventsMu.Lock()
	recentEvents = []SecurityEvent{}
	eventsMu.Unlock()
	return nil
}

// LoginRateLimiter limits failed login attempts per client IP
func LoginRateLimiter(maxAttempts int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path != "/api/auth/login" || c.Request.Method != http.MethodPost {
			c.Next()
			return
		}

		ip := c.ClientIP()

		redisMu.RLock()
		rdb := globalRedisClient
		redisMu.RUnlock()

		if rdb != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			blockedKey := "rate:blocked:login:" + ip
			if exists, _ := rdb.Exists(ctx, blockedKey).Result(); exists > 0 {
				cancel()
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Terlalu banyak percobaan login gagal. Silakan tunggu 1 menit sebelum mencoba lagi.",
				})
				c.Abort()
				return
			}

			loginKey := "rate:login:" + ip
			count, err := rdb.Incr(ctx, loginKey).Result()
			if err == nil {
				if count == 1 {
					rdb.Expire(ctx, loginKey, window)
				}
				if count > int64(maxAttempts) {
					rdb.Set(ctx, blockedKey, "1", window)
					cancel()
					RecordSecurityEvent(c, "BRUTE_FORCE_LOGIN", fmt.Sprintf("Terdeteksi %d percobaan login dalam %v", count, window), int(count), "BLOCKED (429)")
					c.JSON(http.StatusTooManyRequests, gin.H{
						"error": "Terlalu banyak percobaan login gagal. Silakan tunggu 1 menit sebelum mencoba lagi.",
					})
					c.Abort()
					return
				}
				cancel()
				c.Next()
				return
			}
			cancel()
		}

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
			RecordSecurityEvent(c, "BRUTE_FORCE_LOGIN", fmt.Sprintf("Terdeteksi %d percobaan login dalam %v", b.attempts, window), b.attempts, "BLOCKED (429)")
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

// BurstAttackLimiter protects entire API from abnormal rapid request floods (> 30 req / 5 detik)
func BurstAttackLimiter(maxBurst int, window time.Duration, blockDuration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		// Skip static file uploads or barcode images
		if !strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/api/uploads/") {
			c.Next()
			return
		}

		ip := c.ClientIP()
		now := time.Now()

		redisMu.RLock()
		rdb := globalRedisClient
		redisMu.RUnlock()

		if rdb != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			blockedKey := "rate:blocked:burst:" + ip
			if exists, _ := rdb.Exists(ctx, blockedKey).Result(); exists > 0 {
				cancel()
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Akses ditahan sementara karena terdeteksi lonjakan permintaan berlebih (burst attack). Coba lagi beberapa saat.",
				})
				c.Abort()
				return
			}

			burstKey := "rate:burst:" + ip
			count, err := rdb.Incr(ctx, burstKey).Result()
			if err == nil {
				if count == 1 {
					rdb.Expire(ctx, burstKey, window)
				}
				if count > int64(maxBurst) {
					rdb.Set(ctx, blockedKey, "1", blockDuration)
					cancel()
					RecordSecurityEvent(c, "BURST_ATTACK", fmt.Sprintf("Lonjakan %d request dalam %v (ambang batas %d)", count, window, maxBurst), int(count), "BLOCKED (429)")
					c.JSON(http.StatusTooManyRequests, gin.H{
						"error": "Peringatan Keamanan: Lonjakan permintaan berlebih terdeteksi (burst attack). Akses dibatasi sementara.",
					})
					c.Abort()
					return
				}
				cancel()
				c.Next()
				return
			}
			cancel()
		}

		burstMu.Lock()
		b, exists := burstBuckets[ip]
		if !exists {
			burstBuckets[ip] = &burstBucket{count: 1, windowStart: now}
			burstMu.Unlock()
			c.Next()
			return
		}

		// Currently blocked
		if now.Before(b.blockedUntil) {
			burstMu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Akses ditahan sementara karena terdeteksi lonjakan permintaan berlebih (burst attack). Coba lagi beberapa saat.",
			})
			c.Abort()
			return
		}

		// Reset window if past duration
		if now.Sub(b.windowStart) > window {
			b.count = 1
			b.windowStart = now
			burstMu.Unlock()
			c.Next()
			return
		}

		b.count++
		if b.count > maxBurst {
			b.blockedUntil = now.Add(blockDuration)
			burstMu.Unlock()

			RecordSecurityEvent(c, "BURST_ATTACK", fmt.Sprintf("Lonjakan %d request dalam %v (ambang batas %d)", b.count, window, maxBurst), b.count, "BLOCKED (429)")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Peringatan Keamanan: Lonjakan permintaan berlebih terdeteksi (burst attack). Akses dibatasi sementara.",
			})
			c.Abort()
			return
		}

		burstMu.Unlock()
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

		_ = gz.Flush()
		_ = io.Discard
	}
}
