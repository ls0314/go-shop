package task

import (
	"demo-shop-back/src/infra"
	"demo-shop-back/src/infra/mq"
	"demo-shop-back/src/repository"
	"time"
)

func Init() {
	reconcile := NewReconcileService()
	go reconcile.Start(5*time.Minute, 12*time.Hour)
	stockReconcile := NewStockReconcileService()
	go stockReconcile.Start(5 * time.Minute)
	orderTimeoutScan := NewOrderTimeoutScanService()
	go orderTimeoutScan.Start(scanInterval)

	if mqIns := infra.GetMQ(); mqIns != nil {
		lockMgr := NewDistributedLockManager(infra.GetCache())
		mq.StartOutboxDispatcher(mqIns, repository.NewOutboxMessage(), lockMgr)
	}
}
