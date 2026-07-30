CREATE TABLE IF NOT EXISTS user_payment_record (
    payment_id BIGSERIAL PRIMARY KEY ,
    pay_no VARCHAR(64) UNIQUE NOT NULL ,
    order_id BIGINT NOT NULL ,
    user_id BIGINT NOT NULL ,
    pay_method VARCHAR(20) NOT NULL DEFAULT 'mock',
    pay_amount DECIMAL(10, 2) NOT NULL ,
    pay_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    trade_no VARCHAR DEFAULT NULL,
    pay_time TIMESTAMP DEFAULT NULL,
    notify_log JSONB DEFAULT NULL,
    expire_at TIMESTAMP NOT NULL ,
    created_at TIMESTAMP NOT NULL DEFAULT  CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT  CURRENT_TIMESTAMP
);

COMMENT ON COLUMN user_payment_record.payment_id IS '主键，支付记录唯一标识';
COMMENT ON COLUMN user_payment_record.pay_no IS '支付流水号（内部生成），全局唯一';
COMMENT ON COLUMN user_payment_record.order_id IS '关联订单ID';
COMMENT ON COLUMN user_payment_record.user_id IS '支付用户ID';
COMMENT ON COLUMN user_payment_record.pay_method IS '支付方式：mock/alipay/wechat';
COMMENT ON COLUMN user_payment_record.pay_amount IS '支付金额';
COMMENT ON COLUMN user_payment_record.pay_status IS '支付状态';
COMMENT ON COLUMN user_payment_record.trade_no IS '	第三方支付交易号（模拟支付时为空）';
COMMENT ON COLUMN user_payment_record.pay_time IS '支付完成时间';
COMMENT ON COLUMN user_payment_record.notify_log IS '回调通知原始数据';
COMMENT ON COLUMN user_payment_record.expire_at IS '支付过期时间（创建后15分钟）';
COMMENT ON COLUMN user_payment_record.created_at IS '创建时间';
COMMENT ON COLUMN user_payment_record.updated_at IS '更新时间';

-- 索引设计

-- 唯一索引 支付流水号唯一
CREATE UNIQUE INDEX IF NOT EXISTS uk_pay_no ON user_payment_record (pay_no);
-- 普通索引 按订单查支付
CREATE INDEX IF NOT EXISTS idx_payment_order_id ON user_payment_record (order_id);
-- 普通索引 按用户查支付
CREATE INDEX IF NOT EXISTS idx_payment_user_id ON user_payment_record (user_id);
-- 联合索引 查询待支付过期记录
CREATE INDEX IF NOT EXISTS idx_payment_status ON user_payment_record (pay_status, expire_at);

-- 约束设计

-- 支付状态
ALTER TABLE user_payment_record ADD CONSTRAINT ck_pay_status CHECK (pay_status IN ('pending', 'success', 'failed', 'closed'));
-- 支付方式
ALTER TABLE user_payment_record ADD CONSTRAINT ck_pay_method CHECK (pay_method IN ('mock', 'alipay', 'wechat'));
-- 金额校验
ALTER TABLE user_payment_record ADD CONSTRAINT ck_pay_amount CHECK (pay_amount > 0);
-- 外键 关联用户表
ALTER TABLE user_payment_record ADD CONSTRAINT fk_pay_user_id FOREIGN KEY (user_id) REFERENCES sys_user (user_id) ON DELETE CASCADE;
-- 外键 关联订单主表
ALTER TABLE user_payment_record ADD CONSTRAINT fk_pay_order_id FOREIGN KEY (order_id ) REFERENCES user_order_master (order_id ) ON DELETE CASCADE;

-- 创建系统权限
INSERT INTO sys_permission (
    permission_code,
    permission_name,
    permission_type,
    request_method,
    api_path,
    description,
    is_system
) VALUES
      (
          'platform:pay:view',
          '查看支付记录',
          'api',
          'GET',
          '/api/v1/admin/pay/list',
          '所属模块：支付管理',
          TRUE
      )
ON CONFLICT (permission_code, request_method, api_path) DO NOTHING;

-- ====================== 支付管理菜单 ======================
-- 菜单结构：平台管理 → 支付管理 → 支付列表
-- 权限：平台超级管理员 / 平台运营人员 可见
-- 审核人员不涉及支付查询

-- ====================== 创建支付管理父级菜单 ======================
WITH platform_menu AS (
    SELECT menu_id FROM sys_menu WHERE route_path = '/platform' LIMIT 1
),
     pay_menu AS (
         INSERT INTO sys_menu (
                               parent_id, menu_name, menu_type, icon, route_path, component,
                               is_visible, is_cache, sort_order, meta_info
             )
             SELECT
                 p.menu_id,
                 '支付管理',
                 'M',
                 'Money',
                 '/platform/pay',
                 'ParentView',
                 TRUE,
                 TRUE,
                 5,     -- 排在订单管理(sort_order=4)之后
                 '{"title":"支付管理","icon":"Money","noCache":false}'::jsonb
             FROM platform_menu p
             WHERE NOT EXISTS (SELECT 1 FROM sys_menu WHERE route_path = '/platform/pay')
             RETURNING menu_id
     ),
     pay_menu_id AS (
         SELECT menu_id FROM pay_menu
         UNION ALL
         SELECT menu_id FROM sys_menu WHERE route_path = '/platform/pay' LIMIT 1
     )

-- ====================== 创建支付列表子菜单 ======================
INSERT INTO sys_menu (
    parent_id, menu_name, menu_type, icon, route_path, component,
    is_visible, is_cache, sort_order, meta_info
)
SELECT
    pm.menu_id,
    '支付列表',
    'M',
    NULL,
    '/platform/pay/list',
    'platform/pay/list/index',        -- 前端页面组件路径
    TRUE,
    TRUE,
    1,
    '{"title":"支付列表","icon":"","noCache":false}'::jsonb
FROM pay_menu_id pm
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu WHERE route_path = '/platform/pay/list'
);

-- ====================== 角色-权限绑定 ======================
-- 平台超级管理员：拥有全部支付权限
INSERT INTO sys_role_permission (role_id, permission_id)
SELECT r.role_id, p.permission_id
FROM sys_role r
         JOIN sys_permission p
              ON p.permission_code = 'platform:pay:view'
WHERE r.role_name = '平台超级管理员'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 平台运营人员：拥有查看支付记录权限
INSERT INTO sys_role_permission (role_id, permission_id)
SELECT r.role_id, p.permission_id
FROM sys_role r
         JOIN sys_permission p
              ON p.permission_code = 'platform:pay:view'
WHERE r.role_name = '平台运营人员'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- ====================== 角色-菜单关联 ======================
-- 平台超级管理员：拥有所有平台管理相关菜单（/platform% 通配）
-- 注：如果已有 "超级管理员拥有所有 /platform% 菜单" 的通用规则，
--     此处无需重复插入；否则需显式绑定
WITH platform_pay_menus AS (
    SELECT menu_id FROM sys_menu
    WHERE route_path IN ('/platform/pay', '/platform/pay/list')
)
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT r.role_id, m.menu_id
FROM sys_role r
         CROSS JOIN platform_pay_menus m
WHERE r.role_name = '平台超级管理员'
ON CONFLICT (role_id, menu_id) DO NOTHING;

-- 平台运营人员
WITH allowed_menus AS (
    SELECT menu_id FROM sys_menu
    WHERE route_path IN ('/platform', '/platform/pay', '/platform/pay/list')
)
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT r.role_id, m.menu_id
FROM sys_role r
         CROSS JOIN allowed_menus m
WHERE r.role_name = '平台运营人员'
ON CONFLICT (role_id, menu_id) DO NOTHING;

-- 平台审核人员：不绑定（审核人员不涉及支付查询）