package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"trbb/pkg/database"
)

type DashboardHandler struct {
	db *database.DB
}

func NewDashboardHandler(db *database.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

type DashboardStats struct {
	TotalMembers      int     `json:"total_members"`
	NewMembersMonth   int     `json:"new_members_month"`
	TotalEvents       int     `json:"total_events"`
	MonthRegistrations int    `json:"month_registrations"`
	PendingOrders     int     `json:"pending_orders"`
	MonthRevenue      float64 `json:"month_revenue"`
	TotalTrainingLogs int     `json:"total_training_logs"`
	PublicLogs        int     `json:"public_logs"`
}

type RecentRegistration struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	EventName string `json:"event_name"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type RecentOrder struct {
	ID          uint64  `json:"id"`
	MemberName  string  `json:"member_name"`
	ProductName string  `json:"product_name"`
	Amount      float64 `json:"amount"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
}

// GET /v1/admin/dashboard
func (h *DashboardHandler) Get(c *gin.Context) {
	ctx := c.Request.Context()
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastMonthStart := monthStart.AddDate(0, -1, 0)

	var stats DashboardStats

	// 總會員數
	h.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users WHERE deleted_at IS NULL").
		Scan(&stats.TotalMembers)

	// 本月新增會員
	h.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users WHERE created_at >= ? AND deleted_at IS NULL",
		monthStart).Scan(&stats.NewMembersMonth)

	// 上月新增會員（計算趨勢用）
	var lastMonthMembers int
	h.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users WHERE created_at >= ? AND created_at < ? AND deleted_at IS NULL",
		lastMonthStart, monthStart).Scan(&lastMonthMembers)

	// 賽事總數
	h.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM events").Scan(&stats.TotalEvents)

	// 本月報名數
	h.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM event_registrations WHERE created_at >= ?",
		monthStart).Scan(&stats.MonthRegistrations)

	// 待處理訂單
	h.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM orders WHERE status='pending'").
		Scan(&stats.PendingOrders)

	// 本月營收（已付款訂單）
	h.db.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(total_amount),0) FROM orders WHERE status IN ('paid','completed') AND created_at >= ?",
		monthStart).Scan(&stats.MonthRevenue)

	// 訓練日記總數
	h.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM training_logs").Scan(&stats.TotalTrainingLogs)

	// 公開訓練日記
	h.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM training_logs WHERE is_public=1").Scan(&stats.PublicLogs)

	// 最新 5 筆報名
	regRows, _ := h.db.QueryContext(ctx, `
		SELECT r.id,
		       COALESCE(u.real_name, u.display_name, u.email) AS name,
		       e.title,
		       r.status,
		       r.created_at
		FROM event_registrations r
		JOIN users u ON u.id = r.user_id
		JOIN events e ON e.id = r.event_id
		ORDER BY r.created_at DESC LIMIT 5`)
	var recentRegs []RecentRegistration
	if regRows != nil {
		defer regRows.Close()
		for regRows.Next() {
			var rr RecentRegistration
			var createdAt time.Time
			regRows.Scan(&rr.ID, &rr.Name, &rr.EventName, &rr.Status, &createdAt)
			rr.CreatedAt = createdAt.Format("2006-01-02 15:04")
			recentRegs = append(recentRegs, rr)
		}
	}
	if recentRegs == nil {
		recentRegs = []RecentRegistration{}
	}

	// 最新 5 筆訂單
	orderRows, _ := h.db.QueryContext(ctx, `
		SELECT o.id,
		       COALESCE(u.real_name, u.display_name, u.email) AS name,
		       COALESCE(p.title, '商品') AS product,
		       o.total_amount,
		       o.status,
		       o.created_at
		FROM orders o
		JOIN users u ON u.id = o.user_id
		LEFT JOIN order_items oi ON oi.order_id = o.id
		LEFT JOIN products p ON p.id = oi.product_id
		ORDER BY o.created_at DESC LIMIT 5`)
	var recentOrders []RecentOrder
	if orderRows != nil {
		defer orderRows.Close()
		for orderRows.Next() {
			var ro RecentOrder
			var createdAt time.Time
			orderRows.Scan(&ro.ID, &ro.MemberName, &ro.ProductName,
				&ro.Amount, &ro.Status, &createdAt)
			ro.CreatedAt = createdAt.Format("2006-01-02 15:04")
			recentOrders = append(recentOrders, ro)
		}
	}
	if recentOrders == nil {
		recentOrders = []RecentOrder{}
	}

	c.JSON(http.StatusOK, gin.H{
		"stats":           stats,
		"recent_regs":     recentRegs,
		"recent_orders":   recentOrders,
		"last_month_members": lastMonthMembers,
	})
}
