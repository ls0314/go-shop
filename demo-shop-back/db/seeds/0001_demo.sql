-- ============================================================
-- 开发/测试环境种子数据（演示账号与测试角色，生产环境禁止加载！）
--
-- 与迁移的分工（工程规范）：
--   migrations/  → DDL + 系统数据（菜单树、权限点、内置角色），所有环境执行
--   seeds/       → 演示数据（测试账号、测试角色及绑定），仅开发/测试库显式加载
--
-- 加载方式（仅开发/演示实例）：
--   由 db.RunSeeds 按版本账本(sys_seed_history)执行:文件名即版本号,
--   已记账的文件自动跳过;事务由 Runner 统一管理,文件内不得包含 BEGIN/COMMIT。
-- 测试实例（demo_shop_test）不加载本文件：测试数据由各测试用 factory 自建。
--
-- 幂等：全部使用 ON CONFLICT DO NOTHING / DO UPDATE，可重复执行。
-- 注意：测试角色不指定主键（迁移创建的系统角色已占用 1/2/3），
--       所有关联一律按 username / role_name / route_path 关联，与主键无关。
-- ============================================================

-- ------------------------------
-- 1. 演示用户（原 000003 / 000004 拆出）
--    user02      ：普通演示账号
--    platformUser：平台管理演示账号
--    只写最小列（username/密码/状态），email/phone 留空：
--    避免与开发库中已漂移的真实用户数据撞唯一约束；
--    冲突时只重置密码与锁定状态，不覆盖用户可自行修改的字段。
-- ------------------------------
INSERT INTO sys_user (
    username, password_hash, status, failed_attempts, lock_until
) VALUES
    ('user02',      crypt('4545.aaa', gen_salt('bf')), 'active', 0, NULL),
    ('platformUser', crypt('4545.aaa', gen_salt('bf')), 'active', 0, NULL)
ON CONFLICT (username) DO UPDATE SET
    password_hash   = EXCLUDED.password_hash,
    status          = EXCLUDED.status,
    failed_attempts = 0,
    lock_until      = NULL,
    updated_at      = CURRENT_TIMESTAMP;

-- ------------------------------
-- 2. 测试角色（原 000003 拆出）
--    不指定 role_id：全新环境中迁移链已创建系统角色（主键 1/2/3），
--    显式主键会冲突，交给序列自增。
-- ------------------------------
INSERT INTO sys_role (
    role_name, role_type, description, is_system, is_default, data_scope
) VALUES
    ('role01', 'custom', '测试角色1：拥有菜单1、2、3', FALSE, FALSE, 'all'),
    ('role02', 'custom', '测试角色2：拥有菜单2、3、4', FALSE, FALSE, 'self')
ON CONFLICT (role_name) DO UPDATE SET
    role_type    = EXCLUDED.role_type,
    description  = EXCLUDED.description,
    is_system    = EXCLUDED.is_system,
    is_default   = EXCLUDED.is_default,
    data_scope   = EXCLUDED.data_scope,
    updated_at   = CURRENT_TIMESTAMP;

-- ------------------------------
-- 3. 用户-角色绑定（原 000003 / 000004 拆出，按业务名关联）
-- ------------------------------
INSERT INTO sys_user_role (user_id, role_id)
SELECT u.user_id, r.role_id
FROM sys_user u
         JOIN sys_role r ON r.role_name IN ('role01', 'role02')
WHERE u.username = 'user02'
ON CONFLICT (user_id, role_id) DO NOTHING;

-- platformUser 拥有平台超级管理员（系统角色，由迁移链创建）
INSERT INTO sys_user_role (user_id, role_id)
SELECT u.user_id, r.role_id
FROM sys_user u
         JOIN sys_role r ON r.role_name = '平台超级管理员'
WHERE u.username = 'platformUser'
ON CONFLICT (user_id, role_id) DO NOTHING;

-- ------------------------------
-- 4. 测试角色-菜单绑定（原 000003 拆出，按 route_path 关联菜单）
--    role01：后台管理 / 系统管理 / 用户管理
--    role02：系统管理 / 用户管理 / 角色管理
-- ------------------------------
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT r.role_id, m.menu_id
FROM sys_role r
         JOIN sys_menu m ON m.route_path IN ('/admin', '/admin/system', '/admin/system/user')
WHERE r.role_name = 'role01'
ON CONFLICT (role_id, menu_id) DO NOTHING;

INSERT INTO sys_role_menu (role_id, menu_id)
SELECT r.role_id, m.menu_id
FROM sys_role r
         JOIN sys_menu m ON m.route_path IN ('/admin/system', '/admin/system/user', '/admin/system/role')
WHERE r.role_name = 'role02'
ON CONFLICT (role_id, menu_id) DO NOTHING;

-- ------------------------------
-- 5. 测试角色-权限绑定（原 000011 第五节拆出）
--    保留 role01/role02 可进入用户/角色管理页面
-- ------------------------------
INSERT INTO sys_role_permission (role_id, permission_id)
SELECT r.role_id, p.permission_id
FROM sys_role r
         JOIN sys_permission p
              ON p.permission_code IN ('system:user:view', 'system:role:view')
WHERE r.role_name IN ('role01', 'role02')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- ------------------------------
-- 6. 修正序列，避免后续自增冲突
-- ------------------------------
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

-- ------------------------------
-- 7. user02 绑定平台超级管理员,并让该角色持有全部权限与菜单
--    - 角色为迁移链创建的系统角色,这里只做绑定与授权,全段幂等可重跑
--    - 全部权限:与 sys_permission 全表做笛卡尔积入库;后续迁移新增权限点后
--      重跑本段即可补齐(账本已记账的库需手动执行,见文件头加载方式说明)
--    - 全部菜单:API 权限管接口鉴权,菜单绑定管前端可见性,两者都要给满
--    - user02 原有的 role01/role02 绑定保留:多角色取权限并集,不冲突
-- ------------------------------
INSERT INTO sys_user_role (user_id, role_id)
SELECT u.user_id, r.role_id
FROM sys_user u
         JOIN sys_role r ON r.role_name = '平台超级管理员'
WHERE u.username = 'user02'
ON CONFLICT (user_id, role_id) DO NOTHING;

INSERT INTO sys_role_permission (role_id, permission_id)
SELECT r.role_id, p.permission_id
FROM sys_role r
         CROSS JOIN sys_permission p
WHERE r.role_name = '平台超级管理员'
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO sys_role_menu (role_id, menu_id)
SELECT r.role_id, m.menu_id
FROM sys_role r
         CROSS JOIN sys_menu m
WHERE r.role_name = '平台超级管理员'
ON CONFLICT (role_id, menu_id) DO NOTHING;



