package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	MemberID uint64 `json:"member_id,omitempty"`
	AdminID  uint64 `json:"admin_id,omitempty"`
	Role     int    `json:"role"`
	IsAdmin  bool   `json:"is_admin,omitempty"`
	jwt.RegisteredClaims
}

func parseToken(tokenStr, secret string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return claims, nil
}

// JWTAuth - 前台會員 token 驗證
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		claims, err := parseToken(tokenStr, secret)
		if err != nil || claims.IsAdmin {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("member_id", claims.MemberID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// AdminJWTAuth - 後台管理員 token 驗證
func AdminJWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		claims, err := parseToken(tokenStr, secret)
		if err != nil || !claims.IsAdmin {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid admin token"})
			return
		}
		c.Set("admin_id", claims.AdminID)
		c.Set("admin_role", claims.Role)
		c.Next()
	}
}

// AdminOnly - role >= 1（一般管理員以上）
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("admin_role")
		if role.(int) < 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

// SuperAdminOnly - role >= 2（超級管理員）
func SuperAdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("admin_role")
		if role.(int) < 2 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "需要超級管理員權限"})
			return
		}
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}
