package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UploadHandler struct {
	uploadDir  string
	urlPrefix  string
}

func NewUploadHandler(uploadDir, urlPrefix string) *UploadHandler {
	// 確保目錄存在
	os.MkdirAll(uploadDir, 0755)
	os.MkdirAll(filepath.Join(uploadDir, "products"), 0755)
	os.MkdirAll(filepath.Join(uploadDir, "events"), 0755)
	return &UploadHandler{uploadDir: uploadDir, urlPrefix: urlPrefix}
}

func (h *UploadHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "請選擇要上傳的檔案"})
		return
	}
	defer file.Close()

	// 檢查副檔名
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true}
	if !allowed[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只支援 jpg / png / webp / gif"})
		return
	}

	// 檔案大小限制 5MB
	if header.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "檔案大小不能超過 5MB"})
		return
	}

	// 子目錄（products / events，預設 images）
	subDir := c.DefaultPostForm("type", "images")
	subDir = filepath.Base(subDir) // 防止路徑穿越
	dir := filepath.Join(h.uploadDir, subDir)
	os.MkdirAll(dir, 0755)

	// 產生唯一檔名
	filename := fmt.Sprintf("%s-%s%s",
		time.Now().Format("20060102"),
		uuid.New().String()[:8],
		ext,
	)
	destPath := filepath.Join(dir, filename)

	// 讀取並寫入
	buf := make([]byte, header.Size)
	if _, err := file.Read(buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "讀取檔案失敗"})
		return
	}
	if err := os.WriteFile(destPath, buf, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "儲存檔案失敗"})
		return
	}

	url := fmt.Sprintf("%s/%s/%s", strings.TrimRight(h.urlPrefix, "/"), subDir, filename)
	c.JSON(http.StatusOK, gin.H{
		"url":      url,
		"filename": filename,
	})
}
