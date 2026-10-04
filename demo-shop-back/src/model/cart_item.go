package model

// 购物车实体(UserCartItem)已删除。
//
// `user_cart_item` 表已随 C4 迁到 trade-service 的 trade_db,
// `demo_shop` 里的表由
// `db/migrations/000017_retire_order_domain_tables` DROP 掉了。
//
// 购物车实体现在在 `services/trade/internal/model/cart.go`,那边有两处
// **刻意**的差异:
//
//   - 那边把"购物车行"(UserCartItem)与"购物车行 + 商品回填字段"
//     (CartItemView)分成了两个类型。本文件原先只有一个,而
//     name/image/price/stock 这些字段**不落库** —— 混在一个结构体里
//     会让"实体 ↔ 表"的对应失真:一旦有人拿它去 Create/Update,
//     GORM 会尝试写不存在的列。
//   - 那边不给购物车行存商品快照(只存 sku_id 与数量)。若存快照,
//     商品改价后购物车显示旧价而结算按新价 —— 页面与结算对不上。
//
// 下单入口的请求体绑定改用局部结构(见 handler 包),不再借用实体类型:
// 实体是"一行数据"的形状,而请求体是"一次调用"的形状,两者混用会让
// "哪些字段真的会落库"变得看不出来。
