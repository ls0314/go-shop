-- ============================================================
-- 000005: 订单/支付表补用户名快照列
--
-- 落实 DS-A-25 §4.5.2 第 6 条:"订单/支付 JOIN sys_user 取用户名 →
-- **下单时快照**。历史订单不随用户改名而变(产品语义选择,非技术选择)。"
--
-- 与 address_snapshot 同一个模式:调用方(BFF)从权威源(user-service)
-- 拿到值,传给持有方(trade),后者存下来 —— 而不是运行时跨库取。
-- 用户名比地址更简单:CreateOrder 的 user_name 入参本来就传过来了,
-- 只是服务端一直没落库。
--
-- 两条列的语义:
--   user_order_master.username    下单那一刻的用户名
--   user_payment_record.username  支付那一刻的用户名(b 方案:建支付时
--                                 从订单读,保证与订单一致)
--
-- 与 user_id 同库并存是刻意的:user_id 用于关联与归属校验,
-- username 只用于展示 —— 后者是快照,前者是引用。
--
-- **历史数据不处理**(有意为之):存量行的 username 为 ''。
-- 反查 sys_user 需要跨库搬运,而本库是演示数据,不值得为它引入一次
-- 跨库脚本。新订单/新支付从第一行起就是完整的。
--
-- 幂等:ADD COLUMN IF NOT EXISTS;COMMENT 可重复执行。
-- ============================================================

ALTER TABLE user_order_master
    ADD COLUMN IF NOT EXISTS username VARCHAR(100) NOT NULL DEFAULT '';

COMMENT ON COLUMN user_order_master.username IS
    '下单时的用户名快照(DEFAULT 空串是为存量行,新行走 CreateOrder 的入参)';

ALTER TABLE user_payment_record
    ADD COLUMN IF NOT EXISTS username VARCHAR(100) NOT NULL DEFAULT '';

COMMENT ON COLUMN user_payment_record.username IS
    '支付时的用户名快照(建支付时从订单读,与订单一致)';
