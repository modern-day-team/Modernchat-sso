package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/modern-day-team/Modernchat-sso/server/internal/middleware"
	"github.com/modern-day-team/Modernchat-sso/server/internal/service"
)

// UpdateHandler 站长侧的在线更新（状态轮询 / 检查 / 启动）。
// 仅站长可用（router 的 Admin 中间件）；部署方式不支持时统一 400 + 原因。
// SSO 暂无审计表：操作以 slog 记录（管理员用户名 + IP）。
type UpdateHandler struct {
	upd *service.Updater
}

// NewUpdateHandler 构造。upd 可为 nil（WEB_UPDATE=false），此时 Status 报不可用、
// Check/Start 返回 400。
func NewUpdateHandler(upd *service.Updater) *UpdateHandler {
	return &UpdateHandler{upd: upd}
}

// adminName 当前管理员（日志用）。
func adminName(c *gin.Context) string {
	if u := middleware.UserFrom(c); u != nil {
		return u.Username
	}
	return ""
}

// ensureAvailable 部署方式不支持时统一 400。
func (h *UpdateHandler) ensureAvailable(c *gin.Context) bool {
	if h.upd == nil || !h.upd.Available() {
		fail(c, http.StatusBadRequest, "当前部署方式不支持在线更新（需要 systemd 模板单元部署，见 deploy/；Docker / 裸机请手动更新）")
		return false
	}
	return true
}

// Status GET /api/v1/admin/update：状态快照与最近日志（前端 2s 轮询）。
func (h *UpdateHandler) Status(c *gin.Context) {
	if h.upd == nil {
		c.JSON(http.StatusOK, service.StatusView{Available: false})
		return
	}
	c.JSON(http.StatusOK, h.upd.Status())
}

// Check POST /api/v1/admin/update/checks：检查是否有新版本（查询 Release API）。
func (h *UpdateHandler) Check(c *gin.Context) {
	if !h.ensureAvailable(c) {
		return
	}
	res, err := h.upd.Check()
	if err != nil {
		slog.Warn("站长检查更新失败", "admin", adminName(c), "ip", c.ClientIP(), "err", err.Error())
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	slog.Info("站长检查更新", "admin", adminName(c), "ip", c.ClientIP(),
		"upToDate", res.UpToDate, "remote", res.RemoteTag)
	c.JSON(http.StatusOK, res)
}

// Start POST /api/v1/admin/update：启动后台更新任务（下载→校验→替换→重启）。
func (h *UpdateHandler) Start(c *gin.Context) {
	if !h.ensureAvailable(c) {
		return
	}
	if !h.upd.Start() {
		fail(c, http.StatusConflict, "有更新任务正在进行中")
		return
	}
	slog.Info("站长发起在线更新", "admin", adminName(c), "ip", c.ClientIP())
	c.JSON(http.StatusAccepted, h.upd.Status())
}
