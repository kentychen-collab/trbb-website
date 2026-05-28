package handlers

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"trbb/internal/config"
	"trbb/internal/services"
)

type TrainingHandler struct {
	trainingSvc *services.TrainingService
	cfg         *config.Config
}

func NewTrainingHandler(trainingSvc *services.TrainingService, cfg *config.Config) *TrainingHandler {
	return &TrainingHandler{trainingSvc: trainingSvc, cfg: cfg}
}

// GET /v1/api/training  (public feed — only public logs)
func (h *TrainingHandler) ListPublic(c *gin.Context) {
	var in services.ListTrainingInput
	_ = c.ShouldBindQuery(&in)
	in.Public = true
	result, err := h.trainingSvc.List(c.Request.Context(), in, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢失敗"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GET /v1/api/training/:id  (by numeric ID)
func (h *TrainingHandler) GetLog(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }

	viewerID := getOptionalUserID(c)
	log, err := h.trainingSvc.GetByID(c.Request.Context(), id, viewerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "日記不存在或無權限查看"})
		return
	}
	c.JSON(http.StatusOK, log)
}

// GET /v1/api/training/share/:uuid  (public share link)
func (h *TrainingHandler) GetByUUID(c *gin.Context) {
	uid := c.Param("uuid")
	viewerID := getOptionalUserID(c)
	log, err := h.trainingSvc.GetByUUID(c.Request.Context(), uid, viewerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "日記不存在或未公開"})
		return
	}
	c.JSON(http.StatusOK, log)
}

// GET /v1/api/me/training  (my logs)
func (h *TrainingHandler) ListMy(c *gin.Context) {
	userID := mustUserID(c)
	var in services.ListTrainingInput
	_ = c.ShouldBindQuery(&in)
	in.UserID = &userID
	result, err := h.trainingSvc.List(c.Request.Context(), in, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢失敗"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// POST /v1/api/training  (manual entry)
func (h *TrainingHandler) Create(c *gin.Context) {
	var in services.TrainingInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log, err := h.trainingSvc.Create(c.Request.Context(), mustUserID(c), in)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "建立失敗"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "訓練日記已建立", "log": log})
}

// PUT /v1/api/training/:id
func (h *TrainingHandler) Update(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }

	var in services.TrainingInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log, err := h.trainingSvc.Update(c.Request.Context(), id, mustUserID(c), in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失敗"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已更新", "log": log})
}

// DELETE /v1/api/training/:id
func (h *TrainingHandler) Delete(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }

	if err := h.trainingSvc.Delete(c.Request.Context(), id, mustUserID(c)); err != nil {
		if errors.Is(err, services.ErrTrainingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "日記不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "刪除失敗"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已刪除"})
}

// POST /v1/api/training/upload/gpx?log_id=0  (upload GPX, optionally attach to existing log)
func (h *TrainingHandler) UploadGPX(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "請選擇 GPX 檔案"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "讀取檔案失敗"})
		return
	}

	logIDStr := c.DefaultQuery("log_id", "0")
	logID, _ := strconv.ParseUint(logIDStr, 10, 64)

	log, err := h.trainingSvc.UploadGPX(c.Request.Context(), mustUserID(c), logID, header.Filename, data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "GPX 上傳成功", "log": log})
}

// POST /v1/api/training/upload/fit?log_id=0
func (h *TrainingHandler) UploadFIT(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "請選擇 FIT 檔案"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "讀取檔案失敗"})
		return
	}

	logIDStr := c.DefaultQuery("log_id", "0")
	logID, _ := strconv.ParseUint(logIDStr, 10, 64)

	path, err := h.trainingSvc.UploadFIT(c.Request.Context(), mustUserID(c), logID, header.Filename, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "上傳失敗"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "FIT 檔案上傳成功", "path": path})
}

// ── Garmin OAuth 1.0a ─────────────────────────────────────────

// GET /v1/api/me/garmin/status
func (h *TrainingHandler) GarminStatus(c *gin.Context) {
	token, err := h.trainingSvc.GetGarminToken(c.Request.Context(), mustUserID(c))
	if err != nil || token == nil {
		cfgGarmin := h.cfg.Third.Garmin
		apiConfigured := cfgGarmin.ClientID != "" && cfgGarmin.ClientID != "CHANGEME"
		c.JSON(http.StatusOK, gin.H{"connected": false, "api_configured": apiConfigured})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"connected":      true,
		"garmin_user_id": token.GarminUserID,
		"last_sync_at":   token.LastSyncAt,
	})
}

// GET /v1/api/me/garmin/connect  → redirect to Garmin OAuth
func (h *TrainingHandler) GarminConnect(c *gin.Context) {
	cfg := h.cfg.Third.Garmin
	if cfg.ClientID == "" || cfg.ClientID == "CHANGEME" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Garmin API 尚未設定，請聯繫管理員",
		})
		return
	}
	// Garmin uses OAuth 1.0a — need request token first
	// TODO: implement OAuth1 request token flow when credentials are available
	// Step 1: POST to GARMIN_TOKEN_URL for request token
	// Step 2: redirect to GARMIN_AUTH_URL?oauth_token=...
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"error": "Garmin 串接框架已就緒，等待 API 憑證",
		"auth_url":    cfg.AuthURL,
		"callback_url": cfg.RedirectURI,
	})
}

// GET /v1/third/garmin/callback  (Garmin redirects here after user authorizes)
func (h *TrainingHandler) GarminCallback(c *gin.Context) {
	// oauth_token and oauth_verifier from Garmin
	oauthToken    := c.Query("oauth_token")
	oauthVerifier := c.Query("oauth_verifier")

	if oauthToken == "" || oauthVerifier == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing oauth params"})
		return
	}

	// TODO: exchange for access token using Garmin OAuth1
	// POST to GARMIN_TOKEN_URL with oauth_token + oauth_verifier
	// Save access_token + token_secret to garmin_tokens table
	// Redirect to /me/garmin with success message

	c.JSON(http.StatusOK, gin.H{
		"msg":            "Garmin callback received (framework)",
		"oauth_token":    oauthToken,
		"oauth_verifier": oauthVerifier,
		"next_step":      "implement OAuth1 token exchange when credentials ready",
	})
}

// DELETE /v1/api/me/garmin/disconnect
func (h *TrainingHandler) GarminDisconnect(c *gin.Context) {
	if err := h.trainingSvc.DeleteGarminToken(c.Request.Context(), mustUserID(c)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解除連結失敗"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已解除 Garmin 連結"})
}

// POST /v1/api/training/garmin/sync
func (h *TrainingHandler) GarminSync(c *gin.Context) {
	count, err := h.trainingSvc.SyncGarminActivities(c.Request.Context(), mustUserID(c))
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "同步完成", "synced": count})
}

// getOptionalUserID extracts user ID from JWT if present, returns 0 if not logged in
func getOptionalUserID(c *gin.Context) uint64 {
	raw, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	switch v := raw.(type) {
	case float64:
		return uint64(v)
	case string:
		id, _ := strconv.ParseUint(v, 10, 64)
		return id
	}
	return 0
}
