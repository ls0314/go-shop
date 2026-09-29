-- ============================================================
-- 000012: 数据权限(scope)管理菜单 —— 挂在"系统管理"下
-- 权限码 system:scope:* 已在 000011 建好,此处仅补菜单+绑定+分配
-- 幂等:WHERE NOT EXISTS / ON CONFLICT DO NOTHING
-- ============================================================

-- 1. 新增菜单节点:数据权限(/admin/system/scope)
WITH sys_menu_parent AS (
    SELECT menu_id FROM sys_menu WHERE route_path = '/admin/system' LIMIT 1
)
INSERT INTO sys_menu (parent_id, menu_name, menu_type, icon, route_path, component, is_visible, is_cache, sort_order, meta_info)
SELECT
    p.menu_id,
    '数据权限',
    'M',
    'lock',
    '/admin/system/scope',
    'system/scope/index',
    TRUE,
    TRUE,
    6,
    '{"title":"数据权限","icon":"lock","noCache":false}'::jsonb
FROM sys_menu_parent p
WHERE NOT EXISTS (SELECT 1 FROM sys_menu m WHERE m.route_path = '/admin/system/scope');

-- 2. 按钮(F)菜单:挂"数据权限"列表下
WITH list_menu AS (
    SELECT menu_id FROM sys_menu WHERE route_path = '/admin/system/scope' LIMIT 1
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
FROM list_menu l
         CROSS JOIN button_data b
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu m
    WHERE m.parent_id = l.menu_id AND m.menu_name = b.menu_name AND m.menu_type = 'F'
);

-- 3. 列表菜单绑定 system:scope:view,按钮绑定对应动作
INSERT INTO sys_menu_permission (menu_id, permission_id)
SELECT m.menu_id, p.permission_id
FROM sys_menu m
         JOIN sys_permission p ON (
            (m.route_path = '/admin/system/scope' AND p.permission_code = 'system:scope:view')
         )
WHERE m.menu_type = 'M'
ON CONFLICT (menu_id, permission_id) DO NOTHING;

WITH btn_map AS (
    SELECT b.menu_id AS button_menu_id, m.route_path AS parent_route, b.menu_name
    FROM sys_menu b JOIN sys_menu m ON b.parent_id = m.menu_id
    WHERE b.menu_type = 'F'
)
INSERT INTO sys_menu_permission (menu_id, permission_id)
SELECT b.button_menu_id, p.permission_id
FROM btn_map b
         JOIN sys_permission p ON (
            b.parent_route = '/admin/system/scope' AND
            ((b.menu_name = '新增'     AND p.permission_code = 'system:scope:create') OR
             (b.menu_name = '编辑'     AND p.permission_code = 'system:scope:update') OR
             (b.menu_name = '删除'     AND p.permission_code = 'system:scope:delete') OR
             (b.menu_name = '查看详情' AND p.permission_code = 'system:scope:view'))
         )
ON CONFLICT (menu_id, permission_id) DO NOTHING;

-- 4. 平台超级管理员分配该菜单(LIKE '/admin%' 在 000011 已含历史,此处单独补新节点)
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT r.role_id, m.menu_id
FROM sys_role r
         CROSS JOIN (SELECT menu_id FROM sys_menu WHERE route_path LIKE '/admin/system/scope%') m
WHERE r.role_name = '平台超级管理员'
ON CONFLICT (role_id, menu_id) DO NOTHING;

-- 5. 修正序列
SELECT setval(
               pg_get_serial_sequence('sys_menu', 'menu_id'),
               GREATEST((SELECT MAX(menu_id) FROM sys_menu), 1),
               TRUE
       );
