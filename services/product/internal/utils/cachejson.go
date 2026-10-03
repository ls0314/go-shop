package utils

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// GetJSONCtx 读 JSON 并反序列化到 dest。
// 返回值 hit 表示 key 是否存在;不存在时返回 (false, nil),不当作错误。
// go-zero 的 *redis.Redis 只有 GetCtx(返回 string),JSON 组装由此处补齐,
// 与单体 cache.RedisService.GetJSON 语义一致。
func GetJSONCtx(ctx context.Context, r *redis.Redis, key string, dest interface{}) (bool, error) {
	if r == nil {
		return false, nil
	}
	val, err := r.GetCtx(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return true, err
	}
	return true, nil
}

// SetJSONCtx 序列化为 JSON 后写入,ttlSeconds <= 0 时写入永不过期的键。
func SetJSONCtx(ctx context.Context, r *redis.Redis, key string, value interface{}, ttlSeconds int) error {
	if r == nil {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.SetexCtx(ctx, key, string(data), ttlSeconds)
}
