package handler

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type UploadHandler struct {
	pool *pgxpool.Pool
}

func NewUploadHandler(pool *pgxpool.Pool) *UploadHandler {
	return &UploadHandler{pool: pool}
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
	dir := "uploads"
	os.MkdirAll(dir, 0o755)
	path := filepath.Join(dir, name)
	os.WriteFile(path, processed, 0o644)

	c.JSON(http.StatusOK, gin.H{"url": "/uploads/" + name})
}

func ProcessPhotoBytes(src []byte, geo GeoTag, petugas string) ([]byte, error) {
	img, err := imaging.Decode(bytes.NewReader(src))
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if b := img.Bounds(); b.Dx() > 1600 || b.Dy() > 1600 {
		img = imaging.Resize(img, 1600, 0, imaging.Lanczos)
	}
	if geo.Valid {
		img = stampGeoCard(imaging.Clone(img), geo, petugas)
	}
	var out bytes.Buffer
	if err := imaging.Encode(&out, img, imaging.JPEG, imaging.JPEGQuality(85)); err != nil {
		return nil, err
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
	}
	lines := []string{line1, line2}
	if line3 != "" {
		lines = append(lines, line3)
	}
	face := basicfont.Face7x13
	lh := face.Height + 6
	padX, padY := 12, 10
	cardW := 0
	for _, l := range lines {
		if w := len(l) * face.Width; w > cardW {
			cardW = w
		}
	}
	cardW += padX*2 + 22
	cardH := len(lines)*lh - 6 + padY*2
	x0 := b.Min.X + 12
	y0 := b.Max.Y - cardH - 12
	if x0+cardW > b.Max.X {
		x0 = b.Min.X
	}
	// card bg
	overlay := color.RGBA{0, 0, 0, 150}
	for y := y0; y < y0+cardH; y++ {
		for x := x0; x < x0+cardW; x++ {
			if x >= 0 && x < b.Max.X && y >= 0 && y < b.Max.Y {
				o := base.At(x, y)
				r, g, bb, a := o.RGBA()
				blend := func(c uint32, oc uint8) uint8 {
					return uint8((c*65535 + uint32(oc)*uint32(overlay.A)*255) / (65535 + uint32(overlay.A)*255) >> 8)
				}
				base.Set(x, y, color.RGBA{
					blend(r, overlay.R), blend(g, overlay.G), blend(bb, overlay.B), uint8(a >> 8),
				})
			}
		}
	}
	// text
	d := &font.Drawer{Dst: base, Src: image.NewUniform(color.RGBA{255, 255, 255, 255}), Face: face}
	tx := x0 + padX + 22
	ty := y0 + padY + face.Ascent
	for _, l := range lines {
		d.Dot = fixed.P(tx, ty)
		d.DrawString(l)
		ty += lh
	}
	return base
}
