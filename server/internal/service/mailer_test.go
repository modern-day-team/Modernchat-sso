package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/modern-day-team/Modernchat-sso/server/internal/config"
)

func configForMailerTest(provider string) config.Config {
	return config.Config{MailProvider: provider}
}

// fakeMailer 可脚本化失败的假通道（不依赖 DB/网络，纯单测）。
type fakeMailer struct{ err error }

func (f *fakeMailer) Send(_ context.Context, _ MailMessage) error { return f.err }

func TestFailoverMailer_DegradesToNextProvider(t *testing.T) {
	m := &failoverMailer{chain: []Mailer{
		&namedMailer{name: "ses", inner: &fakeMailer{err: errors.New("限流")}},
		&namedMailer{name: "smtp", inner: &fakeMailer{}},
	}}
	if err := m.Send(context.Background(), MailMessage{To: "a@b.c", Subject: "s", Text: "t"}); err != nil {
		t.Fatalf("主通道失败应降级到备用通道，got %v", err)
	}
}

func TestFailoverMailer_AllProvidersFail(t *testing.T) {
	m := &failoverMailer{chain: []Mailer{
		&namedMailer{name: "ses", inner: &fakeMailer{err: errors.New("限流")}},
		&namedMailer{name: "smtp", inner: &fakeMailer{err: errors.New("超时")}},
	}}
	err := m.Send(context.Background(), MailMessage{To: "a@b.c"})
	if err == nil {
		t.Fatal("全部失败应返回错误")
	}
	// 聚合错误要能定位到每个厂商与其失败原因
	for _, want := range []string{"ses", "smtp", "限流", "超时"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("聚合错误应包含 %q，got %v", want, err)
		}
	}
}

func TestNewMailer_UnknownChannelSkipped(t *testing.T) {
	cfg := configForMailerTest("bogus,tencent_ses") // tencent_ses 配置缺失也跳过 → 全空
	m := NewMailer(cfg)
	if _, ok := m.(*ConsoleMailer); !ok {
		t.Fatalf("全部通道不可用应退回 console，got %T", m)
	}
}

func TestNewMailer_SingleChannelNoWrapper(t *testing.T) {
	cfg := configForMailerTest("console")
	m := NewMailer(cfg)
	if _, ok := m.(*failoverMailer); ok {
		t.Fatal("单通道不应包 failoverMailer")
	}
}
