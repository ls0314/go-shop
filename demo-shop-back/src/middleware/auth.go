package middleware

import (
	"context"
	"demo-shop-back/src/infra/cache"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/utils"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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

// getUserPermCodes 获取用户全部权限码：先查 Redis(user:perm:{userId})，miss 查 DB 回写
func getUserPermCodes(permissionRepo *repository.PermissionRepo, cch *cache.RedisService, userId int64) ([]string, error) {
	key := fmt.Sprintf("user:perm:%d", userId)
	if cch != nil {
		var codes []string
		hit, err := cch.GetJSON(context.Background(), key, &codes)
		if err == nil && hit {
			return codes, nil
		}
	}
	codes, err := permissionRepo.GetPermCodesByUserId(userId)
	if err != nil {
		return nil, err
	}
	if cch != nil {
		_ = cch.SetJSON(context.Background(), key, codes, 30*time.Minute)
	}
	return codes, nil
}

// GetUserPermCodes 供 handler 层获取用户全部权限码(按钮级权限展示用)
func GetUserPermCodes(permRepo *repository.PermissionRepo, cch *cache.RedisService, userId int64) ([]string, error) {
	return getUserPermCodes(permRepo, cch, userId)
}

func getApiPermCodes(permissionRepo *repository.PermissionRepo, cch *cache.RedisService, path, method string) ([]string, error) {
	if cch == nil {
		return permissionRepo.GetPermCodesByApi(path, method)
	}

	version := "0"
	v, err := cch.Get(context.Background(), "api:perm:version")
	switch {
	case err == nil:
		version = v
	case errors.Is(err, redis.Nil):
	case err != nil:
		return permissionRepo.GetPermCodesByApi(path, method)
	}

	key := fmt.Sprintf("api:perm:v%s:%s:%s", version, method, path)
	var codes []string
	hit, err := cch.GetJSON(context.Background(), key, &codes)
	if err == nil && hit {
		return codes, nil
	}

	codes, err = permissionRepo.GetPermCodesByApi(path, method)
	if err != nil {
		return nil, err
	}

	// 回写——空结果不缓存
	if len(codes) > 0 {
		_ = cch.SetJSON(context.Background(), key, codes, 30*time.Minute)
	}
	return codes, nil
}

func PermissionMiddleware(permRepo *repository.PermissionRepo, cch *cache.RedisService) gin.HandlerFunc {
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
		codes, err := getApiPermCodes(permRepo, cch, path, method)
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

		// 一次取用户全量权限码(缓存优先)，命中则跳过原先逐 code 的 3 表 JOIN 查询
		userCodes, err := getUserPermCodes(permRepo, cch, userId)
		if err != nil {
			utils.Error(c, 500, model.HasNotPerm.Error())
			c.Abort()
			return
		}
		userCodeSet := make(map[string]struct{}, len(userCodes))
		for _, code := range userCodes {
			userCodeSet[code] = struct{}{}
		}

		// 遍历接口所需 permission_code，用户持有任意一个即放行(语义与原逻辑完全等价)
		var hasPerm bool
		for _, code := range codes {
			if _, ok := userCodeSet[code]; ok {
				hasPerm = true
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
