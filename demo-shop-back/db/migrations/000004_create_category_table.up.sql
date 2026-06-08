CREATE TABLE IF NOT EXISTS sys_category (
    category_id BIGSERIAL PRIMARY KEY,
    parent_id BIGINT NOT NULL DEFAULT 0,
    category_name VARCHAR(100) NOT NULL,
    category_level SMALLINT NOT NULL DEFAULT 1,
    category_path VARCHAR(500) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    icon_url VARCHAR(500),
    is_leaf BOOLEAN NOT NULL DEFAULT FALSE,
    is_visible BOOLEAN NOT NULL DEFAULT TRUE,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    create_by BIGINT,
    update_by BIGINT,

    CONSTRAINT chk_category_level CHECK (category_level BETWEEN 1 AND 4),
    CONSTRAINT chk_category_status CHECK (status IN ('active', 'disabled')),
    CONSTRAINT chk_category_path CHECK (category_path ~ '^0(,[0-9]+)*$')
);

COMMENT ON COLUMN sys_category.category_id IS '主键，类目唯一标识';
COMMENT ON COLUMN sys_category.parent_id IS '父类目ID，0表示根节点';
COMMENT ON COLUMN sys_category.category_name IS '类目名称';
COMMENT ON COLUMN sys_category.category_level IS '类目层级（1-4）';
COMMENT ON COLUMN sys_category.category_path IS '路径枚举，如0,1,5';
COMMENT ON COLUMN sys_category.sort_order IS '排序权重，越小越靠前';
COMMENT ON COLUMN sys_category.icon_url IS '类目图标URL';
COMMENT ON COLUMN sys_category.is_leaf IS '是否叶子节点';
COMMENT ON COLUMN sys_category.is_visible IS '是否可见';
COMMENT ON COLUMN sys_category.status IS '状态：active/disabled';
COMMENT ON COLUMN sys_category.created_at IS '创建时间';
COMMENT ON COLUMN sys_category.updated_at IS '更新时间';
COMMENT ON COLUMN sys_category.create_by IS '创建人ID';
COMMENT ON COLUMN sys_category.update_by IS '更新人ID';

-- 查询子类目
CREATE INDEX IF NOT EXISTS idx_parent_id ON sys_category (parent_id);
-- 树形递归查询
CREATE INDEX IF NOT EXISTS idx_category_path ON sys_category (category_path);
-- 按层级排序展示
CREATE INDEX IF NOT EXISTS idx_level_sort ON sys_category (category_level, sort_order);

-- 创建系统角色
INSERT INTO sys_role (
    role_name,
    role_type,
    description,
    is_system,
    is_default,
    data_scope
) VALUES
      (
          '平台超级管理员',
          'platform',
          '拥有平台类目管理全部权限',
          TRUE,
          FALSE,
          'all'
      ),
      (
          '平台运营人员',
          'platform',
          '拥有平台类目查看、类目树、子类目查看、类目更新权限',
          TRUE,
          FALSE,
          'platform'
      ),
      (
          '平台审核人员',
          'platform',
          '拥有平台类目查看、类目树、子类目查看权限',
          TRUE,
          FALSE,
          'platform'
      )
ON CONFLICT (role_name) DO NOTHING;

-- 创建菜单
WITH platform_menu AS (
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
            0,
            '平台管理',
            'M',
            'Setting',
            '/platform',
            'Layout',
            TRUE,
            TRUE,
            1,
            '{"title":"平台管理","icon":"Setting","noCache":false}'::jsonb
        WHERE NOT EXISTS (
            SELECT 1
            FROM sys_menu
            WHERE parent_id = 0
              AND route_path = '/platform'
        )
        RETURNING menu_id
),
     platform_menu_id AS (
         SELECT menu_id FROM platform_menu
         UNION ALL
         SELECT menu_id
         FROM sys_menu
         WHERE parent_id = 0
           AND route_path = '/platform'
         LIMIT 1
     ),
     category_menu AS (
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
                 '类目管理',
                 'M',
                 'List',
                 '/platform/category',
                 'ParentView',
                 TRUE,
                 TRUE,
                 1,
                 '{"title":"类目管理","icon":"List","noCache":false}'::jsonb
             FROM platform_menu_id p
             WHERE NOT EXISTS (
                 SELECT 1
                 FROM sys_menu
                 WHERE route_path = '/platform/category'
             )
             RETURNING menu_id
     ),
     category_menu_id AS (
         SELECT menu_id FROM category_menu
         UNION ALL
         SELECT menu_id
         FROM sys_menu
         WHERE route_path = '/platform/category'
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
    '类目列表',
    'M',
    NULL,
    '/platform/category/list',
    'platform/category/list/index',
    TRUE,
    TRUE,
    1,
    '{"title":"类目列表","icon":"","noCache":false}'::jsonb
FROM category_menu_id c
WHERE NOT EXISTS (
    SELECT 1
    FROM sys_menu
    WHERE route_path = '/platform/category/list'
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
          'platform:category:create',
          '创建类目',
          'api',
          'POST',
          '/api/v1/platform/category',
          '所属模块：平台类目管理',
          TRUE
      ),
      (
          'platform:category:update',
          '更新类目',
          'api',
          'PUT',
          '/api/v1/platform/category/:id',
          '所属模块：平台类目管理',
          TRUE
      ),
      (
          'platform:category:delete',
          '删除类目',
          'api',
          'DELETE',
          '/api/v1/platform/category/:id',
          '所属模块：平台类目管理',
          TRUE
      ),
      (
          'platform:category:view',
          '查看类目详情',
          'api',
          'GET',
          '/api/v1/platform/category/:id',
          '所属模块：平台类目管理',
          TRUE
      ),
      (
          'platform:category:tree',
          '查看类目树',
          'api',
          'GET',
          '/api/v1/platform/category/tree',
          '所属模块：平台类目管理',
          TRUE
      ),
      (
          'platform:category:children',
          '查看子类目',
          'api',
          'GET',
          '/api/v1/platform/category/children/:id',
          '所属模块：平台类目管理',
          TRUE
      )
ON CONFLICT (permission_code) DO NOTHING;


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
              ON p.permission_code = 'platform:category:view'
WHERE m.route_path IN (
                       '/platform/category',
                       '/platform/category/list'
    )
ON CONFLICT (menu_id, permission_id) DO NOTHING;


-- 平台超级管理员：拥有 platform:category:* 对应的全部类目权限
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
                   'platform:category:create',
                   'platform:category:update',
                   'platform:category:delete',
                   'platform:category:view',
                   'platform:category:tree',
                   'platform:category:children'
                  )
WHERE r.role_name = '平台超级管理员'
ON CONFLICT (role_id, permission_id) DO NOTHING;


-- 平台运营人员：查看、类目树、子类目、更新
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
                   'platform:category:view',
                   'platform:category:tree',
                   'platform:category:children',
                   'platform:category:update'
                  )
WHERE r.role_name = '平台运营人员'
ON CONFLICT (role_id, permission_id) DO NOTHING;


-- 平台审核人员：查看、类目树、子类目
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
                   'platform:category:view',
                   'platform:category:tree',
                   'platform:category:children'
                  )
WHERE r.role_name = '平台审核人员'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 绑定菜单权限
INSERT INTO sys_menu_permission (
    menu_id,
    permission_id
)
SELECT
    m.menu_id,
    p.permission_id
FROM sys_menu m
         JOIN sys_permission p
              ON p.permission_code = 'platform:category:view'
WHERE m.route_path IN (
       '/platform/category',
       '/platform/category/list'
    )
ON CONFLICT (menu_id, permission_id) DO NOTHING;

-- 绑定按钮权限
WITH category_list_menu AS (
    SELECT menu_id
    FROM sys_menu
    WHERE route_path = '/platform/category/list'
    LIMIT 1
),
     button_data AS (
         SELECT *
         FROM (
                  VALUES
                      ('新增类目', 'F', 1, '{"title":"新增类目"}'::jsonb),
                      ('编辑',     'F', 2, '{"title":"编辑"}'::jsonb),
                      ('删除',     'F', 3, '{"title":"删除"}'::jsonb),
                      ('查看详情', 'F', 4, '{"title":"查看详情"}'::jsonb)
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
FROM category_list_menu c
         CROSS JOIN button_data b
WHERE NOT EXISTS (
    SELECT 1
    FROM sys_menu m
    WHERE m.parent_id = c.menu_id
      AND m.menu_name = b.menu_name
      AND m.menu_type = 'F'
);


-- 3. 绑定按钮菜单和权限编码
WITH category_list_menu AS (
    SELECT menu_id
    FROM sys_menu
    WHERE route_path = '/platform/category/list'
    LIMIT 1
),
     button_permission_map AS (
         SELECT *
         FROM (
                  VALUES
                      ('新增类目', 'platform:category:create'),
                      ('编辑',     'platform:category:update'),
                      ('删除',     'platform:category:delete'),
                      ('查看详情', 'platform:category:view')
              ) AS t(menu_name, permission_code)
     )
INSERT INTO sys_menu_permission (
    menu_id,
    permission_id
)
SELECT
    m.menu_id,
    p.permission_id
FROM category_list_menu c
         JOIN sys_menu m
              ON m.parent_id = c.menu_id
         JOIN button_permission_map bpm
              ON bpm.menu_name = m.menu_name
         JOIN sys_permission p
              ON p.permission_code = bpm.permission_code
WHERE m.menu_type = 'F'
ON CONFLICT (menu_id, permission_id) DO NOTHING;