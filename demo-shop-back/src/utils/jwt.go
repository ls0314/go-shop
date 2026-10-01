package utils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"demo-shop-back/src/model"

	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	TokenType string `json:"tokenType"`
	jwt.RegisteredClaims
}

// JWTVerifier 只做验签。签发能力在 user-service,本进程不持有私钥。
type JWTVerifier struct {
	PublicKey *rsa.PublicKey
}

// LoadPublicKey 从 PEM 文件读取 RSA 公钥。
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取 JWT 公钥失败: %w", err)
	}

	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, errors.New("JWT 公钥不是合法的 PEM")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析 JWT 公钥失败: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("JWT 公钥不是 RSA 类型")
	}
	return rsaPub, nil
}

func NewJWTVerifier(publicKey *rsa.PublicKey) *JWTVerifier {
	return &JWTVerifier{PublicKey: publicKey}
}

func (j *JWTVerifier) parse(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return j.PublicKey, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, model.TokenExpired
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, model.TokenNotValidYet
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, model.TokenMalformed
		}
		return nil, model.TokenInvalid
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, model.TokenInvalid
}

// ParseAccessToken 校验访问令牌,并确认令牌类型为 access。
// 类型校验不可省:refresh 令牌有效期长得多,放行等于绕过短有效期设计。
func (j *JWTVerifier) ParseAccessToken(tokenString string) (*CustomClaims, error) {
	claims, err := j.parse(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != "access" {
		return nil, model.TokenInvalid
	}
	return claims, nil
}
