package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"inventariskantor/internal/auth"
	"inventariskantor/internal/config"
	"inventariskantor/internal/db"
	"inventariskantor/internal/handler"
	"inventariskantor/internal/middleware"
	"inventariskantor/internal/telegram"
)

//go:embed all:dist
var frontendDist embed.FS

func dbMigrate(ctx context.Context, pool *pgxpool.Pool) error {
	return db.Migrate(ctx, pool)
}

func seedAdminUser(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config) {
	var count int
	pool.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&count)
	if count > 0 {
		return
	}
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@kantor.id"
	}
	adminPw := os.Getenv("ADMIN_PASSWORD")
	if adminPw == "" {
		adminPw = "admin12345"
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(adminPw), bcrypt.DefaultCost)
	_, err := pool.Exec(ctx,
		`INSERT INTO users (name, email, password_hash, role) VALUES ($1,$2,$3,'admin')`,
		"Admin", adminEmail, string(hash))
	if err != nil {
		slog.Error("seed admin failed", "err", err)
		return
	}
	slog.Info("admin user seeded", "email", adminEmail)
}

func main() {
	createUser := flag.String("create-user", "", "email untuk membuat user (password di-prompt)")
	flag.Parse()

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		slog.Error("config validation failed", "err", err)
		os.Exit(1)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	pool, err := pgxpool.New(context.Background(), cfg.DBURL)
	if err != nil {
		slog.Error("db open failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if *createUser != "" {
		name := readLine("Nama lengkap: ")
		pw := readLine("Password: ")
		hash, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
		role := "admin"
		pool.Exec(context.Background(),
			`INSERT INTO users (name, email, password_hash, role) VALUES ($1,$2,$3,$4)
			 ON CONFLICT (email) DO UPDATE SET name=$1, password_hash=$3, role=$4`,
			name, *createUser, string(hash), role)
		slog.Info("user created", "email", *createUser)
		return
	}

	// Run migrations
	if err := dbMigrate(context.Background(), pool); err != nil {
		slog.Error("migrate failed", "err", err)
		os.Exit(1)
	}

	// Auto-seed admin user from env if DB empty
	seedAdminUser(context.Background(), pool, cfg)

	if cfg.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	enforcer, err := auth.NewEnforcer()
	if err != nil {
		slog.Error("casbin init failed", "err", err)
		os.Exit(1)
	}

	// RBAC middleware applied after AuthMiddleware for all authenticated routes.
	// Public routes (login, scan, barcode) skip auth entirely so never reach RBAC.

	authH := handler.NewAuthHandler(pool, cfg)
	itemH := handler.NewItemHandler(pool)
	mvH := handler.NewMovementHandler(pool)
	userH := handler.NewUserHandler(pool)
	bcH := handler.NewBarcodeHandler(pool)
	expH := handler.NewExportHandler(pool)
	upH := handler.NewUploadHandler(pool, cfg.MinIOEndp, cfg.MinIOUser, cfg.MinIOPass, cfg.MinIOBucket)
	rptH := handler.NewReportHandler(pool)
	bulkH := handler.NewBulkHandler(pool)
	setH := handler.NewSettingHandler(pool)

	r := gin.New()
	allowedOrigins := strings.Split(cfg.CORSOrigins, ",")
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	api.Use(middleware.RBACMiddleware(enforcer))

	api.POST("/auth/login", authH.Login)
	api.GET("/auth/me", authH.Me)
	api.POST("/auth/password", authH.ChangePassword)

	api.GET("/dashboard", itemH.Dashboard)

	api.GET("/items", itemH.List)
	api.POST("/items", itemH.Create)
	api.GET("/items/:id", itemH.Get)
	api.PUT("/items/:id", itemH.Update)
	api.DELETE("/items/:id", itemH.Delete)

	api.POST("/movement/in", mvH.StockIn)
	api.POST("/movement/out", mvH.StockOut)
	api.POST("/movement/adjust", mvH.Adjust)
	api.GET("/transactions", mvH.ListTransactions)

	api.GET("/categories", mvH.Categories)
	api.POST("/categories", userH.CreateCategory)
	api.DELETE("/categories/:id", userH.DeleteCategory)

	api.GET("/locations", mvH.Locations)
	api.POST("/locations", userH.CreateLocation)
	api.DELETE("/locations/:id", userH.DeleteLocation)

	api.GET("/users", userH.List)
	api.POST("/users", userH.Create)
	api.PUT("/users/:email/role", userH.UpdateRole)
	api.DELETE("/users/:email", userH.Delete)

	api.GET("/audit", userH.AuditLog)

	api.GET("/barcode/:id", bcH.PNG)
	api.GET("/barcode/sheet", bcH.Sheet)

	// Public scan endpoint — returns item detail for QR scan landing page
	api.GET("/scan/:id", bcH.ScanItem)

	api.GET("/export/items.csv", expH.ItemsCSV)
	api.GET("/export/tx.csv", expH.TxCSV)
	api.GET("/export/items.xlsx", rptH.ItemsXLSX)
	api.GET("/report.pdf", rptH.PDF)

	api.POST("/upload", upH.Upload)
	api.POST("/adjust/bulk", bulkH.BulkAdjust)

	// Telegram settings (admin only)
	api.GET("/settings/telegram", setH.GetTelegram)
	api.PUT("/settings/telegram", setH.UpdateTelegram)
	api.GET("/settings/telegram/bot", setH.GetBotInfo)
	api.GET("/settings/telegram/subscribers", setH.ListSubscribers)
	api.POST("/settings/telegram/subscribers", setH.AddSubscriber)
	api.PUT("/settings/telegram/subscribers/:chatId", setH.UpdateSubscriber)
	api.DELETE("/settings/telegram/subscribers/:chatId", setH.DeleteSubscriber)
	api.POST("/settings/telegram/test", setH.TestTelegram)
	api.GET("/settings/telegram/logs", setH.TelegramLogs)
	api.GET("/settings/telegram/commands", setH.ListCommands)
	api.POST("/settings/telegram/commands", setH.CreateCommand)
	api.PUT("/settings/telegram/commands/:id", setH.UpdateCommand)
	api.DELETE("/settings/telegram/commands/:id", setH.DeleteCommand)

	// Static files for local uploads (disk fallback)
	r.Static("/uploads", "./uploads")

	// Serve embedded frontend (SPA) — for Zeabur single-container deploy
	frontendFS, _ := fs.Sub(frontendDist, "dist")
	r.NoRoute(func(c *gin.Context) {
		p := c.Request.URL.Path
		if strings.HasPrefix(p, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if p == "/" {
			p = "/index.html"
		}
		cleanPath := path.Clean(strings.TrimPrefix(p, "/"))
		if data, err := fs.ReadFile(frontendFS, cleanPath); err == nil {
			c.Data(http.StatusOK, guessContentType(cleanPath), data)
			return
		}
		indexData, _ := fs.ReadFile(frontendFS, "index.html")
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexData)
	})

	// Serve MinIO uploads via API (when MinIO is active)
	api.GET("/uploads/:name", func(c *gin.Context) {
		name := c.Param("name")
		obj, err := upH.GetObject(c.Request.Context(), name)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		defer obj.Close()
		c.DataFromReader(http.StatusOK, -1, "image/jpeg", obj, nil)
	})

	// Telegram bot
	if cfg.TgToken != "" {
		tgBot := telegram.NewBot(cfg.TgToken, cfg.TgChats, pool)
		tgBot.Start(context.Background())
	}

	slog.Info("server starting", "port", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func readLine(prompt string) string {
	print(prompt)
	var s string
	fmtScanln(&s)
	return s
}

// avoid importing fmt just for Scanln
func fmtScanln(s *string) {
	var b [1024]byte
	n, _ := os.Stdin.Read(b[:])
	*s = string(b[:n])
	// trim newline
	for len(*s) > 0 && ((*s)[len(*s)-1] == '\n' || (*s)[len(*s)-1] == '\r') {
		*s = (*s)[:len(*s)-1]
	}
}

func guessContentType(name string) string {
	switch path.Ext(name) {
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js", ".mjs":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".woff", ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".eot":
		return "application/vnd.ms-fontobject"
	case ".map":
		return "application/json; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}

var _ = fmt.Sprintf
