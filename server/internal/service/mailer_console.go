package service

import (
	"context"
	"log/slog"
)

// ConsoleMailer 本地开发通道：内容打进日志，不发真信。
// ⚠️ Enabled()=false（不具备真实送达能力）：生产必须 MAIL_PROVIDER=smtp 或厂商 API；
// 本地联调邮箱验证建议起本地 SMTP 捕获器（如 MailHog，SMTP_HOST=localhost SMTP_PORT=1025）。
type ConsoleMailer struct{}

// NewConsoleMailer 构造。
func NewConsoleMailer() *ConsoleMailer { return &ConsoleMailer{} }

// Enabled console 通道不具备真实送达能力。
func (m *ConsoleMailer) Enabled() bool { return false }

func (m *ConsoleMailer) Send(_ context.Context, msg MailMessage) error {
	slog.Info("MAIL(console，未真实发送)", "to", msg.To, "subject", msg.Subject, "body", msg.Text)
	return nil
}
