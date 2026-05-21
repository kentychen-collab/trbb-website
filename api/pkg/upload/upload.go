// Package upload handles file uploads.
// MinIO integration is stubbed out - implement when ready.
package upload

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Uploader struct {
	baseURL string
}

func New(endpoint, accessKey, secretKey, bucket string) (*Uploader, error) {
	return &Uploader{baseURL: fmt.Sprintf("http://%s/%s", endpoint, bucket)}, nil
}

func (u *Uploader) Upload(file multipart.File, header *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowed[ext] {
		return "", fmt.Errorf("unsupported file type: %s", ext)
	}
	objectName := fmt.Sprintf("%s/%s%s", time.Now().Format("2006/01"), uuid.New().String(), ext)
	// TODO: implement actual MinIO upload
	_ = file
	return fmt.Sprintf("%s/%s", u.baseURL, objectName), nil
}
