-- ============================================================
-- 000017: 退役 demo_shop 的订单域表
-- ============================================================
--
-- 背景:订单三表的所有权已迁到 trade-service 的独立库 trade_db(C4),
-- 相关的 service / repository / task 代码也都删了。但**表本身留了下来**,
-- 里面是迁移前的测试数据(10+10+20+1 行)。留着它是纯风险:
--
--   - 它会持续被误认为"权威源"。本次就这么发生过一次:旧版 trade 侧
--     CreateOrder 把延迟取消消息写进了**本库**的 sys_outbox_message,
--     于是单体投递器去投一条没人监听的消息(单体消费者已删)。
--   - 数据字典(000008)仍在描述这些"权威表",读者无法从库里看出它们已退役。
--
-- 故本次直接 DROP,而不是改名保留。取舍:改名可逆(一条 ALTER 还原)、
-- 代价为零,但会得到一个 permanent 的 legacy_* 表继续占着数据字典 ——
-- 而用户已明确"demo-shop 的数据并不重要"。选择彻底。
--
-- **不可逆**:down 迁移只能重建空表结构,数据回不来。
--
-- DROP 顺序:先子后父。user_order_detail / user_order_log /
-- user_payment_record 都有指向 user_order_master 的外键,
-- 先删主表会因依赖而失败(除非 CASCADE,但显式排序更清楚)。
--
-- 幂等:全部 IF EXISTS,重复执行安全。
-- ============================================================

-- ---- 1. 先删引用 user_order_master 的三张子表 ----
DROP TABLE IF EXISTS user_order_detail;
DROP TABLE IF EXISTS user_order_log;
DROP TABLE IF EXISTS user_payment_record;

-- ---- 2. 购物车(引用 sys_product_sku / sys_user,无人引用它)----
DROP TABLE IF EXISTS user_cart_item;

-- ---- 3. 最后删主表 ----
DROP TABLE IF EXISTS user_order_master;

-- ---- 4. 清掉订单域遗留的僵尸发件箱消息 ----
--
-- 这些行是旧版 trade 侧 CreateOrder 写进**本库**的(那时它连的还是
-- demo_shop),内容是 OrderDelayCancel + order.delay.queue。
-- 现在:
--   - 本库不再产生订单域消息(trade 侧建单写它自己的 trade_db);
--   - 单体的 order.dead.queue 消费者已删除(见 main.go 的说明)。
-- 于是它们**永远投不出去**,而投递器每轮都会去尝试。
--
-- 只删订单域的事件类型,不用 TRUNCATE:sys_outbox_message 是本库自己的
-- 表,将来可能承载别的发件箱消息,清空会误伤。
DELETE FROM sys_outbox_message
WHERE event_type IN ('OrderDelayCancel', 'A4Test');

COMMENT ON TABLE sys_outbox_message IS
    '事务性发件箱(本库 demo_shop 自己的)。订单域的延迟取消消息已随订单表迁往 trade-service 的 trade_db.sys_outbox_message —— 本表不再承载订单域消息。';
