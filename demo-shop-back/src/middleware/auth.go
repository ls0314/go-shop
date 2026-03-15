package middleware

import (
	"demo-shop-back/src/utils"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var jwtService *utils.JWTService

func InitJWT(secreKey string) *utils.JWTService {
	jwtService = utils.NewJWTService(secreKey)
	return jwtService
}

func AuthMiddleware(jwtService *utils.JWTService) gin.HandlerFunc {

	if jwtService == nil {
		panic("jwt service is nil")
	}

	return func(c *gin.Context) {

		tokenStr := c.GetHeader("Authorization")

		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
			c.Abort()
			return
		}

		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"massage": "格式错误"})
			c.Abort()
			return
		}

		claims, err := jwtService.ParseToken(tokenStr)

		if err != nil {
			var message string
			switch {
			case errors.Is(err, utils.TokenExpired):
				message = "token已过期，请重新登录"
			case errors.Is(err, utils.TokenNotValidYet):
				message = "token尚未生效"
			case errors.Is(err, utils.TokenMalformed):
				message = "token格式错误"
			case errors.Is(err, utils.TokenInvalid):
				fallthrough
			default:
				message = "无效的token"
			}
			c.JSON(http.StatusUnauthorized, gin.H{"message": message})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}
