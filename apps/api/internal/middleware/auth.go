package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/manage-money/api/internal/auth"
	redisstore "github.com/manage-money/api/internal/redis"
	"github.com/manage-money/api/internal/response"
)

const UserIDKey = "userID"
const UserEmailKey = "userEmail"
const UserRoleKey = "userRole"

func Auth(jwtSvc *auth.Service, redis *redisstore.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			response.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Yêu cầu đăng nhập")
			c.Abort()
			return
		}
		claims, err := jwtSvc.ParseAccess(token)
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Phiên đăng nhập không hợp lệ")
			c.Abort()
			return
		}
		c.Set(UserIDKey, claims.UserID)
		c.Set(UserEmailKey, claims.Email)
		c.Set(UserRoleKey, claims.Role)
		c.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		r, _ := c.Get(UserRoleKey)
		if r != role {
			response.Fail(c, http.StatusForbidden, "FORBIDDEN", "Không có quyền truy cập")
			c.Abort()
			return
		}
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	if cookie, err := c.Cookie(auth.CookieAccess); err == nil && cookie != "" {
		return cookie
	}
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return ""
}
