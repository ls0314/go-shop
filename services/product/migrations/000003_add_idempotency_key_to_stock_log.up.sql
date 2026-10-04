-- ============================================================
-- 000003:库存流水的幂等键独立成列(idempotency_key)
--
-- 背景:C4 把下单改成 Saga 之后,库存锁定(S2 锁库存)发生在建单(S3)之前,
-- 而 order_id 是订单表的主键、建单时才分配 —— 也就是说**锁库存时还没有 order_id**。
-- 于是原先把 order_id 当幂等键的做法不再成立。
--
-- 解法不是"换个业务字段兼任幂等键",而是给幂等键一个**独立列**:
-- 它是调用方(前端)生成的一次请求标识,与订单号/订单ID 的职责彻底分开:
--
--	idempotency_key  前端 UUID,全局唯一      → 四操作的幂等判据
--	order_no         服务端雪花号,建单前生成  → 用户可见单号 + 追溯
--	order_id         订单表主键,建单时分配    → 纯主键 + 追溯
--
-- 这样一来 Saga 的补偿只依赖一个"重试时不变"的键,
-- 不再需要"提前生成订单号/订单ID 给下游当幂等键"。
--
-- 三处改动,顺序不能反:
--   ① 加列
--   ② 存量回填:老数据只有 order_id,把它转成字符串补进 idempotency_key,
--      这样老的幂等记录在新索引下依然能挡住重复操作
--   ③ 建新唯一索引
-- ============================================================

ALTER TABLE sys_product_stock_log ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(64);
ALTER TABLE sys_product_stock_log ADD COLUMN IF NOT EXISTS order_no VARCHAR(32);

COMMENT ON COLUMN sys_product_stock_log.idempotency_key IS
    '幂等键(调用方生成,全局唯一):四操作的判据;手工调整为 NULL';
COMMENT ON COLUMN sys_product_stock_log.order_no IS
    '关联订单号(追溯用,服务端雪花号);手工调整为 NULL';

-- ② 存量回填:老流水只有 order_id,把它转成文本补进 idempotency_key。
--    不这么做的话,迁移后同一次操作的重复调用会因为查不到老记录而**再执行一次**
UPDATE sys_product_stock_log
   SET idempotency_key = 'legacy:order:' || order_id::text
 WHERE idempotency_key IS NULL AND order_id IS NOT NULL;

-- ③ 新的幂等索引。**change_type 必须参与**:同一笔订单的四个操作
--    (order_lock / pay_deduct / order_release / refund_release)共用同一个
--    idempotency_key,少了 change_type 会让第二个操作被误判为重复
CREATE UNIQUE INDEX IF NOT EXISTS uk_stock_log_idem
    ON sys_product_stock_log (idempotency_key, sku_id, change_type)
    WHERE idempotency_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_stock_log_order_no ON sys_product_stock_log (order_no);

-- 老索引 uk_stock_log_order **保留**:CreateInventoryLog 用的是
-- OnConflict{DoNothing},GORM 会匹配任意唯一索引,老索引还在就仍能挡住重复。
-- 等确认没有回滚需求后再单独删。
