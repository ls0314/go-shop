package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// ============================================================
// 预扣减闸门(DS-A-19):Redis 闸门挡量 + DB 账本保真
// ============================================================
// 角色分工:
//   - 闸门(Redis Lua): 原子完成「读余量-校验-扣减」,无效流量 O(1) 拒绝,不操作数据库
//   - 账本(PostgreSQL): 悲观锁 + 条件 UPDATE 原样保留,作为对账依据
//   - 对账(task/stock_reconcile.go): 定时以 DB 为准收敛闸门计数
//
// 键设计:
//
//	coupon:stock:{tid}        模板剩余可领数(回填自 DB: total_count - received_count)
//	coupon:limit:{tid}        每人限领数 —— 创建后不可变(无更新接口),SETNX 一次即可,
//	                          之后预扣请求全程零 DB
//	coupon:ucnt:{tid}:{uid}   用户已领数。与 DB 口径对齐:已使用未过期的券仍占限领名额
//	                          (CountUserCoupon 统计 status != 'expired'),所以这里
//	                          只在领取时 INCR、核销/归还时不动
//	sku:stock:{skuId}         SKU 可售库存 —— 刻意复用商品详情的既有缓存键:
//	                          它本来就存在(详情页在读),预扣只是让它从"可丢的缓存"
//	                          升级为"可对账的计数器"。一套键一种语义,排查时不用对表
//
// 返回码约定(跨 Lua/Go 的统一语言):0=需回填,-1=售罄/不足,-2=超限;
// 成功=正数(值为「扣减后剩余量 + 1」—— 必须偏移 1,否则发完最后一张时
// DECR 返回 0,会被误判为 GateBackfill 导致最后一张永远发不出去)
const (
	GateBackfill      = 0  // 键未回填,调用方从 DB 读权威值回填后重试
	GateSoldOut       = -1 // 售罄 / 库存不足
	GateLimitExceeded = -2 // 超过每人限领
)

// couponDeductScript 领券预扣。
// KEYS[1]=coupon:stock:{tid}  KEYS[2]=coupon:ucnt:{tid}:{uid}  KEYS[3]=coupon:limit:{tid}
//
// 逻辑次序不可调整(三道检查的顺序就是防超发的顺序):
//  1. stock 缺失 → BACKFILL;stock<=0 → 售罄拒绝 —— 都必须发生在任何写操作之前,
//     售罄路径零写入(否则每次拒绝都会污染计数器)
//  2. limit 缺失 → BACKFILL(回填未完成,tonumber(nil) 会直接脚本报错)
//  3. 「先 INCR 再比较再回滚」:INCR 返回值是排队后的最新值,以它为准才原子
var couponDeductScript = redis.NewScript(`
local stock = redis.call('GET', KEYS[1])
if not stock then
	return 0
end
if tonumber(stock) <= 0 then
	return -1
end
local limit = redis.call('GET', KEYS[3])
if not limit then
	return 0
end
local used = redis.call('INCR', KEYS[2])
if used > tonumber(limit) then
	redis.call('DECR', KEYS[2])
	return -2
end
-- 返回 剩余量+1 保证恒为正(0 已被 BACKFILL 占用,见顶部约定)
return 1 + tonumber(redis.call('DECR', KEYS[1]))
`)

// couponCompensateScript 领券补偿:DB 事务失败后把闸门的扣减还回去。
var couponCompensateScript = redis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 1 then
	redis.call('INCR', KEYS[1])
end
local used = redis.call('GET', KEYS[2])
if used and tonumber(used) > 0 then
	redis.call('DECR', KEYS[2])
end
return 1
`)

// skuDeductScript 库存预扣:余量 >= qty 才一次 DECRBY;不足时零写入直接拒绝。
// KEYS[1]=sku:stock:{skuId}  ARGV[1]=qty
var skuDeductScript = redis.NewScript(`
local stock = redis.call('GET', KEYS[1])
if not stock then
	return 0
end
if tonumber(stock) < tonumber(ARGV[1]) then
	return -1
end
-- 同领券脚本:剩余量+1,避开 BACKFILL=0 的返回码
return 1 + tonumber(redis.call('DECRBY', KEYS[1], ARGV[1]))
`)

// skuCompensateScript 库存补偿(与领券补偿同款存在性守卫)
var skuCompensateScript = redis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 1 then
	return redis.call('INCRBY', KEYS[1], ARGV[1])
end
return 0
`)

// DeductCouponReceive 领券预扣闸门。
// 返回值:code>0 预扣成功(剩余量),继续走 DB 双防线;
func (c *RedisService) DeductCouponReceive(ctx context.Context, templateId, userId int64) (int64, error) {
	keys := []string{
		fmt.Sprintf("coupon:stock:%d", templateId),
		fmt.Sprintf("coupon:ucnt:%d:%d", templateId, userId),
		fmt.Sprintf("coupon:limit:%d", templateId),
	}
	return couponDeductScript.Run(ctx, c.Client, keys).Int64()
}

// CompensateCouponReceive 领券补偿。只在「预扣成功过 + DB 事务失败」时调用,
func (c *RedisService) CompensateCouponReceive(ctx context.Context, templateId, userId int64) error {
	keys := []string{
		fmt.Sprintf("coupon:stock:%d", templateId),
		fmt.Sprintf("coupon:ucnt:%d:%d", templateId, userId),
	}
	return couponCompensateScript.Run(ctx, c.Client, keys).Err()
}

// DeductSkuStock 库存预扣闸门(返回码语义同领券)
func (c *RedisService) DeductSkuStock(ctx context.Context, skuId, qty int64) (int64, error) {
	key := fmt.Sprintf("sku:stock:%d", skuId)
	return skuDeductScript.Run(ctx, c.Client, []string{key}, qty).Int64()
}

// CompensateSkuStock 库存补偿
func (c *RedisService) CompensateSkuStock(ctx context.Context, skuId, qty int64) error {
	key := fmt.Sprintf("sku:stock:%d", skuId)
	return skuCompensateScript.Run(ctx, c.Client, []string{key}, qty).Err()
}

// FillGateCounter 以 DB 权威值回填闸门计数器。
// 用 SETNX 而不是 SET:并发回填(回填风暴)时只有一个请求真正写入,
// 其余请求直接复用先行者的值;且不覆盖已有值 —— 闸门里可能已有本周期
// 的预扣进度,盲目覆盖会把它清零造成超放
func (c *RedisService) FillGateCounter(ctx context.Context, key string, value int64, ttl time.Duration) bool {
	ok, err := c.SetNX(ctx, key, strconv.FormatInt(value, 10), ttl)
	if err != nil {
		return false
	}
	return ok
}
