package handlers

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"inventariskantor/internal/views"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func (h *Handlers) DriveSyncPage(w http.ResponseWriter, r *http.Request) {
	d := views.DriveSyncData{
		User: userInfo(r), Configured: driveConfigured(),
		LastRun: h.driveLastRun, Result: h.driveLastResult,
	}
	h.show(w, r, "Sinkronisasi", views.DriveSync(d))
}

func (h *Handlers) DriveSyncRun(w http.ResponseWriter, r *http.Request) {
	if !driveConfigured() {
		http.Redirect(w, r, "/drivesync", http.StatusSeeOther)
		return
	}
	res, err := h.driveImport(r.Context())
	if err != nil {
		http.Error(w, "Drive sync gagal: "+err.Error(), http.StatusInternalServerError)
		return
	}
	h.driveLastRun = time.Now().In(wibLoc()).Format("02-01-2006 15:04 WIB")
	h.driveLastResult = res
	http.Redirect(w, r, "/drivesync", http.StatusSeeOther)
}

func driveConfigured() bool {
	return os.Getenv("GOOGLE_SA_JSON") != "" && os.Getenv("GOOGLE_FOLDER_ID") != ""
}

// driveImport lists CSV/Sheet files in GOOGLE_FOLDER_ID, parses them, and
// upserts inventory_items by SKU. One-way: Drive → DB.
func (h *Handlers) driveImport(ctx context.Context) (string, error) {
	sa, err := os.ReadFile(os.Getenv("GOOGLE_SA_JSON"))
	if err != nil {
		return "", err
	}
	conf, err := google.JWTConfigFromJSON(sa, drive.DriveReadonlyScope)
	if err != nil {
		return "", err
	}
	srv, err := drive.NewService(ctx, option.WithTokenSource(conf.TokenSource(ctx)))
	if err != nil {
		return "", err
	}
	q := fmt.Sprintf("'%s' in parents and trashed = false", os.Getenv("GOOGLE_FOLDER_ID"))
	files, err := srv.Files.List().Q(q).Fields("files(id,name,mimeType)").Do()
	if err != nil {
		return "", err
	}

	var ins, upd int
	for _, f := range files.Files {
		var rows [][]string
		switch {
		case f.MimeType == "application/vnd.google-apps.spreadsheet":
			resp, err := srv.Files.Export(f.Id, "text/csv").Download()
			if err != nil {
				continue
			}
			rows, err = readCSV(resp)
			if err != nil {
				continue
			}
		case strings.HasSuffix(strings.ToLower(f.Name), ".csv"):
			resp, err := srv.Files.Get(f.Id).Download()
			if err != nil {
				continue
			}
			rows, err = readCSV(resp)
			if err != nil {
				continue
			}
		default:
			continue
		}
		i, u := h.upsertCSVRows(ctx, rows)
		ins += i
		upd += u
	}
	return fmt.Sprintf("insert: %d, update: %d", ins, upd), nil
}

func readCSV(resp *http.Response) ([][]string, error) {
	defer resp.Body.Close()
	rd := csv.NewReader(resp.Body)
	rd.TrimLeadingSpace = true
	return rd.ReadAll()
}

// upsertCSVRows maps MASTER_HEADERS-style CSV into inventory_items by SKU column.
func (h *Handlers) upsertCSVRows(ctx context.Context, rows [][]string) (ins, upd int) {
	if len(rows) < 2 {
		return 0, 0
	}
	header := rows[0]
	col := map[string]int{}
	for i, hh := range header {
		col[strings.TrimSpace(strings.ToLower(hh))] = i
	}
	skuI := colOr(col, "kode barcode / sku")
	if skuI < 0 {
		skuI = colOr(col, "sku")
	}
	if skuI < 0 {
		return 0, 0
	}
	nameI := colOr(col, "nama barang")
	catI := colOr(col, "kategori")
	locI := colOr(col, "lokasi simpan / ruangan")
	unitI := colOr(col, "satuan")

	for _, row := range rows[1:] {
		if len(row) <= skuI || strings.TrimSpace(row[skuI]) == "" {
			continue
		}
		sku := strings.TrimSpace(row[skuI])
		name := cellAt(row, nameI)
		if name == "" {
			continue
		}
		cat := cellDef(row, catI, "Lainnya")
		loc := cellDef(row, locI, "-")
		unit := cellDef(row, unitI, "Unit")

		tag, err := h.pool.Exec(ctx, `INSERT INTO inventory_items (sku, name, category, location, unit)
			VALUES ($1,$2,$3,$4,$5)
			ON CONFLICT (sku) DO UPDATE SET name=$2, category=$3, location=$4, unit=$5, updated_at=now()`,
			sku, name, cat, loc, unit)
		if err != nil {
			continue
		}
		if tag.Insert() {
			ins++
		} else {
			upd++
		}
	}
	return ins, upd
}

func colOr(col map[string]int, key string) int { return col[key] }

func cellAt(row []string, i int) string {
	if i < 0 || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func cellDef(row []string, i int, def string) string {
	if s := cellAt(row, i); s != "" {
		return s
	}
	return def
}

func wibLoc() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.FixedZone("WIB", 7*3600)
	}
	return loc
}

var (
	_ = io.EOF
	_ = bytes.MinRead
)
