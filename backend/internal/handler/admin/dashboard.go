package admin

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	db *sql.DB
}

func NewDashboardHandler(db *sql.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

func (h *DashboardHandler) Stats(c *gin.Context) {
	var todayOrders, activeMembers, pendingMembers, monthlyEvents int

	h.db.QueryRow(`SELECT COUNT(*) FROM orders WHERE DATE(created_at)=CURDATE()`).Scan(&todayOrders)
	h.db.QueryRow(`SELECT COUNT(*) FROM members WHERE status=1`).Scan(&activeMembers)
	h.db.QueryRow(`SELECT COUNT(*) FROM members WHERE status=0`).Scan(&pendingMembers)
	h.db.QueryRow(`SELECT COUNT(*) FROM events WHERE YEAR(event_date)=YEAR(NOW()) AND MONTH(event_date)=MONTH(NOW())`).Scan(&monthlyEvents)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"today_orders":    todayOrders,
			"active_members":  activeMembers,
			"pending_members": pendingMembers,
			"monthly_events":  monthlyEvents,
		},
	})
}
