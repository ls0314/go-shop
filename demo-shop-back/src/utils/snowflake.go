package utils

import (
	"errors"
	"sync"
	"time"
)

// Snowflake 雪花ID生成器
//
// 标准 64 位雪花算法：
//
//	[1 位未使用] [41 位时间戳(毫秒)] [10 位工作节点] [12 位序列号]
//	 0           timestamp            worker         sequence
//
// 起始时间：2026-01-01 00:00:00 UTC（epoch），可用约 69 年
// workerId：0~1023，支持部署到 1024 个节点
// sequence：0~4095，单节点每毫秒最多生成 4096 个 ID
//
// 使用示例：
//
//	sf, err := utils.NewSnowflake(1)
//	id := sf.NextId()  // → 7801234567890123456
type Snowflake struct {
	mu       sync.Mutex
	epoch    int64 // 起始时间戳（毫秒）
	workerId int64 // 工作节点 ID（0~1023）
	sequence int64 // 毫秒内序列号（0~4095）
	lastTime int64 // 上次生成 ID 的时间戳（毫秒）
}

const (
	// 起始时间：2026-01-01 00:00:00 UTC
	defaultEpoch int64 = 1767225600000

	workerBits     = 10                        // 工作节点占 10 位
	sequenceBits   = 12                        // 序列号占 12 位
	workerMax      = -1 ^ (-1 << workerBits)   // 1023
	sequenceMax    = -1 ^ (-1 << sequenceBits) // 4095
	workerShift    = sequenceBits              // 12
	timestampShift = workerBits + sequenceBits // 22
)

// NewSnowflake 创建雪花算法生成器
// 接收值：workerId - 工作节点 ID，范围 0~1023
// 返回值：*Snowflake - 生成器实例，error - workerId 越界时返回错误
func NewSnowflake(workerId int64) (*Snowflake, error) {
	if workerId < 0 || workerId > workerMax {
		return nil, errors.New("workerId超出范围(0~1023)")
	}
	return &Snowflake{
		epoch:    defaultEpoch,
		workerId: workerId,
	}, nil
}

// NextId 生成下一个雪花ID
// 接收值：无
// 返回值：int64 - 唯一雪花ID
func (s *Snowflake) NextId() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()

	// 时钟回拨检测：如果当前时间小于上次生成时间，等待追上
	if now < s.lastTime {
		now = s.lastTime
	}

	if now == s.lastTime {
		// 同一毫秒内，序列号递增
		s.sequence = (s.sequence + 1) & sequenceMax
		if s.sequence == 0 {
			// 序列号溢出，等待下一毫秒
			for now <= s.lastTime {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		s.sequence = 0
	}

	s.lastTime = now

	// 组装雪花ID
	// [41位: timestamp-epoch] [10位: workerId] [12位: sequence]
	id := ((now - s.epoch) << timestampShift) |
		(s.workerId << workerShift) |
		s.sequence

	return id
}

// GenerateOrderNo 生成订单号
// 格式：DS + 雪花ID的36进制大写表示
// 示例：DS2L9P7XQ3R8A
// 接收值：sf - 雪花生成器实例
// 返回值：string - 订单业务号
func GenerateOrderNo(sf *Snowflake) string {
	id := sf.NextId()
	// 转 36 进制（0-9 + A-Z），取大写
	return "DS" + formatBase36(id)
}

// formatBase36 将 int64 转为 36 进制大写字符串
func formatBase36(n int64) string {
	if n == 0 {
		return "0"
	}
	const base36Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var result [13]byte // int64 最大约 13 位 36 进制
	i := len(result)
	for n > 0 {
		i--
		result[i] = base36Chars[n%36]
		n /= 36
	}
	return string(result[i:])
}

// 全局默认雪花生成器（单节点，workerId=1）
var defaultSnowflake *Snowflake

// InitSnowflake 初始化默认雪花生成器
// 接收值：workerId - 工作节点 ID
func InitSnowflake(workerId int64) {
	sf, err := NewSnowflake(workerId)
	if err != nil {
		panic(err)
	}
	defaultSnowflake = sf
}

// NextOrderNo 使用默认生成器生成订单号
// 接收值：无
// 返回值：string - 订单号
func NextOrderNo() string {
	if defaultSnowflake == nil {
		InitSnowflake(1)
	}
	return GenerateOrderNo(defaultSnowflake)
}
