package utils

import (
	"demo-shop-back/src/model"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	TokenType string `json:"tokenType"`
	jwt.RegisteredClaims
}
type JWTService struct {
	SigningKey []byte
}

func NewJWTService(signingKey string) *JWTService {
	return &JWTService{
		SigningKey: []byte(signingKey),
	}
}

func (j *JWTService) GenerateAccessToken(userID int64, username string) (string, error) {
	claims := CustomClaims{
		UserID:    userID,
		Username:  username,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

func (j *JWTService) GenerateRefreshToken(userID int64, username string) (string, error) {
	claims := CustomClaims{
		UserID:    userID,
		Username:  username,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
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

	if token != nil {
		if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
			return claims, nil
		}
	}

	return nil, model.TokenInvalid
}
