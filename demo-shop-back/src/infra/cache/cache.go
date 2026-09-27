package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	Client *redis.Client
}

func NewRedisService(opts *redis.Options) (*RedisService, error) {
	// 基于配置初始化 Redis 客户端
	rdb := redis.NewClient(opts)

	// 带超时的连通性校验，避免网络异常时初始化阻塞
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Ping 测试连接有效性
	if err := rdb.Ping(ctx).Err(); err != nil {
		// 连接失败时主动关闭客户端，释放底层连接资源
		_ = rdb.Close()
		return nil, err
	}

	return &RedisService{
		Client: rdb,
	}, nil
}

// 基础操作

// Get 读取字符串值；key 不存在时返回 redis.Nil 错误
func (c *RedisService) Get(ctx context.Context, key string) (string, error) {
	return c.Client.Get(ctx, key).Result()
}

// Set 写入字符串值并设置过期时间；ttl <= 0 时写入永不过期的键
func (c *RedisService) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.Client.Set(ctx, key, value, ttl).Err()
}

// Del 删除一个或多个键
func (c *RedisService) Del(ctx context.Context, keys ...string) error {
	return c.Client.Del(ctx, keys...).Err()
}

// Exists 判断 key 是否存在
func (c *RedisService) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// SetNX 仅在 key 不存在时写入，返回是否写入成功。
// 用于幂等操作或分布式锁——锁场景必须传 ttl > 0，否则键永不过期
func (c *RedisService) SetNX(ctx context.Context, key, value string, ttl time.Duration) (bool, error) {
	return c.Client.SetNX(ctx, key, value, ttl).Result()
}

// Incr 对 key 自增并返回自增后的值（key 不存在时从 0 开始）
func (c *RedisService) Incr(ctx context.Context, key string) (int64, error) {
	return c.Client.Incr(ctx, key).Result()
}

// GetJSON 读取 JSON 并反序列化到 dest。
// 返回值 bool 表示 key 是否存在（命中缓存）；不存在时返回 (false, nil)
func (c *RedisService) GetJSON(ctx context.Context, key string, dest interface{}) (bool, error) {
	val, err := c.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return true, err
	}
	return true, nil
}

// SetJSON 将 value 序列化为 JSON 后写入并设置过期时间；ttl <= 0 时写入永不过期的键
func (c *RedisService) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.Client.Set(ctx, key, data, ttl).Err()
}

func (c *RedisService) Close() {
	c.Client.Close()
}
