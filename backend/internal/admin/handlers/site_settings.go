package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"trbb/internal/services"
	"trbb/pkg/storage"
)

// ── Public handler ────────────────────────────────────────────

type SiteSettingsHandler struct {
	svc *services.SiteSettingsService
}

func NewSiteSettingsHandler(svc *services.SiteSettingsService) *SiteSettingsHandler {
	return &SiteSettingsHandler{svc: svc}
}

// GET /v1/api/settings  — 前台取得所有設定
func (h *SiteSettingsHandler) GetPublic(c *gin.Context) {
	settings, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取得設定失敗"})
		return
	}
	c.JSON(http.StatusOK, settings)
}

// ── Admin handler ─────────────────────────────────────────────

type AdminSiteSettingsHandler struct {
	svc   *services.SiteSettingsService
	store *storage.Storage
}

func NewAdminSiteSettingsHandler(svc *services.SiteSettingsService, store *storage.Storage) *AdminSiteSettingsHandler {
	return &AdminSiteSettingsHandler{svc: svc, store: store}
}

// GET /v1/admin/settings
func (h *AdminSiteSettingsHandler) GetAll(c *gin.Context) {
	grouped, err := h.svc.GetGrouped(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取得設定失敗"})
		return
	}
	c.JSON(http.StatusOK, grouped)
}

// POST /v1/admin/settings  — 批次更新
func (h *AdminSiteSettingsHandler) BatchUpdate(c *gin.Context) {
	var body map[string]string
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updaterID := mustAdminUserID(c)
	if err := h.svc.BatchSet(c.Request.Context(), body, updaterID); err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新設定失敗"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "設定已更新"})
}

// PUT /v1/admin/settings/:key  — 更新單一設定
func (h *AdminSiteSettingsHandler) SetOne(c *gin.Context) {
	key := c.Param("key")
	var body struct {
		Value string `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updaterID := mustAdminUserID(c)
	if err := h.svc.Set(c.Request.Context(), key, body.Value, updaterID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失敗"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已更新"})
}

// POST /v1/admin/settings/upload/:purpose
// purpose: logo | banner | banner2
func (h *AdminSiteSettingsHandler) UploadSettingImage(c *gin.Context) {
	purpose := c.Param("purpose")
	keyMap := map[string]string{
		"logo":    "logo_image",
		"banner":  "banner_image",
		"banner2": "banner_image_2",
		"icon":    "site_icon",
		"icon_lg": "site_icon_lg",
	}
	settingKey, ok := keyMap[purpose]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支援的用途"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "請選擇要上傳的圖片"})
		return
	}
	defer file.Close()

	if header.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "圖片大小不能超過 5MB"})
		return
	}

	objectPath, err := h.store.UploadImage(
		c.Request.Context(),
		"general", header.Filename, "image/jpeg",
		file, header.Size,
	)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "上傳失敗"})
		return
	}

	fullURL := h.store.ImageURL(objectPath)

	// 自動更新對應設定
	updaterID := mustAdminUserID(c)
	_ = h.svc.Set(c.Request.Context(), settingKey, fullURL, updaterID)

	c.JSON(http.StatusOK, gin.H{
		"path":        objectPath,
		"url":         fullURL,
		"setting_key": settingKey,
	})
}

func mustAdminUID(c *gin.Context) uint64 {
	raw, _ := c.Get("user_id")
	switch v := raw.(type) {
	case float64:
		return uint64(v)
	case string:
		id, _ := strconv.ParseUint(v, 10, 64)
		return id
	}
	return 0
}
