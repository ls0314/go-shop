CREATE TABLE IF NOT EXISTS user_order_master (
    order_id BIGSERIAL PRIMARY KEY ,
    order_no VARCHAR(32) NOT NULL ,
    user_id BIGINT NOT NULL ,
    order_status VARCHAR(30) NOT NULL DEFAULT 'pending_pay',
    total_amount DECIMAL(10, 2) NOT NULL ,
    pay_amount DECIMAL(10, 2) NOT NULL ,
    pay_method VARCHAR(20) DEFAULT NULL,
    pay_time TIMESTAMP DEFAULT NULL,
    address_snapshot jsonb NOT NULL ,
    buyer_remark VARCHAR(500) DEFAULT NULL,
    idempotent_key VARCHAR(64) NOT NULL ,
    is_deleted BOOLEAN NOT NULL  DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    detail_count BIGINT NOT NULL ,
    first_image VARCHAR(500) NOT NULL
);

COMMENT ON COLUMN user_order_master.order_id IS '主键，订单内部唯一标识';
COMMENT ON COLUMN user_order_master.order_no IS '订单业务号';
COMMENT ON COLUMN user_order_master.user_id IS '下单用户ID，关联sys_user';
COMMENT ON COLUMN user_order_master.order_status IS '订单状态';
COMMENT ON COLUMN user_order_master.total_amount IS '商品总金额（原价合计）';
COMMENT ON COLUMN user_order_master.pay_amount IS '实付金额';
COMMENT ON COLUMN user_order_master.pay_method IS '支付方式：mock/alipay/wechat';
COMMENT ON COLUMN user_order_master.pay_time IS '支付完成时间';
COMMENT ON COLUMN user_order_master.address_snapshot IS '下单时收货地址完整快照';
COMMENT ON COLUMN user_order_master.buyer_remark IS '买家备注';
COMMENT ON COLUMN user_order_master.idempotent_key IS '幂等键，用于防重复下单';
COMMENT ON COLUMN user_order_master.is_deleted IS '软删除标记（用户端删除后隐藏）';
COMMENT ON COLUMN user_order_master.created_at IS '下单时间';
COMMENT ON COLUMN user_order_master.updated_at IS '最后更新时间';

COMMENT ON COLUMN user_order_master.detail_count IS '订单包含的商品件数 冗余字段';
COMMENT ON COLUMN user_order_master.first_image IS '首张商品图片 冗余字段';

-- 唯一索引
-- 业务号唯一
CREATE UNIQUE INDEX IF NOT EXISTS uk_order_no ON user_order_master (order_no);
-- 幂等键唯一
CREATE UNIQUE INDEX IF NOT EXISTS uk_idempotent_key ON user_order_master (idempotent_key);

-- 联合索引
-- 用户订单列表
CREATE INDEX IF NOT EXISTS idx_order_status ON user_order_master (user_id,order_status);
-- 管理端按状态筛选
CREATE INDEX IF NOT EXISTS idx_order_user_id ON user_order_master (order_status,created_at);

-- 普通索引
-- 按时间范围查询
CREATE INDEX IF NOT EXISTS idx_order_created ON user_order_master (created_at);



CREATE TABLE IF NOT EXISTS user_order_detail (
    detail_id BIGSERIAL PRIMARY KEY ,
    order_id BIGINT NOT NULL ,
    sku_id BIGINT NOT NULL ,
    spu_name VARCHAR(200) NOT NULL ,
    sku_name VARCHAR(300) NOT NULL ,
    spec_values jsonb NOT NULL ,
    main_image VARCHAR(500) DEFAULT NULL,
    quantity INT NOT NULL ,
    unit_price DECIMAL(10, 2) NOT NULL ,
    total_price DECIMAL(10,2) NOT NULL ,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN user_order_detail.detail_id IS '主键';
COMMENT ON COLUMN user_order_detail.order_id IS '所属订单ID，关联order_master';
COMMENT ON COLUMN user_order_detail.sku_id IS '购买SKU ID';
COMMENT ON COLUMN user_order_detail.spu_name IS '商品名称快照';
COMMENT ON COLUMN user_order_detail.sku_name IS 'SKU名称快照';
COMMENT ON COLUMN user_order_detail.spec_values IS '规格值快照';
COMMENT ON COLUMN user_order_detail.main_image IS '商品图片快照';
COMMENT ON COLUMN user_order_detail.quantity IS '购买数量';
COMMENT ON COLUMN user_order_detail.unit_price IS '购买时单价快照';
COMMENT ON COLUMN user_order_detail.total_price IS '明细总价（unit_price × quantity）';
COMMENT ON COLUMN user_order_detail.created_at IS '创建时间';

-- 查订单明细
CREATE INDEX IF NOT EXISTS idx_detail_order_id ON user_order_detail (order_id);

CREATE TABLE IF NOT EXISTS user_order_log(
    log_id BIGSERIAL PRIMARY KEY ,
    order_id BIGINT NOT NULL ,
    order_status VARCHAR(30) NOT NULL ,
    action VARCHAR(50) NOT NULL ,
    operator VARCHAR(100) NOT NULL ,
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

-- 查订单日志
CREATE INDEX IF NOT EXISTS idx_log_order_id ON user_order_log (order_id, created_at);

-- 约束设计
-- 订单状态
ALTER TABLE user_order_master ADD CONSTRAINT ck_order_status CHECK (order_status IN ('pending_pay','paid','shipped','completed','cancelled'));
-- 金额非负
ALTER TABLE user_order_master ADD CONSTRAINT ck_price_amount CHECK (total_amount >= 0 AND pay_amount >= 0);

ALTER TABLE user_order_master ADD CONSTRAINT fk_order_user_id FOREIGN KEY (user_id) REFERENCES sys_user (user_id) ON DELETE CASCADE;
ALTER TABLE user_order_detail ADD CONSTRAINT fk_detail_order_id FOREIGN KEY (order_id) REFERENCES user_order_master (order_id) ON DELETE CASCADE;
ALTER TABLE user_order_log ADD CONSTRAINT fk_log_order_id FOREIGN KEY (order_id) REFERENCES user_order_master (order_id) ON DELETE CASCADE;


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
          'platform:order:view',
          '查看订单',
          'api',
          'GET',
          '/api/v1/platform/orders',
          '所属模块：订单管理',
          TRUE
      ),
      (
          'platform:order:view',
          '查看订单',
          'api',
          'GET',
          '/api/v1/platform/orders/:id',
          '所属模块：订单管理',
          TRUE
      ),
      (
          'platform:order:ship',
          '订单发货',
          'api',
          'PUT',
          '/api/v1/platform/orders/:id/ship',
          '所属模块：订单管理',
          TRUE
      )
ON CONFLICT (permission_code, request_method, api_path) DO NOTHING;

-- 创建菜单：平台管理 → 订单管理 → 订单列表
WITH
    platform_menu AS (
        SELECT menu_id FROM sys_menu WHERE route_path = '/platform' LIMIT 1
    ),
    order_menu AS (
        INSERT INTO sys_menu (
            parent_id, menu_name, menu_type, icon, route_path, component,
            is_visible, is_cache, sort_order, meta_info
        )
        SELECT
            p.menu_id,
            '订单管理',
            'M',
            'Document',
            '/platform/order',
            'ParentView',
            TRUE,
            TRUE,
            4,
            '{"title":"订单管理","icon":"Document","noCache":false}'::jsonb
        FROM platform_menu p
        WHERE NOT EXISTS (SELECT 1 FROM sys_menu WHERE route_path = '/platform/order')
        RETURNING menu_id
    ),
    order_menu_id AS (
        SELECT menu_id FROM order_menu
        UNION ALL
        SELECT menu_id FROM sys_menu WHERE route_path = '/platform/order' LIMIT 1
    )
INSERT INTO sys_menu (
    parent_id, menu_name, menu_type, icon, route_path, component,
    is_visible, is_cache, sort_order, meta_info
)
SELECT
    o.menu_id,
    '订单列表',
    'M',
    NULL,
    '/platform/order/list',
    'platform/order/list/index',
    TRUE,
    TRUE,
    1,
    '{"title":"订单列表","icon":"","noCache":false}'::jsonb
FROM order_menu_id o
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu WHERE route_path = '/platform/order/list'
);

-- 菜单权限绑定数据
INSERT INTO sys_menu_permission (menu_id, permission_id)
SELECT m.menu_id, p.permission_id
FROM sys_menu m
         JOIN sys_permission p
              ON p.permission_code = 'platform:order:view'
WHERE m.route_path IN ('/platform/order', '/platform/order/list')
ON CONFLICT (menu_id, permission_id) DO NOTHING;

-- 角色-权限绑定
-- 平台超级管理员：拥有全部订单权限
INSERT INTO sys_role_permission (role_id, permission_id)
SELECT r.role_id, p.permission_id
FROM sys_role r
         JOIN sys_permission p
              ON p.permission_code IN ('platform:order:view', 'platform:order:ship')
WHERE r.role_name = '平台超级管理员'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 平台运营人员：拥有查看和发货权限
INSERT INTO sys_role_permission (role_id, permission_id)
SELECT r.role_id, p.permission_id
FROM sys_role r
         JOIN sys_permission p
              ON p.permission_code IN ('platform:order:view', 'platform:order:ship')
WHERE r.role_name = '平台运营人员'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 平台审核人员：仅拥有查看权限
INSERT INTO sys_role_permission (role_id, permission_id)
SELECT r.role_id, p.permission_id
FROM sys_role r
         JOIN sys_permission p
              ON p.permission_code IN ('platform:order:view')
WHERE r.role_name = '平台审核人员'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 绑定按钮权限
WITH order_list_menu AS (
    SELECT menu_id FROM sys_menu WHERE route_path = '/platform/order/list' LIMIT 1
),
     button_data AS (
         SELECT *
         FROM (
                  VALUES
                      ('查看详情', 'F', 1, '{"title":"查看详情"}'::jsonb),
                      ('发货',     'F', 2, '{"title":"发货"}'::jsonb)
              ) AS t(menu_name, menu_type, sort_order, meta_info)
     )
INSERT INTO sys_menu (
    parent_id, menu_name, menu_type, icon, route_path, component,
    is_visible, is_cache, sort_order, meta_info
)
SELECT
    c.menu_id, b.menu_name, b.menu_type, NULL, NULL, NULL,
    FALSE, FALSE, b.sort_order, b.meta_info
FROM order_list_menu c
         CROSS JOIN button_data b
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu m
    WHERE m.parent_id = c.menu_id
      AND m.menu_name = b.menu_name
      AND m.menu_type = 'F'
);

-- 绑定按钮菜单和权限编码
WITH order_list_menu AS (
    SELECT menu_id FROM sys_menu WHERE route_path = '/platform/order/list' LIMIT 1
),
     button_permission_map AS (
         SELECT *
         FROM (
                  VALUES
                      ('查看详情', 'platform:order:view'),
                      ('发货',     'platform:order:ship')
              ) AS t(menu_name, permission_code)
     )
INSERT INTO sys_menu_permission (menu_id, permission_id)
SELECT m.menu_id, p.permission_id
FROM order_list_menu c
         JOIN sys_menu m ON m.parent_id = c.menu_id
         JOIN button_permission_map bpm ON bpm.menu_name = m.menu_name
         JOIN sys_permission p ON p.permission_code = bpm.permission_code
WHERE m.menu_type = 'F'
ON CONFLICT (menu_id, permission_id) DO NOTHING;

-- ====================== 角色-菜单关联 ======================
-- 平台超级管理员：拥有所有平台管理相关菜单
WITH platform_menus AS (
    SELECT menu_id FROM sys_menu WHERE route_path LIKE '/platform%'
)
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT r.role_id, m.menu_id
FROM sys_role r
         CROSS JOIN platform_menus m
WHERE r.role_name = '平台超级管理员'
ON CONFLICT (role_id, menu_id) DO NOTHING;

-- 平台运营人员：拥有平台管理和订单管理菜单
WITH allowed_menus AS (
    SELECT menu_id FROM sys_menu
    WHERE route_path IN ('/platform', '/platform/order', '/platform/order/list')
)
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT r.role_id, m.menu_id
FROM sys_role r
         CROSS JOIN allowed_menus m
WHERE r.role_name = '平台运营人员'
ON CONFLICT (role_id, menu_id) DO NOTHING;

-- 平台审核人员：拥有平台管理和订单管理菜单
WITH allowed_menus AS (
    SELECT menu_id FROM sys_menu
    WHERE route_path IN ('/platform', '/platform/order', '/platform/order/list')
)
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT r.role_id, m.menu_id
FROM sys_role r
         CROSS JOIN allowed_menus m
WHERE r.role_name = '平台审核人员'
ON CONFLICT (role_id, menu_id) DO NOTHING;


