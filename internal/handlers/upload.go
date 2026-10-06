package handlers

import (
	"bytes"
	"io"
	"net/http"
)

// UploadPhoto endpoint (dipakai bot Telegram / external): POST multipart "photo".
func (h *Handlers) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(6 << 20); err != nil {
		http.Error(w, "form terlalu besar (max 6MB)", http.StatusRequestEntityTooLarge)
		return
	}
	file, hdr, err := r.FormFile("photo")
	if err != nil || file == nil || hdr.Size == 0 {
		http.Error(w, "file photo wajib", http.StatusBadRequest)
		return
	}
	defer file.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	geo := geoFromForm(r.PostForm)
	processed, err := processPhoto(buf.Bytes(), geo, authUserName(r))
	if err != nil {
		http.Error(w, "foto gagal diproses: "+err.Error(), http.StatusBadRequest)
		return
	}
	url, err := h.storePhoto(r, processed, hdr.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"url":"` + url + `"}`))
}
