package service

import (
	"context"
	"log/slog"
)

// ConsoleMailer 本地开发通道：内容打进日志，不发真信。
// ⚠️ 仅限开发环境（日志里能看到验证码正是它的用途）；生产必须 MAIL_PROVIDER=smtp 或厂商 API。
type ConsoleMailer struct{}

// NewConsoleMailer 构造。
func NewConsoleMailer() *ConsoleMailer { return &ConsoleMailer{} }

func (m *ConsoleMailer) Send(_ context.Context, msg MailMessage) error {
	slog.Info("MAIL(console，未真实发送)", "to", msg.To, "subject", msg.Subject, "body", msg.Text)
	return nil
}
