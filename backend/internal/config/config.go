package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port         string
	DBURL        string
	JWTSecret    string
	CORSOrigins  string
	MinIOEndp    string
	MinIOUser    string
	MinIOPass    string
	MinIOBucket  string
	TgToken      string
	TgChats      string
	Env           string
	RedisURL      string
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
}

func Load() *Config {
	rDB, _ := strconv.Atoi(get("REDIS_DB", "0"))
	cfg := &Config{
		Port:          get("PORT", "8080"),
		DBURL:         get("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/inventaris?sslmode=disable"),
		JWTSecret:     get("JWT_SECRET", "dev-insecure-secret-inventaris-kantor-2024-change-in-prod"),
		CORSOrigins:   get("CORS_ORIGINS", "http://localhost:3000,http://localhost:5173"),
		MinIOEndp:     get("MINIO_ENDPOINT", "localhost:9000"),
		MinIOUser:     get("MINIO_ROOT_USER", "minioadmin"),
		MinIOPass:     get("MINIO_ROOT_PASSWORD", "minioadmin"),
		MinIOBucket:   get("MINIO_BUCKET", "inventaris"),
		TgToken:       get("TELEGRAM_BOT_TOKEN", ""),
		TgChats:       get("TELEGRAM_ALLOWED_CHAT_IDS", ""),
		Env:           get("APP_ENV", "dev"),
		RedisURL:      get("REDIS_URL", ""),
		RedisHost:     get("REDIS_HOST", "localhost"),
		RedisPort:     get("REDIS_PORT", "6379"),
		RedisPassword: get("REDIS_PASSWORD", ""),
		RedisDB:       rDB,
	}
	return cfg
}

// Validate checks critical config in production.
func (c *Config) Validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET must not be empty")
	}
	if c.IsProd() {
		if len(c.JWTSecret) < 32 || c.JWTSecret == "dev-insecure-secret-inventaris-kantor-2024-change-in-prod" {
			return fmt.Errorf("JWT_SECRET must be set to a random string of at least 32 characters in production")
		}
		if c.CORSOrigins == "" || c.CORSOrigins == "http://localhost:3000,http://localhost:5173" {
			return fmt.Errorf("CORS_ORIGINS must be set to production origins in production")
		}
		if c.RedisPassword == "" && c.RedisURL == "" {
			return fmt.Errorf("REDIS_PASSWORD must be configured in production")
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