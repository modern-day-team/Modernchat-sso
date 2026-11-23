// Package router 组合根：装配 db → store → service → handler/middleware。
// 依赖方向单向：handler/middleware →（窄接口）→ service → store；本包是唯一知道全部具体类型的地方。
package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/modern-day-team/Modernchat-sso/server/internal/config"
	"github.com/modern-day-team/Modernchat-sso/server/internal/handler"
	"github.com/modern-day-team/Modernchat-sso/server/internal/middleware"
	"github.com/modern-day-team/Modernchat-sso/server/internal/service"
	"github.com/modern-day-team/Modernchat-sso/server/internal/store"
	"github.com/modern-day-team/Modernchat-sso/server/internal/version"
)

// Setup 装配全部路由。upd 为在线更新服务（WEB_UPDATE=false 时传 nil，接口报不可用）。
func Setup(cfg config.Config, db *gorm.DB, upd *service.Updater) *gin.Engine {
	r := gin.Default()

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 版本信息（公开）：更新流程的前置自检与运维排查用
	r.GET("/api/v1/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, version.Current())
	})

	// stores
	users := store.NewUserStore(db)
	settings := store.NewSettingStore(db)
	emailCodes := store.NewEmailCodeStore(db, cfg.JWTSecret)
	loginAttempts := store.NewLoginAttemptStore(db)
	resets := store.NewPasswordResetStore(db)

	settingsH := handler.NewSettingsHandler(settings)

	// services
	mailer := service.NewMailer(cfg)
	accounts := service.NewAccountService(users, emailCodes, mailer, loginAttempts, settings, resets, cfg.PublicURL, cfg.JWTSecret)
	authH := handler.NewAuthHandler(accounts, cfg.CookieSecure)
	authMw := middleware.Auth(accounts) // *service.AccountService 满足窄接口 Authenticator

	// 限流：认证类公开端点 IP 10/min（主仓库同款；多实例共享配额可换 Redis 实现）
	authLimit := middleware.RateLimit(middleware.NewLimiterStore(10, 10), middleware.KeyByIP)
	// 维护模式门禁：登录/登出/站长接口豁免
	maint := middleware.Maintenance(settings)

	v1 := r.Group("/api/v1")
	// 公开站点开关：前端渲染依据
	v1.GET("/settings", settingsH.Public)

	v1.POST("/users", authLimit, maint, authH.Register)               // 注册
	v1.POST("/sessions", authLimit, authH.Login)                      // 登录（创建会话；豁免维护）
	v1.POST("/email-codes", authLimit, maint, authH.SendEmailCode)    // 发送邮箱验证码
	v1.POST("/password-resets", authLimit, maint, authH.Forgot)       // 忘记密码
	v1.POST("/password-resets/:token", authLimit, maint, authH.Reset) // 重置密码

	authed := v1.Group("", authMw)
	authed.GET("/users/me", authH.Me)                               // 当前用户
	authed.PATCH("/users/me/password", maint, authH.ChangePassword) // 自身改密
	authed.DELETE("/sessions/current", authH.Logout)                // 登出（吊销全部旧 token；豁免维护）

	// 站长后台：Auth + Admin（角色每请求查库，变更即时生效）
	admin := v1.Group("/admin", authMw, middleware.Admin())
	admin.GET("/settings", settingsH.AdminGet) // 读取完整站点开关（后台表单回填）
	admin.PUT("/settings", settingsH.Update)   // 修改站点开关
	adminUsersH := handler.NewAdminUsersHandler(service.NewAdminService(users, loginAttempts))
	admin.GET("/users", adminUsersH.List)                              // 用户分页列表
	admin.POST("/users/:id/ban", adminUsersH.Ban)                      // 封禁（tv+1）
	admin.POST("/users/:id/unban", adminUsersH.Unban)                  // 解封
	admin.POST("/users/:id/reset-password", adminUsersH.ResetPassword) // 重置密码（返回新密码）
	admin.POST("/users/:id/role", adminUsersH.SetRole)                 // 改角色（最后站长保护）
	updateH := handler.NewUpdateHandler(upd)
	admin.GET("/update", updateH.Status)        // 状态与最近日志（前端轮询）
	admin.POST("/update/checks", updateH.Check) // 检查新版本
	admin.POST("/update", updateH.Start)        // 发起更新任务

	return r
}
