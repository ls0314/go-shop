-- ============================================================
-- 000001: 券域表(coupon_template / user_coupon)
--
-- 从单体 db/migrations/000010 平移,两处**刻意的删减**:
--   ① 去掉 user_coupon.user_id → sys_user 的外键:
--      sys_user 在 user_db、本服务在 marketing_db,跨库外键建不出来;
--      归属改由应用层保证(核销时校验 user_coupon.user_id == 调用方传入的 user_id)。
--   ② 去掉权限点 / 菜单 / 角色绑定的 seed:
--      那些表(sys_permission / sys_menu / sys_role_*)全部归 user-service,
--      本库没有它们;权限点的补齐见 user-service 的迁移 000017/000018。
--
-- 幂等:CREATE ... IF NOT EXISTS;约束包在 DO 块里(重跑不报重复)。
-- ============================================================

CREATE TABLE IF NOT EXISTS coupon_template (
    template_id BIGSERIAL PRIMARY KEY,
    coupon_name VARCHAR(100) NOT NULL,
    coupon_type VARCHAR(20) NOT NULL,
    threshold_amount DECIMAL(10, 2) NOT NULL DEFAULT 0,
    discount_amount DECIMAL(10, 2) NOT NULL,
    total_count INT NOT NULL,
    received_count INT NOT NULL DEFAULT 0,
    per_user_limit INT NOT NULL DEFAULT 1,
    usable_days INT NOT NULL DEFAULT 30,
    start_time TIMESTAMP DEFAULT NULL,
    end_time TIMESTAMP DEFAULT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN coupon_template.template_id IS '主键';
COMMENT ON COLUMN coupon_template.coupon_name IS '优惠券名称';
COMMENT ON COLUMN coupon_template.coupon_type IS 'full_reduction(满减)/ direct_discount(直减)';
COMMENT ON COLUMN coupon_template.threshold_amount IS '使用门槛金额,0 表示无门槛';
COMMENT ON COLUMN coupon_template.discount_amount IS '优惠金额(满减)或折扣率(直减)';
COMMENT ON COLUMN coupon_template.total_count IS '发放总量';
COMMENT ON COLUMN coupon_template.received_count IS '已领取数量(原子扣减维护)';
COMMENT ON COLUMN coupon_template.per_user_limit IS '每人限领数量';
COMMENT ON COLUMN coupon_template.usable_days IS '领取后有效天数';
COMMENT ON COLUMN coupon_template.start_time IS '固定有效期-开始';
COMMENT ON COLUMN coupon_template.end_time IS '固定有效期-结束';
COMMENT ON COLUMN coupon_template.is_deleted IS '软删除';
COMMENT ON COLUMN coupon_template.created_at IS '创建时间';
COMMENT ON COLUMN coupon_template.updated_at IS '更新时间';

CREATE TABLE IF NOT EXISTS user_coupon (
    user_coupon_id BIGSERIAL NOT NULL PRIMARY KEY,
    template_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'unused',
    order_no VARCHAR(100) DEFAULT NULL,
    used_at TIMESTAMP DEFAULT NULL,
    expire_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN user_coupon.user_coupon_id IS '主键';
COMMENT ON COLUMN user_coupon.template_id IS '关联模板';
COMMENT ON COLUMN user_coupon.user_id IS '所属用户(跨库引用 user_db.sys_user,无外键)';
COMMENT ON COLUMN user_coupon.status IS 'unused / used / expired';
COMMENT ON COLUMN user_coupon.order_no IS '使用的订单号';
COMMENT ON COLUMN user_coupon.used_at IS '使用时间';
COMMENT ON COLUMN user_coupon.expire_at IS '过期时间(领取时计算)';
COMMENT ON COLUMN user_coupon.created_at IS '创建时间';

CREATE INDEX IF NOT EXISTS idx_user_coupon_status ON user_coupon (user_id, status);
CREATE INDEX IF NOT EXISTS idx_coupon_expire ON user_coupon (expire_at);
CREATE INDEX IF NOT EXISTS idx_coupon_template_id ON user_coupon (template_id);

-- 核销幂等 + "按订单反查券"都靠它:同一订单号最多对应一张券。
-- 单体没有这条索引(核销只在本地事务里做,不怕重放);拆出后有 RPC 重试,必须补。
CREATE UNIQUE INDEX IF NOT EXISTS uk_user_coupon_order_no
    ON user_coupon (order_no) WHERE order_no IS NOT NULL;

-- 约束:重跑迁移时不报 "constraint already exists"
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_discount_amount') THEN
        ALTER TABLE coupon_template ADD CONSTRAINT ck_discount_amount CHECK (discount_amount > 0);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_threshold_amount') THEN
        ALTER TABLE coupon_template ADD CONSTRAINT ck_threshold_amount CHECK (threshold_amount >= 0);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_coupon_type') THEN
        ALTER TABLE coupon_template ADD CONSTRAINT ck_coupon_type
            CHECK (coupon_type IN ('full_reduction', 'direct_discount'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_coupon_status') THEN
        ALTER TABLE user_coupon ADD CONSTRAINT ck_coupon_status
            CHECK (status IN ('unused', 'used', 'expired'));
    END IF;
END $$;
