package model

// 订单域的实体(UserOrder / UserOrderDetail / UserOrderLog)已删除。
//
// 三张表已随 C4 迁到 trade-service 的独立库 trade_db,对应的
// service / repository / task 代码也都删了,而 `demo_shop` 里的表本身
// 由 `db/migrations/000017_retire_order_domain_tables` DROP 掉了。
//
// 留着这些结构体是有害的:它们带着
// `TableName() = "user_order_master"` 与 `primary_key;AUTO_INCREMENT` 这类
// "权威表"的写法,读者无法从代码看出那已经是张不存在的表 ——
// 本次就这么被误导过一次(旧版 trade 侧 CreateOrder 把延迟取消消息
// 写进了本库的 sys_outbox_message,而单体投递器去投一条没人监听的消息)。
//
// 订单域的实体现在在 `services/trade/internal/model/order.go`,
// 落库形状与本文件原先那份有两处**刻意**的差异:
//
//   - trade 侧多了 `ExpireAt`(支付截止时间)。单体时代这个阈值散在
//     三处各自解释(MQ 延迟 TTL 15 分钟、rabbitmq.go 注释写 2 分钟、
//     order_service.go 又是 15 分钟),迁出后收敛为"下单时算一次写进
//     expire_at,一切判据都读那一列"。
//   - trade 侧**去掉了** `IsDeleted`。订单是历史凭证,不能软删 ——
//     软删会让"这张单去哪了"变成需要过滤条件才能回答的问题,
//     而订单表本来也没有删除入口。
//
// 地址快照结构(原先的 AddressSnap)也一并删除:它只被下单流程使用,
// 而那个结构现在在 `tradeclient.AddressSnapshot`(HTTP 层 → trade 的入参)
// 与 `trade/internal/model.AddressSnap`(落库形状)。两处形状一致,
// 但分属两侧的契约,不再共用本包的类型。
