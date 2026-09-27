package middleware

import (
	"bytes"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"
	"encoding/json"
	"io"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var logQueue = make(chan *model.OperationLog, 1024)

// truncateStr 按字符截断：PostgreSQL VARCHAR(n) 按字符数计，用 rune 截断避免切坏中文
func truncateStr(s string, maxLen int) string {
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	return string(r[:maxLen])
}

// isSensitiveKey 判断键名是否敏感（小写包含匹配，覆盖 password/secret/token 等变体）
func isSensitiveKey(k string) bool {
	k = strings.ToLower(k)
	return strings.Contains(k, "password") ||
		strings.Contains(k, "secret") ||
		strings.Contains(k, "token") ||
		strings.Contains(k, "authorization")
}

// redactSensitive 递归遍历 JSON 结构，将敏感键的值替换为 ***
func redactSensitive(v interface{}) {
	switch t := v.(type) {
	case map[string]interface{}:
		for k, val := range t {
			if isSensitiveKey(k) {
				t[k] = "***"
			} else {
				redactSensitive(val)
			}
		}
	case []interface{}:
		for _, item := range t {
			redactSensitive(item)
		}
	}
}

// sanitizeParams 请求体脱敏 + 截断：JSON 可解析时递归脱敏敏感键，再统一截断 2000 字符
// 背景：handler 在 bcrypt 哈希前绑定明文密码，日志快照若不脱敏会存明文密码副本
func sanitizeParams(raw string) string {
	if raw == "" {
		return ""
	}
	var obj interface{}
	if err := json.Unmarshal([]byte(raw), &obj); err == nil {
		redactSensitive(obj)
		if data, err := json.Marshal(obj); err == nil {
			raw = string(data)
		}
	}
	return truncateStr(raw, 2000)
}

func firstErrorMessage(c *gin.Context) string {
	if len(c.Errors) == 0 {
		return ""
	}
	return c.Errors[len(c.Errors)-1].Error()
}

func OperationLogMiddleware(module string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		var body []byte
		if c.Request.Body != nil {
			body, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		c.Next()

		userIdVal, hasUserId := c.Get("user_id")
		usernameVal, hasUsername := c.Get("username")

		if !hasUserId || !hasUsername {
			return
		}
		userId, ok := userIdVal.(int64)
		if !ok {
			return
		}
		username, _ := usernameVal.(string)

		operationLog := &model.OperationLog{
			UserId:        userId,
			Username:      username,
			Module:        module,
			Operation:     c.Request.Method,
			RequestMethod: c.Request.Method,
			RequestUrl:    c.FullPath(),
			RequestParams: sanitizeParams(string(body)),
			IpAddress:     c.ClientIP(),
			UserAgent:     truncateStr(c.Request.UserAgent(), 500),
			ExecuteTime:   time.Since(start).Milliseconds(),
			Status:        c.Writer.Status() < 400,
			ErrorMessage:  firstErrorMessage(c),
		}
		select {
		case logQueue <- operationLog:
		default:
			log.Printf("[WARN] 操作日志队列已满，丢弃: module=%s url=%s", module, c.FullPath())
		}
	}
}

func InitLogWorker(repo *repository.OperationLogRepo) {
	for i := 0; i < 2; i++ {
		go func() {
			for operationLog := range logQueue {
				if err := repo.CreateOperationLog(operationLog); err != nil {
					log.Printf("[WARN] 操作日志写入失败: %v", err)
				}
			}
		}()
	}
}
