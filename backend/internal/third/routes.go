package third

import (
	"github.com/gin-gonic/gin"
	apiHandlers "trbb/internal/api/handlers"
	"trbb/internal/config"
	"trbb/pkg/database"
	"trbb/pkg/logger"
	"trbb/pkg/storage"
	"trbb/internal/services"
)

func RegisterRoutes(r *gin.RouterGroup, db *database.DB,
	minio *storage.Storage, cfg *config.Config, log *logger.Logger) {

	trainingSvc := services.NewTrainingService(db, minio)
	trainingH   := apiHandlers.NewTrainingHandler(trainingSvc, cfg)

	// Garmin OAuth 1.0a callback
	r.GET("/garmin/callback", trainingH.GarminCallback)

	// Placeholder stubs for future third-party callbacks
	r.POST("/ecpay/callback",    func(c *gin.Context) { c.JSON(200, gin.H{"msg": "TODO: ecpay"}) })
	r.POST("/linepay/callback",  func(c *gin.Context) { c.JSON(200, gin.H{"msg": "TODO: linepay"}) })
}
