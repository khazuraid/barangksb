package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-telegram/bot"

	"inventariskantor/internal/auth"
	"inventariskantor/internal/db"
	"inventariskantor/internal/handlers"
	"inventariskantor/internal/telegram"
)

func main() {
	createUser := flag.String("create-user", "", "email untuk membuat/memperbarui user (password di-prompt)")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	addr := env("APP_PORT", "8080")
	dsn := env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/inventaris?sslmode=disable")
	sessionKey := env("SESSION_KEY", "dev-only-insecure-key")

	// F2: guard produksi — tolak SESSION_KEY default saat APP_ENV=prod
	if os.Getenv("APP_ENV") == "prod" && sessionKey == "dev-only-insecure-key" {
		slog.Error("SESSION_KEY harus diset saat APP_ENV=prod")
		os.Exit(1)
	}

	pool, err := db.Open(dsn)
	if err != nil {
		slog.Error("db open failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := db.Migrate(context.Background(), pool); err != nil {
		slog.Error("migrate failed", "err", err)
		os.Exit(1)
	}

	if *createUser != "" {
		if err := auth.CreateUserInteractive(context.Background(), pool, *createUser); err != nil {
			slog.Error("create user failed", "err", err)
			os.Exit(1)
		}
		return
	}

	sess := auth.NewSessionManager(sessionKey)
	h := handlers.New(pool, sess)
	core := &handlers.Core{Pool: pool}
	h.StartSchedulers()

	// Telegram bot (opsional)
	if tgToken := os.Getenv("TELEGRAM_BOT_TOKEN"); tgToken != "" {
		if tgBot, err := bot.New(tgToken); err == nil {
			tgHandler := telegram.NewHandler(core, os.Getenv("TELEGRAM_ALLOWED_CHAT_IDS"))
			tgHandler.Register(tgBot)
			tgAdv := telegram.NewAdvanced(core, tgToken, os.Getenv("TELEGRAM_ALLOWED_CHAT_IDS"))
			tgAdv.Register(tgBot)
			notif := &telegram.Notifier{Bot: tgBot, Core: core, ChatIDs: tgHandler.ChatIDList()}
			notif.StartDailyAlert(context.Background())
			go tgBot.Start(context.Background())
			slog.Info("telegram bot started")
		} else {
			slog.Error("telegram init failed", "err", err)
		}
	}

	r := chi.NewRouter()
	r.Use(securityHeaders)
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer)
	r.Use(sess.LoadAndSave)

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("web/uploads"))))
	r.Get("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("User-agent: *\nDisallow: /\n"))
	})

	r.Get("/login", h.LoginForm)
	r.Post("/login", h.RateLimitedLogin(h.LoginPost))
	r.Post("/logout", h.LogoutPost)

	// Halaman publik hasil scan QR (tanpa login, read-only)
	r.Get("/scan/{itemID}", h.ScanDetail)
	r.Get("/scan/sku/{sku}", h.ScanBySKU)

	r.Group(func(a chi.Router) {
		a.Use(auth.RequireAuth(sess))
		a.Get("/", h.Dashboard)
		a.Get("/events", h.SSE)
		a.Get("/items", h.ItemsList)
		a.Post("/items", h.ItemsCreate)
		a.Get("/items/new", h.ItemNewForm)
		a.Get("/items/{id}", h.ItemDetail)
		a.Get("/items/{id}/edit", h.ItemEditForm)
		a.Post("/items/{id}", h.ItemsUpdate)
		a.Post("/items/{id}/delete", h.ItemDelete)
		a.Get("/categories", h.CategoriesList)
		a.Post("/categories", h.CategoriesCreate)
		a.Post("/categories/{id}/rename", h.CategoriesRename)
		a.Post("/categories/{id}/delete", h.CategoriesDelete)
		a.Get("/locations", h.LocationsList)
		a.Post("/locations", h.LocationsCreate)
		a.Post("/locations/{id}/rename", h.LocationsRename)
		a.Post("/locations/{id}/delete", h.LocationsDelete)
		a.Get("/movement", h.MovementForm)
		a.Post("/movement", h.MovementPost)
		a.Post("/movement/out", h.MovementOut)
		a.Post("/movement/adjust", h.MovementAdjust)
		a.Get("/history", h.History)
		a.Get("/barcode", h.BarcodePage)
		a.Get("/barcode/{itemID}.png", h.BarcodePNG)
		a.Get("/barcode/sheet", h.BarcodeSheet)
		a.Post("/upload", h.UploadPhoto)
		a.Get("/export/items.csv", h.ExportItemsCSV)
		a.Get("/export/tx.csv", h.ExportTXCSV)
		a.Get("/export/items.xlsx", h.ExportItemsXLSX)
		a.Get("/report.pdf", h.ReportPDF)
		a.Get("/drivesync", h.DriveSyncPage)
		a.Post("/drivesync", h.DriveSyncRun)

		// User management (admin)
		a.Get("/users", h.UsersList)
		a.Post("/users", h.UsersCreate)
		a.Post("/users/{email}/role", h.UsersRole)
		a.Post("/users/{email}/delete", h.UsersDelete)

		// Ganti password sendiri
		a.Get("/password", h.PasswordForm)
		a.Post("/password", h.PasswordPost)

		// F3: opname massal, F4: audit log, F6: arsip foto
		a.Get("/adjust/bulk", h.BulkAdjustForm)
		a.Post("/adjust/bulk", h.BulkAdjustRun)
		a.Get("/audit", h.AuditLogPage)
		a.Get("/photozip", h.PhotoZip)

		// API JSON
		a.Get("/api/items", h.APIItems)
		a.Get("/api/items/{id}", h.APIItem)
		a.Get("/api/transactions", h.APITransactions)
		a.Post("/api/movement", h.APIMovement)
	})

	slog.Info("listening", "addr", ":"+addr)
	if err := http.ListenAndServe(":"+addr, r); err != nil {
		slog.Error("server exited", "err", err)
		os.Exit(1)
	}
}

// securityHeaders sets basic hardening headers.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		next.ServeHTTP(w, r)
	})
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
