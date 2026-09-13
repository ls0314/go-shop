package tests

import (
	"context"
	"demo-shop-back/src/service"
	"demo-shop-back/src/task"
	"fmt"
	"testing"
	"time"
)

func TestReceiveCoupon_Concurrent_NoOversell(t *testing.T) {
	const (
		total      = 100
		goroutines = 300
	)
	templateID := createTemplate(t, total, 1)
	userIDs := mustCreateUsers(t, goroutines)
	svc := service.NewCouponService()

	stats, _ := runConcurrent(t, goroutines, func(i int) (int64, error) {
		_, err := svc.ReceiveCoupon(userIDs[i], templateID)
		return 0, err
	})

	if stats["success"] != total || stats["sold_out"] != goroutines-total || stats["other"] != 0 {
		t.Fatalf("C1 断言失败: %v (期望 success=%d sold_out=%d other=0)", stats, total, goroutines-total)
	}

	if got := queryInt64(t, `SELECT received_count FROM coupon_template WHERE template_id = ?`, templateID); got != total {
		t.Fatalf("received_count 应为 %d, 实际 %d", total, got)
	}
	if got := queryInt64(t, `SELECT COUNT(*) FROM user_coupon WHERE template_id = ?`, templateID); got != total {
		t.Fatalf("user_coupon 行数应为 %d, 实际 %d", total, got)
	}
}

func TestReceiveCoupon_Concurrent_PerUserLimit(t *testing.T) {
	const (
		limit      = 3
		goroutines = 30
	)
	templateID := createTemplate(t, 100, limit)
	userID := mustCreateUser(t)

	svc := service.NewCouponService()
	stats, _ := runConcurrent(t, goroutines, func(i int) (int64, error) {
		_, err := svc.ReceiveCoupon(userID, templateID)
		return 0, err
	})

	if stats["success"] != limit || stats["limit_exceeded"] != goroutines-limit || stats["other"] != 0 {
		t.Fatalf("C2 断言失败: %v (期望 success=%d limit_exceeded=%d other=0)", stats, limit, goroutines-limit)
	}
	if got := queryInt64(t, `SELECT COUNT(*) FROM user_coupon WHERE user_id = ? AND template_id = ?`, userID, templateID); got != limit {
		t.Fatalf("该用户实际持券应为 %d, 实际 %d", limit, got)
	}
}

// coupon_concurrent_test.go —— 新增 import: "context" "fmt" "time" "demo-shop-back/src/infra/cache"

// C5 闸门端到端(两阶段,匹配闸门的设计契约):
//
//	阶段一(突发): 闸门允许瞬时偏差(SETNX 回填竞态可致少放行 1 个),但绝不超发 ——
//	             发放计数 / 账本行数 / 成功响应三方一致,且 ≤ total
//	阶段二(收敛): 对账任务以 DB 为准把真实余量还给闸门后,剩余额度必须能被领完 ——
//	             突发期的可用性损失必须被最终一致修复
func TestReceiveCoupon_WithGate_NoOversell(t *testing.T) {
	rdb := newTestRedis(t)
	templateID := createTemplate(t, 10, 1)
	flushGateKeys(t, rdb, templateID) // 清除跨轮残留/跨环境串台的闸门键(见 harness 注释)
	userIDs := mustCreateUsers(t, 50)
	svc := service.NewCouponServiceWithCache(rdb)

	stats, _ := runConcurrent(t, 50, func(i int) (int64, error) {
		_, err := svc.ReceiveCoupon(userIDs[i], templateID)
		return 0, err
	})

	// ---- 阶段一:突发期不超发 ----
	if stats["other"] != 0 || stats["success"] > 10 {
		t.Fatalf("C5 阶段一断言失败: %v (要求 other=0 且 success ≤ 10)", stats)
	}
	rows := queryInt64(t, `SELECT COUNT(*) FROM user_coupon WHERE template_id = ?`, templateID)
	received := queryInt64(t, `SELECT received_count FROM coupon_template WHERE template_id = ?`, templateID)
	if rows != received || rows != int64(stats["success"]) {
		t.Fatalf("C5 三方不一致: 账本=%d 计数=%d 响应成功=%d", rows, received, stats["success"])
	}

	// ---- 阶段二:对账收敛 ----
	task.NewStockReconcileServiceWithCache(rdb).ReconcileOnce()
	extra := 10 - int(received) // 突发期被闸门误拦的余量
	if extra > 0 {
		ids := mustCreateUsers(t, extra)
		s2, _ := runConcurrent(t, extra, func(i int) (int64, error) {
			_, err := svc.ReceiveCoupon(ids[i], templateID)
			return 0, err
		})
		if s2["success"] != extra || s2["other"] != 0 {
			t.Fatalf("C5 阶段二断言失败(对账后余量未释放): %v (期望 success=%d)", s2, extra)
		}
	}
	if got := queryInt64(t, `SELECT COUNT(*) FROM user_coupon WHERE template_id = ?`, templateID); got != 10 {
		t.Fatalf("C5 最终账本应为 10 行, 实际 %d", got)
	}
}

// C5b 闸门 O(1) 拒绝:DB 有量、闸门计数为 0 → 全部在闸门被拒,DB 零写入。
// 这是对「闸门挡量」的直接证明:30 次拒绝没有产生任何 DB 查询/写入
func TestReceiveCoupon_GateRejectsWithoutDB(t *testing.T) {
	rdb := newTestRedis(t)
	templateID := createTemplate(t, 100, 1)
	flushGateKeys(t, rdb, templateID) // 同上:先清残留,再人为置 0 才有"DB 有量而闸门无"的语义
	userIDs := mustCreateUsers(t, 30)
	svc := service.NewCouponServiceWithCache(rdb)

	ctx := context.Background()
	ttl := time.Hour
	// 直接把闸门视图置 0(DB 实际仍有 100 张):模拟"闸门认为售罄"
	rdb.FillGateCounter(ctx, fmt.Sprintf("coupon:stock:%d", templateID), 0, ttl)
	rdb.FillGateCounter(ctx, fmt.Sprintf("coupon:limit:%d", templateID), 1, ttl)

	stats, _ := runConcurrent(t, 30, func(i int) (int64, error) {
		_, err := svc.ReceiveCoupon(userIDs[i], templateID)
		return 0, err
	})

	if stats["sold_out"] != 30 || stats["other"] != 0 {
		t.Fatalf("C5b 断言失败: %v (期望 30 次全部在闸门拒绝)", stats)
	}
	if got := queryInt64(t, `SELECT received_count FROM coupon_template WHERE template_id = ?`, templateID); got != 0 {
		t.Fatalf("闸门拒绝不应触碰 DB: received_count=%d", got)
	}
}
