package admin

import (
	"net/http"
	"strconv"
	"sports-platform/internal/repository"

	"github.com/gin-gonic/gin"
)

type OrderAdminHandler struct {
	orderRepo *repository.OrderRepo
}

func NewOrderAdminHandler(orderRepo *repository.OrderRepo) *OrderAdminHandler {
	return &OrderAdminHandler{orderRepo: orderRepo}
}

func (h *OrderAdminHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 { page = 1 }
	list, total, err := h.orderRepo.AdminList(limit, (page-1)*limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total})
}

func (h *OrderAdminHandler) UpdateStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.orderRepo.UpdateStatus(id, body.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}
