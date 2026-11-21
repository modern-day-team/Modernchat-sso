// 邮件发送层：SSO 的邮箱验证会接入不同厂商。
// 业务只依赖 Mailer 接口；新增厂商 = 加一个实现 + NewMailer 工厂加一行，业务代码零改动。
// 通道：console（默认，仅本地开发：内容打日志、不发真信）| smtp（465 隐式 TLS / 587 STARTTLS）；
// 厂商 API（阿里邮件推送、SendGrid、Resend 等）预留同一接口接入。
package service

import (
	"context"
	"log/slog"

	"github.com/modern-day-team/Modernchat-sso/server/internal/config"
)

// MailMessage 一封纯文本邮件（SSO 的验证码/重置链接都是纯文本；HTML 需要时再扩展）。
type MailMessage struct {
	To      string
	Subject string
	Text    string
}

// Mailer 邮件发送抽象。
type Mailer interface {
	Send(ctx context.Context, msg MailMessage) error
}

// NewMailer 按配置选通道；配置不完整时退回 console（并告警），保证服务能起。
// ⚠️ 生产误配会静默退化为不发真信：靠启动日志的告警暴露，务必核对。
func NewMailer(cfg config.Config) Mailer {
	switch cfg.MailProvider {
	case "smtp":
		if cfg.SMTPHost == "" {
			slog.Warn("MAIL_PROVIDER=smtp 但 SMTP_HOST 为空，退回 console（不发真信）")
			return NewConsoleMailer()
		}
		return NewSMTPMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)
	case "tencent_ses":
		m, err := NewTencentSESMailer(TencentSESConfig{
			SecretID:   cfg.TencentSESSecretID,
			SecretKey:  cfg.TencentSESSecretKey,
			Region:     cfg.TencentSESRegion,
			From:       cfg.TencentSESFrom,
			TemplateID: cfg.TencentSESTemplateID,
		})
		if err != nil {
			slog.Error("腾讯云 SES 配置不完整，退回 console（不发真信）", "err", err)
			return NewConsoleMailer()
		}
		return m
	default:
		return NewConsoleMailer()
	}
}
