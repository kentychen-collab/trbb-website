package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"trbb/internal/services"
)

type AdminShopHandler struct {
	shopSvc *services.ShopService
}

func NewAdminShopHandler(shopSvc *services.ShopService) *AdminShopHandler {
	return &AdminShopHandler{shopSvc: shopSvc}
}

// ── Products ──────────────────────────────────────────────────

// GET /v1/admin/products
func (h *AdminShopHandler) ListProducts(c *gin.Context) {
	var in services.ListProductsInput
	_ = c.ShouldBindQuery(&in)
	result, err := h.shopSvc.ListAdmin(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢失敗"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GET /v1/admin/products/:id
func (h *AdminShopHandler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }
	p, err := h.shopSvc.GetProductByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "商品不存在"})
		return
	}
	c.JSON(http.StatusOK, p)
}

// POST /v1/admin/products
func (h *AdminShopHandler) CreateProduct(c *gin.Context) {
	var in services.ProductInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	creatorID := mustAdminUserID(c)
	p, err := h.shopSvc.CreateProduct(c.Request.Context(), in, creatorID)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "建立失敗"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "商品已建立", "product": p})
}

// PUT /v1/admin/products/:id
func (h *AdminShopHandler) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }
	var in services.ProductInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.shopSvc.UpdateProduct(c.Request.Context(), id, in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失敗"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已更新", "product": p})
}

// DELETE /v1/admin/products/:id
func (h *AdminShopHandler) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }
	if err := h.shopSvc.DeleteProduct(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "刪除失敗"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已刪除"})
}

// ── Orders ────────────────────────────────────────────────────

// GET /v1/admin/orders
func (h *AdminShopHandler) ListOrders(c *gin.Context) {
	var in services.ListOrdersInput
	_ = c.ShouldBindQuery(&in)
	result, err := h.shopSvc.ListOrders(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢失敗"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GET /v1/admin/orders/:id
func (h *AdminShopHandler) GetOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }
	order, err := h.shopSvc.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "訂單不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢失敗"})
		}
		return
	}
	c.JSON(http.StatusOK, order)
}

// PUT /v1/admin/orders/:id
func (h *AdminShopHandler) UpdateOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }
	var in services.UpdateOrderInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order, err := h.shopSvc.UpdateOrder(c.Request.Context(), id, in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失敗"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已更新", "order": order})
}
