package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"trbb/pkg/logger"
)

// Logger — JSON request log，500 時印出錯誤詳情
func Logger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		status := c.Writer.Status()
		fields := []any{
			"method",     c.Request.Method,
			"path",       c.Request.URL.Path,
			"status",     status,
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip",  c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		}
		// 500 時把 gin errors 也印出來
		if status >= 500 {
			if errs := c.Errors.String(); errs != "" {
				fields = append(fields, "gin_errors", errs)
			}
			log.Error("request", fields...)
		} else {
			log.Info("request", fields...)
		}
	}
}

// Recovery — panic recovery，把錯誤附加到 gin context
func Recovery(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("panic recovered",
					"error", err,
					"path",  c.Request.URL.Path,
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError,
					gin.H{"error": "internal server error"})
			}
		}()
		c.Next()
	}
}

// CORS
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin,Authorization,Content-Type,Accept")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// RealIP — 取得真實 client IP
func RealIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		if ip := c.GetHeader("X-Real-IP"); ip != "" {
			c.Request.RemoteAddr = ip
		} else if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
			c.Request.RemoteAddr = strings.Split(ip, ",")[0]
		}
		c.Next()
	}
}

// JWT — 驗證 Bearer token
func JWT(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("user_id", claims["sub"])
		c.Set("user_role", claims["role"])
		c.Next()
	}
}

// RequireRole — 最低角色限制
func RequireRole(minRole int) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		r, ok := role.(float64)
		if !ok || int(r) < minRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient role"})
			return
		}
		c.Next()
	}
}

// OptionalJWT — parses token if present but does not abort if missing
func OptionalJWT(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr != "" {
			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
				return []byte(secret), nil
			})
			if err == nil && token.Valid {
				c.Set("user_id", claims["sub"])
				c.Set("user_role", claims["role"])
			}
		}
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return c.Query("token")
}
