package handlers

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"inventariskantor/internal/drive"
)

var errEmptyPhoto = errors.New("empty photo")

// storePhoto uploads to Google Drive when credentials exist; otherwise saves to
// web/uploads and returns a local URL. Drive folder: GOOGLE_DRIVE_PHOTO_FOLDER_ID.
func (h *Handlers) storePhoto(r *http.Request, data []byte, filename string) (string, error) {
	if data == nil {
		return "", errEmptyPhoto
	}
	saPath := os.Getenv("GOOGLE_SA_JSON")
	folder := os.Getenv("GOOGLE_DRIVE_PHOTO_FOLDER_ID")
	if saPath != "" && folder != "" {
		if link, err := drive.UploadPhoto(saPath, folder, filename, data, "image/jpeg"); err == nil {
			return link, nil
		}
		// fallthrough to local on any drive failure
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		ext = ".jpg"
	}
	name := uuid.NewString() + ext
	dir := "web/uploads"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
		return "", err
	}
	return "/uploads/" + name, nil
}
