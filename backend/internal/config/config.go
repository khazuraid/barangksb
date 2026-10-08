package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port       string
	DBURL      string
	JWTSecret  string
	MinIOEndp  string
	MinIOUser  string
	MinIOPass  string
	MinIOBucket string
	TgToken    string
	TgChats    string
	Env        string
}

func Load() *Config {
	return &Config{
		Port:       get("PORT", "8080"),
		DBURL:      get("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/inventaris?sslmode=disable"),
		JWTSecret:  get("JWT_SECRET", "dev-secret-change-in-prod"),
		MinIOEndp:  get("MINIO_ENDPOINT", "localhost:9000"),
		MinIOUser:  get("MINIO_ROOT_USER", "minioadmin"),
		MinIOPass:  get("MINIO_ROOT_PASSWORD", "minioadmin"),
		MinIOBucket: get("MINIO_BUCKET", "inventaris"),
		TgToken:    get("TELEGRAM_BOT_TOKEN", ""),
		TgChats:    get("TELEGRAM_ALLOWED_CHAT_IDS", ""),
		Env:        get("APP_ENV", "dev"),
	}
}

func get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func (c *Config) IsProd() bool { return c.Env == "prod" }

// PerPage options
var PerPageOptions = []int{20, 25, 50, 100}

func ParsePerPage(s string) int {
	n, _ := strconv.Atoi(s)
	for _, opt := range PerPageOptions {
		if n == opt {
			return n
		}
	}
	return 25
}
