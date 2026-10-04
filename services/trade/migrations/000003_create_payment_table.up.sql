-- ============================================================
-- 000003: 支付流水表(user_payment_record)
--
-- 从单体 db/migrations/000009 平移,两处**刻意的删减**:
--   ① 去掉 fk_pay_user_id → sys_user(user_db);
--   ② 去掉权限点/菜单/角色绑定 seed(那些表归 user-service)。
--   保留 fk_pay_order_id → user_order_master:同库,是真正的完整性约束。
--
-- 保留 uk_pay_no 与 uk_pay_trade_no 两条唯一索引 —— 支付回调的幂等靠它们:
-- 同一笔渠道交易号只能落一次,重复回调由索引拒绝。
--
-- 幂等:CREATE ... IF NOT EXISTS;约束包在 DO 块里(重跑不报重复)。
-- ============================================================

CREATE TABLE IF NOT EXISTS user_payment_record (
    payment_id BIGSERIAL PRIMARY KEY,
    pay_no VARCHAR(64) NOT NULL,
    order_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    pay_method VARCHAR(20) NOT NULL DEFAULT 'mock',
    pay_amount DECIMAL(10, 2) NOT NULL,
    pay_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    trade_no VARCHAR(64) DEFAULT NULL,
    pay_time TIMESTAMP DEFAULT NULL,
    notify_log JSONB DEFAULT NULL,
    expire_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN user_payment_record.payment_id IS '主键，支付记录唯一标识';
COMMENT ON COLUMN user_payment_record.pay_no IS '支付流水号（内部生成），全局唯一';
COMMENT ON COLUMN user_payment_record.order_id IS '关联订单ID，关联user_order_master';
COMMENT ON COLUMN user_payment_record.user_id IS '支付用户ID(跨库引用 user_db.sys_user,无外键)';
COMMENT ON COLUMN user_payment_record.pay_method IS '支付方式：mock/alipay/wechat';
COMMENT ON COLUMN user_payment_record.pay_amount IS '支付金额';
COMMENT ON COLUMN user_payment_record.pay_status IS '支付状态';
COMMENT ON COLUMN user_payment_record.trade_no IS '第三方支付交易号（模拟支付时为空）';
COMMENT ON COLUMN user_payment_record.pay_time IS '支付完成时间';
COMMENT ON COLUMN user_payment_record.notify_log IS '回调通知原始数据';
COMMENT ON COLUMN user_payment_record.expire_at IS '支付过期时间(与订单 expire_at 同源)';
COMMENT ON COLUMN user_payment_record.created_at IS '创建时间';
COMMENT ON COLUMN user_payment_record.updated_at IS '更新时间';

-- 支付流水号唯一
CREATE UNIQUE INDEX IF NOT EXISTS uk_pay_no ON user_payment_record (pay_no);
-- 渠道交易号唯一(部分索引):**支付回调幂等的凭据**。
-- 同一笔渠道交易只能落一次;NULL(未回调)不参与唯一性判断。
-- 单体只有 pay_no 唯一,回调重复投递靠"状态已 success 就跳过"这种软判断 ——
-- 拆出后回调是跨服务入口,必须有硬约束兜底。
CREATE UNIQUE INDEX IF NOT EXISTS uk_pay_trade_no
    ON user_payment_record (trade_no) WHERE trade_no IS NOT NULL;
-- 按订单查支付
CREATE INDEX IF NOT EXISTS idx_payment_order_id ON user_payment_record (order_id);
-- 按用户查支付
CREATE INDEX IF NOT EXISTS idx_payment_user_id ON user_payment_record (user_id);
-- 查询待支付过期记录
CREATE INDEX IF NOT EXISTS idx_payment_status ON user_payment_record (pay_status, expire_at);

-- 约束:重跑迁移时不报 "constraint already exists"
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_pay_status') THEN
        ALTER TABLE user_payment_record ADD CONSTRAINT ck_pay_status
            CHECK (pay_status IN ('pending', 'success', 'failed', 'closed'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_pay_method') THEN
        ALTER TABLE user_payment_record ADD CONSTRAINT ck_pay_method
            CHECK (pay_method IN ('mock', 'alipay', 'wechat'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_pay_amount') THEN
        ALTER TABLE user_payment_record ADD CONSTRAINT ck_pay_amount CHECK (pay_amount > 0);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_pay_order_id') THEN
        ALTER TABLE user_payment_record ADD CONSTRAINT fk_pay_order_id
            FOREIGN KEY (order_id) REFERENCES user_order_master (order_id) ON DELETE CASCADE;
    END IF;
END $$;
