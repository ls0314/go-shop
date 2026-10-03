package task

import (
	"demo-shop-back/src/infra/mq"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/service"
	"time"
)

// Init 启动全部后台定时任务。
// 接收值：deps - 服务层依赖（由 main 的 composition root 构造后注入）
//
// 已迁 product-service 的两个任务(阶段 C2,DS-A-26 §3)不在这里:
//   - ES 数据对账(原 reconcile.go):它要读 sys_product_spu / sys_product_sku,
//     而这两张表的所有权在 product-service 的库里;
//   - 库存闸门对账(原 stock_reconcile.go):同理。
//
// 仍在单体的两个任务都是**订单/券域**的,与商品表所有权无关:
//   - 券闸门对账:读 user_coupon / coupon_template(C3 迁 marketing-service 时随迁);
//   - 订单超时扫描:读 user_order_master(C4 迁 trade-service 时随迁)。
func Init(deps service.ServiceDeps) {
	couponReconcile := NewCouponReconcileService(deps)
	go couponReconcile.Start(5 * time.Minute)
	orderTimeoutScan := NewOrderTimeoutScanService(deps)
	go orderTimeoutScan.Start(scanInterval)

	if mqIns := deps.MQ; mqIns != nil {
		lockMgr := NewDistributedLockManager(deps.Cache)
		mq.StartOutboxDispatcher(mqIns, repository.NewOutboxMessage(deps.DB), lockMgr)
	}
}
