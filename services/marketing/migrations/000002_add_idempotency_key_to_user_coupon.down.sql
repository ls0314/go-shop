-- 000002 down:回滚幂等键列
DROP INDEX IF EXISTS uk_user_coupon_idem;
ALTER TABLE user_coupon DROP COLUMN IF EXISTS idempotency_key;
