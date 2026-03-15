package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}
type JWTService struct {
	SigningKey []byte
}

var (
	TokenExpired     = errors.New("token已过期")
	TokenNotValidYet = errors.New("token无")
	TokenMalformed   = errors.New("token格式错误")
	TokenInvalid     = errors.New("token错误")
)

func NewJWTService(signingKey string) *JWTService {
	return &JWTService{}
}

func (j *JWTService) GenerateToken(userID int64, username string, duration time.Duration) (string, error) {

	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

func (j *JWTService) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return j.SigningKey, nil
		})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, TokenExpired
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, TokenNotValidYet
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, TokenMalformed
		}
		return nil, TokenInvalid
	}

	if token != nil {
		if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
			return claims, nil
		}
	}

	return nil, TokenInvalid
}
