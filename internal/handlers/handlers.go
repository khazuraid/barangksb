package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"inventariskantor/internal/auth"
	"inventariskantor/internal/imaging"
	"inventariskantor/internal/service"
)

type Handlers struct {
	pool            *pgxpool.Pool
	svc             *service.Service
	sess            *scs.SessionManager
	broker          *Broker
	driveLastRun    string
	driveLastResult string
}

func New(pool *pgxpool.Pool, sess *scs.SessionManager) *Handlers {
	return &Handlers{pool: pool, svc: service.NewService(pool), sess: sess, broker: NewBroker()}
}

func parseUUID(s string) (uuid.UUID, error) { return uuid.Parse(s) }

// processUpload reads the "photo" file from a multipart form, compresses it,
// stamps the geotag card, uploads to Drive (or falls back to local disk), and
// returns the public URL. keepOld is used on update when no new photo is given.
func (h *Handlers) processUpload(r *http.Request, keepOld string) string {
	file, hdr, err := r.FormFile("photo")
	if err != nil || file == nil || hdr.Size == 0 {
		return keepOld
	}
	defer file.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		return keepOld
	}
	geo := geoFromForm(r.PostForm)
	processed, err := processPhoto(buf.Bytes(), geo, authUserName(r))
	if err != nil {
		return keepOld
	}
	photoURL, err := h.storePhoto(r, processed, hdr.Filename)
	if err != nil {
		return keepOld
	}
	return photoURL
}

// processPhoto compresses + stamps geotag card on raw photo bytes.
func processPhoto(src []byte, geo geoData, petugas string) ([]byte, error) {
	return imaging.ProcessPhotoBytes(src, imaging.GeoTag{
		Lat: geo.lat.Float64, Lng: geo.lng.Float64, Acc: geo.acc.Float64,
		Valid: geo.lat.Valid, At: geo.at.Time, Name: geo.name.String,
	}, petugas)
}

var (
	_ = templ.Component(nil)
	_ = context.Background
	_ = io.EOF
	_ = fmt.Sprintf
	_ = templ.Component(nil)
	_ = pgtype.Text{}
	_ = pgxpool.Pool{}
	_ = chi.URLParam
	_ = auth.User{}
)
