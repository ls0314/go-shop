package tests

import (
	"testing"

	"demo-shop-back/src/repository"
	"demo-shop-back/src/service"
)

// C3 并发核销:20 并发核销同一张券 → RowsAffected 总和恰为 1,终态 used
// 断言姿势关键点:UseCoupon 的败方返回 (0, nil) 而非 error —— 乐观锁的约定是
// "败者无错,只是没赢",所以不能数 success,必须数 RowsAffected 总和
func TestUseCoupon_Concurrent_OnlyOneWins(t *testing.T) {
	const (
		orderNo    = "TESTORDER-C3"
		goroutines = 20
	)

	userID := mustCreateUser(t)
	// 前置:走真实链路领一张券(保证 expire_at 等字段合法)
	templateID := createTemplate(t, 10, 5)
	svc := service.NewCouponService()
	resp, err := svc.ReceiveCoupon(userID, templateID)
	if err != nil {
		t.Fatalf("领取前置券失败: %v", err)
	}

	repo := repository.NewCouponRepo()
	stats, affectedSum := runConcurrent(t, goroutines, func(i int) (int64, error) {
		return repo.UseCoupon(resp.UserCouponId, orderNo)
	})

	if stats["other"] != 0 {
		t.Fatalf("出现未分类错误: %v", stats)
	}
	if affectedSum != 1 {
		t.Fatalf("并发核销应恰好 1 个事务成功, RowsAffected 总和 %d (stats=%v)", affectedSum, stats)
	}
	if status := queryString(t, `SELECT status FROM user_coupon WHERE user_coupon_id = ?`, resp.UserCouponId); status != "used" {
		t.Fatalf("终态应为 used, 实际 %s", status)
	}
	if got := queryString(t, `SELECT COALESCE(order_no,'') FROM user_coupon WHERE user_coupon_id = ?`, resp.UserCouponId); got != orderNo {
		t.Fatalf("核销凭据应为 %s, 实际 %q", orderNo, got)
	}
}

// C4 重复归还幂等:第一次归还真归还,第二次条件(status='used')不再匹配 → (0,nil) 天然幂等
func TestRefundCoupon_Idempotent(t *testing.T) {
	const (
		orderNo = "TESTORDER-C4"
	)

	userID := mustCreateUser(t)
	templateID := createTemplate(t, 10, 5)
	svc := service.NewCouponService()
	resp, err := svc.ReceiveCoupon(userID, templateID)
	if err != nil {
		t.Fatalf("领取前置券失败: %v", err)
	}
	repo := repository.NewCouponRepo()
	if n, err := repo.UseCoupon(resp.UserCouponId, orderNo); err != nil || n != 1 {
		t.Fatalf("前置核销失败: affected=%d err=%v", n, err)
	}

	if n, err := repo.RefundCoupon(resp.UserCouponId); err != nil || n != 1 {
		t.Fatalf("第一次归也应成功: affected=%d err=%v", n, err)
	}
	if n, err := repo.RefundCoupon(resp.UserCouponId); err != nil || n != 0 {
		t.Fatalf("重复归还必须返回 (0,nil): affected=%d err=%v", n, err)
	}
	status := queryString(t, `SELECT status FROM user_coupon WHERE user_coupon_id = ?`, resp.UserCouponId)
	got := queryString(t, `SELECT COALESCE(order_no,'') FROM user_coupon WHERE user_coupon_id = ?`, resp.UserCouponId)
	if status != "unused" || got != "" {
		t.Fatalf("归还终态异常: status=%s order_no=%q", status, got)
	}
}

//func TestUserCouponFK_RejectsOrphanUser(t *testing.T) {
//	templateID := createTemplate(t, 10, 1)
//	uc := model.UserCoupon{
//		TemplateId: templateID,
//		UserId:     int64(100000),
//		Status:     "unused",
//		ExpireAt:   time.Now().Add(time.Hour),
//	}
//
//	if err := db.DB.Select("template_id", "user_id", "status", "expire_at").Create(&uc).Error; err == nil {
//		t.Fatal("user_coupon.user_id 外键未生效: 孤儿 userId 插入成功了")
//	}
//}
