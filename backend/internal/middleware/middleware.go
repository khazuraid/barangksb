package middleware

import (
	"net/http"
	"path"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"

	"inventariskantor/internal/auth"
)

// isPublicRoute checks whether a path bypasses authentication and RBAC.
// Barcode template management requires authentication and RBAC.
func isPublicRoute(p string) bool {
	if p == "/api/auth/login" ||
		strings.HasPrefix(p, "/api/scan/") ||
		strings.HasPrefix(p, "/api/public/") {
		return true
	}
	if strings.HasPrefix(p, "/api/barcode/") && !strings.HasPrefix(p, "/api/barcode/templates") {
		return true
	}
	return false
}

func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := path.Clean(c.Request.URL.Path)
		if isPublicRoute(p) {
			c.Next()
			return
		}

		tokenStr := ""
		header := c.GetHeader("Authorization")
		if strings.HasPrefix(header, "Bearer ") {
			tokenStr = strings.TrimPrefix(header, "Bearer ")
		} else if q := c.Query("token"); q != "" {
			tokenStr = q
		}

		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
			c.Abort()
			return
		}
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
		p := path.Clean(c.Request.URL.Path)
		if isPublicRoute(p) {
			c.Next()
			return
		}

		role, _ := c.Get("role")
		roleStr, _ := role.(string)
		if roleStr == "" {
			roleStr = "anonymous"
		}
		method := c.Request.Method

		allowed, err := enforcer.Enforce(roleStr, p, method)
		if err != nil || !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}
