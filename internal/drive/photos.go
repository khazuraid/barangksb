package drive

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// UploadPhoto uploads a file to a folder using a service-account JSON file.
// Returns the webViewLink.
func UploadPhoto(saJSONPath, folderID, name string, data []byte, mimeType string) (string, error) {
	sa, err := os.ReadFile(saJSONPath)
	if err != nil {
		return "", err
	}
	conf, err := google.JWTConfigFromJSON(sa, drive.DriveFileScope)
	if err != nil {
		return "", err
	}
	srv, err := drive.NewService(context.Background(), option.WithTokenSource(conf.TokenSource(context.Background())))
	if err != nil {
		return "", err
	}
	f := &drive.File{Name: name, Parents: []string{folderID}, MimeType: mimeType}
	res, err := srv.Files.Create(f).Media(bytes.NewReader(data)).Fields("webViewLink").Do()
	if err != nil {
		return "", err
	}
	if res.WebViewLink == "" {
		return fmt.Sprintf("https://drive.google.com/file/d/%s/view", res.Id), nil
	}
	return res.WebViewLink, nil
}
