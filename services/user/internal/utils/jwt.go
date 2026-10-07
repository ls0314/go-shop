package utils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"demo-shop/services/user/internal/model"

	"github.com/golang-jwt/jwt/v5"
)

const (
	accessTokenTTL  = 30 * time.Minute
	refreshTokenTTL = 24 * time.Hour
)

// CustomClaims 是 access / refresh 令牌的载荷。
//
// 字段名与 json tag **都必须保持稳定**:BFF 与各服务只持公钥做验签,
// 它们按这些名字读值,改名等于让所有在途令牌失效。
type CustomClaims struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	TokenType string `json:"tokenType"`
	jwt.RegisteredClaims
}

// JWTIssuer 签发 access/refresh 令牌,并校验 refresh 令牌。
// 私钥只在本服务加载;其他进程只持公钥做验签。
type JWTIssuer struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取 JWT 私钥失败: %w", err)
	}

	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, errors.New("JWT 私钥不是合法的 PEM")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析 JWT 私钥失败: %w", err)
	}
	return key, nil
}

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

// NewJWTIssuer 加载私钥,并由私钥导出公钥。
// 公钥不必单独配置:它可以从私钥推导,少一处配置不一致的可能。
func NewJWTIssuer(privateKeyPath string) (*JWTIssuer, error) {
	priv, err := LoadPrivateKey(privateKeyPath)
	if err != nil {
		return nil, err
	}
	return &JWTIssuer{
		privateKey: priv,
		publicKey:  &priv.PublicKey,
	}, nil
}

func (j *JWTIssuer) GenerateAccessToken(userID int64, username string) (string, error) {
	return j.generate(userID, username, "access", accessTokenTTL)
}

func (j *JWTIssuer) GenerateRefreshToken(userID int64, username string) (string, error) {
	return j.generate(userID, username, "refresh", refreshTokenTTL)
}

func (j *JWTIssuer) generate(userID int64, username, tokenType string, ttl time.Duration) (string, error) {
	claims := CustomClaims{
		UserID:    userID,
		Username:  username,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(j.privateKey)
}

// ParseRefreshToken 校验刷新令牌,并确认令牌类型为 refresh。
func (j *JWTIssuer) ParseRefreshToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return j.publicKey, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, model.TokenExpired
		}
		return nil, model.TokenInvalid
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, model.TokenInvalid
	}
	if claims.TokenType != "refresh" {
		return nil, model.TokenInvalid
	}
	return claims, nil
}
