package tests

import (
	"demo-shop-back/src/service"
	"testing"
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
