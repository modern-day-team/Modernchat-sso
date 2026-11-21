// Package router 组合根：装配 store → service → handler/middleware。
// 依赖方向单向：handler/middleware →（窄接口）→ service → store；本包是唯一知道全部具体类型的地方。
package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/modern-day-team/Modernchat-sso/server/internal/config"
	"github.com/modern-day-team/Modernchat-sso/server/internal/handler"
	"github.com/modern-day-team/Modernchat-sso/server/internal/middleware"
	"github.com/modern-day-team/Modernchat-sso/server/internal/service"
	"github.com/modern-day-team/Modernchat-sso/server/internal/store"
)

// Setup 装配全部路由。
func Setup(cfg config.Config, users *store.UserStore) *gin.Engine {
	r := gin.Default()

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	accounts := service.NewAccountService(users, cfg.JWTSecret)
	authH := handler.NewAuthHandler(accounts, cfg.CookieSecure)
	authMw := middleware.Auth(accounts) // *service.AccountService 满足窄接口 Authenticator

	v1 := r.Group("/api/v1")
	// TODO(限流)：注册/登录挂 IP 级限流 + 账号级失败锁定（参照主仓库 middleware/limiter.go 的 go-redis 方案）；
	// 落地时做成独立中间件，经窄接口注入，不侵入 handler/service
	v1.POST("/users", authH.Register) // 注册
	v1.POST("/sessions", authH.Login) // 登录（创建会话）

	authed := v1.Group("", authMw)
	authed.GET("/users/me", authH.Me)                // 当前用户
	authed.DELETE("/sessions/current", authH.Logout) // 登出（吊销全部旧 token）

	return r
}
