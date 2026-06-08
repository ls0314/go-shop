BEGIN;
CREATE EXTENSION IF NOT EXISTS pgcrypto;
-- 用户：user_id = 1，username = user02，password = 4545.aaa
INSERT INTO sys_user (
    user_id,
    username,
    password_hash,
    email,
    phone,
    status,
    failed_attempts,
    lock_until,
    created_at,
    updated_at
) VALUES (
             'user02',
             crypt('4545.aaa', gen_salt('bf')),
             'user02@example.com',
             '15944167678',
             'active',
             0,
             NULL,
             CURRENT_TIMESTAMP,
             CURRENT_TIMESTAMP
         )
ON CONFLICT (user_id) DO UPDATE SET
                username = EXCLUDED.username,
                password_hash = EXCLUDED.password_hash,
                email = EXCLUDED.email,
                phone = EXCLUDED.phone,
                status = EXCLUDED.status,
                failed_attempts = 0,
                lock_until = NULL,
                updated_at = CURRENT_TIMESTAMP;

-- 角色 1
INSERT INTO sys_role (
    role_id,
    role_name,
    role_type,
    description,
    is_system,
    is_default,
    data_scope,
    created_at,
    create_by,
    updated_at,
    update_by
) VALUES (
             1,
             'role01',
             'custom',
             '测试角色1：拥有菜单1、2、3',
             FALSE,
             FALSE,
             'all',
             CURRENT_TIMESTAMP,
             1,
             CURRENT_TIMESTAMP,
             1
         )
ON CONFLICT (role_id) DO UPDATE SET
                                    role_name = EXCLUDED.role_name,
                                    role_type = EXCLUDED.role_type,
                                    description = EXCLUDED.description,
                                    is_system = EXCLUDED.is_system,
                                    is_default = EXCLUDED.is_default,
                                    data_scope = EXCLUDED.data_scope,
                                    updated_at = CURRENT_TIMESTAMP,
                                    update_by = EXCLUDED.update_by;

-- 角色 2
INSERT INTO sys_role (
    role_id,
    role_name,
    role_type,
    description,
    is_system,
    is_default,
    data_scope,
    created_at,
    create_by,
    updated_at,
    update_by
) VALUES (
             2,
             'role02',
             'custom',
             '测试角色2：拥有菜单2、3、4',
             FALSE,
             FALSE,
             'self',
             CURRENT_TIMESTAMP,
             1,
             CURRENT_TIMESTAMP,
             1
         )
ON CONFLICT (role_id) DO UPDATE SET
                                    role_name = EXCLUDED.role_name,
                                    role_type = EXCLUDED.role_type,
                                    description = EXCLUDED.description,
                                    is_system = EXCLUDED.is_system,
                                    is_default = EXCLUDED.is_default,
                                    data_scope = EXCLUDED.data_scope,
                                    updated_at = CURRENT_TIMESTAMP,
                                    update_by = EXCLUDED.update_by;

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

-- 用户 1 关联角色 1 和角色 2
INSERT INTO sys_user_role (
    user_id,
    role_id
) VALUES
      (1, 1),
      (1, 2)
ON CONFLICT (user_id, role_id) DO NOTHING;

-- 角色 1 关联菜单 1、2、3
INSERT INTO sys_role_menu (
    role_id,
    menu_id
) VALUES
      (1, 1),
      (1, 2),
      (1, 3)
ON CONFLICT (role_id, menu_id) DO NOTHING;

-- 角色 2 关联菜单 2、3、4
INSERT INTO sys_role_menu (
    role_id,
    menu_id
) VALUES
      (2, 2),
      (2, 3),
      (2, 4)
ON CONFLICT (role_id, menu_id) DO NOTHING;

-- 修正 BIGSERIAL 序列，避免后续自增 ID 冲突
SELECT setval(
               pg_get_serial_sequence('sys_user', 'user_id'),
               GREATEST((SELECT MAX(user_id) FROM sys_user), 1),
               TRUE
       );

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

COMMIT;