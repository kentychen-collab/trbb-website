package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"trbb/internal/services"
)

type StravaWebhookHandler struct {
	trainingSvc *services.TrainingService
}

func NewStravaWebhookHandler(svc *services.TrainingService) *StravaWebhookHandler {
	return &StravaWebhookHandler{trainingSvc: svc}
}

// GET /v1/third/strava/webhook  — Strava 訂閱驗證
func (h *StravaWebhookHandler) Verify(c *gin.Context) {
	mode      := c.Query("hub.mode")
	token     := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	verifyToken := os.Getenv("STRAVA_WEBHOOK_VERIFY_TOKEN")
	if verifyToken == "" {
		verifyToken = "trbb_strava_webhook_2024"
	}

	if mode == "subscribe" && token == verifyToken {
		c.JSON(http.StatusOK, gin.H{"hub.challenge": challenge})
		return
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "invalid verify token"})
}

// POST /v1/third/strava/webhook  — Strava 推送活動事件
func (h *StravaWebhookHandler) Event(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)

	var event struct {
		ObjectType     string `json:"object_type"`  // "activity" | "athlete"
		ObjectID       int64  `json:"object_id"`    // activity ID
		AspectType     string `json:"aspect_type"`  // "create" | "update" | "delete"
		OwnerID        int64  `json:"owner_id"`     // Strava athlete ID
		SubscriptionID int    `json:"subscription_id"`
	}

	if err := json.Unmarshal(body, &event); err != nil {
		c.JSON(http.StatusOK, gin.H{}) // 200 必須，否則 Strava 重試
		return
	}

	// 只處理活動新增事件
	if event.ObjectType != "activity" || event.AspectType != "create" {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	// 根據 Strava athlete ID 找到對應用戶
	go func() {
		ctx := c.Request.Context()
		token, err := h.trainingSvc.GetStravaTokenByAthleteID(ctx, event.OwnerID)
		if err != nil || token == nil {
			return
		}

		// 同步這一筆活動
		count, err := h.trainingSvc.SyncStravaActivityByID(ctx, token.UserID, event.ObjectID)
		if err != nil || count == 0 {
			return
		}

		// 如果用戶設定了 sync_public，批次 PATCH
		if token.SyncPublic {
			h.trainingSvc.SetLatestStravaActivityPublic(ctx, token.UserID, event.ObjectID)
		}

		fmt.Printf("[Strava Webhook] synced activity %d for user %d\n",
			event.ObjectID, token.UserID)
	}()

	c.JSON(http.StatusOK, gin.H{})
}
