-- ============================================================
-- 000011: RBAC 系统管理模块 —— 角色 + 权限点 + 菜单 + 按钮 + 角色分配
-- 模块:用户管理 / 角色管理 / 权限管理 / 菜单管理 / 部门管理 / 数据权限
-- 权限码命名:system:<module>:<action>
-- 幂等:全部使用 ON CONFLICT DO NOTHING / WHERE NOT EXISTS
-- ============================================================

-- ------------------------------------------------------------
-- 一、系统角色 seed(sys_role)
-- 角色是 RBAC 的基础参照数据,先于角色-权限/角色-菜单绑定建立
-- 显式指定 role_id,保证跨库迁移后角色 ID 稳定(绑定关系靠它关联)
-- ------------------------------------------------------------
INSERT INTO sys_role (role_id, role_name, role_type, description, is_system, is_default, data_scope) VALUES
    (1, '平台超级管理员', 'platform', '拥有平台全部权限', TRUE, FALSE, 'all'),
    (2, '平台运营人员',   'platform', '负责日常运营操作', TRUE, FALSE, 'all'),
    (3, '平台审核人员',   'platform', '负责内容与订单审核', TRUE, FALSE, 'all')
ON CONFLICT (role_name) DO NOTHING;

-- ------------------------------------------------------------
-- 二、权限点 seed(sys_permission)
-- api_path 必须与后端 c.FullPath() 完全一致(含 :id 占位符)
-- 同一权限码可对应多个 (method, path);同一 (code,method,path) 唯一
-- ------------------------------------------------------------

-- system:user:view —— 用户列表/详情/角色回显/部门回显
INSERT INTO sys_permission (permission_code, permission_name, permission_type, request_method, api_path, is_system) VALUES
    ('system:user:view',        '查看用户',     'api', 'GET',    '/api/v1/user',              TRUE),
    ('system:user:view',        '查看用户',     'api', 'GET',    '/api/v1/user/:id',          TRUE),
    ('system:user:view',        '查看用户',     'api', 'GET',    '/api/v1/user/:id/role',     TRUE),
    ('system:user:view',        '查看用户',     'api', 'GET',    '/api/v1/user/:id/dept',     TRUE)
ON CONFLICT (permission_code, request_method, api_path) DO NOTHING;

-- system:user:create / update / delete
INSERT INTO sys_permission (permission_code, permission_name, permission_type, request_method, api_path, is_system) VALUES
    ('system:user:create',      '新增用户',     'api', 'POST',   '/api/v1/user',              TRUE),
    ('system:user:update',      '编辑用户',     'api', 'PUT',    '/api/v1/user/:id',          TRUE),
    ('system:user:delete',      '删除用户',     'api', 'DELETE', '/api/v1/user/:id',          TRUE)
ON CONFLICT (permission_code, request_method, api_path) DO NOTHING;

-- system:user:assign-role / assign-dept(用户-角色/部门 分配)
INSERT INTO sys_permission (permission_code, permission_name, permission_type, request_method, api_path, is_system) VALUES
    ('system:user:assign-role', '分配用户角色', 'api', 'POST',   '/api/v1/user/assign-role',        TRUE),
    ('system:user:assign-role', '分配用户角色', 'api', 'DELETE', '/api/v1/user/:id/clear-role',     TRUE),
    ('system:user:assign-dept', '分配用户部门', 'api', 'POST',   '/api/v1/user/assign-dept',        TRUE),
    ('system:user:assign-dept', '分配用户部门', 'api', 'DELETE', '/api/v1/user/:id/clear-dept',     TRUE)
ON CONFLICT (permission_code, request_method, api_path) DO NOTHING;

-- system:role:* —— 角色 CRUD + 角色-菜单/权限分配
INSERT INTO sys_permission (permission_code, permission_name, permission_type, request_method, api_path, is_system) VALUES
    ('system:role:view',        '查看角色',     'api', 'GET',    '/api/v1/role',              TRUE),
    ('system:role:view',        '查看角色',     'api', 'GET',    '/api/v1/role/:id',          TRUE),
    ('system:role:view',        '查看角色',     'api', 'GET',    '/api/v1/role/:id/menu',     TRUE),
    ('system:role:view',        '查看角色',     'api', 'GET',    '/api/v1/role/:id/perm',     TRUE),
    ('system:role:create',      '新增角色',     'api', 'POST',   '/api/v1/role',              TRUE),
    ('system:role:update',      '编辑角色',     'api', 'PUT',    '/api/v1/role/:id',          TRUE),
    ('system:role:delete',      '删除角色',     'api', 'DELETE', '/api/v1/role/:id',          TRUE),
    ('system:role:assign-menu', '分配角色菜单', 'api', 'POST',   '/api/v1/role/assign-menu',        TRUE),
    ('system:role:assign-menu', '分配角色菜单', 'api', 'DELETE', '/api/v1/role/:id/clear-menu',     TRUE),
    ('system:role:assign-perm', '分配角色权限', 'api', 'POST',   '/api/v1/role/assign-perm',        TRUE),
    ('system:role:assign-perm', '分配角色权限', 'api', 'DELETE', '/api/v1/role/:id/clear-perm',     TRUE)
ON CONFLICT (permission_code, request_method, api_path) DO NOTHING;

-- system:perm:* —— 权限点 CRUD
INSERT INTO sys_permission (permission_code, permission_name, permission_type, request_method, api_path, is_system) VALUES
    ('system:perm:view',        '查看权限',     'api', 'GET',    '/api/v1/permissions',       TRUE),
    ('system:perm:view',        '查看权限',     'api', 'GET',    '/api/v1/permissions/:id',   TRUE),
    ('system:perm:create',      '新增权限',     'api', 'POST',   '/api/v1/permissions',       TRUE),
    ('system:perm:update',      '编辑权限',     'api', 'PUT',    '/api/v1/permissions/:id',   TRUE),
    ('system:perm:delete',      '删除权限',     'api', 'DELETE', '/api/v1/permissions/:id',   TRUE)
ON CONFLICT (permission_code, request_method, api_path) DO NOTHING;

-- system:menu:* —— 菜单 CRUD + 菜单-权限分配
INSERT INTO sys_permission (permission_code, permission_name, permission_type, request_method, api_path, is_system) VALUES
    ('system:menu:view',        '查看菜单',     'api', 'GET',    '/api/v1/menu',              TRUE),
    ('system:menu:view',        '查看菜单',     'api', 'GET',    '/api/v1/menu/:id',          TRUE),
    ('system:menu:view',        '查看菜单',     'api', 'GET',    '/api/v1/menu/:id/tree',     TRUE),
    ('system:menu:view',        '查看菜单',     'api', 'GET',    '/api/v1/menu/:id/perm',     TRUE),
    ('system:menu:create',      '新增菜单',     'api', 'POST',   '/api/v1/menu',              TRUE),
    ('system:menu:update',      '编辑菜单',     'api', 'PUT',    '/api/v1/menu/:id',          TRUE),
    ('system:menu:delete',      '删除菜单',     'api', 'DELETE', '/api/v1/menu/:id',          TRUE),
    ('system:menu:assign-perm', '分配菜单权限', 'api', 'POST',   '/api/v1/menu/assign-perm',        TRUE),
    ('system:menu:assign-perm', '分配菜单权限', 'api', 'DELETE', '/api/v1/menu/:id/clear-perm',     TRUE)
ON CONFLICT (permission_code, request_method, api_path) DO NOTHING;

-- system:dept:* —— 部门 CRUD
INSERT INTO sys_permission (permission_code, permission_name, permission_type, request_method, api_path, is_system) VALUES
    ('system:dept:view',        '查看部门',     'api', 'GET',    '/api/v1/dept',              TRUE),
    ('system:dept:view',        '查看部门',     'api', 'GET',    '/api/v1/dept/:id',          TRUE),
    ('system:dept:view',        '查看部门',     'api', 'GET',    '/api/v1/dept/tree/:userId', TRUE),
    ('system:dept:create',      '新增部门',     'api', 'POST',   '/api/v1/dept',              TRUE),
    ('system:dept:update',      '编辑部门',     'api', 'PUT',    '/api/v1/dept/:id',          TRUE),
    ('system:dept:delete',      '删除部门',     'api', 'DELETE', '/api/v1/dept/:id',          TRUE)
ON CONFLICT (permission_code, request_method, api_path) DO NOTHING;

-- system:scope:* —— 数据权限 CRUD
INSERT INTO sys_permission (permission_code, permission_name, permission_type, request_method, api_path, is_system) VALUES
    ('system:scope:view',       '查看数据权限', 'api', 'GET',    '/api/v1/scope',             TRUE),
    ('system:scope:view',       '查看数据权限', 'api', 'GET',    '/api/v1/scope/:id',         TRUE),
    ('system:scope:create',     '新增数据权限', 'api', 'POST',   '/api/v1/scope',             TRUE),
    ('system:scope:update',     '编辑数据权限', 'api', 'PUT',    '/api/v1/scope/:id',         TRUE),
    ('system:scope:delete',     '删除数据权限', 'api', 'DELETE', '/api/v1/scope/:id',         TRUE)
ON CONFLICT (permission_code, request_method, api_path) DO NOTHING;

-- ------------------------------------------------------------
-- 三、系统管理菜单树补充(挂在"系统管理" /admin/system 下)
-- 现有:用户管理(3)、角色管理(4);新增:权限/菜单/部门管理
-- 新增节点不指定 menu_id,由序列自增;WHERE NOT EXISTS 幂等
-- ------------------------------------------------------------
WITH sys_menu_parent AS (
    SELECT menu_id FROM sys_menu WHERE route_path = '/admin/system' LIMIT 1
)
INSERT INTO sys_menu (parent_id, menu_name, menu_type, icon, route_path, component, is_visible, is_cache, sort_order, meta_info)
SELECT
    p.menu_id,
    d.menu_name,
    'M',
    d.icon,
    d.route_path,
    d.component,
    TRUE,
    TRUE,
    d.sort_order,
    d.meta_info::jsonb
FROM sys_menu_parent p
         CROSS JOIN (VALUES
            ('权限管理', 'lock',     '/admin/system/permission', 'system/permission/index', 3, '{"title":"权限管理","icon":"lock","noCache":false}'),
            ('菜单管理', 'menu',     '/admin/system/menu',       'system/menu/index',       4, '{"title":"菜单管理","icon":"menu","noCache":false}'),
            ('部门管理', 'office',   '/admin/system/dept',       'system/dept/index',       5, '{"title":"部门管理","icon":"office","noCache":false}')
         ) AS d(menu_name, icon, route_path, component, sort_order, meta_info)
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu m WHERE m.route_path = d.route_path
);

-- ------------------------------------------------------------
-- 四、按钮(F)菜单:权限/菜单/部门三个列表页各配 新增/编辑/删除/查看详情
-- ------------------------------------------------------------
WITH list_menus AS (
    SELECT menu_id, route_path FROM sys_menu
    WHERE route_path IN ('/admin/system/permission', '/admin/system/menu', '/admin/system/dept')
),
button_data AS (
    SELECT * FROM (VALUES
        ('新增',   'F', 1),
        ('编辑',   'F', 2),
        ('删除',   'F', 3),
        ('查看详情', 'F', 4)
    ) AS t(menu_name, menu_type, sort_order)
)
INSERT INTO sys_menu (parent_id, menu_name, menu_type, icon, route_path, component, is_visible, is_cache, sort_order, meta_info)
SELECT
    l.menu_id,
    b.menu_name,
    b.menu_type,
    NULL,
    NULL,
    NULL,
    FALSE,
    FALSE,
    b.sort_order,
    ('{"title":"' || b.menu_name || '"}' )::jsonb
FROM list_menus l
         CROSS JOIN button_data b
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu m
    WHERE m.parent_id = l.menu_id AND m.menu_name = b.menu_name AND m.menu_type = 'F'
);

-- ------------------------------------------------------------
-- 五、菜单-权限绑定(sys_menu_permission)
-- 列表菜单 → 对应模块 view;按钮 → 对应动作权限码
-- 绑定映射:menu_name 决定 action(view/create/update/delete)
-- ------------------------------------------------------------

-- 4.1 列表菜单绑 view
INSERT INTO sys_menu_permission (menu_id, permission_id)
SELECT m.menu_id, p.permission_id
FROM sys_menu m
         JOIN sys_permission p ON (
            (m.route_path = '/admin/system/permission' AND p.permission_code = 'system:perm:view') OR
            (m.route_path = '/admin/system/menu'        AND p.permission_code = 'system:menu:view') OR
            (m.route_path = '/admin/system/dept'        AND p.permission_code = 'system:dept:view')
         )
WHERE m.menu_type = 'M'
ON CONFLICT (menu_id, permission_id) DO NOTHING;

-- 4.2 按钮绑 create/update/delete/view(父菜单 route_path 决定模块前缀)
WITH btn_map AS (
    SELECT
        b.menu_id AS button_menu_id,
        m.route_path AS parent_route,
        b.menu_name
    FROM sys_menu b
             JOIN sys_menu m ON b.parent_id = m.menu_id
    WHERE b.menu_type = 'F'
)
INSERT INTO sys_menu_permission (menu_id, permission_id)
SELECT b.button_menu_id, p.permission_id
FROM btn_map b
         JOIN sys_permission p ON (
            (b.parent_route = '/admin/system/permission' AND
             ((b.menu_name = '新增'     AND p.permission_code = 'system:perm:create') OR
              (b.menu_name = '编辑'     AND p.permission_code = 'system:perm:update') OR
              (b.menu_name = '删除'     AND p.permission_code = 'system:perm:delete') OR
              (b.menu_name = '查看详情' AND p.permission_code = 'system:perm:view'))) OR
            (b.parent_route = '/admin/system/menu' AND
             ((b.menu_name = '新增'     AND p.permission_code = 'system:menu:create') OR
              (b.menu_name = '编辑'     AND p.permission_code = 'system:menu:update') OR
              (b.menu_name = '删除'     AND p.permission_code = 'system:menu:delete') OR
              (b.menu_name = '查看详情' AND p.permission_code = 'system:menu:view'))) OR
            (b.parent_route = '/admin/system/dept' AND
             ((b.menu_name = '新增'     AND p.permission_code = 'system:dept:create') OR
              (b.menu_name = '编辑'     AND p.permission_code = 'system:dept:update') OR
              (b.menu_name = '删除'     AND p.permission_code = 'system:dept:delete') OR
              (b.menu_name = '查看详情' AND p.permission_code = 'system:dept:view')))
         )
ON CONFLICT (menu_id, permission_id) DO NOTHING;

-- ------------------------------------------------------------
-- 六、角色-权限分配(sys_role_permission)
-- 平台超级管理员:全部 system:* 权限
-- ------------------------------------------------------------
INSERT INTO sys_role_permission (role_id, permission_id)
SELECT r.role_id, p.permission_id
FROM sys_role r
         CROSS JOIN sys_permission p
WHERE r.role_name = '平台超级管理员'
  AND p.permission_code LIKE 'system:%'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- ------------------------------------------------------------
-- 七、角色-菜单分配(sys_role_menu)
-- 平台超级管理员:整棵 /admin 后台管理树(含本次新增),使其能访问系统管理
-- ------------------------------------------------------------
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT r.role_id, m.menu_id
FROM sys_role r
         CROSS JOIN (
    SELECT menu_id FROM sys_menu WHERE route_path LIKE '/admin%'
) m
WHERE r.role_name = '平台超级管理员'
ON CONFLICT (role_id, menu_id) DO NOTHING;

-- ------------------------------------------------------------
-- 八、修正序列,避免后续自增冲突
-- 角色 seed 显式指定了 role_id,序列不会自动跟进,必须补
-- ------------------------------------------------------------
SELECT setval(
               pg_get_serial_sequence('sys_role', 'role_id'),
               GREATEST((SELECT MAX(role_id) FROM sys_role), 1),
               TRUE
       );
SELECT setval(
               pg_get_serial_sequence('sys_menu', 'menu_id'),
               GREATEST((SELECT MAX(menu_id) FROM sys_menu), 1),
               TRUE
       );
SELECT setval(
               pg_get_serial_sequence('sys_permission', 'permission_id'),
               GREATEST((SELECT MAX(permission_id) FROM sys_permission), 1),
               TRUE
       );
