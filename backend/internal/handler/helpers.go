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

	"inventariskantor/internal/service"
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

// SanitizeSpreadsheetCell neutralizes CSV/XLSX formula injection payloads.
func SanitizeSpreadsheetCell(s string) string {
	return service.SanitizeSpreadsheetCell(s)
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

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func isAllZeros(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] != '0' {
			return false
		}
	}
	return true
}

func parseGroupedInt(parts []string) (string, bool) {
	if len(parts) < 2 || len(parts[0]) < 1 || len(parts[0]) > 3 || !isAllDigits(parts[0]) {
		return "", false
	}
	var b strings.Builder
	b.WriteString(parts[0])
	for _, p := range parts[1:] {
		if len(p) != 3 || !isAllDigits(p) {
			return "", false
		}
		b.WriteString(p)
	}
	return b.String(), true
}

// parseStrictInt64 parses spreadsheet integers safely across Indonesian (1.500.000,00)
// and English (1,500,000.00) formats, rejecting non-integer decimals like "1.5" or "1,5".
func parseStrictInt64(raw string) (int64, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, errors.New("nilai angka kosong")
	}
	upper := strings.ToUpper(s)
	if strings.HasPrefix(upper, "RP") {
		s = strings.TrimSpace(strings.TrimPrefix(s[2:], "."))
	} else if strings.HasPrefix(upper, "IDR") {
		s = strings.TrimSpace(s[3:])
	}
	sign := ""
	if strings.HasPrefix(s, "-") || strings.HasPrefix(s, "+") {
		sign = s[:1]
		s = strings.TrimSpace(s[1:])
	}
	if s == "" {
		return 0, fmt.Errorf("format angka tidak valid: %q", raw)
	}

	hasDot := strings.Contains(s, ".")
	hasComma := strings.Contains(s, ",")
	var digits string

	switch {
	case hasDot && hasComma:
		lastDot := strings.LastIndex(s, ".")
		lastComma := strings.LastIndex(s, ",")
		decSep := ","
		thouSep := "."
		if lastDot > lastComma {
			decSep = "."
			thouSep = ","
		}
		decParts := strings.Split(s, decSep)
		if len(decParts) != 2 || !isAllZeros(decParts[1]) {
			return 0, fmt.Errorf("angka desimal tidak diizinkan: %q", raw)
		}
		intPart := decParts[0]
		if strings.Contains(intPart, thouSep) {
			g, ok := parseGroupedInt(strings.Split(intPart, thouSep))
			if !ok {
				return 0, fmt.Errorf("format pemisah ribuan tidak valid: %q", raw)
			}
			digits = g
		} else if isAllDigits(intPart) {
			digits = intPart
		} else {
			return 0, fmt.Errorf("format angka tidak valid: %q", raw)
		}

	case hasDot || hasComma:
		sep := "."
		if hasComma {
			sep = ","
		}
		parts := strings.Split(s, sep)
		if len(parts) > 2 {
			g, ok := parseGroupedInt(parts)
			if !ok {
				return 0, fmt.Errorf("format pemisah ribuan tidak valid: %q", raw)
			}
			digits = g
		} else {
			left, right := parts[0], parts[1]
			if !isAllDigits(left) || !isAllDigits(right) {
				return 0, fmt.Errorf("format angka tidak valid: %q", raw)
			}
			// Single separator with 3-digit group (e.g. "1.000", "25,500") is a thousands separator
			// unless left is "0" (e.g. "0.000" -> 0, "0.500" -> decimal error).
			if len(right) == 3 && len(left) >= 1 && len(left) <= 3 && left != "0" {
				digits = left + right
			} else if isAllZeros(right) {
				digits = left
			} else {
				return 0, fmt.Errorf("angka harus berupa bilangan bulat: %q", raw)
			}
		}

	default:
		if !isAllDigits(s) {
			return 0, fmt.Errorf("format angka tidak valid: %q", raw)
		}
		digits = s
	}

	return strconv.ParseInt(sign+digits, 10, 64)
}

func parseStrictInt32(raw string) (int32, error) {
	v, err := parseStrictInt64(raw)
	if err != nil {
		return 0, err
	}
	if v < -2147483648 || v > 2147483647 {
		return 0, fmt.Errorf("angka di luar rentang int32: %q", raw)
	}
	return int32(v), nil
}
