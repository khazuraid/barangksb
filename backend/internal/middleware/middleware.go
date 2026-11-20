package middleware

import (
	"net/http"
	"path"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

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
	if strings.HasPrefix(p, "/api/barcode/") &&
		!strings.HasPrefix(p, "/api/barcode/templates") &&
		!strings.HasPrefix(p, "/api/barcode/units") {
		return true
	}
	return false
}

func AuthMiddleware(secret string, pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := path.Clean(c.Request.URL.Path)
		if isPublicRoute(p) {
			c.Next()
			return
		}

		tokenStr := ""
		header := c.GetHeader("Authorization")
		if strings.HasPrefix(header, "Bearer ") {
			tokenStr = strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
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

		role := claims.Role
		name := claims.Name
		email := claims.Email
		if pool != nil {
			var dbRole, dbName, dbEmail string
			var dbTokenVersion int
			err := pool.QueryRow(c,
				`SELECT email, name, role, COALESCE(token_version, 1) FROM users WHERE id::text = $1`,
				claims.UserID,
			).Scan(&dbEmail, &dbName, &dbRole, &dbTokenVersion)
			if err != nil || claims.TokenVersion != dbTokenVersion {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "sesi tidak valid atau telah dicabut, silakan login ulang"})
				c.Abort()
				return
			}
			role = dbRole
			name = dbName
			email = dbEmail
		}

		c.Set("uid", claims.UserID)
		c.Set("email", email)
		c.Set("role", role)
		c.Set("name", name)
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
