package handler

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// NotImplemented returns 501 for features not yet built (金流, Garmin, etc.)
func NotImplemented(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error":   "not implemented",
			"feature": feature,
			"message": feature + " is not yet integrated. Coming soon.",
		})
	}
}
