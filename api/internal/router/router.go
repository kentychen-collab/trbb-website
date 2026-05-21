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

	// Repositories
	memberRepo  := repository.NewMemberRepo(db)
	eventRepo   := repository.NewEventRepo(db)
	orderRepo   := repository.NewOrderRepo(db)
	adminRepo   := repository.NewAdminRepo(db)
	productRepo := repository.NewProductRepo(db)

	// Services
	authSvc    := service.NewAuthService(memberRepo, cfg.JWTSecret, cfg.JWTExpireHours)
	memberSvc  := service.NewMemberService(memberRepo)
	eventSvc   := service.NewEventService(eventRepo, orderRepo)
	adminSvc   := service.NewAdminService(adminRepo, cfg)
	productSvc := service.NewProductService(productRepo)

	// Handlers
	authH    := handler.NewAuthHandler(authSvc)
	memberH  := handler.NewMemberHandler(memberSvc)
	eventH   := handler.NewEventHandler(eventSvc)
	productH := handler.NewProductHandler(productSvc)
	uploadH  := handler.NewUploadHandler(cfg.UploadDir, cfg.UploadURLPrefix)

	// Admin handlers
	adminAuthH    := admin.NewAuthAdminHandler(adminSvc)
	adminUserH    := admin.NewAdminUserHandler(adminSvc)
	adminMemberH  := admin.NewMemberAdminHandler(memberRepo)
	adminOrderH   := admin.NewOrderAdminHandler(orderRepo)
	adminEventH   := admin.NewEventAdminHandler(eventSvc)
	adminProductH := admin.NewProductAdminHandler(productSvc)
	adminDashH    := admin.NewDashboardHandler(db)

	// ── 前台 API  /api/v1 ──────────────────────────────
	api := r.Group("/api/v1")

	api.POST("/auth/register", authH.Register)
	api.POST("/auth/login",    authH.Login)

	api.GET("/products",        productH.List)
	api.GET("/products/:slug",  productH.GetBySlug)
	api.GET("/categories",      productH.ListCategories)
	api.GET("/events",          eventH.List)
	api.GET("/events/:slug",    eventH.GetBySlug)
	api.GET("/races",           handler.NotImplemented("race calendar"))
	api.GET("/second-hand",     handler.NotImplemented("second-hand list"))
	api.GET("/second-hand/:id", handler.NotImplemented("second-hand detail"))
	api.POST("/payment/ecpay/callback", handler.NotImplemented("ECPay callback"))

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

		authG.POST("/orders",          handler.NotImplemented("create order"))
		authG.GET("/orders",           handler.NotImplemented("list orders"))
		authG.GET("/orders/:order_no", handler.NotImplemented("get order"))

		authG.POST("/races/:id/save",    handler.NotImplemented("save race"))
		authG.DELETE("/races/:id/save",  handler.NotImplemented("unsave race"))

		authG.GET("/garmin/auth-url",     handler.NotImplemented("Garmin OAuth"))
		authG.GET("/garmin/callback",     handler.NotImplemented("Garmin callback"))
		authG.GET("/me/activities",       handler.NotImplemented("Garmin activities"))
		authG.POST("/me/activities/sync", handler.NotImplemented("Garmin sync"))

		authG.POST("/upload/image", uploadH.Upload)
	}

	// ── 後台 API  /admin-api/v1 ────────────────────────
	adminAPI := r.Group("/admin-api/v1")
	adminAPI.POST("/auth/login", adminAuthH.Login)

	adminAuth := adminAPI.Group("/")
	adminAuth.Use(middleware.AdminJWTAuth(cfg.JWTSecret))
	{
		adminAuth.GET("/me", adminAuthH.Me)
		adminAuth.GET("/dashboard/stats", adminDashH.Stats)

		// 上傳（後台共用）
		adminAuth.POST("/upload/image", uploadH.Upload)

		// 會員
		adminAuth.GET("/members",              adminMemberH.List)
		adminAuth.PUT("/members/:id/approve",  adminMemberH.Approve)
		adminAuth.PUT("/members/:id/reject",   adminMemberH.Reject)
		adminAuth.PUT("/members/:id/status",   adminMemberH.SetStatus)

		// 訂單
		adminAuth.GET("/orders",              adminOrderH.List)
		adminAuth.PUT("/orders/:id/status",   adminOrderH.UpdateStatus)

		// 活動
		adminAuth.GET("/events",      adminEventH.List)
		adminAuth.GET("/events/:id",  adminEventH.Get)
		adminAuth.POST("/events",     adminEventH.Create)
		adminAuth.PUT("/events/:id",  adminEventH.Update)

		// 商品
		adminAuth.GET("/products",            adminProductH.List)
		adminAuth.GET("/products/:id",        adminProductH.Get)
		adminAuth.POST("/products",           adminProductH.Create)
		adminAuth.PUT("/products/:id",        adminProductH.Update)
		adminAuth.PUT("/products/:id/status", adminProductH.SetStatus)
		adminAuth.GET("/categories",          adminProductH.ListCategories)

		// 管理員
		adminAuth.GET("/admins",                   adminUserH.List)
		adminAuth.POST("/admins",                  adminUserH.Create)
		adminAuth.GET("/admins/:id/permissions",   adminUserH.GetPermissions)
		adminAuth.PUT("/admins/:id/permissions",   adminUserH.SetPermissions)
		adminAuth.PUT("/admins/:id/status",        adminUserH.SetStatus)

		// Stubs
		adminAuth.GET("/second-hand",              handler.NotImplemented("admin second-hand"))
		adminAuth.PUT("/second-hand/:id/approve",  handler.NotImplemented("admin approve second-hand"))
		adminAuth.GET("/races",                    handler.NotImplemented("admin races"))
		adminAuth.POST("/races",                   handler.NotImplemented("admin create race"))
		adminAuth.PUT("/races/:id",                handler.NotImplemented("admin update race"))
	}

	return r
}
