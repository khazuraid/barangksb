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
		// public routes: login, scan landing, barcode images, barcode sheet
		path := c.Request.URL.Path
		if path == "/api/auth/login" ||
			strings.HasPrefix(path, "/api/scan/") ||
			strings.HasPrefix(path, "/api/barcode/") {
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
		// skip public routes (same set as AuthMiddleware)
		path := c.Request.URL.Path
		if path == "/api/auth/login" ||
			strings.HasPrefix(path, "/api/scan/") ||
			strings.HasPrefix(path, "/api/barcode/") {
			c.Next()
			return
		}

		role, _ := c.Get("role")
		roleStr, _ := role.(string)
		if roleStr == "" {
			roleStr = "anonymous"
		}
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
