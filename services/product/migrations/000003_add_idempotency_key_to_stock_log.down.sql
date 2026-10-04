-- 000003 down:回滚幂等键独立列
DROP INDEX IF EXISTS idx_stock_log_order_no;
DROP INDEX IF EXISTS uk_stock_log_idem;
ALTER TABLE sys_product_stock_log DROP COLUMN IF EXISTS order_no;
ALTER TABLE sys_product_stock_log DROP COLUMN IF EXISTS idempotency_key;
