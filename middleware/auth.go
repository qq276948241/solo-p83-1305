package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"groupbuy/config"
)

func TokenAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "缺少 Authorization 头"})
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization 格式错误，需 Bearer <token>"})
			return
		}
		if parts[1] != cfg.AuthToken {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "token 无效"})
			return
		}
		c.Next()
	}
}
