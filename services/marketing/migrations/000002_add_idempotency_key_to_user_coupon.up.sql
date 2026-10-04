-- ============================================================
-- 000002:用户券的幂等键独立成列(idempotency_key)
--
-- 与 product 侧 000003 同一决策(C4):**幂等键是独立的请求标识**,
-- 不让它去兼任订单号。
--
--	order_no         业务追溯字段:这张券用在哪张单上(会被清空、会被复用)
--	idempotency_key  幂等判据:一次核销请求只生效一次
--
-- 为什么不能继续用 order_no 当幂等键:
--   归还(取消订单)会把 order_no 清空,而幂等预检需要"重放同一个键时
--   还能找到那张券"。order_no 一旦清空,反查就落空,重复的补偿请求
--   会找不到目标而报错 —— 补偿必须幂等且宽容,报错会让已经取消成功的
--   订单被报成失败。
--
-- 三处改动,顺序不能反:① 加列 → ② 存量回填 → ③ 建索引
-- ============================================================

ALTER TABLE user_coupon ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(64);

COMMENT ON COLUMN user_coupon.idempotency_key IS
    '幂等键(调用方生成,全局唯一):核销与归还的判据;归还时清空以允许键被复用';

-- ② 存量回填:老数据只有 order_no,把它补进 idempotency_key,
--    这样老记录的重复核销在新索引下依然能被挡住
UPDATE user_coupon
   SET idempotency_key = 'legacy:order:' || order_no
 WHERE idempotency_key IS NULL AND order_no IS NOT NULL;

-- ③ 幂等索引:同一个幂等键最多让一张券生效一次。
--    **必须是部分索引**:归还(或撤销)时把 idempotency_key 置 NULL,
--    NULL 不参与唯一性判断,该键才可能被下一次核销复用
CREATE UNIQUE INDEX IF NOT EXISTS uk_user_coupon_idem
    ON user_coupon (idempotency_key) WHERE idempotency_key IS NOT NULL;

-- 老索引 uk_user_coupon_order_no **保留不动**:
-- 它管的是另一件事 —— "一个订单号最多对应一张券"(防止一张券被记到两张单上)。
-- 这两条约束互相独立,不是替代关系。
