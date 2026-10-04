-- 回滚 000017:重建 demo_shop 的订单域空表
--
-- **注意:只能重建表结构,数据回不来** —— 000017 是 DROP,不是改名。
-- 若将来需要这些数据,只能从备份恢复,或从 trade_db 的对应表回灌
-- (但那是另一套 schema:trade 侧多了 expire_at、去了 is_deleted,
-- 见 services/trade/migrations/000002_create_order_table.up.sql)。
--
-- 表结构以 000008_create_order_table.up.sql 为准,此处**不复制一份**:
-- 迁移文件是历史记录,复制会得到两份会各自漂移的 DDL。
-- 需要完整回滚时,把 000008 的 up 段重跑一遍,或按下面的顺序手工重建。
--
-- 重建顺序:先父后子(与 up 相反),因为有外键依赖。

-- ---- 1. 主表(被下面三张引用)----
-- user_order_master:见 000008

-- ---- 2. 子表 ----
-- user_order_detail / user_order_log / user_payment_record:见 000008

-- ---- 3. 购物车 ----
-- user_cart_item:见 000008

-- 清掉的消息无法恢复(它们本来也投不出去):
--   DELETE FROM sys_outbox_message WHERE event_type IN ('OrderDelayCancel','A4Test');
COMMENT ON TABLE sys_outbox_message IS
    '事务性发件箱。000017 起订单域的延迟取消消息迁往 trade_db.sys_outbox_message。';
