package task

import (
	"demo-shop-back/src/infra/mq"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/service"
	"time"
)

// Init 启动全部后台定时任务。
// 接收值：deps - 服务层依赖（由 main 的 composition root 构造后注入）
func Init(deps service.ServiceDeps) {
	reconcile := NewReconcileService(deps)
	go reconcile.Start(5*time.Minute, 12*time.Hour)
	stockReconcile := NewStockReconcileService(deps)
	go stockReconcile.Start(5 * time.Minute)
	couponReconcile := NewCouponReconcileService(deps)
	go couponReconcile.Start(5 * time.Minute)
	orderTimeoutScan := NewOrderTimeoutScanService(deps)
	go orderTimeoutScan.Start(scanInterval)

	if mqIns := deps.MQ; mqIns != nil {
		lockMgr := NewDistributedLockManager(deps.Cache)
		mq.StartOutboxDispatcher(mqIns, repository.NewOutboxMessage(deps.DB), lockMgr)
	}
}
