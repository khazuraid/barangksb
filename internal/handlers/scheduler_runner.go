package handlers

import (
	"context"
	"log/slog"
	"net/http/httptest"
	"os"
	"time"

	"inventariskantor/internal/pdf"
	"inventariskantor/internal/scheduler"
)

// StartSchedulers runs cron jobs: hourly drive import, monthly PDF to Drive.
func (h *Handlers) StartSchedulers() {
	ctx := context.Background()

	// Import CSV dari Drive tiap jam (jika dikonfigurasi)
	if os.Getenv("GOOGLE_SA_JSON") != "" && os.Getenv("GOOGLE_FOLDER_ID") != "" {
		scheduler.Hourly(ctx, func(c context.Context) {
			res, err := h.driveImport(c)
			if err != nil {
				slog.Error("scheduled drive import failed", "err", err)
				return
			}
			h.driveLastRun = timeNowWIB()
			h.driveLastResult = res
			slog.Info("scheduled drive import done", "result", res)
		})
	}

	// PDF laporan bulanan → Drive (tanggal 1 tiap bulan, 06:00 WIB)
	scheduler.Monthly(1, func(c context.Context) {
		items, err := h.svc.AllItems(c)
		if err != nil {
			slog.Error("monthly pdf: load items", "err", err)
			return
		}
		txs, err := h.svc.AllTX(c)
		if err != nil {
			slog.Error("monthly pdf: load txs", "err", err)
			return
		}
		data, err := pdf.BuildReport(items, txs)
		if err != nil {
			slog.Error("monthly pdf: build", "err", err)
			return
		}
		folder := os.Getenv("GOOGLE_DRIVE_REPORT_FOLDER_ID")
		sa := os.Getenv("GOOGLE_SA_JSON")
		if folder == "" || sa == "" {
			slog.Info("monthly pdf: drive not configured, skip upload")
			return
		}
		name := "laporan-inventaris-" + time.Now().Format("2006-01") + ".pdf"
		if _, err := driveUpload(sa, folder, name, data, "application/pdf"); err != nil {
			slog.Error("monthly pdf: upload", "err", err)
			return
		}
		slog.Info("monthly pdf uploaded", "name", name)
	})
}

func timeNowWIB() string {
	return time.Now().In(time.FixedZone("WIB", 7*3600)).Format("02-01-2006 15:04")
}

var _ = httptest.NewServer
