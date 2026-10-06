package handlers

import (
	"net/http"
	"strings"
)

// parseForm handles both urlencoded posts and multipart uploads (photo).
// Returns 413-style error only when a real multipart body is malformed/too big.
func parseForm(w http.ResponseWriter, r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		if err := r.ParseMultipartForm(6 << 20); err != nil {
			http.Error(w, "form terlalu besar (max 6MB)", http.StatusRequestEntityTooLarge)
			return false
		}
		return true
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form tidak valid", http.StatusBadRequest)
		return false
	}
	return true
}
