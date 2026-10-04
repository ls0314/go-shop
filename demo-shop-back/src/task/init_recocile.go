package task

import (
	"demo-shop-back/src/service"

	"github.com/zeromicro/go-zero/core/logx"
)

// Init 启动全部后台定时任务。
// 接收值：deps - 服务层依赖（由 main 的 composition root 构造后注入）
//
// **当前没有任何任务可启动** —— 全部随表迁出了。这个函数保留是因为
// 它是 composition root 的固定挂载点(将来本库若有新的本地任务,加在这里)。
//
// 迁出的任务清单,以及**为什么"留着有害"而不只是"没用了"**:
//
//   - ES 数据对账 + 库存闸门对账 → product-service(C2),读 sys_product_*;
//   - 券闸门对账 → marketing-service(C3),读 coupon_template / user_coupon。
//     它曾按"以 DB 为准"把 coupon:stock:* 收敛成 0,于是 marketing 侧
//     刚回填的余量每 5 分钟被抹掉一次,领券全被闸门拒;两边还共用
//     同一个锁名 coupon:reconcile,等于抢同一把锁互相打断。
//   - 订单超时扫描 → trade-service(C4),读 user_order_master。
//     它扫不到任何单(订单已落 trade_db),但两边锁名同为 order:time:scan,
//     会让 trade 侧真正该跑的扫描有一半轮次直接跳过。
//
// **outbox 投递器也已停止**。原先它投递本库 sys_outbox_message 的消息,
// 但那些消息全是订单域的 OrderDelayCancel(旧版 trade 侧建单时写进来的),
// 而单体的 order.dead.queue 消费者已删除 —— 投出去没人收。
// 000017 迁移已清空那张表,这条投递路径也一并停掉。
//
// 订单域的发件箱现在归 trade-service:
//
//	写:trade 建单时与订单同事务写 trade_db.sys_outbox_message
//	发:trade 的 outbox 投递器(internal/task/outboxdispatch.go)
//	收:trade 的死信队列消费端(internal/task/orderdelayconsumer.go)
//
// 注意 `deps.MQ` 与 `deps.Cache` 因此在本服务内已无消费者。
// 它们暂时保留在 ServiceDeps 里:profile/其它中间件可能仍会用到,
// 且删掉会让 deps 的构造与配置文件一起改动,收益不足以匹配风险。
func Init(deps service.ServiceDeps) {
	logx.Info("单体无本地后台任务(订单域已迁 trade-service);outbox 投递器与超时扫描均已停止")
}
