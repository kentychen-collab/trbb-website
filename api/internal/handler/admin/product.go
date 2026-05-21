package admin

import (
	"net/http"
	"strconv"
	"sports-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type ProductAdminHandler struct {
	productSvc *service.ProductService
}

func NewProductAdminHandler(productSvc *service.ProductService) *ProductAdminHandler {
	return &ProductAdminHandler{productSvc: productSvc}
}

func (h *ProductAdminHandler) List(c *gin.Context) {
	page, _  := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 { page = 1 }

	var categoryID *int
	if v := c.Query("category_id"); v != "" {
		id, err := strconv.Atoi(v)
		if err == nil { categoryID = &id }
	}
	var status *int
	if v := c.Query("status"); v != "" {
		s, err := strconv.Atoi(v)
		if err == nil { status = &s }
	}

	list, total, err := h.productSvc.AdminList(page, limit, categoryID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total})
}

func (h *ProductAdminHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	p, err := h.productSvc.GetByID(id)
	if err != nil || p == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "商品不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": p})
}

func (h *ProductAdminHandler) Create(c *gin.Context) {
	var input service.ProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.productSvc.Create(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": p})
}

func (h *ProductAdminHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var input service.ProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.productSvc.Update(id, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": p})
}

func (h *ProductAdminHandler) SetStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct{ Status int `json:"status"` }
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.productSvc.SetStatus(id, body.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "狀態已更新"})
}

func (h *ProductAdminHandler) ListCategories(c *gin.Context) {
	list, err := h.productSvc.ListCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}
