/*
负责MQ宕机后的兜底，超时扫描
*/

package task

import (
	"demo-shop-back/src/infra"
	"demo-shop-back/src/infra/metrics"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/service"
	"errors"
	"log"
	"time"
)

const (
	scanInterval   = 1 * time.Minute
	scanLockExpiry = 5 * time.Minute
	scanBatchSize  = 100
	scanLockName   = "order:time:scan"
)

type OrderTimeoutSacnService struct {
	orderRepo *repository.OrderRepo
	orderSvc  *service.OrderService
	locks     *DistributedLockManager
}

// NewOrderTimeoutScanService 创建扫描任务
func NewOrderTimeoutScanService() *OrderTimeoutSacnService {
	return &OrderTimeoutSacnService{
		orderRepo: repository.NewOrderRepo(),
		orderSvc:  service.NewOrderService(),
		locks:     NewDistributedLockManager(infra.GetCache()),
	}
}

// Start 启动周期扫描(阻塞调用方,通常放 goroutine)
func (s *OrderTimeoutSacnService) Start(interval time.Duration) {
	log.Printf("[INFO] 启动订单超时扫描: 周期 %v, 阈值 %v", interval, model.OrderPayTTL)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		go s.run()
	}
}

// ScanOnce 立即执行一轮扫描测试 / 手动触发入口
// 返回值: cancelled - 本轮真实取消的单数; skipped - 已被 MQ 处理而跳过的单数
func (s *OrderTimeoutSacnService) ScanOnce() (cancelled, skipped int) {
	release, ok := s.locks.TryLock(scanLockName, scanLockExpiry)
	if !ok {
		return 0, 0
	}
	defer release()

	ids, err := s.orderRepo.ListExpirePendingPay(model.OrderPayTTL, scanBatchSize)
	if err != nil {
		log.Printf("[WARN] 超时扫描: 查询超时订单失败: %v", err)
		return 0, 0
	}
	if len(ids) == 0 {
		return 0, 0
	}
	log.Printf("[INFO] 超时扫描: 命中 %d 张超时未支付订单", len(ids))

	for _, orderId := range ids {
		_, err := s.orderSvc.CancelOrderBySystem(orderId, "系统")
		switch {
		case err == nil:
			cancelled++
			metrics.OrderTimeoutCancelTotal.WithLabelValues("cancelled").Inc()

		case errors.Is(err, model.ErrOrderAlreadyCancelled):
			//订单已取消,目标已达成
			skipped++
			metrics.OrderTimeoutCancelTotal.WithLabelValues("skipped").Inc()

		case errors.Is(err, model.ErrOrderCannotCancel):
			// 状态已不是 pending_pay。绝大多数是"MQ 先取消成功"的正常竞态,
			// 少数是"用户在超时瞬间付款成功" —— 后者值得留意,故打一行日志区分。
			skipped++
			metrics.OrderTimeoutCancelTotal.WithLabelValues("skipped").Inc()
			log.Printf("[INFO] 超时扫描: 订单状态已变更,跳过 orderId=%d err=%v", orderId, err)

		default:
			// 真故障(DB 超时等):下轮还会重扫到,值得告警
			metrics.OrderTimeoutCancelTotal.WithLabelValues("failed").Inc()
			log.Printf("[WARN] 超时扫描: 取消失败(下轮重试) orderId=%d err=%v", orderId, err)
		}
	}
	return cancelled, skipped
}

func (s *OrderTimeoutSacnService) run() {
	s.ScanOnce()
}
