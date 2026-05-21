package admin

import (
	"net/http"
	"sports-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthAdminHandler struct {
	adminSvc *service.AdminService
}

func NewAuthAdminHandler(adminSvc *service.AdminService) *AuthAdminHandler {
	return &AuthAdminHandler{adminSvc: adminSvc}
}

func (h *AuthAdminHandler) Login(c *gin.Context) {
	var input service.AdminLoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.adminSvc.Login(input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AuthAdminHandler) Me(c *gin.Context) {
	adminID := c.MustGet("admin_id").(uint64)
	// Return basic info from token context
	c.JSON(http.StatusOK, gin.H{
		"admin_id": adminID,
		"role":     c.MustGet("admin_role"),
	})
}
