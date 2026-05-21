package handler

import (
	"net/http"
	"sports-platform/internal/model"
	"sports-platform/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MemberHandler struct {
	memberSvc *service.MemberService
}

func NewMemberHandler(memberSvc *service.MemberService) *MemberHandler {
	return &MemberHandler{memberSvc: memberSvc}
}

func (h *MemberHandler) GetProfile(c *gin.Context) {
	memberID := c.MustGet("member_id").(uint64)
	m, err := h.memberSvc.GetProfile(memberID)
	if err != nil || m == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": m})
}

func (h *MemberHandler) UpdateProfile(c *gin.Context) {
	memberID := c.MustGet("member_id").(uint64)
	var input service.UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m, err := h.memberSvc.UpdateProfile(memberID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": m})
}

func (h *MemberHandler) ListAddresses(c *gin.Context) {
	memberID := c.MustGet("member_id").(uint64)
	list, err := h.memberSvc.ListAddresses(memberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *MemberHandler) CreateAddress(c *gin.Context) {
	memberID := c.MustGet("member_id").(uint64)
	var a model.MemberAddress
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.memberSvc.CreateAddress(memberID, a); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "address created"})
}

func (h *MemberHandler) UpdateAddress(c *gin.Context) {
	memberID := c.MustGet("member_id").(uint64)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var a model.MemberAddress
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	a.ID = id
	if err := h.memberSvc.UpdateAddress(memberID, a); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "address updated"})
}

func (h *MemberHandler) DeleteAddress(c *gin.Context) {
	memberID := c.MustGet("member_id").(uint64)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.memberSvc.DeleteAddress(id, memberID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "address deleted"})
}
