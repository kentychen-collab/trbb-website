package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"trbb/pkg/storage"
)

type UploadHandler struct {
	store *storage.Storage
}

func NewUploadHandler(store *storage.Storage) *UploadHandler {
	return &UploadHandler{store: store}
}

// POST /v1/admin/upload/image?folder=events|products|avatars|general
func (h *UploadHandler) UploadImage(c *gin.Context) {
	folder := c.DefaultQuery("folder", storage.FolderGeneral)

	// 白名單驗證
	allowed := map[string]bool{
		storage.FolderEvents:   true,
		storage.FolderProducts: true,
		storage.FolderAvatars:  true,
		storage.FolderGeneral:  true,
		storage.FolderTraining: true,
	}
	if !allowed[folder] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支援的上傳目錄"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "請選擇要上傳的檔案"})
		return
	}
	defer file.Close()

	// 驗證副檔名
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
	}
	contentType, ok := allowedExts[ext]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只支援 JPG、PNG、GIF、WebP 格式"})
		return
	}

	// 限制大小 10MB
	if header.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "檔案大小不能超過 10MB"})
		return
	}

	// 產生唯一檔名，保留副檔名
	filename := fmt.Sprintf("%d_%s%s",
		time.Now().UnixMilli(),
		sanitizeFilename(strings.TrimSuffix(header.Filename, ext)),
		ext,
	)

	// 上傳到 MinIO images bucket
	objectPath, err := h.store.UploadImage(
		c.Request.Context(),
		folder, filename, contentType,
		file, header.Size,
	)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "上傳失敗，請稍後再試"})
		return
	}

	// 回傳 object path（前端用環境變數拼完整 URL）
	c.JSON(http.StatusOK, gin.H{
		"path":     objectPath,                    // e.g. "events/123_cover.jpg"
		"url":      h.store.ImageURL(objectPath),  // e.g. "https://images.trbbtw.com/images/events/..."
		"filename": filename,
		"size":     header.Size,
	})
}

// sanitizeFilename removes unsafe characters from filename
func sanitizeFilename(name string) string {
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	result := b.String()
	if len(result) > 30 {
		result = result[:30]
	}
	if result == "" {
		result = "image"
	}
	return result
}
