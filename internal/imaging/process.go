package imaging

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	maxSide   = 1600
	jpegQ     = 85
	padX      = 12
	padY      = 10
	lineGap   = 6
	cornerR   = 14
	cardAlpha = 150
)

// GeoTag is the data stamped onto the photo card.
type GeoTag struct {
	Lat, Lng, Acc float64
	Valid         bool
	At            time.Time
	Name          string
}

// ProcessPhotoBytes compresses and stamps a geotag card, returns JPEG bytes.
func ProcessPhotoBytes(src []byte, geo GeoTag, petugas string) ([]byte, error) {
	img, err := imaging.Decode(bytes.NewReader(src))
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	// 1. kompres: resize sisi terpanjang → maxSide
	if b := img.Bounds(); b.Dx() > maxSide || b.Dy() > maxSide {
		img = imaging.Resize(img, maxSide, 0, imaging.Lanczos)
	}

	if geo.Valid {
		img = stampCard(imaging.Clone(img), geo, petugas)
	}

	var out bytes.Buffer
	if err := imaging.Encode(&out, img, imaging.JPEG, imaging.JPEGQuality(jpegQ)); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// stampCard draws the rounded geotag card at bottom-left of the image.
func stampCard(base *image.NRGBA, geo GeoTag, petugas string) *image.NRGBA {
	b := base.Bounds()
	loc := time.FixedZone("WIB", 7*3600)
	at := geo.At
	if at.IsZero() {
		at = time.Now()
	}

	line1 := fmt.Sprintf("📍 %.6f, %.6f (±%.0fm)", geo.Lat, geo.Lng, geo.Acc)
	line2 := at.In(loc).Format("02-01-2006 15:04 WIB")
	line3 := strings.TrimSpace(strings.Join(nonEmpty(geo.Name, petugas), " — "))

	lines := []string{line1, line2}
	if line3 != "" {
		lines = append(lines, line3)
	}

	face := basicfont.Face7x13
	lh := face.Height + lineGap
	textW := 0
	for _, l := range lines {
		if w := len(l) * face.Width; w > textW {
			textW = w
		}
	}
	cardW := textW + padX*2 + 22 // +22 = pin icon
	cardH := len(lines)*lh - lineGap + padY*2

	x0, y0 := b.Min.X+12, b.Max.Y-cardH-12
	if x0+cardW > b.Max.X { // foto sempit: geser ke kiri
		x0 = b.Min.X
	}

	// rounded card background
	drawRoundedRect(base, image.Rect(x0, y0, x0+cardW, y0+cardH), cornerR,
		color.RGBA{0, 0, 0, cardAlpha})

	// GPS pin icon (ring + triangle tail)
	drawPin(base, x0+padX+8, y0+cardH/2, 8, color.RGBA{255, 255, 255, 255})

	// text lines (basicfont bitmap — deterministic, no font files needed)
	tx := x0 + padX + 22
	ty := y0 + padY + face.Ascent
	white := color.RGBA{255, 255, 255, 255}
	d := &font.Drawer{Dst: base, Src: image.NewUniform(white), Face: face}
	for _, l := range lines {
		d.Dot = fixed.P(tx, ty)
		d.DrawString(l)
		ty += lh
	}
	return base
}

func nonEmpty(ss ...string) []string {
	out := make([]string, 0, len(ss))
	for _, s := range ss {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// drawRoundedRect fills a rounded rectangle with radius r and color c.
func drawRoundedRect(dst *image.NRGBA, rect image.Rectangle, r int, c color.RGBA) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			if insideRounded(x, y, rect, r) {
				dst.Set(x, y, c)
			}
		}
	}
}

func insideRounded(x, y int, rect image.Rectangle, r int) bool {
	if x >= rect.Min.X+r && x < rect.Max.X-r {
		return true
	}
	if y >= rect.Min.Y+r && y < rect.Max.Y-r {
		return true
	}
	for _, p := range [][2]int{
		{rect.Min.X + r, rect.Min.Y + r}, {rect.Max.X - r, rect.Min.Y + r},
		{rect.Min.X + r, rect.Max.Y - r}, {rect.Max.X - r, rect.Max.Y - r},
	} {
		if math.Hypot(float64(x-p[0]), float64(y-p[1])) <= float64(r) {
			return true
		}
	}
	return false
}

func drawPin(dst *image.NRGBA, cx, cy, rad int, c color.RGBA) {
	for y := -rad; y <= rad; y++ {
		for x := -rad; x <= rad; x++ {
			d := math.Hypot(float64(x), float64(y))
			if d <= float64(rad) && d >= float64(rad-3) {
				dst.Set(cx+x, cy+y, c)
			}
		}
	}
	for t := 0; t <= rad; t++ {
		for dx := -t/2; dx <= t/2; dx++ {
			dst.Set(cx+dx, cy+rad-1+t, c)
		}
	}
}
