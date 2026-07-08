package middleware

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/utils"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var globalJWTService *utils.JWTService

func InitJWT(secretKey string) {
	if globalJWTService == nil {
		globalJWTService = utils.NewJWTService(secretKey)
	}
}

func GetJWTService() *utils.JWTService {
	if globalJWTService == nil {
		panic("jwt service not initialized, call InitJWT first")
	}
	return globalJWTService
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
			c.Abort()
			return
		}

		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "格式错误"})
			c.Abort()
			return
		}

		claims, err := GetJWTService().ParseToken(tokenStr)
		if err != nil {
			var message string
			switch {
			case errors.Is(err, model.TokenExpired):
				message = "token已过期，请重新登录"
			case errors.Is(err, model.TokenNotValidYet):
				message = "token尚未生效"
			case errors.Is(err, model.TokenMalformed):
				message = "token格式错误"
			case errors.Is(err, model.TokenInvalid):
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

func PermissionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		path := c.FullPath() // Gin 路由模板，如 /api/v1/platform/products/:id
		method := c.Request.Method

		userIdValue, exist := c.Get("user_id")
		if !exist {
			utils.Fail(c, 400, model.UserNotLogin.Error())
			c.Abort()
			return
		}

		userId, ok := userIdValue.(int64)
		if !ok {
			utils.Fail(c, 400, model.UserInfoError.Error())
			c.Abort()
			return
		}
		var permissionRepo = repository.NewPermissionRepo()
		codes, err := permissionRepo.GetPermCodesByApi(path, method)
		if err != nil {
			utils.Fail(c, 400, err.Error())
			c.Abort()
			return
		}
		if len(codes) == 0 {
			utils.Fail(c, 400, model.PermissionNotExist.Error())
			c.Abort()
			return
		}

		// 遍历所有 permission_code，用户持有任意一个即放行
		var hasPerm bool
		for _, code := range codes {
			hasPerm, err = permissionRepo.HasPermission(userId, code)
			if err != nil {
				utils.Error(c, 500, model.HasNotPerm.Error())
				c.Abort()
				return
			}
			if hasPerm {
				break
			}
		}

		if !hasPerm {
			utils.Fail(c, 400, model.UserHasNotPerm.Error())
			c.Abort()
			return
		}
		c.Next()

	}
}
