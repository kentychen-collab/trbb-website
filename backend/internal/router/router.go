package router

import (
	"database/sql"
	"net/http"
	"sports-platform/internal/config"
	"sports-platform/internal/handler"
	"sports-platform/internal/handler/admin"
	"sports-platform/internal/middleware"
	"sports-platform/internal/repository"
	"sports-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func Setup(cfg *config.Config, db *sql.DB) *gin.Engine {
	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// --- Repositories ---
	memberRepo := repository.NewMemberRepo(db)
	eventRepo  := repository.NewEventRepo(db)
	orderRepo  := repository.NewOrderRepo(db)
	adminRepo  := repository.NewAdminRepo(db)

	// --- Services ---
	authSvc   := service.NewAuthService(memberRepo, cfg.JWTSecret, cfg.JWTExpireHours)
	memberSvc := service.NewMemberService(memberRepo)
	eventSvc  := service.NewEventService(eventRepo, orderRepo)
	adminSvc  := service.NewAdminService(adminRepo, cfg)

	// --- Handlers ---
	authH   := handler.NewAuthHandler(authSvc)
	memberH := handler.NewMemberHandler(memberSvc)
	eventH  := handler.NewEventHandler(eventSvc)

	// Admin handlers
	adminAuthH    := admin.NewAuthAdminHandler(adminSvc)
	adminUserH    := admin.NewAdminUserHandler(adminSvc)
	adminMemberH  := admin.NewMemberAdminHandler(memberRepo)
	adminOrderH   := admin.NewOrderAdminHandler(orderRepo)
	adminEventH   := admin.NewEventAdminHandler(eventSvc)
	adminDashH    := admin.NewDashboardHandler(db)

	// =============================================
	// 前台 API  /api/v1/...
	// =============================================
	api := r.Group("/api/v1")

	// Public
	api.POST("/auth/register", authH.Register)
	api.POST("/auth/login",    authH.Login)
	api.POST("/auth/refresh",  handler.NotImplemented("JWT refresh"))

	api.GET("/events",       eventH.List)
	api.GET("/events/:slug", eventH.GetBySlug)
	api.GET("/products",     handler.NotImplemented("products list"))
	api.GET("/products/:slug", handler.NotImplemented("product detail"))
	api.GET("/categories",   handler.NotImplemented("categories"))
	api.GET("/races",        handler.NotImplemented("race calendar"))
	api.GET("/second-hand",  handler.NotImplemented("second-hand list"))
	api.GET("/second-hand/:id", handler.NotImplemented("second-hand detail"))
	api.POST("/payment/ecpay/callback", handler.NotImplemented("ECPay callback"))

	// Authenticated
	authG := api.Group("/")
	authG.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		authG.GET("/me",          memberH.GetProfile)
		authG.PUT("/me",          memberH.UpdateProfile)
		authG.PUT("/me/password", authH.ChangePassword)

		authG.GET("/me/addresses",        memberH.ListAddresses)
		authG.POST("/me/addresses",       memberH.CreateAddress)
		authG.PUT("/me/addresses/:id",    memberH.UpdateAddress)
		authG.DELETE("/me/addresses/:id", memberH.DeleteAddress)

		authG.POST("/events/:id/register", eventH.Register)
		authG.GET("/me/registrations",     eventH.MyRegistrations)

		authG.GET("/cart",        handler.NotImplemented("cart"))
		authG.POST("/cart",       handler.NotImplemented("add to cart"))
		authG.PUT("/cart/:id",    handler.NotImplemented("update cart"))
		authG.DELETE("/cart/:id", handler.NotImplemented("remove from cart"))
		authG.POST("/orders",     handler.NotImplemented("create order"))
		authG.GET("/orders",      handler.NotImplemented("list orders"))
		authG.GET("/orders/:order_no", handler.NotImplemented("get order"))

		authG.POST("/second-hand",           handler.NotImplemented("create second-hand"))
		authG.PUT("/second-hand/:id",        handler.NotImplemented("update second-hand"))
		authG.POST("/second-hand/:id/inquire", handler.NotImplemented("inquire second-hand"))

		authG.POST("/races/:id/save",    handler.NotImplemented("save race"))
		authG.DELETE("/races/:id/save",  handler.NotImplemented("unsave race"))

		authG.GET("/garmin/auth-url",     handler.NotImplemented("Garmin OAuth"))
		authG.GET("/garmin/callback",     handler.NotImplemented("Garmin callback"))
		authG.GET("/me/activities",       handler.NotImplemented("Garmin activities"))
		authG.POST("/me/activities/sync", handler.NotImplemented("Garmin sync"))

		authG.POST("/upload/image", handler.NotImplemented("image upload"))
	}

	// =============================================
	// 後台 API  /admin-api/v1/...
	// =============================================
	adminAPI := r.Group("/admin-api/v1")

	// Public: admin login only
	adminAPI.POST("/auth/login", adminAuthH.Login)

	// All other admin routes require admin JWT
	adminAuth := adminAPI.Group("/")
	adminAuth.Use(middleware.AdminJWTAuth(cfg.JWTSecret))
	{
		adminAuth.GET("/me", adminAuthH.Me)

		// Dashboard
		adminAuth.GET("/dashboard/stats", adminDashH.Stats)

		// 會員管理（需要 manage_members 或超管）
		adminAuth.GET("/members",                adminMemberH.List)
		adminAuth.PUT("/members/:id/approve",    adminMemberH.Approve)
		adminAuth.PUT("/members/:id/reject",     adminMemberH.Reject)
		adminAuth.PUT("/members/:id/status",     adminMemberH.SetStatus)

		// 訂單管理
		adminAuth.GET("/orders",             adminOrderH.List)
		adminAuth.PUT("/orders/:id/status",  adminOrderH.UpdateStatus)

		// 活動管理
		adminAuth.GET("/events",      adminEventH.List)
		adminAuth.POST("/events",     adminEventH.Create)
		adminAuth.PUT("/events/:id",  adminEventH.Update)

		// 管理員帳號管理（僅超管）
		adminAuth.GET("/admins",                       adminUserH.List)
		adminAuth.POST("/admins",                      adminUserH.Create)
		adminAuth.GET("/admins/:id/permissions",       adminUserH.GetPermissions)
		adminAuth.PUT("/admins/:id/permissions",       adminUserH.SetPermissions)
		adminAuth.PUT("/admins/:id/status",            adminUserH.SetStatus)

		// Stubs
		adminAuth.GET("/products",               handler.NotImplemented("admin products"))
		adminAuth.POST("/products",              handler.NotImplemented("admin create product"))
		adminAuth.PUT("/products/:id",           handler.NotImplemented("admin update product"))
		adminAuth.DELETE("/products/:id",        handler.NotImplemented("admin delete product"))
		adminAuth.GET("/second-hand",            handler.NotImplemented("admin second-hand"))
		adminAuth.PUT("/second-hand/:id/approve", handler.NotImplemented("admin approve second-hand"))
		adminAuth.GET("/races",                  handler.NotImplemented("admin races"))
		adminAuth.POST("/races",                 handler.NotImplemented("admin create race"))
		adminAuth.PUT("/races/:id",              handler.NotImplemented("admin update race"))
	}

	return r
}
