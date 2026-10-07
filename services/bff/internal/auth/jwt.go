// Package auth 提供访问令牌的**验签**能力(RS256 公钥)。
//
// ============================================================
// 为什么在 BFF 内部,而不是独立的共享包
// ============================================================
//
// 拆分期间它曾是 `pkg/auth`,由 BFF 与单体**共用** —— 因为过渡期同一时间
// 两处都要验签,而验签的正确性依赖三件事完全一致:公钥如何解析、
// Claims 的字段名、access/refresh 如何区分。任何一处漂移的后果不是编译
// 错误,而是"某些令牌在 BFF 通过、在单体被拒"(或反之),表现为随机的
// 401,极难排查。
//
// **单体退役后这个理由消失了**:全仓只有 BFF 验签,于是并回 BFF 内部
// (正是当时包注释里写好的计划)。
//
// ============================================================
// 仍然存在的跨服务契约:字段名必须与签发方一致
// ============================================================
//
// 签发在 user-service(它持私钥),本包只验签。CustomClaims 的字段名
// (user_id / username / tokenType)是 **JSON 键**,改这里等于改跨服务
// 契约,会让所有在途令牌失效。
//
// 所以本包**不是**可以随意重命名的内部实现 —— 它与
// services/user/internal/utils/jwt.go 之间有一条隐式约定。
//
// ============================================================
// 错误在这里自定义,不依赖任何业务 model
// ============================================================
//
// 调用方(HTTP 中间件)用 errors.Is 把它们翻成不同提示语:
// 「已过期」要引导用户重新登录,而「格式错误」不该给这个提示,
// 否则用户会白跑一次登录流程。
package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

// 验签失败的几种原因
var (
	// ErrTokenExpired 令牌已过期
	ErrTokenExpired = errors.New("token已过期")
	// ErrTokenNotValidYet 令牌尚未生效(签发时间在未来)
	ErrTokenNotValidYet = errors.New("token尚未生效")
	// ErrTokenMalformed 格式错误(不是合法 JWT,或签名算法不对)
	ErrTokenMalformed = errors.New("token格式错误")
	// ErrTokenInvalid 签名验证失败或 Claims 不合法
	ErrTokenInvalid = errors.New("token无效")
	// ErrWrongTokenType 令牌类型不对(用 refresh 令牌访问业务接口)
	ErrWrongTokenType = errors.New("令牌类型错误")
)

// TokenTypeAccess / TokenTypeRefresh 两种令牌的类型标记。
//
// 为什么要区分:refresh 令牌的有效期长得多(现状 24h),
// 若它能直接访问业务接口,等于把长有效期的凭据当短的有效期用 ——
// 而这正是"access 30min / refresh 24h"这个设计要避免的。
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// CustomClaims 访问令牌的载荷。
//
// 字段名必须与签发方(user-service)逐字一致,见包注释。
type CustomClaims struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	TokenType string `json:"tokenType"`
	jwt.RegisteredClaims
}

// Verifier 用公钥验签。
//
// 只有公钥 —— 签发(需要私钥)在 user-service 手里。BFF 能验签但签不出来,
// 这是刻意的:私钥分散到多个服务等于把"谁能冒充任何人"这个问题变成
// "哪个服务的机器被攻破"。
type Verifier struct {
	publicKey *rsa.PublicKey
}

// LoadPublicKey 从 PEM 文件读公钥。
//
// 支持两种 PEM 块:PKIX("PUBLIC KEY",openssl 的现代默认)与
// PKCS1("RSA PUBLIC KEY",老格式)。两种都认是因为密钥文件是运维产的,
// 不该因为 openssl 版本不同就启动失败。
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取公钥文件失败: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("公钥文件不是合法的 PEM 格式")
	}

	if pub, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		rsaPub, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("公钥不是 RSA 类型(本项目的令牌用 RS256)")
		}
		return rsaPub, nil
	}

	rsaPub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析公钥失败(已尝试 PKIX 与 PKCS1): %w", err)
	}
	return rsaPub, nil
}

// NewVerifier 用已解析的公钥构造验签器
func NewVerifier(publicKey *rsa.PublicKey) *Verifier {
	return &Verifier{publicKey: publicKey}
}

// NewVerifierFromFile 从 PEM 文件构造验签器(启动时用)
func NewVerifierFromFile(path string) (*Verifier, error) {
	pub, err := LoadPublicKey(path)
	if err != nil {
		return nil, err
	}
	return NewVerifier(pub), nil
}

// parse 解析并验签,返回 Claims。
//
// **必须校验签名算法**:不校验的话,攻击者可以把 alg 改成 "none"
// 或换成 HMAC(用公钥当对称密钥),从而伪造任意令牌。
// 这是 JWT 最经典的一个坑,故这里显式限定 RS256。
func (v *Verifier) parse(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("非预期的签名算法: %v", token.Header["alg"])
			}
			return v.publicKey, nil
		},
		jwt.WithValidMethods([]string{"RS256"}),
	)
	if err != nil {
		// 把 jwt 库的错误翻成本包的哨兵。顺序重要:
		// 「已过期」与「尚未生效」都要先判,否则会被笼统的 ErrTokenInvalid 吃掉,
		// 而前端就分不出"该重新登录"与"时钟不对"
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, ErrTokenExpired
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			return nil, ErrTokenNotValidYet
		case errors.Is(err, jwt.ErrTokenMalformed):
			return nil, ErrTokenMalformed
		default:
			return nil, ErrTokenInvalid
		}
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}

// ParseAccessToken 解析**访问令牌**。
//
// 用 refresh 令牌访问业务接口会返回 ErrWrongTokenType ——
// 这不是"令牌无效",而是"令牌用错了地方",两者的处置不同
// (后者该提示前端去刷新,而不是让用户重新登录)。
func (v *Verifier) ParseAccessToken(tokenString string) (*CustomClaims, error) {
	claims, err := v.parse(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != TokenTypeAccess {
		return nil, ErrWrongTokenType
	}
	return claims, nil
}
