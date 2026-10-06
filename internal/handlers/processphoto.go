package handlers

import (
	"bytes"
	"io"

	"inventariskantor/internal/imaging"
)

// processPhoto compresses + stamps geotag card on raw photo bytes.
func processPhoto(src []byte, geo geoData, petugas string) ([]byte, error) {
	return imaging.ProcessPhotoBytes(src, imaging.GeoTag{
		Lat: geo.lat.Float64, Lng: geo.lng.Float64, Acc: geo.acc.Float64,
		Valid: geo.lat.Valid, At: geo.at.Time, Name: geo.name.String,
	}, petugas)
}

var (
	_ = io.EOF
	_ = bytes.MinRead
)
