package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"inventariskantor/internal/auth"
	"inventariskantor/internal/config"
	"inventariskantor/internal/handler"
	"inventariskantor/internal/middleware"
)

func main() {
	createUser := flag.String("create-user", "", "email untuk membuat user (password di-prompt)")
	flag.Parse()

	cfg := config.Load()
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

	if cfg.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	enforcer, err := auth.NewEnforcer()
	if err != nil {
		slog.Error("casbin init failed", "err", err)
		os.Exit(1)
	}

	_ = enforcer // RBAC ready — wire to middleware later

	authH := handler.NewAuthHandler(pool, cfg)
	itemH := handler.NewItemHandler(pool)
	mvH := handler.NewMovementHandler(pool)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
	}))

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(cfg.JWTSecret))

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
	api.GET("/locations", mvH.Locations)

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
