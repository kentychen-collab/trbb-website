package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"trbb/internal/services"
)

type SiteSettingsHandler struct {
	svc *services.SiteSettingsService
}

func NewSiteSettingsHandler(svc *services.SiteSettingsService) *SiteSettingsHandler {
	return &SiteSettingsHandler{svc: svc}
}

// GET /v1/api/settings
func (h *SiteSettingsHandler) GetPublic(c *gin.Context) {
	settings, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取得設定失敗"})
		return
	}
	c.JSON(http.StatusOK, settings)
}
