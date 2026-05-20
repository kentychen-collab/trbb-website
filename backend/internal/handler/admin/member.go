package admin

import (
	"net/http"
	"strconv"
	"sports-platform/internal/repository"

	"github.com/gin-gonic/gin"
)

type MemberAdminHandler struct {
	memberRepo *repository.MemberRepo
}

func NewMemberAdminHandler(memberRepo *repository.MemberRepo) *MemberAdminHandler {
	return &MemberAdminHandler{memberRepo: memberRepo}
}

func (h *MemberAdminHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 { page = 1 }

	// 可選 status filter: pending/active/all
	var statusPtr *int
	statusStr := c.Query("status")
	if statusStr != "" {
		s, err := strconv.Atoi(statusStr)
		if err == nil {
			statusPtr = &s
		}
	}

	list, total, err := h.memberRepo.AdminList(limit, (page-1)*limit, statusPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total})
}

func (h *MemberAdminHandler) Approve(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.memberRepo.UpdateStatus(id, 1); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "會員已核准"})
}

func (h *MemberAdminHandler) Reject(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.memberRepo.UpdateStatus(id, 2); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "會員已停用"})
}

func (h *MemberAdminHandler) SetStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct{ Status int `json:"status"` }
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.memberRepo.UpdateStatus(id, body.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "狀態已更新"})
}
