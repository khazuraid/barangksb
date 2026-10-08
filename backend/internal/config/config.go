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
	Env          string
}

func Load() *Config {
	cfg := &Config{
		Port:        get("PORT", "8080"),
		DBURL:       get("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/inventaris?sslmode=disable"),
		JWTSecret:   get("JWT_SECRET", ""),
		CORSOrigins: get("CORS_ORIGINS", "http://localhost:3000,http://localhost:5173"),
		MinIOEndp:   get("MINIO_ENDPOINT", "localhost:9000"),
		MinIOUser:   get("MINIO_ROOT_USER", "minioadmin"),
		MinIOPass:   get("MINIO_ROOT_PASSWORD", "minioadmin"),
		MinIOBucket: get("MINIO_BUCKET", "inventaris"),
		TgToken:     get("TELEGRAM_BOT_TOKEN", ""),
		TgChats:     get("TELEGRAM_ALLOWED_CHAT_IDS", ""),
		Env:         get("APP_ENV", "dev"),
	}
	return cfg
}

// Validate checks critical config in production.
func (c *Config) Validate() error {
	if c.IsProd() {
		if c.JWTSecret == "" || len(c.JWTSecret) < 32 {
			return fmt.Errorf("JWT_SECRET must be set to a random string of at least 32 characters in production")
		}
		if c.CORSOrigins == "" || c.CORSOrigins == "http://localhost:3000,http://localhost:5173" {
			return fmt.Errorf("CORS_ORIGINS must be set to production origins in production")
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