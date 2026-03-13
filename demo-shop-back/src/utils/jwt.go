package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("demo-shop-secret")

type CustomClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

func GenerateToken(userID, username string, duration time.Duration) (string, error) {

	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtSecret)
}
