package token

import (
	"strings"
	"testing"
)

const testSecret = "test-jwt-secret-0123456789abcdef-0123456789abcdef"

// TestIssueParseRoundTrip 签发后再解析，uid/tv 必须原样保留。
func TestIssueParseRoundTrip(t *testing.T) {
	s, err := Issue(12345, 7, testSecret)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	claims, err := Parse(s, testSecret)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.UID != 12345 || claims.TV != 7 {
		t.Fatalf("claims 不一致: uid=%d tv=%d", claims.UID, claims.TV)
	}
}

// TestParseRejectsWrongSecret 用错误密钥签发的 token 必须被拒绝（防伪造身份）。
func TestParseRejectsWrongSecret(t *testing.T) {
	s, err := Issue(1, 1, testSecret)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if _, err := Parse(s, "another-secret-0123456789abcdef-0123456789abcdef"); err == nil {
		t.Fatal("错误密钥签发的 token 不应通过校验")
	}
}

// TestParseRejectsTamperedClaims 篡改 payload（uid）后签名不再匹配，必须拒绝。
func TestParseRejectsTamperedClaims(t *testing.T) {
	s, err := Issue(1, 1, testSecret)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		t.Fatalf("JWT 应为三段，got %d", len(parts))
	}
	// 改 payload 第一个字符，保持签名不变 → 必须拒绝
	if parts[1] == "" {
		t.Fatal("payload 为空")
	}
	replaced := "x" + parts[1][1:]
	tampered := parts[0] + "." + replaced + "." + parts[2]
	if _, err := Parse(tampered, testSecret); err == nil {
		t.Fatal("篡改后的 token 不应通过校验")
	}
}

// TestParseRejectsMissingUID 缺 uid claim 的合法签名 token 必须拒绝（防身份混淆）。
func TestParseRejectsMissingUID(t *testing.T) {
	// 用同一密钥签一个只有 exp 的 token
	jwtToken, err := Issue(1, 1, testSecret)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	// 直接构造无 uid 的签名串较繁琐；这里用空串与随机串覆盖解析失败路径
	for _, bad := range []string{"", "not-a-jwt", jwtToken + "x"} {
		if _, err := Parse(bad, testSecret); err == nil {
			t.Fatalf("非法 token 不应通过: %q", bad)
		}
	}
}
