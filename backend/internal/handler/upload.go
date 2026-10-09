package handler

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type UploadHandler struct {
	pool    *pgxpool.Pool
	minio   *minio.Client
	bucket  string
	useDisk bool // fallback when MinIO not configured
}

func NewUploadHandler(pool *pgxpool.Pool, endpoint, accessKey, secretKey, bucket string) *UploadHandler {
	h := &UploadHandler{pool: pool, bucket: bucket, useDisk: true}
	// Only initialize MinIO if explicitly configured, not default localhost/docker
	if endpoint != "" && accessKey != "" && secretKey != "" &&
		endpoint != "minio:9000" && endpoint != "localhost:9000" && endpoint != "127.0.0.1:9000" {
		mc, err := minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: false,
		})
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			exists, errBucket := mc.BucketExists(ctx, bucket)
			if errBucket == nil {
				if !exists {
					_ = mc.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
				}
				h.minio = mc
				h.useDisk = false
				slog.Info("MinIO storage initialized", "endpoint", endpoint, "bucket", bucket)
			} else {
				slog.Warn("MinIO not reachable, using disk storage", "endpoint", endpoint, "err", errBucket)
			}
		}
	} else {
		slog.Info("using disk storage for uploads", "dir", getUploadDir())
	}
	return h
}

type GeoTag struct {
	Lat, Lng, Acc float64
	Valid         bool
	At            time.Time
	Name          string
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func (h *UploadHandler) Upload(c *gin.Context) {
	// Restrict max upload size to 12 MB to prevent memory exhaustion DoS
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 12<<20)

	file, hdr, err := c.Request.FormFile("photo")
	if err != nil || file == nil || hdr.Size == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file photo wajib (maksimal 12 MB)"})
		return
	}
	defer file.Close()

	var buf bytes.Buffer
	buf.ReadFrom(file)

	geo := GeoTag{
		Lat:  parseFloat(c.PostForm("geo_lat")),
		Lng:  parseFloat(c.PostForm("geo_lng")),
		Acc:  parseFloat(c.PostForm("geo_acc")),
		Name: c.PostForm("geo_name"),
		At:   time.Now(),
	}
	geo.Valid = geo.Lat != 0 && geo.Lng != 0
	petugas := c.GetString("name")
	clientStamped := c.PostForm("client_stamped") == "true"

	processed, err := ProcessPhotoBytes(buf.Bytes(), geo, petugas, clientStamped)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := uuid.NewString() + ".jpg"

	if !h.useDisk && h.minio != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()
		_, err := h.minio.PutObject(ctx, h.bucket, name, bytes.NewReader(processed), int64(len(processed)), minio.PutObjectOptions{
			ContentType: "image/jpeg",
		})
		if err != nil {
			// fallback to disk
			h.saveDisk(name, processed)
		}
		resp := gin.H{"url": fmt.Sprintf("/api/uploads/%s", name)}
		if geo.Valid {
			resp["geo_lat"] = geo.Lat
			resp["geo_lng"] = geo.Lng
			resp["geo_acc"] = geo.Acc
			resp["geo_name"] = geo.Name
		}
		c.JSON(http.StatusOK, resp)
		return
	}

	// disk fallback
	url := h.saveDisk(name, processed)
	resp := gin.H{"url": url}
	if geo.Valid {
		resp["geo_lat"] = geo.Lat
		resp["geo_lng"] = geo.Lng
		resp["geo_acc"] = geo.Acc
		resp["geo_name"] = geo.Name
	}
	c.JSON(http.StatusOK, resp)
}

func getUploadDir() string {
	if fi, err := os.Stat("/uploads"); err == nil && fi.IsDir() {
		return "/uploads"
	}
	_ = os.MkdirAll("uploads", 0o755)
	return "uploads"
}

func (h *UploadHandler) saveDisk(name string, data []byte) string {
	dir := getUploadDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		slog.Error("failed writing to primary upload dir", "path", path, "err", err)
		_ = os.MkdirAll("uploads", 0o755)
		fallbackPath := filepath.Join("uploads", name)
		if err2 := os.WriteFile(fallbackPath, data, 0o644); err2 != nil {
			slog.Error("failed writing to fallback upload dir", "path", fallbackPath, "err", err2)
		} else {
			slog.Info("saved photo to fallback dir", "path", fallbackPath, "bytes", len(data))
		}
	} else {
		slog.Info("saved photo to primary dir", "path", path, "bytes", len(data))
	}
	return "/api/uploads/" + name
}

// GetObject retrieves a file from MinIO (or nil if using disk)
func (h *UploadHandler) GetObject(ctx context.Context, name string) (io.ReadCloser, error) {
	if h.minio != nil && !h.useDisk {
		obj, err := h.minio.GetObject(ctx, h.bucket, name, minio.GetObjectOptions{})
		if err == nil {
			if _, statErr := obj.Stat(); statErr == nil {
				return obj, nil
			}
			obj.Close()
		}
	}
	// disk fallback - check both getUploadDir() and local uploads/
	dir := getUploadDir()
	f, err := os.Open(filepath.Join(dir, name))
	if err == nil {
		return f, nil
	}
	fFallback, errFallback := os.Open(filepath.Join("uploads", name))
	if errFallback == nil {
		return fFallback, nil
	}
	slog.Error("photo file not found on disk", "name", name, "primaryDir", dir, "err", err, "fallbackErr", errFallback)
	return nil, err
}

// Serve streams uploaded image from MinIO or disk
func (h *UploadHandler) Serve(c *gin.Context) {
	name := filepath.Base(c.Param("name"))
	if name == "" || name == "." {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid file name"})
		return
	}
	reader, err := h.GetObject(c.Request.Context(), name)
	if err != nil {
		slog.Error("serve image failed: file not found", "name", name, "err", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found: " + name})
		return
	}
	defer reader.Close()

	c.Header("Content-Type", "image/jpeg")
	c.Header("Cache-Control", "public, max-age=86400")
	if c.Query("download") == "true" || c.Query("download") == "1" {
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", name))
	}
	io.Copy(c.Writer, reader)
}

// DeleteUploadedPhoto deletes physical file from storage to keep disk clean
func DeleteUploadedPhoto(photoURL string) {
	if photoURL == "" {
		return
	}
	name := filepath.Base(photoURL)
	if name == "" || name == "." || name == "/" {
		return
	}
	dir := getUploadDir()
	_ = os.Remove(filepath.Join(dir, name))
	_ = os.Remove(filepath.Join("uploads", name))
	slog.Info("deleted uploaded photo file", "name", name)
}

type PhotoReference struct {
	Type  string `json:"type"` // "item", "transaction", "maintenance"
	ID    string `json:"id"`
	Title string `json:"title"`
	Sub   string `json:"sub"`
}

type PhotoItem struct {
	Filename   string           `json:"filename"`
	URL        string           `json:"url"`
	Size       int64            `json:"size"`
	ModTime    time.Time        `json:"mod_time"`
	IsOrphan   bool             `json:"is_orphan"`
	References []PhotoReference `json:"references"`
}

type PhotoStats struct {
	TotalFiles  int   `json:"total_files"`
	TotalSize   int64 `json:"total_size"`
	UsedFiles   int   `json:"used_files"`
	UsedSize    int64 `json:"used_size"`
	OrphanFiles int   `json:"orphan_files"`
	OrphanSize  int64 `json:"orphan_size"`
}

type diskFileInfo struct {
	name    string
	size    int64
	modTime time.Time
}

func (h *UploadHandler) getPhysicalFiles(ctx context.Context) map[string]diskFileInfo {
	result := make(map[string]diskFileInfo)

	dirs := []string{getUploadDir(), "uploads"}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(e.Name()))
			if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			result[e.Name()] = diskFileInfo{
				name:    e.Name(),
				size:    info.Size(),
				modTime: info.ModTime(),
			}
		}
	}

	if !h.useDisk && h.minio != nil {
		objCh := h.minio.ListObjects(ctx, h.bucket, minio.ListObjectsOptions{Recursive: true})
		for obj := range objCh {
			if obj.Err != nil {
				continue
			}
			name := filepath.Base(obj.Key)
			if name == "" || name == "." {
				continue
			}
			result[name] = diskFileInfo{
				name:    name,
				size:    obj.Size,
				modTime: obj.LastModified,
			}
		}
	}

	return result
}

func (h *UploadHandler) getDBReferences(ctx context.Context) map[string][]PhotoReference {
	refMap := make(map[string][]PhotoReference)

	// 1. inventory_items
	rows1, err1 := h.pool.Query(ctx, `
		SELECT id::text, name, COALESCE(sku, ''), photo_url
		FROM inventory_items
		WHERE photo_url IS NOT NULL AND photo_url != ''
	`)
	if err1 == nil {
		defer rows1.Close()
		for rows1.Next() {
			var id, name, sku, photoURL string
			if err := rows1.Scan(&id, &name, &sku, &photoURL); err == nil {
				base := filepath.Base(photoURL)
				if base != "" && base != "." {
					refMap[base] = append(refMap[base], PhotoReference{
						Type:  "item",
						ID:    id,
						Title: name,
						Sub:   "Barang · SKU: " + sku,
					})
				}
			}
		}
	}

	// 2. stock_transactions
	rows2, err2 := h.pool.Query(ctx, `
		SELECT t.id::text, COALESCE(i.name, 'Barang'), COALESCE(t.type, 'Mutasi'), t.photo_url
		FROM stock_transactions t
		LEFT JOIN inventory_items i ON t.item_id = i.id
		WHERE t.photo_url IS NOT NULL AND t.photo_url != ''
	`)
	if err2 == nil {
		defer rows2.Close()
		for rows2.Next() {
			var id, itemName, txType, photoURL string
			if err := rows2.Scan(&id, &itemName, &txType, &photoURL); err == nil {
				base := filepath.Base(photoURL)
				if base != "" && base != "." {
					refMap[base] = append(refMap[base], PhotoReference{
						Type:  "transaction",
						ID:    id,
						Title: itemName,
						Sub:   "Riwayat Stok: " + txType,
					})
				}
			}
		}
	}

	// 3. item_maintenances
	rows3, err3 := h.pool.Query(ctx, `
		SELECT m.id::text, COALESCE(i.name, 'Aset'), m.service_type, m.photo_url
		FROM item_maintenances m
		LEFT JOIN inventory_items i ON m.item_id = i.id
		WHERE m.photo_url IS NOT NULL AND m.photo_url != ''
	`)
	if err3 == nil {
		defer rows3.Close()
		for rows3.Next() {
			var id, itemName, sType, photoURL string
			if err := rows3.Scan(&id, &itemName, &sType, &photoURL); err == nil {
				base := filepath.Base(photoURL)
				if base != "" && base != "." {
					refMap[base] = append(refMap[base], PhotoReference{
						Type:  "maintenance",
						ID:    id,
						Title: itemName,
						Sub:   "Pemeliharaan: " + sType,
					})
				}
			}
		}
	}

	return refMap
}

func (h *UploadHandler) ListPhotos(c *gin.Context) {
	ctx := c.Request.Context()
	files := h.getPhysicalFiles(ctx)
	refMap := h.getDBReferences(ctx)

	var allPhotos []PhotoItem
	var stats PhotoStats
	stats.TotalFiles = len(files)

	for _, f := range files {
		stats.TotalSize += f.size
		refs := refMap[f.name]
		isOrphan := len(refs) == 0

		if isOrphan {
			stats.OrphanFiles++
			stats.OrphanSize += f.size
		} else {
			stats.UsedFiles++
			stats.UsedSize += f.size
		}

		allPhotos = append(allPhotos, PhotoItem{
			Filename:   f.name,
			URL:        "/api/uploads/" + f.name,
			Size:       f.size,
			ModTime:    f.modTime,
			IsOrphan:   isOrphan,
			References: refs,
		})
	}

	sort.Slice(allPhotos, func(i, j int) bool {
		return allPhotos[i].ModTime.After(allPhotos[j].ModTime)
	})

	q := strings.ToLower(strings.TrimSpace(c.Query("q")))
	status := c.DefaultQuery("status", "all")

	var filtered []PhotoItem
	for _, p := range allPhotos {
		if status == "orphan" && !p.IsOrphan {
			continue
		}
		if status == "used" && p.IsOrphan {
			continue
		}
		if q != "" {
			match := strings.Contains(strings.ToLower(p.Filename), q)
			if !match {
				for _, r := range p.References {
					if strings.Contains(strings.ToLower(r.Title), q) || strings.Contains(strings.ToLower(r.Sub), q) {
						match = true
						break
					}
				}
			}
			if !match {
				continue
			}
		}
		filtered = append(filtered, p)
	}

	total := len(filtered)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "24"))
	if perPage < 1 {
		perPage = 24
	}

	start := (page - 1) * perPage
	end := start + perPage
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paged := filtered[start:end]
	if paged == nil {
		paged = []PhotoItem{}
	}

	c.JSON(http.StatusOK, gin.H{
		"stats":    stats,
		"data":     paged,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

func (h *UploadHandler) DeletePhoto(c *gin.Context) {
	name := filepath.Base(c.Param("filename"))
	if name == "" || name == "." || name == "/" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nama file tidak valid"})
		return
	}

	dir := getUploadDir()
	_ = os.Remove(filepath.Join(dir, name))
	_ = os.Remove(filepath.Join("uploads", name))

	if !h.useDisk && h.minio != nil {
		_ = h.minio.RemoveObject(c.Request.Context(), h.bucket, name, minio.RemoveObjectOptions{})
	}

	slog.Info("photo explicitly deleted by user", "filename", name)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Foto berhasil dihapus dari penyimpanan"})
}

func (h *UploadHandler) CleanOrphaned(c *gin.Context) {
	ctx := c.Request.Context()
	files := h.getPhysicalFiles(ctx)
	refMap := h.getDBReferences(ctx)

	var deletedCount int
	var freedBytes int64

	dir := getUploadDir()
	for _, f := range files {
		if len(refMap[f.name]) == 0 {
			_ = os.Remove(filepath.Join(dir, f.name))
			_ = os.Remove(filepath.Join("uploads", f.name))
			if !h.useDisk && h.minio != nil {
				_ = h.minio.RemoveObject(ctx, h.bucket, f.name, minio.RemoveObjectOptions{})
			}
			deletedCount++
			freedBytes += f.size
		}
	}

	slog.Info("cleaned orphaned photos", "count", deletedCount, "freed_bytes", freedBytes)
	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"deleted_count": deletedCount,
		"freed_bytes":   freedBytes,
		"message":       fmt.Sprintf("%d foto tidak terpakai (%s) berhasil dibersihkan", deletedCount, formatBytes(freedBytes)),
	})
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func ProcessPhotoBytes(src []byte, geo GeoTag, petugas string, clientStamped bool) ([]byte, error) {
	img, err := imaging.Decode(bytes.NewReader(src))
	if err != nil {
		return src, nil
	}
	if b := img.Bounds(); b.Dx() > 1920 || b.Dy() > 1920 {
		img = imaging.Resize(img, 1920, 0, imaging.Lanczos)
	}
	if geo.Valid && !clientStamped {
		img = stampGeoCard(imaging.Clone(img), geo, petugas)
	}
	var out bytes.Buffer
	if err := imaging.Encode(&out, img, imaging.JPEG, imaging.JPEGQuality(88)); err != nil {
		return src, nil
	}
	return out.Bytes(), nil
}

func stampGeoCard(base *image.NRGBA, geo GeoTag, petugas string) *image.NRGBA {
	b := base.Bounds()
	loc := time.FixedZone("WIB", 7*3600)
	at := geo.At
	if at.IsZero() {
		at = time.Now()
	}
	line1 := fmt.Sprintf("GPS %.6f, %.6f (±%.0fm)", geo.Lat, geo.Lng, geo.Acc)
	line2 := at.In(loc).Format("02-01-2006 15:04 WIB")
	line3 := ""
	if geo.Name != "" && petugas != "" {
		line3 = geo.Name + " — " + petugas
	} else if geo.Name != "" {
		line3 = geo.Name
	} else if petugas != "" {
		line3 = "Petugas: " + petugas
	}
	lines := []string{line1, line2}
	if line3 != "" {
		lines = append(lines, line3)
	}
	face := basicfont.Face7x13
	lh := face.Height + 5
	padX, padY := 14, 10
	cardW := 0
	for _, l := range lines {
		if w := len(l) * face.Width; w > cardW {
			cardW = w
		}
	}
	cardW += padX*2 + 16
	cardH := len(lines)*lh - 5 + padY*2

	// Render card to separate image with alpha
	card := image.NewNRGBA(image.Rect(0, 0, cardW, cardH))
	bg := color.NRGBA{0, 0, 0, 185}
	accent := color.NRGBA{245, 165, 36, 255} // Amber accent stripe

	for y := 0; y < cardH; y++ {
		for x := 0; x < cardW; x++ {
			if x < 4 {
				card.Set(x, y, accent)
			} else {
				card.Set(x, y, bg)
			}
		}
	}

	d := &font.Drawer{Dst: card, Src: image.NewUniform(color.RGBA{255, 255, 255, 255}), Face: face}
	tx := padX + 6
	ty := padY + face.Ascent
	for _, l := range lines {
		d.Dot = fixed.P(tx, ty)
		d.DrawString(l)
		ty += lh
	}

	scale := b.Dx() / 320
	if scale < 2 {
		scale = 2
	} else if scale > 4 {
		scale = 4
	}

	scaledCard := card
	if scale > 1 {
		scaledCard = imaging.Resize(card, cardW*scale, cardH*scale, imaging.NearestNeighbor)
	}

	sw := scaledCard.Bounds().Dx()
	sh := scaledCard.Bounds().Dy()
	x0 := b.Min.X + 16*scale
	y0 := b.Max.Y - sh - 16*scale
	if x0+sw > b.Max.X {
		x0 = b.Min.X
	}
	if y0 < b.Min.Y {
		y0 = b.Min.Y
	}

	return imaging.Overlay(base, scaledCard, image.Pt(x0, y0), 1.0)
}