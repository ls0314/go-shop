package gate

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"demo-shop/services/marketing/internal/model"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// ============================================================
// 领券预扣减闸门(DS-A-19):Redis 闸门挡量 + DB 账本保真
//
// 从单体 src/infra/cache/deduct.go 的券部分平移。角色分工不变:
//   - 闸门(Redis Lua): 原子完成「读余量-校验-扣减」,无效流量 O(1) 拒绝,不碰数据库
//   - 账本(PostgreSQL): 悲观锁 + 条件 UPDATE 原样保留,作为对账依据
//   - 对账(task/couponreconcile.go): 定时以 DB 为准收敛闸门计数
//
// 键设计(与单体同名同义,便于对账与排查):
//
//	coupon:stock:{tid}        模板剩余可领数(回填自 DB: total_count - received_count)
//	coupon:limit:{tid}        每人限领数 —— 由模板定义,回填后不变
//	coupon:ucnt:{tid}:{uid}   用户已领数。与 DB 口径对齐:已使用未过期的券仍占名额
//	                          (CountUserCoupon 统计 status != 'expired'),故只在
//	                          领取时 INCR、核销/归还时不动
//
// 返回码约定(跨 Lua/Go 的统一语言):0=需回填,-1=售罄/不足,-2=超限;
// 成功=正数(值为「扣减后剩余量 + 1」—— **必须偏移 1**,否则发完最后一张时
// DECR 返回 0,会被误判为 GateBackfill 导致最后一张永远发不出去)。
// ============================================================
const (
	GateBackfill      = 0  // 键未回填,调用方从 DB 读权威值回填后重试
	GateSoldOut       = -1 // 售罄 / 库存不足
	GateLimitExceeded = -2 // 超过每人限领
)

// couponDeductScript 领券预扣。
// KEYS[1]=coupon:stock:{tid}  KEYS[2]=coupon:ucnt:{tid}:{uid}  KEYS[3]=coupon:limit:{tid}
//
// 逻辑次序不可调整(三道检查的顺序就是防超发的顺序):
//  1. stock 缺失 → BACKFILL;stock<=0 → 售罄拒绝。都必须发生在任何写操作之前,
//     售罄路径零写入(否则每次拒绝都会污染计数器)
//  2. limit 缺失 → BACKFILL(回填未完成,tonumber(nil) 会直接让脚本报错)
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
//
// 存在性守卫(EXISTS)不可省:闸门键可能已过期被清,此时 INCR 会凭空造出一个
// 计数为 1 的键,让"剩余量"凭空多出 1 —— 超发正是这么来的。
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

// CouponGate 领券闸门。Redis 未配置时 Enabled 为 false,调用方降级直走 DB。
type CouponGate struct {
	store *redis.Redis
}

func NewCouponGate(store *redis.Redis) *CouponGate {
	return &CouponGate{store: store}
}

// Enabled Redis 是否可用。false 时整个闸门旁路,由 DB 双防线独立保证正确性。
func (g *CouponGate) Enabled() bool {
	return g != nil && g.store != nil
}

// StockKey / LimitKey / UserCountKey 键名构造,对账任务与回填共用
func StockKey(templateId int64) string {
	return fmt.Sprintf("coupon:stock:%d", templateId)
}

func LimitKey(templateId int64) string {
	return fmt.Sprintf("coupon:limit:%d", templateId)
}

func UserCountKey(templateId, userId int64) string {
	return fmt.Sprintf("coupon:ucnt:%d:%d", templateId, userId)
}

// Deduct 领券预扣。返回值即上面的返回码约定:
// code>0 预扣成功(值为剩余量),继续走 DB 双防线;0 需回填;-1 售罄;-2 超限。
func (g *CouponGate) Deduct(ctx context.Context, templateId, userId int64) (int64, error) {
	if !g.Enabled() {
		return GateBackfill, nil
	}
	res, err := g.store.ScriptRunCtx(ctx, couponDeductScript, []string{
		StockKey(templateId),
		UserCountKey(templateId, userId),
		LimitKey(templateId),
	})
	if err != nil {
		return 0, err
	}
	return toInt64(res), nil
}

// Compensate 领券补偿。**只在「预扣成功过 + DB 事务失败」时调用** ——
// 无条件补偿会把没扣过的量加回去,导致超发。
func (g *CouponGate) Compensate(ctx context.Context, templateId, userId int64) error {
	if !g.Enabled() {
		return nil
	}
	_, err := g.store.ScriptRunCtx(ctx, couponCompensateScript, []string{
		StockKey(templateId),
		UserCountKey(templateId, userId),
	})
	return err
}

// Fill 以 DB 权威值回填闸门计数器。
//
// 用 SETNX 而不是 SET:并发回填(回填风暴)时只有一个请求真正写入,
// 其余请求直接复用先行者的值;且不覆盖已有值 —— 闸门里可能已有本周期
// 的预扣进度,盲目覆盖会把它清零造成超放。
func (g *CouponGate) Fill(ctx context.Context, key string, value int64, ttl time.Duration) bool {
	if !g.Enabled() {
		return false
	}
	ok, err := g.store.SetnxExCtx(ctx, key, strconv.FormatInt(value, 10), int(ttl.Seconds()))
	if err != nil {
		return false
	}
	return ok
}

// GateTTL 闸门键存活时间 = 模板剩余有效期 + 1 天缓冲,下限 1 小时。
// 与单体 getTTL 一致:闸门是加速器不是账本,过期后由对账重建即可。
func GateTTL(tpl *model.CouponTemplate) time.Duration {
	var d time.Duration
	if tpl.UsableDays > 0 {
		d = time.Duration(tpl.UsableDays)*24*time.Hour + 24*time.Hour
	} else if !tpl.EndTime.IsZero() {
		d = time.Until(tpl.EndTime) + 24*time.Hour
	}
	if d < time.Hour {
		d = time.Hour
	}
	return d
}

// toInt64 Lua 脚本的返回在 go-zero 里是 any(int64 / string 均可能),统一收敛。
// 解析不出来按 GateBackfill 处理 —— 走 DB 兜底总比误判售罄安全。
func toInt64(v any) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case string:
		if parsed, err := strconv.ParseInt(n, 10, 64); err == nil {
			return parsed
		}
	}
	return GateBackfill
}
