package task

import (
	"demo-shop-back/src/infra/mq"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/service"
)

// Init 启动全部后台定时任务。
// 接收值：deps - 服务层依赖（由 main 的 composition root 构造后注入）
//
// 已迁出的任务都不在这里:
//   - ES 数据对账(原 reconcile.go)+ 库存闸门对账(原 stock_reconcile.go)
//     → product-service(C2)。理由:它们要读 sys_product_spu / sys_product_sku,
//     那两张表的所有权在 product-service 的库里;
//   - **券闸门对账(原 coupon_reconcile.go)→ marketing-service(C3)**。
//     它读 coupon_template / user_coupon,表已迁 marketing_db。
//
// 券对账**必须**随表一起走,留在单体是有害的而不只是无用:它读本库那张
// 永远为空的表,会把 coupon:stock:* 一致地"收敛"成 0 —— 于是
// marketing-service 刚回填的余量每 5 分钟被抹掉一次,领券全被闸门拒。
// 两边还共用同一个锁名 coupon:reconcile,等于抢同一把锁互相打断。
//
// 仍在单体的只有订单超时扫描(读 user_order_master,C4 迁 trade-service 时随迁)。
func Init(deps service.ServiceDeps) {
	orderTimeoutScan := NewOrderTimeoutScanService(deps)
	go orderTimeoutScan.Start(scanInterval)

	if mqIns := deps.MQ; mqIns != nil {
		lockMgr := NewDistributedLockManager(deps.Cache)
		mq.StartOutboxDispatcher(mqIns, repository.NewOutboxMessage(deps.DB), lockMgr)
	}
}
