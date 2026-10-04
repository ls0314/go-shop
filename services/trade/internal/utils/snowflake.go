package utils

import (
	"errors"
	"sync"
	"time"
)

// ============================================================
// 雪花 ID 与订单号生成
//
// 从单体 utils/snowflake.go 平移。放在 trade 侧是因为:订单号必须由本服务
// 在建单之前生成 —— 它是券核销的追溯字段,也是用户看到的单号,
// 而库存/券的幂等已改用调用方传入的 idempotency_key(不再依赖订单号)。
// ============================================================

// Snowflake 雪花ID生成器
//
// 标准 64 位雪花算法:
//
//	[1 位未使用] [41 位时间戳(毫秒)] [10 位工作节点] [12 位序列号]
//
// 起始时间:2026-01-01 00:00:00 UTC(epoch),可用约 69 年
// workerId:0~1023,支持部署到 1024 个节点
// sequence:0~4095,单节点每毫秒最多生成 4096 个 ID
type Snowflake struct {
	mu       sync.Mutex
	epoch    int64
	workerId int64
	sequence int64
	lastTime int64
}

const (
	defaultEpoch int64 = 1767225600000 // 2026-01-01 00:00:00 UTC

	workerBits     = 10
	sequenceBits   = 12
	workerMax      = -1 ^ (-1 << workerBits)   // 1023
	sequenceMax    = -1 ^ (-1 << sequenceBits) // 4095
	workerShift    = sequenceBits
	timestampShift = workerBits + sequenceBits
)

// NewSnowflake 创建雪花算法生成器
// 接收值：workerId - 工作节点 ID，范围 0~1023
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
func (s *Snowflake) NextId() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()
	if now < s.lastTime {
		now = s.lastTime
	}

	if now == s.lastTime {
		s.sequence = (s.sequence + 1) & sequenceMax
		if s.sequence == 0 {
			// 序列号溢出(同一毫秒内超过 4096 个),等到下一毫秒
			for now <= s.lastTime {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		s.sequence = 0
	}

	s.lastTime = now

	return ((now - s.epoch) << timestampShift) |
		(s.workerId << workerShift) |
		s.sequence
}

// OrderNoPrefix 订单号前缀,便于人工识别
const OrderNoPrefix = "DS"

// FormatOrderNo 由雪花 ID 生成订单号:DS + 36 进制大写。
//
// 长度约 15 字符,远小于 user_order_master.order_no 的 VARCHAR(32) ——
// 留足余量,将来换发号策略(比如加业务后缀)也不会越界。
func FormatOrderNo(id int64) string {
	return OrderNoPrefix + formatBase36(id)
}

// formatBase36 将 int64 转为 36 进制大写字符串(0-9 + A-Z)
func formatBase36(n int64) string {
	if n <= 0 {
		return "0"
	}
	const base36Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var buf [13]byte // int64 最大约 13 位 36 进制
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = base36Chars[n%36]
		n /= 36
	}
	return string(buf[i:])
}
