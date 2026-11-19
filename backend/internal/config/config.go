package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port                    string
	DBURL                   string
	JWTSecret               string
	JWTTTLHours             int
	CORSOrigins             string
	PublicURL               string
	TrustedProxies          []string
	AllowPrivateTelegramAPI bool
	MinIOEndp               string
	MinIOUser               string
	MinIOPass               string
	MinIOBucket             string
	TgToken                 string
	TgChats                 string
	Env                     string
	RedisURL                string
	RedisHost               string
	RedisPort               string
	RedisPassword           string
	RedisDB                 int
}

func Load() *Config {
	rDB, _ := strconv.Atoi(get("REDIS_DB", "0"))
	jwtTTL, _ := strconv.Atoi(get("JWT_TTL_HOURS", "24"))
	if jwtTTL <= 0 || jwtTTL > 168 {
		jwtTTL = 24
	}
	proxiesRaw := get("TRUSTED_PROXIES", "127.0.0.1,::1,172.16.0.0/12")
	var trustedProxies []string
	for _, p := range strings.Split(proxiesRaw, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			trustedProxies = append(trustedProxies, p)
		}
	}
	cfg := &Config{
		Port:                    get("PORT", "8080"),
		DBURL:                   get("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/inventaris?sslmode=disable"),
		JWTSecret:               get("JWT_SECRET", "dev-insecure-secret-inventaris-kantor-2024-change-in-prod"),
		JWTTTLHours:             jwtTTL,
		CORSOrigins:             get("CORS_ORIGINS", "http://localhost:3000,http://localhost:5173"),
		PublicURL:               strings.TrimRight(strings.TrimSpace(os.Getenv("APP_PUBLIC_URL")), "/"),
		TrustedProxies:          trustedProxies,
		AllowPrivateTelegramAPI: strings.EqualFold(os.Getenv("ALLOW_PRIVATE_TELEGRAM_API"), "true"),
		MinIOEndp:               get("MINIO_ENDPOINT", "localhost:9000"),
		MinIOUser:               get("MINIO_ROOT_USER", "minioadmin"),
		MinIOPass:               get("MINIO_ROOT_PASSWORD", "minioadmin"),
		MinIOBucket:             get("MINIO_BUCKET", "inventaris"),
		TgToken:                 get("TELEGRAM_BOT_TOKEN", ""),
		TgChats:                 get("TELEGRAM_ALLOWED_CHAT_IDS", ""),
		Env:                     get("APP_ENV", "dev"),
		RedisURL:                get("REDIS_URL", ""),
		RedisHost:               get("REDIS_HOST", "localhost"),
		RedisPort:               get("REDIS_PORT", "6379"),
		RedisPassword:           get("REDIS_PASSWORD", ""),
		RedisDB:                 rDB,
	}
	return cfg
}

// Validate checks critical config in production.
func (c *Config) Validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET must not be empty")
	}
	if !c.IsProd() && c.JWTSecret == "dev-insecure-secret-inventaris-kantor-2024-change-in-prod" {
		slog.Warn("using default development JWT_SECRET — set APP_ENV=prod and a random 32+ char JWT_SECRET in production")
	}
	if c.IsProd() {
		if len(c.JWTSecret) < 32 || c.JWTSecret == "dev-insecure-secret-inventaris-kantor-2024-change-in-prod" {
			return fmt.Errorf("JWT_SECRET must be set to a random string of at least 32 characters in production")
		}
		if c.CORSOrigins == "" || c.CORSOrigins == "*" || c.CORSOrigins == "http://localhost:3000,http://localhost:5173" {
			return fmt.Errorf("CORS_ORIGINS must be set to production origins in production")
		}
		if (c.RedisPassword == "" || c.RedisPassword == "dev-redis-secure-password-2024") && c.RedisURL == "" {
			return fmt.Errorf("REDIS_PASSWORD must be configured to a strong password in production")
		}
		if strings.Contains(c.RedisURL, "dev-redis-secure-password-2024") {
			return fmt.Errorf("REDIS_URL must not use default dev password in production")
		}
		if os.Getenv("ADMIN_PASSWORD") == "admin12345" {
			return fmt.Errorf("ADMIN_PASSWORD must not use default dev password in production")
		}
	}
	return nil
}

func get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func (c *Config) IsProd() bool { return c.Env == "prod" }

var PerPageOptions = []int{25, 50, 100}

func ParsePerPage(s string) int {
	n, _ := strconv.Atoi(s)
	for _, opt := range PerPageOptions {
		if n == opt {
			return n
		}
	}
	return 25
}