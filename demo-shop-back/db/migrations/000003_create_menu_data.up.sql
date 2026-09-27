-- ============================================================
-- 000003: 后台管理菜单树（系统数据，生产必需）
-- 幂等：ON CONFLICT (menu_id) DO UPDATE，可重复执行。
--
-- 说明：
--   1. pgcrypto 扩展声明已前移至 000001（每个库随迁移链自动安装）；
--   2. 演示用户 user02、测试角色 role01/role02 及其绑定等种子数据
--      已拆分至 db/seeds/seed.sql，由开发/测试环境显式加载，
--      生产环境不加载，避免预置弱密码账号。
-- ============================================================
BEGIN;

-- 菜单 1：后台管理，第一层
INSERT INTO sys_menu (
    menu_id,
    parent_id,
    menu_name,
    menu_type,
    icon,
    route_path,
    component,
    is_visible,
    is_cache,
    sort_order,
    meta_info,
    created_at
) VALUES (
             1,
             0,
             '后台管理',
             'M',
             'dashboard',
             '/admin',
             'Layout',
             TRUE,
             TRUE,
             1,
             '{"title": "后台管理", "icon": "dashboard", "noCache": false}'::jsonb,
             CURRENT_TIMESTAMP
         )
ON CONFLICT (menu_id) DO UPDATE SET
                                parent_id = EXCLUDED.parent_id,
                                menu_name = EXCLUDED.menu_name,
                                menu_type = EXCLUDED.menu_type,
                                icon = EXCLUDED.icon,
                                route_path = EXCLUDED.route_path,
                                component = EXCLUDED.component,
                                is_visible = EXCLUDED.is_visible,
                                is_cache = EXCLUDED.is_cache,
                                sort_order = EXCLUDED.sort_order,
                                meta_info = EXCLUDED.meta_info;

-- 菜单 2：系统管理，第二层
INSERT INTO sys_menu (
    menu_id,
    parent_id,
    menu_name,
    menu_type,
    icon,
    route_path,
    component,
    is_visible,
    is_cache,
    sort_order,
    meta_info,
    created_at
) VALUES (
             2,
             1,
             '系统管理',
             'M',
             'system',
             '/admin/system',
             'ParentView',
             TRUE,
             TRUE,
             1,
             '{"title": "系统管理", "icon": "system", "noCache": false}'::jsonb,
             CURRENT_TIMESTAMP
         )
ON CONFLICT (menu_id) DO UPDATE SET
                                parent_id = EXCLUDED.parent_id,
                                menu_name = EXCLUDED.menu_name,
                                menu_type = EXCLUDED.menu_type,
                                icon = EXCLUDED.icon,
                                route_path = EXCLUDED.route_path,
                                component = EXCLUDED.component,
                                is_visible = EXCLUDED.is_visible,
                                is_cache = EXCLUDED.is_cache,
                                sort_order = EXCLUDED.sort_order,
                                meta_info = EXCLUDED.meta_info;

-- 菜单 3：用户管理，第三层
INSERT INTO sys_menu (
    menu_id,
    parent_id,
    menu_name,
    menu_type,
    icon,
    route_path,
    component,
    is_visible,
    is_cache,
    sort_order,
    meta_info,
    created_at
) VALUES (
             3,
             2,
             '用户管理',
             'M',
             'user',
             '/admin/system/user',
             'system/user/index',
             TRUE,
             TRUE,
             1,
             '{"title": "用户管理", "icon": "user", "noCache": false}'::jsonb,
             CURRENT_TIMESTAMP
         )
ON CONFLICT (menu_id) DO UPDATE SET
                                parent_id = EXCLUDED.parent_id,
                                menu_name = EXCLUDED.menu_name,
                                menu_type = EXCLUDED.menu_type,
                                icon = EXCLUDED.icon,
                                route_path = EXCLUDED.route_path,
                                component = EXCLUDED.component,
                                is_visible = EXCLUDED.is_visible,
                                is_cache = EXCLUDED.is_cache,
                                sort_order = EXCLUDED.sort_order,
                                meta_info = EXCLUDED.meta_info;

-- 菜单 4：角色管理，第三层
INSERT INTO sys_menu (
    menu_id,
    parent_id,
    menu_name,
    menu_type,
    icon,
    route_path,
    component,
    is_visible,
    is_cache,
    sort_order,
    meta_info,
    created_at
) VALUES (
             4,
             2,
             '角色管理',
             'M',
             'role',
             '/admin/system/role',
             'system/role/index',
             TRUE,
             TRUE,
             2,
             '{"title": "角色管理", "icon": "role", "noCache": false}'::jsonb,
             CURRENT_TIMESTAMP
         )
ON CONFLICT (menu_id) DO UPDATE SET
                                parent_id = EXCLUDED.parent_id,
                                menu_name = EXCLUDED.menu_name,
                                menu_type = EXCLUDED.menu_type,
                                icon = EXCLUDED.icon,
                                route_path = EXCLUDED.route_path,
                                component = EXCLUDED.component,
                                is_visible = EXCLUDED.is_visible,
                                is_cache = EXCLUDED.is_cache,
                                sort_order = EXCLUDED.sort_order,
                                meta_info = EXCLUDED.meta_info;

-- 修正 sys_menu 序列，避免后续自增 ID 冲突
SELECT setval(
               pg_get_serial_sequence('sys_menu', 'menu_id'),
               GREATEST((SELECT MAX(menu_id) FROM sys_menu), 1),
               TRUE
       );

COMMIT;
