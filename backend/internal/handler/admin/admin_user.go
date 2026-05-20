package admin

import (
	"net/http"
	"sports-platform/internal/model"
	"sports-platform/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminUserHandler struct {
	adminSvc *service.AdminService
}

func NewAdminUserHandler(adminSvc *service.AdminService) *AdminUserHandler {
	return &AdminUserHandler{adminSvc: adminSvc}
}

func (h *AdminUserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 { page = 1 }
	list, total, err := h.adminSvc.ListAdmins(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total})
}

func (h *AdminUserHandler) Create(c *gin.Context) {
	creatorID := c.MustGet("admin_id").(uint64)
	creatorRole := c.MustGet("admin_role").(int)
	var input service.CreateAdminInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	admin, err := h.adminSvc.CreateAdmin(input, creatorID, creatorRole)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": admin})
}

func (h *AdminUserHandler) GetPermissions(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	perms, err := h.adminSvc.GetPermissions(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": perms})
}

func (h *AdminUserHandler) SetPermissions(c *gin.Context) {
	creatorRole := c.MustGet("admin_role").(int)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var p model.AdminPermissions
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p.AdminID = id
	if err := h.adminSvc.SetPermissions(creatorRole, &p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "permissions updated"})
}

func (h *AdminUserHandler) SetStatus(c *gin.Context) {
	creatorRole := c.MustGet("admin_role").(int)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct{ Status int `json:"status"` }
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.adminSvc.SetStatus(creatorRole, id, body.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}
