package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Admin 站长校验：必须挂在 Auth 之后。
// 角色变更即时生效——Auth 中间件每请求都查库拿最新用户，这里的 Role 就是库里的现值。
func Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := UserFrom(c)
		if u == nil || u.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "需要站长权限"})
			return
		}
		c.Next()
	}
}
