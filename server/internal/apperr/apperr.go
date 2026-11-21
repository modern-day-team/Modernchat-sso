// Package apperr 领域错误哨兵：各层共用的「错误语言」。
// 有了它，handler/middleware 只认错误语义、不依赖 service 具体实现，
// 即可各自映射 HTTP 状态码，避免业务层与传输层横向耦合。
package apperr

import "errors"

var (
	ErrInvalidInput       = errors.New("参数不合法")
	ErrInvalidEmail       = errors.New("邮箱格式不正确")
	ErrUsernameTaken      = errors.New("用户名已存在")
	ErrEmailTaken         = errors.New("邮箱已被使用")
	ErrInvalidCredentials = errors.New("用户名或密码错误")
	ErrInvalidToken       = errors.New("未登录或登录已过期")
	ErrBanned             = errors.New("账号封禁中")

	// 站点开关与邮箱验证（阶段 1）
	ErrRegisterClosed    = errors.New("站点已关闭注册，请联系管理员")
	ErrEmailRequired     = errors.New("请填写邮箱")
	ErrEmailCodeRequired = errors.New("请填写邮箱验证码")
	ErrEmailCodeInvalid  = errors.New("验证码错误或已过期，请重新获取")
	ErrMailNotReady      = errors.New("站点尚未配置邮件服务，暂时无法发送验证码")
	ErrMailSendFailed    = errors.New("验证邮件发送失败，请稍后重试")
	ErrAccountLocked     = errors.New("账号已锁定") // 消息由 service 拼接剩余分钟数
)
