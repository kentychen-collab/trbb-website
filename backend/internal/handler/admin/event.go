package admin

import (
	"net/http"
	"strconv"
	"sports-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type EventAdminHandler struct {
	eventSvc *service.EventService
}

func NewEventAdminHandler(eventSvc *service.EventService) *EventAdminHandler {
	return &EventAdminHandler{eventSvc: eventSvc}
}

func (h *EventAdminHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 { page = 1 }
	list, total, err := h.eventSvc.AdminList(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total})
}

func (h *EventAdminHandler) Create(c *gin.Context) {
	var input service.CreateEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	e, err := h.eventSvc.AdminCreate(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": e})
}

func (h *EventAdminHandler) Update(c *gin.Context) {
	// TODO: implement update
	c.JSON(http.StatusOK, gin.H{"message": "TODO: update event"})
}
