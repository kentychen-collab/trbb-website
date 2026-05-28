package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"trbb/internal/services"
)

type AdminTrainingHandler struct {
	trainingSvc *services.TrainingService
}

func NewAdminTrainingHandler(trainingSvc *services.TrainingService) *AdminTrainingHandler {
	return &AdminTrainingHandler{trainingSvc: trainingSvc}
}

// GET /v1/admin/training
func (h *AdminTrainingHandler) ListTraining(c *gin.Context) {
	var in services.AdminListTrainingInput
	_ = c.ShouldBindQuery(&in)

	// optional filter by user_id from query param
	if uid := c.Query("user_id"); uid != "" {
		if id, err := strconv.ParseUint(uid, 10, 64); err == nil {
			in.UserID = &id
		}
	}

	result, err := h.trainingSvc.AdminList(c.Request.Context(), in)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢失敗"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GET /v1/admin/training/stats
func (h *AdminTrainingHandler) Stats(c *gin.Context) {
	stats, err := h.trainingSvc.AdminStats(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "統計查詢失敗"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// GET /v1/admin/training/:id
func (h *AdminTrainingHandler) GetTraining(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	// viewerID=0 → admin can see everything
	log, err := h.trainingSvc.GetByID(c.Request.Context(), id, 0)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "日記不存在"})
		return
	}
	c.JSON(http.StatusOK, log)
}
