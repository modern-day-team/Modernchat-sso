package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/modern-day-team/Modernchat-sso/server/internal/store"
)

// Maintenance 维护模式门禁：非站长的公开业务端点 503，登录/登出与站长接口豁免
// （站长需在维护期间能登录进来关掉维护，主仓库同款设计）。
func Maintenance(settings *store.SettingStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		if settings == nil {
			c.Next()
			return
		}
		ops := settings.Ops()
		if !ops.MaintenanceMode {
			c.Next()
			return
		}
		u := UserFrom(c)
		if u != nil && u.Role == "admin" {
			c.Next()
			return
		}
		msg := ops.MaintenanceMessage
		if msg == "" {
			msg = "站点维护中，请稍后再试"
		}
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": msg})
	}
}
