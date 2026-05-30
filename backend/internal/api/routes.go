package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apiHandlers "trbb/internal/api/handlers"
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

	userSvc     := services.NewUserService(db, cfg.App.SecretKey)
	sitesSvc    := services.NewSiteSettingsService(db)
	eventSvc    := services.NewEventService(db)
	shopSvc     := services.NewShopService(db)
	trainingSvc := services.NewTrainingService(db, minio)
	authH       := apiHandlers.NewAuthHandler(userSvc)
	eventH      := apiHandlers.NewEventHandler(eventSvc)
	shopH       := apiHandlers.NewShopHandler(shopSvc)
	trainingH   := apiHandlers.NewTrainingHandler(trainingSvc, cfg)
	uploadH     := apiHandlers.NewUploadHandler(minio)
	siteH       := apiHandlers.NewSiteSettingsHandler(sitesSvc)

	// ── Site Settings (public) ───────────────────────────────
	r.GET("/settings", siteH.GetPublic)

	// ── Auth ────────────────────────────────────────────────
	auth := r.Group("/auth")
	{
		auth.POST("/register",        authH.Register)
		auth.POST("/login",           authH.Login)
		auth.POST("/refresh",         todo("refresh"))
		auth.POST("/forgot-password", todo("forgot-password"))
		auth.POST("/reset-password",  todo("reset-password"))
	}

	// ── Public GET ───────────────────────────────────────────
	r.GET("/events",                  eventH.ListEvents)
	r.GET("/events/:id",              eventH.GetEvent)
	r.GET("/products",                shopH.ListProducts)
	r.GET("/products/:id",            shopH.GetProduct)
	r.GET("/announcements",           todo("announcements"))
	r.GET("/announcements/:id",       todo("announcement-detail"))
	r.GET("/secondhand",              todo("secondhand-list"))
	r.GET("/secondhand/:id",          todo("secondhand-detail"))

	// ── Training (public-but-auth-aware: optionalJWT) ────────
	optJWT := middleware.OptionalJWT(cfg.App.SecretKey)
	r.GET("/training",                trainingH.ListPublic)
	r.GET("/training/share/:uuid",    optJWT, trainingH.GetByUUID)
	r.GET("/training/:id",            optJWT, trainingH.GetLog)

	p := r.Group("", middleware.JWT(cfg.App.SecretKey))
	{
		// 個人資料
		me := p.Group("/me")
		{
			me.GET("",                        authH.GetProfile)
			me.PUT("",                        authH.UpdateProfile)
			me.PUT("/password",               authH.ChangePassword)
			me.GET("/registration-profile",   authH.GetRegistrationProfile)
			me.POST("/avatar",                todo("upload-avatar"))
			me.GET("/notifications",          todo("notifications"))
			me.PUT("/notifications/:id/read", todo("mark-read"))
			// 訓練日記
			me.GET("/training",               trainingH.ListMy)
			// Garmin
			me.GET("/garmin/status",          trainingH.GarminStatus)
			me.GET("/garmin/connect",         trainingH.GarminConnect)
			me.DELETE("/garmin/disconnect",   trainingH.GarminDisconnect)

			// Strava
			me.GET("/strava/status",          trainingH.StravaStatus)
			me.GET("/strava/connect",         trainingH.StravaConnect)
			me.DELETE("/strava/disconnect",   trainingH.StravaDisconnect)
		}

		// 賽事
		p.GET("/events/:id/register",              eventH.GetMyRegistration)
		p.POST("/events/:id/register",             eventH.Register)
		p.DELETE("/events/:id/register",           eventH.CancelRegistration)

		// 商城
		p.GET("/orders",                           shopH.ListMyOrders)
		p.GET("/orders/:id",                       shopH.GetMyOrder)
		p.POST("/orders",                          shopH.CreateOrder)
		p.POST("/orders/:id/pay",                  todo("orders-pay"))

		// 訓練日記
		p.POST("/training",                        trainingH.Create)
		p.PUT("/training/:id",                     trainingH.Update)
		p.PATCH("/training/:id/public",            trainingH.PatchPublic)
		p.DELETE("/training/:id",                  trainingH.Delete)
		p.POST("/training/upload/gpx",             trainingH.UploadGPX)
		p.POST("/training/upload/fit",             trainingH.UploadFIT)
		p.POST("/training/garmin/sync",            trainingH.GarminSync)
		p.POST("/training/strava/sync",            trainingH.StravaSync)

		// 圖片上傳
		p.POST("/upload/image",                    uploadH.UploadImage)

		// 二手
		p.POST("/secondhand",                      todo("secondhand-create"))
		p.PUT("/secondhand/:id",                   todo("secondhand-update"))
		p.DELETE("/secondhand/:id",                todo("secondhand-delete"))
	}
}

func todo(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"msg": "TODO: " + name})
	}
}
