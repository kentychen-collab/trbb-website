package handler

import (
	"net/http"
	"strconv"
	"sports-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	eventSvc *service.EventService
}

func NewEventHandler(eventSvc *service.EventService) *EventHandler {
	return &EventHandler{eventSvc: eventSvc}
}

func (h *EventHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 { page = 1 }
	if limit < 1 || limit > 50 { limit = 10 }

	list, total, err := h.eventSvc.List(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total, "page": page, "limit": limit})
}

func (h *EventHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	e, err := h.eventSvc.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if e == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": e})
}

func (h *EventHandler) Register(c *gin.Context) {
	memberID := c.MustGet("member_id").(uint64)
	eventID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}
	var input service.RegisterEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	reg, err := h.eventSvc.Register(eventID, memberID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": reg, "message": "registration successful"})
}

func (h *EventHandler) MyRegistrations(c *gin.Context) {
	memberID := c.MustGet("member_id").(uint64)
	list, err := h.eventSvc.MyRegistrations(memberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}
