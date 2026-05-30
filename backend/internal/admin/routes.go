package admin

import (
	"github.com/gin-gonic/gin"
	adminHandlers "trbb/internal/admin/handlers"
	"trbb/internal/config"
	"trbb/internal/middleware"
	"trbb/internal/services"
	"trbb/pkg/cache"
	"trbb/pkg/database"
	"trbb/pkg/logger"
	"trbb/pkg/storage"
)

func RegisterRoutes(r *gin.RouterGroup, db *database.DB, rdb *cache.Cache,
	minio *storage.Storage, cfg *config.Config, log *logger.Logger) {

	userSvc   := services.NewUserService(db, cfg.App.SecretKey)
	eventSvc  := services.NewEventService(db)
	shopSvc      := services.NewShopService(db)
	sitesSvc     := services.NewSiteSettingsService(db)
	trainingSvc  := services.NewTrainingService(db, minio)
	userH        := adminHandlers.NewAdminUserHandler(userSvc)
	trainingH    := adminHandlers.NewAdminTrainingHandler(trainingSvc)
	eventH    := adminHandlers.NewAdminEventHandler(eventSvc)
	shopH     := adminHandlers.NewAdminShopHandler(shopSvc)
	uploadH   := adminHandlers.NewUploadHandler(minio)
	siteH     := adminHandlers.NewAdminSiteSettingsHandler(sitesSvc, minio)

	// ── Auth ────────────────────────────────────────────────
	auth := r.Group("/auth")
	{
		auth.POST("/login",   userH.Login)
		auth.POST("/refresh", func(c *gin.Context) { c.JSON(501, gin.H{"msg": "TODO"}) })
	}

	// ── Protected (role >= 8) ────────────────────────────────
	a := r.Group("", middleware.JWT(cfg.App.SecretKey), middleware.RequireRole(8))
	{
		a.GET("/dashboard", func(c *gin.Context) { c.JSON(200, gin.H{"msg": "ok"}) })

		// ── 一般會員管理 ─────────────────────────────────────
		members := a.Group("/members")
		{
			members.GET("",              userH.ListMembers)
			members.POST("",             userH.CreateMember)      // 後台新增一般會員
			members.GET("/:id",          userH.GetUser)
			members.PUT("/:id/profile",  userH.UpdateUserProfile) // 修改資料（不含帳號）
			members.PUT("/:id/password", userH.SetUserPassword)   // 直接設定密碼
			members.PUT("/:id/status",   userH.UpdateUserStatus)  // 核准/停用/拒絕
		}

		// ── 管理員管理（超級管理員限定部分操作） ────────────
		admins := a.Group("/admins")
		{
			admins.GET("",    userH.ListAdmins)
			admins.GET("/:id",         userH.GetUser)
			admins.PUT("/:id/profile", userH.UpdateUserProfile)
			admins.PUT("/:id/password",userH.SetUserPassword)
			// 以下僅超級管理員可用
			admins.POST("",    middleware.RequireRole(9), userH.CreateAdmin)
			admins.DELETE("/:id", middleware.RequireRole(9), userH.DeleteAdmin)
		}

		// ── 賽事管理 ─────────────────────────────────────────
		events := a.Group("/events")
		{
			events.GET("",                          eventH.ListEvents)
			events.POST("",                         eventH.CreateEvent)
			events.GET("/:id",                      eventH.GetEvent)
			events.PUT("/:id",                      eventH.UpdateEvent)
			events.DELETE("/:id",                   eventH.DeleteEvent)
			events.GET("/:id/registrations",        eventH.ListRegistrations)
			events.PUT("/:id/registrations/:regId", eventH.UpdateRegistration)
			events.GET("/:id/registrations/export", eventH.ExportRegistrations)
		}

		// ── 訓練日記管理 ─────────────────────────────────────
		training := a.Group("/training")
		{
			training.GET("",       trainingH.ListTraining)
			training.GET("/stats", trainingH.Stats)
			training.GET("/:id",   trainingH.GetTraining)
		}

		// ── 網站設定 ─────────────────────────────────────────
		ss := a.Group("/settings")
		{
			ss.GET("",              siteH.GetAll)
			ss.POST("",             siteH.BatchUpdate)
			ss.PUT("/:key",         siteH.SetOne)
			ss.POST("/upload/:purpose", siteH.UploadSettingImage)
		}

		// ── TODO stubs ───────────────────────────────────────
		// 商品管理
		products := a.Group("/products")
		{
			products.GET("",      shopH.ListProducts)
			products.POST("",     shopH.CreateProduct)
			products.GET("/:id",  shopH.GetProduct)
			products.PUT("/:id",  shopH.UpdateProduct)
			products.DELETE("/:id", shopH.DeleteProduct)
		}

		// 訂單管理
		orders := a.Group("/orders")
		{
			orders.GET("",      shopH.ListOrders)
			orders.GET("/:id",  shopH.GetOrder)
			orders.PUT("/:id",  shopH.UpdateOrder)
		}
		a.GET("/announcements",  func(c *gin.Context) { c.JSON(501, gin.H{"msg": "TODO"}) })
		a.POST("/announcements", func(c *gin.Context) { c.JSON(501, gin.H{"msg": "TODO"}) })
		// 圖片上傳
		a.POST("/upload/image", uploadH.UploadImage)
	}
}
