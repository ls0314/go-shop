package task

import (
	"context"
	"errors"
	"time"

	"demo-shop/services/trade/internal/infra/lock"
	"demo-shop/services/trade/internal/model"
	"demo-shop/services/trade/internal/repository"
	tradesvc "demo-shop/services/trade/internal/service"

	"github.com/zeromicro/go-zero/core/logx"
	gozeroservice "github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// orderScanLockName 锁名与单体保持一致:搬迁前后不能出现两把不同的锁同时跑
const orderScanLockName = "order:time:scan"

// OrderTimeoutService 订单超时扫描。
type OrderTimeoutService struct {
	orderRepo *repository.OrderRepo
	orderSvc  *tradesvc.OrderService
	locks     *lock.TaskLockManager
}

// NewOrderTimeoutService 构造超时扫描任务。
func NewOrderTimeoutService(orderRepo *repository.OrderRepo, orderSvc *tradesvc.OrderService,
	store *redis.Redis) *OrderTimeoutService {
	return &OrderTimeoutService{
		orderRepo: orderRepo,
		orderSvc:  orderSvc,
		locks:     lock.NewTaskLockManager(store),
	}
}

// AsService 包装成可被 go-zero service group 托管的服务(支持优雅退出)
func (s *OrderTimeoutService) AsService(interval time.Duration) gozeroservice.Service {
	return newTaskService("order-timeout-scan", func(ctx context.Context) {
		_ = s.Start(ctx, interval)
	})
}

// Start 周期扫描(阻塞)
func (s *OrderTimeoutService) Start(ctx context.Context, interval time.Duration) error {
	logx.Infof("启动订单超时扫描: 周期 %v(判据为订单表 expire_at)", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logx.Info("订单超时扫描任务退出")
			return nil
		case <-ticker.C:
			s.ScanOnce(ctx)
		}
	}
}

// ScanOnce 执行一轮扫描(测试/手动触发入口)。
// 返回值:cancelled - 本轮真实取消的单数;skipped - 已被 MQ 处理而跳过的单数。
func (s *OrderTimeoutService) ScanOnce(ctx context.Context) (cancelled, skipped int) {
	release, ok := s.locks.TryLock(orderScanLockName, 5*time.Minute)
	if !ok {
		return 0, 0
	}
	defer release()
	orders, err := s.orderRepo.ListExpirePendingPay(time.Now(), model.OrderScanBatch)
	if err != nil {
		logx.Errorf("超时扫描: 查询超时订单失败: %v", err)
		return 0, 0
	}
	if len(orders) == 0 {
		return 0, 0
	}
	logx.Infof("超时扫描: 命中 %d 张超时未支付订单", len(orders))

	for _, order := range orders {
		_, err := s.orderSvc.CancelOrderBySystem(ctx, order.OrderId, "系统")
		switch {
		case err == nil:
			cancelled++
		case errors.Is(err, model.ErrOrderAlreadyCancelled):
			// 订单已取消,目标已达成
			skipped++
		case errors.Is(err, model.ErrOrderCannotCancel):
			// 状态已不是 pending_pay。绝大多数是"MQ 先取消成功"的正常竞态,
			// 少数是"用户在超时瞬间付款成功" —— 后者值得留意,故留一行日志区分。
			skipped++
			logx.Infof("超时扫描: 订单状态已变更,跳过 orderId=%d err=%v", order.OrderId, err)
		default:
			// 真故障(DB 超时 / 下游不可用):下轮还会重扫到,值得告警
			logx.Errorf("超时扫描: 取消失败(下轮重试) orderId=%d err=%v", order.OrderId, err)
		}
	}
	return cancelled, skipped
}
