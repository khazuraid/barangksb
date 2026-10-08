package middleware

import (
	"net/http"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"

	"inventariskantor/internal/auth"
)

func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// skip public routes
		path := c.Request.URL.Path
		if path == "/api/auth/login" || strings.HasPrefix(path, "/api/scan/") ||
			strings.HasSuffix(path, ".png") || path == "/api/barcode/sheet" {
			c.Next()
			return
		}

		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
			c.Abort()
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := auth.ParseToken(secret, tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}
		c.Set("uid", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("name", claims.Name)
		c.Next()
	}
}

func RBACMiddleware(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		roleStr, _ := role.(string)
		if roleStr == "" {
			roleStr = "anonymous"
		}
		path := c.Request.URL.Path
		method := c.Request.Method

		allowed, err := enforcer.Enforce(roleStr, path, method)
		if err != nil || !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}
