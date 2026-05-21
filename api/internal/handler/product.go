package handler

import (
	"net/http"
	"strconv"
	"sports-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productSvc *service.ProductService
}

func NewProductHandler(productSvc *service.ProductService) *ProductHandler {
	return &ProductHandler{productSvc: productSvc}
}

func (h *ProductHandler) List(c *gin.Context) {
	page, _  := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 { page = 1 }

	var categoryID *int
	if v := c.Query("category_id"); v != "" {
		id, err := strconv.Atoi(v)
		if err == nil { categoryID = &id }
	}
	featured := c.Query("featured") == "1"

	list, total, err := h.productSvc.List(page, limit, categoryID, featured)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total})
}

func (h *ProductHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	p, err := h.productSvc.GetBySlug(slug)
	if err != nil || p == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "商品不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": p})
}

func (h *ProductHandler) ListCategories(c *gin.Context) {
	list, err := h.productSvc.ListCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}
