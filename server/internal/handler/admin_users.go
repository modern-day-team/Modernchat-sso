package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/modern-day-team/Modernchat-sso/server/internal/model"
)

// AdminUsersHandler 站长账号管理（列表/封禁/解封/重置密码/改角色）。
// 仅站长可用（router 的 Admin 中间件）；SSO 暂无审计表，操作以 slog 记录。
type AdminUsersHandler struct {
	admins AdminUserService
}

// AdminUserService 消费方定义的窄接口（由 service.AdminService 实现）。
type AdminUserService interface {
	List(page, pageSize int, kw, role, status string) ([]model.User, int64, error)
	Ban(id uint, days int, reason string) error
	Unban(id uint) error
	ResetPassword(id uint) (newPassword string, err error)
	SetRole(id uint, role string) error
}

// NewAdminUsersHandler 构造。
func NewAdminUsersHandler(admins AdminUserService) *AdminUsersHandler {
	return &AdminUsersHandler{admins: admins}
}

// List GET /api/v1/admin/users?page=&pageSize=&kw=&role=&status=
func (h *AdminUsersHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	users, total, err := h.admins.List(page, pageSize, c.Query("kw"), c.Query("role"), c.Query("status"))
	if err != nil {
		slog.Error("用户列表查询失败", "err", err)
		fail(c, http.StatusInternalServerError, "服务器开小差了，请稍后再试")
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users, "total": total, "page": page, "pageSize": pageSize})
}

// parseUserID 路径参数 :id 解析（0 或非数字一律不合法）。
func parseUserID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		fail(c, http.StatusBadRequest, "用户 id 不合法")
		return 0, false
	}
	return uint(id), true
}

// Ban POST /api/v1/admin/users/:id/ban
func (h *AdminUsersHandler) Ban(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		return
	}
	var req struct {
		Days   int    `json:"days"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Days = -1 // 触发天数校验错误
	}
	if err := h.admins.Ban(id, req.Days, req.Reason); err != nil {
		mapErr(c, err)
		return
	}
	slog.Warn("站长封禁用户", "admin", adminName(c), "ip", c.ClientIP(), "target_id", id, "days", req.Days)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Unban POST /api/v1/admin/users/:id/unban
func (h *AdminUsersHandler) Unban(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		return
	}
	if err := h.admins.Unban(id); err != nil {
		mapErr(c, err)
		return
	}
	slog.Warn("站长解封用户", "admin", adminName(c), "ip", c.ClientIP(), "target_id", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ResetPassword POST /api/v1/admin/users/:id/reset-password
func (h *AdminUsersHandler) ResetPassword(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		return
	}
	newPassword, err := h.admins.ResetPassword(id)
	if err != nil {
		mapErr(c, err)
		return
	}
	slog.Warn("站长重置用户密码", "admin", adminName(c), "ip", c.ClientIP(), "target_id", id)
	c.JSON(http.StatusOK, gin.H{"newPassword": newPassword})
}

// SetRole POST /api/v1/admin/users/:id/role
func (h *AdminUsersHandler) SetRole(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		return
	}
	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "role 只能是 admin 或 user")
		return
	}
	if err := h.admins.SetRole(id, req.Role); err != nil {
		mapErr(c, err)
		return
	}
	slog.Warn("站长变更用户角色", "admin", adminName(c), "ip", c.ClientIP(), "target_id", id, "role", req.Role)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
