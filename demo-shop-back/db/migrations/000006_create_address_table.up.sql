CREATE TABLE IF NOT EXISTS user_address (
    address_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL ,
    receiver_name VARCHAR(50) NOT NULL ,
    receiver_phone VARCHAR(20) NOT NULL ,
    province VARCHAR(50) NOT NULL ,
    city VARCHAR(50) NOT NULL ,
    district VARCHAR(50) NOT NULL ,
    detail_address VARCHAR(200) NOT NULL ,
    postal_code VARCHAR(10) DEFAULT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    address_tag VARCHAR(20) DEFAULT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL  DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL  DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN user_address.address_id IS '主键，地址唯一标识';
COMMENT ON COLUMN user_address.user_id IS '所属用户ID，关联sys_user';
COMMENT ON COLUMN user_address.receiver_name IS '收货人姓名';
COMMENT ON COLUMN user_address.receiver_phone IS '收货人手机号';
COMMENT ON COLUMN user_address.province IS '省份';
COMMENT ON COLUMN user_address.city IS '城市';
COMMENT ON COLUMN user_address.district IS '区/县';
COMMENT ON COLUMN user_address.detail_address IS '详细地址（街道、门牌号等）';
COMMENT ON COLUMN user_address.postal_code IS '邮政编码';
COMMENT ON COLUMN user_address.is_default IS '是否默认地址';
COMMENT ON COLUMN user_address.address_tag IS '地址标签（家/公司/学校）';
COMMENT ON COLUMN user_address.is_deleted IS '软删除标记';
COMMENT ON COLUMN user_address.created_at IS '创建时间';
COMMENT ON COLUMN user_address.updated_at IS '更新时间';

-- 联合索引
-- 按用户查询有效地址
CREATE INDEX IF NOT EXISTS idx_address_user_id ON user_address(user_id, is_deleted);
-- 快速定位默认地址
CREATE INDEX IF NOT EXISTS idx_address_default ON user_address(user_id,  is_default);

-- 约束说明
-- 手机号格式
ALTER TABLE user_address ADD CONSTRAINT ck_receiver_phone CHECK (receiver_phone ~ '^1[3-9]\d{9}$');
-- 收货人非空
ALTER TABLE user_address ADD CONSTRAINT ck_receiver_name CHECK (char_length(receiver_name) > 0);

-- 外键约束：禁止级联删除（用户有地址时禁止删用户）
-- 创建外键，设置级联策略
ALTER TABLE user_address ADD CONSTRAINT fk_address_user_id FOREIGN KEY (user_id) REFERENCES sys_user (user_id) ON DELETE RESTRICT  ON UPDATE CASCADE;

-- 新建系统菜单
WITH
    platform_menu AS(
        SELECT menu_id
        FROM sys_menu
        WHERE route_path = '/platform'
        LIMIT 1
    ),
    product_menu AS (
        INSERT INTO sys_menu(
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
                '库存管理',
                'M',
                'inventory',
                '/platform/inventory',
                'ParentView',
                TRUE,
                TRUE,
                3,
                '{"title":"库存管理","icon":"inventory","noCache":false}'::jsonb
            FROM platform_menu p
            WHERE NOT EXISTS (SELECT 1 FROM sys_menu WHERE route_path = '/platform/inventory')
            RETURNING menu_id
    ),
    product_menu_id AS (
        SELECT menu_id FROM product_menu
        UNION ALL
        SELECT menu_id FROM sys_menu WHERE route_path = '/platform/inventory' LIMIT 1
    ),
    insert_list_menu AS (
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
                pr.menu_id,
                '库存列表',
                'M',
                NULL,
                '/platform/inventory/list',
                'platform/inventory/list/index',
                TRUE,
                TRUE,
                1,
                '{"title":"库存列表","icon":"","noCache":false}'::jsonb
            FROM product_menu_id pr
            WHERE NOT EXISTS (
                SELECT 1 FROM sys_menu WHERE route_path = '/platform/inventory/list'
            )
            RETURNING menu_id
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
        pr.menu_id,
        '库存日志',
        'M',
        NULL,
        '/platform/inventory/log',
        'platform/inventory/log/index',
        TRUE,
        TRUE,
        2,
        '{"title":"库存日志","icon":"","noCache":false}'::jsonb
    FROM product_menu_id pr
    WHERE NOT EXISTS (
        SELECT 1 FROM sys_menu WHERE route_path = '/platform/inventory/log'
    );


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
          'platform:inventory:adjust',
          '手动调整库存',
          'api',
          'POST',
          '/api/v1/admin/inventory/adjust',
          '所属模块：库存管理',
          TRUE
      ),
      (
          'platform:inventory:log',
          '查看库存日志',
          'api',
          'GET',
          '/api/v1/admin/inventory/log',
          '所属模块：库存管理',
          TRUE
      ),
      (
          'platform:inventory:view',
          '查看库存信息',
          'api',
          'GET',
          '/api/v1/admin/inventory/sku/:id',
          '所属模块：库存管理',
          TRUE
      ),
      (
          'platform:inventory:view',
          '查看库存信息',
          'api',
          'GET',
          '/api/v1/admin/inventory/spu/:id',
          '所属模块：库存管理',
          TRUE
      ),
      (
          'platform:product:view',
          '查看低库存预警信息',
          'api',
          'GET',
          '/api/v1/admin/inventory/warning',
          '所属模块：库存管理',
          TRUE
      )
ON CONFLICT (permission_code, request_method, api_path) DO NOTHING;

-- 菜单权限绑定数据
INSERT INTO sys_menu_permission (
    menu_id,
    permission_id
)
SELECT
    m.menu_id,
    p.permission_id
FROM sys_menu m
         JOIN sys_permission p
              ON p.permission_code = 'platform:inventory:view'
WHERE m.route_path IN (
                       '/platform/inventory',
                       '/platform/inventory/list'
    )
ON CONFLICT (menu_id, permission_id) DO NOTHING;

INSERT INTO sys_menu_permission (
        menu_id,
        permission_id)
SELECT
    m.menu_id,
    p.permission_id
FROM sys_menu m,
     sys_permission p
WHERE m.route_path = '/platform/inventory/log'
  AND p.permission_code = 'platform:inventory:log'
ON CONFLICT (menu_id, permission_id) DO NOTHING;

-- 平台超级管理员：拥有 platform:inventory:* 对应的全部库存权限
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
                                       'platform:inventory:adjust',
                                       'platform:inventory:log',
                                       'platform:inventory:view'
                  )
WHERE r.role_name = '平台超级管理员'
ON CONFLICT (role_id, permission_id) DO NOTHING;


-- 平台运营人员：查看库存更改日志
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
                                       'platform:inventory:log'
                  )
WHERE r.role_name = '平台运营人员'
ON CONFLICT (role_id, permission_id) DO NOTHING;


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

-- 平台运营人员：拥有平台管理和商品管理菜单
WITH allowed_menus AS (
    SELECT menu_id FROM sys_menu
    WHERE route_path IN ('/platform', '/platform/inventory', '/platform/inventory/log')
)
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT r.role_id, m.menu_id
FROM sys_role r
         CROSS JOIN allowed_menus m
WHERE r.role_name = '平台运营人员'
ON CONFLICT (role_id, menu_id) DO NOTHING;






