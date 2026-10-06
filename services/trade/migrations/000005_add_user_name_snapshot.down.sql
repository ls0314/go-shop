-- 回滚 000005:去掉两张表的用户名快照列。
--
-- 注意这会**丢数据** —— username 是快照,没有其它地方能还原
-- (反查 sys_user 只能拿到当前用户名,拿不到"下单那一刻"的)。

ALTER TABLE user_order_master
    DROP COLUMN IF EXISTS username;

ALTER TABLE user_payment_record
    DROP COLUMN IF EXISTS username;
