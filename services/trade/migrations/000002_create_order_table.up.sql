-- ============================================================
-- 000002: 订单三表(user_order_master / user_order_detail / user_order_log)
--
-- 从单体 db/migrations/000008 平移,一处**刻意的删减** + 一处**新增列**:
--
-- ① 删:fk_order_user_id → sys_user(user_db)、以及权限点/菜单/角色绑定的 seed
--    (sys_permission / sys_menu / sys_role_* 全部归 user-service)。明细与日志
--    两张表对订单主表的外键**保留** —— 它们同库,是真正的完整性约束。
--
-- ② 新增 expire_at(支付截止时间,NOT NULL):
--    超时阈值的**单一事实源**。单体时代这个阈值散在三处(MQ 延迟 TTL 15 分钟、
--    rabbitmq.go 注释写 2 分钟、order_service.go 又是 15 分钟),三处并存必然漂移。
--    迁出后收敛为:MQ 延迟 TTL 与响应 PayExpireAt 都从本列派生。
--
-- ③ 去掉 is_deleted:单体的软删除列在订单域**没有任何查询用到**,
--    带上它只会让每个查询都要记得加条件。订单的"删除"语义由 cancelled 状态承担。
--
-- 幂等:CREATE ... IF NOT EXISTS;约束包在 DO 块里(重跑不报重复)。
-- ============================================================

CREATE TABLE IF NOT EXISTS user_order_master (
    order_id BIGSERIAL PRIMARY KEY,
    order_no VARCHAR(32) NOT NULL,
    user_id BIGINT NOT NULL,
    order_status VARCHAR(30) NOT NULL DEFAULT 'pending_pay',
    total_amount DECIMAL(10, 2) NOT NULL,
    pay_amount DECIMAL(10, 2) NOT NULL,
    pay_method VARCHAR(20) DEFAULT NULL,
    pay_time TIMESTAMP DEFAULT NULL,
    address_snapshot jsonb NOT NULL,
    buyer_remark VARCHAR(500) DEFAULT NULL,
    idempotent_key VARCHAR(64) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    detail_count BIGINT NOT NULL,
    first_image VARCHAR(500) NOT NULL,

    -- 支付截止时间。超时扫描、MQ 延迟消息、前端倒计时三处都由它派生
    expire_at TIMESTAMP NOT NULL
);

COMMENT ON COLUMN user_order_master.order_id IS '主键，订单内部唯一标识';
COMMENT ON COLUMN user_order_master.order_no IS '订单业务号';
COMMENT ON COLUMN user_order_master.user_id IS '下单用户ID(跨库引用 user_db.sys_user,无外键)';
COMMENT ON COLUMN user_order_master.order_status IS '订单状态';
COMMENT ON COLUMN user_order_master.total_amount IS '商品总金额（原价合计）';
COMMENT ON COLUMN user_order_master.pay_amount IS '实付金额';
COMMENT ON COLUMN user_order_master.pay_method IS '支付方式：mock/alipay/wechat';
COMMENT ON COLUMN user_order_master.pay_time IS '支付完成时间';
COMMENT ON COLUMN user_order_master.address_snapshot IS '下单时收货地址完整快照';
COMMENT ON COLUMN user_order_master.buyer_remark IS '买家备注';
COMMENT ON COLUMN user_order_master.idempotent_key IS '幂等键，用于防重复下单';
COMMENT ON COLUMN user_order_master.created_at IS '下单时间';
COMMENT ON COLUMN user_order_master.updated_at IS '最后更新时间';
COMMENT ON COLUMN user_order_master.detail_count IS '订单包含的商品件数 冗余字段';
COMMENT ON COLUMN user_order_master.first_image IS '首张商品图片 冗余字段';
COMMENT ON COLUMN user_order_master.expire_at IS '支付截止时间(超时阈值的单一事实源)';

-- 业务号唯一
CREATE UNIQUE INDEX IF NOT EXISTS uk_order_no ON user_order_master (order_no);
-- 幂等键唯一。**同一用户同一 key 只建一单**靠它兜底:
-- 服务端先查后插,并发下查不到也不能保证唯一,最终由索引拒绝重复插入
CREATE UNIQUE INDEX IF NOT EXISTS uk_idempotent_key ON user_order_master (idempotent_key);
-- 用户订单列表
CREATE INDEX IF NOT EXISTS idx_order_status ON user_order_master (user_id, order_status);
-- 管理端按状态筛选
CREATE INDEX IF NOT EXISTS idx_order_user_id ON user_order_master (order_status, created_at);
-- 按时间范围查询
CREATE INDEX IF NOT EXISTS idx_order_created ON user_order_master (created_at);
-- 超时扫描:**这条索引是扫描任务的性能依据**。
-- 没有它,每轮扫描都要全表扫 pending_pay 才能找出过期单
CREATE INDEX IF NOT EXISTS idx_order_expire ON user_order_master (order_status, expire_at);

CREATE TABLE IF NOT EXISTS user_order_detail (
    detail_id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    sku_id BIGINT NOT NULL,
    spu_name VARCHAR(200) NOT NULL,
    sku_name VARCHAR(300) NOT NULL,
    spec_values jsonb NOT NULL,
    main_image VARCHAR(500) DEFAULT NULL,
    quantity INT NOT NULL,
    unit_price DECIMAL(10, 2) NOT NULL,
    total_price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN user_order_detail.detail_id IS '主键';
COMMENT ON COLUMN user_order_detail.order_id IS '所属订单ID，关联user_order_master';
COMMENT ON COLUMN user_order_detail.sku_id IS '购买SKU ID';
COMMENT ON COLUMN user_order_detail.spu_name IS '商品名称快照';
COMMENT ON COLUMN user_order_detail.sku_name IS 'SKU名称快照';
COMMENT ON COLUMN user_order_detail.spec_values IS '规格值快照';
COMMENT ON COLUMN user_order_detail.main_image IS '商品图片快照';
COMMENT ON COLUMN user_order_detail.quantity IS '购买数量';
COMMENT ON COLUMN user_order_detail.unit_price IS '购买时单价快照';
COMMENT ON COLUMN user_order_detail.total_price IS '明细总价（unit_price × quantity）';
COMMENT ON COLUMN user_order_detail.created_at IS '创建时间';

CREATE INDEX IF NOT EXISTS idx_detail_order_id ON user_order_detail (order_id);

CREATE TABLE IF NOT EXISTS user_order_log (
    log_id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    order_status VARCHAR(30) NOT NULL,
    action VARCHAR(50) NOT NULL,
    operator VARCHAR(100) NOT NULL,
    detail VARCHAR(500) DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN user_order_log.log_id IS '主键';
COMMENT ON COLUMN user_order_log.order_id IS '关联订单ID';
COMMENT ON COLUMN user_order_log.order_status IS '变更后的订单状态';
COMMENT ON COLUMN user_order_log.action IS '操作类型';
COMMENT ON COLUMN user_order_log.operator IS '操作人（用户/管理员/系统）';
COMMENT ON COLUMN user_order_log.detail IS '操作详情';
COMMENT ON COLUMN user_order_log.created_at IS '操作时间';

CREATE INDEX IF NOT EXISTS idx_log_order_id ON user_order_log (order_id, created_at);

-- 约束:重跑迁移时不报 "constraint already exists"
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_order_status') THEN
        ALTER TABLE user_order_master ADD CONSTRAINT ck_order_status
            CHECK (order_status IN ('pending_pay', 'paid', 'shipped', 'completed', 'cancelled'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_price_amount') THEN
        ALTER TABLE user_order_master ADD CONSTRAINT ck_price_amount
            CHECK (total_amount >= 0 AND pay_amount >= 0);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_detail_order_id') THEN
        ALTER TABLE user_order_detail ADD CONSTRAINT fk_detail_order_id
            FOREIGN KEY (order_id) REFERENCES user_order_master (order_id) ON DELETE CASCADE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_log_order_id') THEN
        ALTER TABLE user_order_log ADD CONSTRAINT fk_log_order_id
            FOREIGN KEY (order_id) REFERENCES user_order_master (order_id) ON DELETE CASCADE;
    END IF;
END $$;
