package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func fmtSprintf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}

func configPerPage(c *gin.Context) int {
	pp, err := strconv.Atoi(c.DefaultQuery("per_page", "25"))
	if err != nil || pp <= 0 {
		return 25
	}
	switch pp {
	case 12, 20, 24, 25, 48, 50, 96, 100:
		return pp
	default:
		if pp > 1000 {
			return 1000
		}
		return pp
	}
}

type paginatedResp struct {
	Data    any `json:"data"`
	Total   int `json:"total"`
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Pages   int `json:"pages"`
}

func nullable(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// internalError logs the full error server-side and returns a generic message to the client.
func internalError(c *gin.Context, publicMsg string, err error) {
	if err != nil {
		slog.Error(publicMsg, "path", c.Request.URL.Path, "method", c.Request.Method, "err", err)
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": publicMsg})
}

// SanitizeSpreadsheetCell neutralizes CSV/XLSX formula injection payloads
// (values starting with =, +, -, @, tab, or carriage return).
func SanitizeSpreadsheetCell(s string) string {
	if s == "" {
		return s
	}
	trimmed := strings.TrimLeft(s, " \t\r\n")
	if trimmed == "" {
		return s
	}
	switch s[0] {
	case '=', '+', '-', '@', '\t', '\r', '\n':
		return "'" + s
	}
	switch trimmed[0] {
	case '=', '+', '-', '@':
		return "'" + s
	}
	return s
}

// SanitizePhotoURL validates a stored photo URL to prevent javascript:/data: Stored XSS
// and protocol-relative open redirects.
// Permits empty string, relative /api/uploads/... or /uploads/... paths, and http/https URLs.
func SanitizePhotoURL(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", nil
	}
	if len(s) > 2048 {
		return "", errors.New("photo_url terlalu panjang")
	}
	// Reject control chars / newlines
	if strings.ContainsAny(s, "\r\n\x00") {
		return "", errors.New("photo_url tidak valid")
	}
	if strings.HasPrefix(s, "/") {
		if strings.HasPrefix(s, "//") || strings.Contains(s, "..") {
			return "", errors.New("photo_url relatif tidak valid")
		}
		if strings.HasPrefix(s, "/api/uploads/") || strings.HasPrefix(s, "/uploads/") {
			return s, nil
		}
		return "", errors.New("photo_url harus berupa path /api/uploads/ atau URL http(s)")
	}
	u, err := url.Parse(s)
	if err != nil || u.Host == "" {
		return "", errors.New("photo_url tidak valid")
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "https" && scheme != "http" {
		return "", errors.New("photo_url hanya mengizinkan skema http atau https")
	}
	return u.String(), nil
}

// isBlockedIP returns true if ip is loopback, private, link-local, multicast, unspecified, or CGNAT.
func isBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() || ip.IsUnspecified() || ip.IsInterfaceLocalMulticast() {
		return true
	}
	// CGNAT 100.64.0.0/10 and cloud metadata 169.254.0.0/16 (also covered by link-local)
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127 {
			return true
		}
		if ip4[0] == 0 {
			return true
		}
	}
	return false
}

// ValidateSafeExternalURL validates an external API URL (e.g. Telegram Bot API base URL)
// against SSRF: requires HTTPS (unless allowPrivate is true) and blocks private/loopback/metadata IPs.
func ValidateSafeExternalURL(raw string, allowPrivate bool) (*url.URL, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil, nil
	}
	u, err := url.Parse(s)
	if err != nil || u.Host == "" {
		return nil, errors.New("URL tidak valid")
	}
	scheme := strings.ToLower(u.Scheme)
	if !allowPrivate && scheme != "https" {
		return nil, errors.New("URL API Telegram harus menggunakan protokol https://")
	}
	if allowPrivate && scheme != "https" && scheme != "http" {
		return nil, errors.New("URL harus menggunakan protokol http:// atau https://")
	}
	if u.User != nil {
		return nil, errors.New("URL tidak boleh mengandung kredensial userinfo")
	}
	host := u.Hostname()
	if host == "" {
		return nil, errors.New("hostname URL tidak valid")
	}
	if !allowPrivate {
		lowerHost := strings.ToLower(host)
		if lowerHost == "localhost" || strings.HasSuffix(lowerHost, ".localhost") || strings.HasSuffix(lowerHost, ".internal") || strings.HasSuffix(lowerHost, ".local") {
			return nil, errors.New("hostname internal/localhost tidak diizinkan")
		}
		if ip := net.ParseIP(host); ip != nil {
			if isBlockedIP(ip) {
				return nil, errors.New("alamat IP privat/internal tidak diizinkan")
			}
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil || len(ips) == 0 {
				return nil, errors.New("hostname tidak dapat di-resolve")
			}
			for _, ipAddr := range ips {
				if isBlockedIP(ipAddr.IP) {
					return nil, errors.New("hostname mengarah ke alamat IP privat/internal yang diblokir")
				}
			}
		}
	}
	return u, nil
}

// NewSafeHTTPClient builds an HTTP client that blocks connections to private/loopback/metadata IPs
// at dial time (mitigating DNS rebinding SSRF) and disables cross-host redirects to internal IPs.
func NewSafeHTTPClient(timeout time.Duration, allowPrivate bool) *http.Client {
	baseDialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConns:        20,
		IdleConnTimeout:     60 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if allowPrivate {
				return baseDialer.DialContext(ctx, network, addr)
			}
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil || len(ips) == 0 {
				return nil, fmt.Errorf("resolve host failed")
			}
			var safeIP net.IP
			for _, ipAddr := range ips {
				if isBlockedIP(ipAddr.IP) {
					return nil, errors.New("connection to private/internal IP blocked")
				}
				if safeIP == nil {
					safeIP = ipAddr.IP
				}
			}
			return baseDialer.DialContext(ctx, network, net.JoinHostPort(safeIP.String(), port))
		},
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
