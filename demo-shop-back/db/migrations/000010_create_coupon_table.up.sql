CREATE TABLE  IF NOT EXISTS coupon_template (
    template_id BIGSERIAL PRIMARY KEY ,
    coupon_name VARCHAR(100) NOT NULL ,
    coupon_type VARCHAR(20) NOT NULL ,
    threshold_amount DECIMAL(10, 2) NOT NULL DEFAULT 0,
    discount_amount DECIMAL(10, 2) NOT NULL,
    total_count INT NOT NULL ,
    received_count INT NOT NULL DEFAULT 0,
    per_user_limit INT NOT NULL DEFAULT 1,
    usable_days INT NOT NULL DEFAULT 30,
    start_time TIMESTAMP  DEFAULT NULL,
    end_time TIMESTAMP  DEFAULT NULL,
    is_deleted BOOLEAN NOT NULL  DEFAULT  FALSE,
    created_at TIMESTAMP NOT NULL  DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL  DEFAULT CURRENT_TIMESTAMP
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
COMMENT ON COLUMN coupon_template.is_deleted IS '固定有效期-软删除';
COMMENT ON COLUMN coupon_template.created_at IS '创建时间';
COMMENT ON COLUMN coupon_template.updated_at IS '更新时间';




CREATE TABLE IF NOT EXISTS user_coupon(
    user_coupon_id BIGSERIAL NOT NULL PRIMARY KEY ,
    template_id BIGINT NOT NULL ,
    user_id BIGINT NOT NULL ,
    status VARCHAR(20) NOT NULL  DEFAULT  'unused',
    order_no VARCHAR(100) DEFAULT NULL,
    used_at TIMESTAMP DEFAULT NULL,
    expire_at TIMESTAMP NOT NULL ,
    created_at TIMESTAMP NOT NULL  DEFAULT  CURRENT_TIMESTAMP

);

COMMENT ON COLUMN user_coupon.user_coupon_id IS '主键';
COMMENT ON COLUMN user_coupon.template_id IS '关联模板';
COMMENT ON COLUMN user_coupon.user_id IS '所属用户,外键→sys_user';
COMMENT ON COLUMN user_coupon.status IS 'unused / used / expired';
COMMENT ON COLUMN user_coupon.order_no IS '使用的订单号';
COMMENT ON COLUMN user_coupon.used_at IS '使用时间';
COMMENT ON COLUMN user_coupon.expire_at IS '过期时间(领取时计算)';
COMMENT ON COLUMN user_coupon.created_at IS '创建时间';

CREATE INDEX IF NOT EXISTS idx_user_coupon_status ON user_coupon (user_id, status);
CREATE INDEX IF NOT EXISTS idx_coupon_expire ON user_coupon (expire_at);
CREATE INDEX IF NOT EXISTS idx_coupon_template_id ON user_coupon (template_id);


ALTER TABLE coupon_template ADD CONSTRAINT ck_discount_amount CHECK (discount_amount > 0);
ALTER TABLE coupon_template ADD CONSTRAINT ck_threshold_amount CHECK (threshold_amount >= 0);
ALTER TABLE user_coupon ADD CONSTRAINT ck_coupon_status CHECK (status IN ('unused','used','expired'));
ALTER TABLE coupon_template ADD CONSTRAINT ck_coupon_type CHECK (coupon_type IN ('full_reduction','direct_discount'));

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
        'platform:coupon:create',
        '创建优惠券模板',
        'api',
        'POST',
        '/api/v1/admin/platform/coupons',
        '所属模块：优惠券管理',
        TRUE
    ),
    (
        'platform:coupon:view',
        '查看优惠券模板列表',
        'api',
        'GET',
        '/api/v1/admin/platform/coupons',
        '所属模块：优惠券管理',
        TRUE
    )
ON CONFLICT (permission_code, request_method, api_path) DO NOTHING;

-- =============================================
-- 优惠券模块：菜单 + 权限关联 seed
-- 幂等：全部使用 WHERE NOT EXISTS / ON CONFLICT DO NOTHING
-- =============================================

-- ------------------------------
-- 1. 创建菜单：优惠券管理（挂载在已有的"平台管理"下）
-- 层级：平台管理(/platform) → 优惠券管理(/platform/coupon) → 优惠券列表(/platform/coupon/list)
-- ------------------------------
WITH platform_menu_id AS (
    SELECT menu_id
    FROM sys_menu
    WHERE parent_id = 0
      AND route_path = '/platform'
    LIMIT 1
),
     coupon_menu AS (
         INSERT INTO sys_menu (
             parent_id,
             menu_name,
             menu_type,
             icon,
             route_path,
             component,
             is_visible,
             is_cache,
             sort_order,
             meta_info
         )
         SELECT
             p.menu_id,
             '优惠券管理',
             'M',
             'Ticket',
             '/platform/coupon',
             'ParentView',
             TRUE,
             TRUE,
             3,
             '{"title":"优惠券管理","icon":"Ticket","noCache":false}'::jsonb
         FROM platform_menu_id p
         WHERE NOT EXISTS (
             SELECT 1
             FROM sys_menu
             WHERE route_path = '/platform/coupon'
         )
         RETURNING menu_id
     ),
     coupon_menu_id AS (
         SELECT menu_id FROM coupon_menu
         UNION ALL
         SELECT menu_id
         FROM sys_menu
         WHERE route_path = '/platform/coupon'
         LIMIT 1
     )
INSERT INTO sys_menu (
    parent_id,
    menu_name,
    menu_type,
    icon,
    route_path,
    component,
    is_visible,
    is_cache,
    sort_order,
    meta_info
)
SELECT
    c.menu_id,
    '优惠券列表',
    'M',
    NULL,
    '/platform/coupon/list',
    'platform/coupon/list/index',
    TRUE,
    TRUE,
    1,
    '{"title":"优惠券列表","icon":"","noCache":false}'::jsonb
FROM coupon_menu_id c
WHERE NOT EXISTS (
    SELECT 1
    FROM sys_menu
    WHERE route_path = '/platform/coupon/list'
);

-- ------------------------------
-- 2. 创建按钮菜单（F 类型，挂在"优惠券列表"下）
-- ------------------------------
WITH coupon_list_menu AS (
    SELECT menu_id
    FROM sys_menu
    WHERE route_path = '/platform/coupon/list'
    LIMIT 1
),
     button_data AS (
         SELECT *
         FROM (
                  VALUES
                      ('新增优惠券', 'F', 1, '{"title":"新增优惠券"}'::jsonb),
                      ('查看详情',   'F', 2, '{"title":"查看详情"}'::jsonb)
              ) AS t(menu_name, menu_type, sort_order, meta_info)
     )
INSERT INTO sys_menu (
    parent_id,
    menu_name,
    menu_type,
    icon,
    route_path,
    component,
    is_visible,
    is_cache,
    sort_order,
    meta_info
)
SELECT
    c.menu_id,
    b.menu_name,
    b.menu_type,
    NULL,
    NULL,
    NULL,
    FALSE,
    FALSE,
    b.sort_order,
    b.meta_info
FROM coupon_list_menu c
         CROSS JOIN button_data b
WHERE NOT EXISTS (
    SELECT 1
    FROM sys_menu m
    WHERE m.parent_id = c.menu_id
      AND m.menu_name = b.menu_name
      AND m.menu_type = 'F'
);

-- ------------------------------
-- 3. 菜单-权限绑定：列表菜单绑定 view 权限
-- ------------------------------
INSERT INTO sys_menu_permission (
    menu_id,
    permission_id
)
SELECT
    m.menu_id,
    p.permission_id
FROM sys_menu m
         JOIN sys_permission p
              ON p.permission_code = 'platform:coupon:view'
WHERE m.route_path IN (
    '/platform/coupon',
    '/platform/coupon/list'
)
ON CONFLICT (menu_id, permission_id) DO NOTHING;

-- ------------------------------
-- 4. 按钮菜单-权限绑定：按钮绑定对应权限编码
-- ------------------------------
WITH coupon_list_menu AS (
    SELECT menu_id
    FROM sys_menu
    WHERE route_path = '/platform/coupon/list'
    LIMIT 1
),
     button_permission_map AS (
         SELECT *
         FROM (
                  VALUES
                      ('新增优惠券', 'platform:coupon:create'),
                      ('查看详情',   'platform:coupon:view')
              ) AS t(menu_name, permission_code)
     )
INSERT INTO sys_menu_permission (
    menu_id,
    permission_id
)
SELECT
    m.menu_id,
    p.permission_id
FROM coupon_list_menu c
         JOIN sys_menu m
              ON m.parent_id = c.menu_id
         JOIN button_permission_map bpm
              ON bpm.menu_name = m.menu_name
         JOIN sys_permission p
              ON p.permission_code = bpm.permission_code
WHERE m.menu_type = 'F'
ON CONFLICT (menu_id, permission_id) DO NOTHING;

-- ------------------------------
-- 5. 角色-权限分配
-- 平台超级管理员：全部优惠券权限
-- 平台运营人员 / 平台审核人员：仅查看
-- ------------------------------
INSERT INTO sys_role_permission (
    role_id,
    permission_id
)
SELECT
    r.role_id,
    p.permission_id
FROM sys_role r
         JOIN sys_permission p
              ON p.permission_code IN (
                   'platform:coupon:create',
                   'platform:coupon:view'
                  )
WHERE r.role_name = '平台超级管理员'
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO sys_role_permission (
    role_id,
    permission_id
)
SELECT
    r.role_id,
    p.permission_id
FROM sys_role r
         JOIN sys_permission p
              ON p.permission_code = 'platform:coupon:view'
WHERE r.role_name IN ('平台运营人员', '平台审核人员')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- ------------------------------
-- 6. 角色-菜单分配
-- 平台超级管理员：LIKE '/platform%' 自动包含新菜单
-- 平台运营人员 / 平台审核人员：平台管理 + 优惠券管理 + 优惠券列表
-- ------------------------------
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT r.role_id, m.menu_id
FROM sys_role r
         CROSS JOIN (
    SELECT menu_id FROM sys_menu WHERE route_path LIKE '/platform%'
) m
WHERE r.role_name = '平台超级管理员'
ON CONFLICT (role_id, menu_id) DO NOTHING;

INSERT INTO sys_role_menu (role_id, menu_id)
SELECT r.role_id, m.menu_id
FROM sys_role r
         CROSS JOIN (
    SELECT menu_id FROM sys_menu
    WHERE route_path IN ('/platform', '/platform/coupon', '/platform/coupon/list')
) m
WHERE r.role_name IN ('平台运营人员', '平台审核人员')
ON CONFLICT (role_id, menu_id) DO NOTHING;


