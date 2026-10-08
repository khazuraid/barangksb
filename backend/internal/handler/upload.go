package handler

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
	if endpoint != "" && accessKey != "" && secretKey != "" && endpoint != "minio:9000" {
		mc, err := minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: false, // internal docker network
		})
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			// Ensure bucket exists
			exists, _ := mc.BucketExists(ctx, bucket)
			if !exists {
				mc.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
			}
			h.minio = mc
			h.useDisk = false
		}
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
	file, hdr, err := c.Request.FormFile("photo")
	if err != nil || file == nil || hdr.Size == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file photo wajib"})
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

	processed, err := ProcessPhotoBytes(buf.Bytes(), geo, petugas)
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
		_ = os.MkdirAll("uploads", 0o755)
		_ = os.WriteFile(filepath.Join("uploads", name), data, 0o644)
	}
	return "/api/uploads/" + name
}

// GetObject retrieves a file from MinIO (or nil if using disk)
func (h *UploadHandler) GetObject(ctx context.Context, name string) (io.ReadCloser, error) {
	if h.minio != nil && !h.useDisk {
		obj, err := h.minio.GetObject(ctx, h.bucket, name, minio.GetObjectOptions{})
		if err != nil {
			return nil, err
		}
		return obj, nil
	}
	// disk fallback
	dir := getUploadDir()
	f, err := os.Open(filepath.Join(dir, name))
	if err != nil {
		return os.Open(filepath.Join("uploads", name))
	}
	return f, nil
}

// Serve streams uploaded image from MinIO or disk
func (h *UploadHandler) Serve(c *gin.Context) {
	name := filepath.Base(c.Param("name"))
	if name == "" || name == "." {
		c.Status(http.StatusNotFound)
		return
	}
	reader, err := h.GetObject(c.Request.Context(), name)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	defer reader.Close()

	c.Header("Content-Type", "image/jpeg")
	c.Header("Cache-Control", "public, max-age=86400")
	io.Copy(c.Writer, reader)
}

func ProcessPhotoBytes(src []byte, geo GeoTag, petugas string) ([]byte, error) {
	img, err := imaging.Decode(bytes.NewReader(src))
	if err != nil {
		return src, nil
	}
	if b := img.Bounds(); b.Dx() > 1600 || b.Dy() > 1600 {
		img = imaging.Resize(img, 1600, 0, imaging.Lanczos)
	}
	if geo.Valid {
		img = stampGeoCard(imaging.Clone(img), geo, petugas)
	}
	var out bytes.Buffer
	if err := imaging.Encode(&out, img, imaging.JPEG, imaging.JPEGQuality(85)); err != nil {
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

	scale := b.Dx() / 450
	if scale < 1 {
		scale = 1
	} else if scale > 3 {
		scale = 3
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