package model

// 支付实体(UserPayment)已删除。
//
// `user_payment_record` 表已随 C4 迁到 trade-service 的 trade_db,
// `demo_shop` 里的表由
// `db/migrations/000017_retire_order_domain_tables` DROP 掉了。
//
// 支付域实体现在在 `services/trade/internal/model/payment.go`。
// 落库形状有一处**刻意**的差异:trade 侧把 `notify_log` 保留为
// 原始回调报文(排障刚需 —— 渠道说"我回调过了"时能拿出当时收到什么),
// 而上面的 `UpdateAdt` 字段名是拼写错误(应为 `UpdatedAt`),
// 迁出时一并修正了。
