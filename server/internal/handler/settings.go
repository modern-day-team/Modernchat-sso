package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/modern-day-team/Modernchat-sso/server/internal/store"
)

// SettingsHandler 站点开关：公开读取（前端渲染依据）+ 站长修改。
type SettingsHandler struct {
	settings *store.SettingStore
}

// NewSettingsHandler 构造。settings 可为 nil（启动异常时按默认值运行，公开读取兜底）。
func NewSettingsHandler(settings *store.SettingStore) *SettingsHandler {
	return &SettingsHandler{settings: settings}
}

// opsView GET /api/v1/settings 的公开响应（只暴露前端需要的开关，不泄内部配置）。
type opsView struct {
	AllowRegister      bool `json:"allowRegister"`
	RequireEmailVerify bool `json:"requireEmailVerify"`
}

// Public 公开读取站点开关。
func (h *SettingsHandler) Public(c *gin.Context) {
	ops := store.DefaultOps()
	if h.settings != nil {
		ops = h.settings.Ops()
	}
	c.JSON(http.StatusOK, opsView{
		AllowRegister:      ops.AllowRegister,
		RequireEmailVerify: ops.RequireEmailVerify,
	})
}

// updateReq PUT /api/v1/admin/settings 的请求体。
type updateReq struct {
	AllowRegister      *bool   `json:"allowRegister"`
	RequireEmailVerify *bool   `json:"requireEmailVerify"`
	RequireInviteCode  *bool   `json:"requireInviteCode"`
	MaintenanceMode    *bool   `json:"maintenanceMode"`
	MaintenanceMessage *string `json:"maintenanceMessage"`
	LoginMaxAttempts   *int    `json:"loginMaxAttempts"`
	LoginLockMinutes   *int    `json:"loginLockMinutes"`
}

// Update 站长修改站点开关（nil 字段保持原值）。
func (h *SettingsHandler) Update(c *gin.Context) {
	var req updateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数不合法")
		return
	}
	if h.settings == nil {
		fail(c, http.StatusInternalServerError, "服务器开小差了，请稍后再试")
		return
	}
	ops := h.settings.Ops()
	if req.AllowRegister != nil {
		ops.AllowRegister = *req.AllowRegister
	}
	if req.RequireEmailVerify != nil {
		ops.RequireEmailVerify = *req.RequireEmailVerify
	}
	if req.RequireInviteCode != nil {
		ops.RequireInviteCode = *req.RequireInviteCode
	}
	if req.MaintenanceMode != nil {
		ops.MaintenanceMode = *req.MaintenanceMode
	}
	if req.MaintenanceMessage != nil {
		ops.MaintenanceMessage = *req.MaintenanceMessage
	}
	if req.LoginMaxAttempts != nil {
		ops.LoginMaxAttempts = *req.LoginMaxAttempts
	}
	if req.LoginLockMinutes != nil {
		ops.LoginLockMinutes = *req.LoginLockMinutes
	}
	if err := h.settings.UpdateOps(ops); err != nil {
		fail(c, http.StatusInternalServerError, "服务器开小差了，请稍后再试")
		return
	}
	c.JSON(http.StatusOK, opsView{
		AllowRegister:      ops.AllowRegister,
		RequireEmailVerify: ops.RequireEmailVerify,
	})
}
