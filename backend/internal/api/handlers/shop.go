package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"trbb/internal/services"
)

type ShopHandler struct {
	shopSvc *services.ShopService
}

func NewShopHandler(shopSvc *services.ShopService) *ShopHandler {
	return &ShopHandler{shopSvc: shopSvc}
}

// GET /v1/api/products
func (h *ShopHandler) ListProducts(c *gin.Context) {
	var in services.ListProductsInput
	_ = c.ShouldBindQuery(&in)
	result, err := h.shopSvc.ListPublic(c.Request.Context(), in)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢商品失敗"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GET /v1/api/products/:id
func (h *ShopHandler) GetProduct(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }
	p, err := h.shopSvc.GetProductByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "商品不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢失敗"})
		}
		return
	}
	if p.Status != 1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "商品不存在"})
		return
	}
	c.JSON(http.StatusOK, p)
}

// POST /v1/api/orders
func (h *ShopHandler) CreateOrder(c *gin.Context) {
	var in services.OrderInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order, err := h.shopSvc.CreateOrder(c.Request.Context(), mustUserID(c), in)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrOutOfStock):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrProductNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "商品不存在"})
		default:
			_ = c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "建立訂單失敗"})
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "訂單已建立", "order": order})
}

// GET /v1/api/orders  (my orders)
func (h *ShopHandler) ListMyOrders(c *gin.Context) {
	userID := mustUserID(c)
	var in services.ListOrdersInput
	_ = c.ShouldBindQuery(&in)
	in.UserID = &userID
	result, err := h.shopSvc.ListOrders(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢失敗"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GET /v1/api/orders/:id
func (h *ShopHandler) GetMyOrder(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }
	order, err := h.shopSvc.GetOrderByIDAndUser(c.Request.Context(), id, mustUserID(c))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "訂單不存在"})
		return
	}
	c.JSON(http.StatusOK, order)
}
