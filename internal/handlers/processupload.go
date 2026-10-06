package handlers

import (
	"bytes"
	"io"
	"net/http"
)

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
	petugas := ""
	if u := authUserOf(r); u != nil {
		petugas = u.Name
	}
	processed, err := processPhoto(buf.Bytes(), geo, petugas)
	if err != nil {
		return keepOld
	}
	url, err := h.storePhoto(r, processed, hdr.Filename)
	if err != nil {
		return keepOld
	}
	return url
}
