// Package token JWT 签发与解析：口径与主仓库 Modernchat-new 完全一致
// （HS256，claims 仅 uid + tv），两个服务共享 JWT_SECRET 即可互相校验，实现平滑接管。
// 本包是叶子依赖：不 import 项目内其他包。
package token

import (
	"errors"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

// TTL token 有效期（与主仓库一致：7 天）。
const TTL = 7 * 24 * time.Hour

// AuthCookieName 与主仓库同名：浏览器 <img>/WebSocket 握手带不了 Authorization 头，
// 靠同源 cookie 兜底。
const AuthCookieName = "mc_auth"

// Claims 解析出的身份。
type Claims struct {
	UID uint
	TV  int64 // token_version
}

// Issue 签发 HS256 JWT（claims: uid, tv）。
func Issue(uid uint, tv int64, secret string) (string, error) {
	now := time.Now()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": uid,
		"tv":  tv,
		"iat": now.Unix(),
		"exp": now.Add(TTL).Unix(),
	})
	return t.SignedString([]byte(secret))
}

// Parse 校验并解析 JWT（HTTP 认证与未来的服务间校验共用）。
func Parse(tokenStr, secret string) (*Claims, error) {
	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !tok.Valid {
		return nil, errors.New("token 无效")
	}
	mc, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("claims 类型异常")
	}
	uidF, ok := mc["uid"].(float64)
	if !ok {
		return nil, errors.New("claims 缺少 uid")
	}
	tvF, _ := mc["tv"].(float64)
	return &Claims{UID: uint(uidF), TV: int64(tvF)}, nil
}
