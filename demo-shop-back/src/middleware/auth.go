package middleware

import (
	"demo-shop-back/src/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		tokenStr := c.GetHeader("Authorization")

		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
			c.Abort()
			return
		}

		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

		token, err := jwt.ParseWithClaims(tokenStr, &utils.CustomClaims{},
			func(token *jwt.Token) (interface{}, error) {
				return []byte("demo-shop-secret"), nil
			})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "token无效"})
			c.Abort()
			return
		}

		claims := token.Claims.(*utils.CustomClaims)

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("phone", claims.Phone)
		c.Set("email", claims.Email)

		c.Next()
	}
}
