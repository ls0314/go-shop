-- ============================================================
-- 000016 权限表 api_path 对齐统一后的路由命名
--
-- 统一对外路由命名(管理端 /api/v1/admin/**、用户端 /api/v1/**),
-- ============================================================

-- ---------- 管理端:platform → admin ----------
UPDATE sys_permission SET api_path = '/api/v1/admin/category'
 WHERE api_path = '/api/v1/platform/category';
UPDATE sys_permission SET api_path = '/api/v1/admin/category/:id'
 WHERE api_path = '/api/v1/platform/category/:id';
UPDATE sys_permission SET api_path = '/api/v1/admin/category/tree'
 WHERE api_path = '/api/v1/platform/category/tree';
UPDATE sys_permission SET api_path = '/api/v1/admin/category/children/:id'
 WHERE api_path = '/api/v1/platform/category/children/:id';

-- 商品:同时修正 :id 位置写法
UPDATE sys_permission SET api_path = '/api/v1/admin/products'
 WHERE api_path = '/api/v1/platform/products';
UPDATE sys_permission SET api_path = '/api/v1/admin/products/:id'
 WHERE api_path = '/api/v1/platform/product/:id';
UPDATE sys_permission SET api_path = '/api/v1/admin/products/:id/publish'
 WHERE api_path = '/api/v1/platform/product/publish/:id';
UPDATE sys_permission SET api_path = '/api/v1/admin/products/:id/withdraw'
 WHERE api_path = '/api/v1/platform/product/withdraw/:id';

UPDATE sys_permission SET api_path = '/api/v1/admin/orders'
 WHERE api_path = '/api/v1/platform/orders';
UPDATE sys_permission SET api_path = '/api/v1/admin/orders/:id'
 WHERE api_path = '/api/v1/platform/orders/:id';
UPDATE sys_permission SET api_path = '/api/v1/admin/orders/:id/ship'
 WHERE api_path = '/api/v1/platform/orders/:id/ship';

UPDATE sys_permission SET api_path = '/api/v1/admin/coupons'
 WHERE api_path = '/api/v1/admin/platform/coupons';
UPDATE sys_permission SET api_path = '/api/v1/admin/log'
 WHERE api_path = '/api/v1/admin/platform/log';

-- ---------- 管理端:裸前缀 → admin  ----------
UPDATE sys_permission SET api_path = '/api/v1/admin/dept'            WHERE api_path = '/api/v1/dept';
UPDATE sys_permission SET api_path = '/api/v1/admin/dept/:id'        WHERE api_path = '/api/v1/dept/:id';
UPDATE sys_permission SET api_path = '/api/v1/admin/dept/tree/:userId' WHERE api_path = '/api/v1/dept/tree/:userId';

UPDATE sys_permission SET api_path = '/api/v1/admin/scope'      WHERE api_path = '/api/v1/scope';
UPDATE sys_permission SET api_path = '/api/v1/admin/scope/:id'  WHERE api_path = '/api/v1/scope/:id';

UPDATE sys_permission SET api_path = '/api/v1/admin/permissions'      WHERE api_path = '/api/v1/permissions';
UPDATE sys_permission SET api_path = '/api/v1/admin/permissions/:id'  WHERE api_path = '/api/v1/permissions/:id';

UPDATE sys_permission SET api_path = '/api/v1/admin/role'                WHERE api_path = '/api/v1/role';
UPDATE sys_permission SET api_path = '/api/v1/admin/role/:id'            WHERE api_path = '/api/v1/role/:id';
UPDATE sys_permission SET api_path = '/api/v1/admin/role/assign-perm'    WHERE api_path = '/api/v1/role/assign-perm';
UPDATE sys_permission SET api_path = '/api/v1/admin/role/assign-menu'    WHERE api_path = '/api/v1/role/assign-menu';
UPDATE sys_permission SET api_path = '/api/v1/admin/role/:id/perm'       WHERE api_path = '/api/v1/role/:id/perm';
UPDATE sys_permission SET api_path = '/api/v1/admin/role/:id/menu'       WHERE api_path = '/api/v1/role/:id/menu';
UPDATE sys_permission SET api_path = '/api/v1/admin/role/:id/clear-perm' WHERE api_path = '/api/v1/role/:id/clear-perm';
UPDATE sys_permission SET api_path = '/api/v1/admin/role/:id/clear-menu' WHERE api_path = '/api/v1/role/:id/clear-menu';

UPDATE sys_permission SET api_path = '/api/v1/admin/menu'                WHERE api_path = '/api/v1/menu';
UPDATE sys_permission SET api_path = '/api/v1/admin/menu/:id'            WHERE api_path = '/api/v1/menu/:id';
UPDATE sys_permission SET api_path = '/api/v1/admin/menu/:id/tree'       WHERE api_path = '/api/v1/menu/:id/tree';
UPDATE sys_permission SET api_path = '/api/v1/admin/menu/assign-perm'    WHERE api_path = '/api/v1/menu/assign-perm';
UPDATE sys_permission SET api_path = '/api/v1/admin/menu/:id/perm'       WHERE api_path = '/api/v1/menu/:id/perm';
UPDATE sys_permission SET api_path = '/api/v1/admin/menu/:id/clear-perm' WHERE api_path = '/api/v1/menu/:id/clear-perm';

-- ---------- 管理端:用户 CRUD 与用户关联 ----------
-- 公开认证端点 /api/v1/user/register|login|refresh 与登录态自查 /api/v1/user/info|perms 不在此列
UPDATE sys_permission SET api_path = '/api/v1/admin/user'                 WHERE api_path = '/api/v1/user';
UPDATE sys_permission SET api_path = '/api/v1/admin/user/:id'             WHERE api_path = '/api/v1/user/:id';
UPDATE sys_permission SET api_path = '/api/v1/admin/user/assign-role'     WHERE api_path = '/api/v1/user/assign-role';
UPDATE sys_permission SET api_path = '/api/v1/admin/user/:id/role'        WHERE api_path = '/api/v1/user/:id/role';
UPDATE sys_permission SET api_path = '/api/v1/admin/user/:id/clear-role'  WHERE api_path = '/api/v1/user/:id/clear-role';
UPDATE sys_permission SET api_path = '/api/v1/admin/user/assign-dept'     WHERE api_path = '/api/v1/user/assign-dept';
UPDATE sys_permission SET api_path = '/api/v1/admin/user/:id/dept'        WHERE api_path = '/api/v1/user/:id/dept';
UPDATE sys_permission SET api_path = '/api/v1/admin/user/:id/clear-dept'  WHERE api_path = '/api/v1/user/:id/clear-dept';

-- 以下路径本就已符合统一后的命名,无需更新:
--   /api/v1/admin/inventory/{sku/:id, spu/:id, adjust, log, warning}
--   /api/v1/admin/pay/list
--   /api/v1/upload/chunk
--   /api/v1/pay/callback/mock

-- ============================================================
-- 补齐权限点缺口
--
-- 补充原则(与既有种子一致):
--   - 权限码 = 资源域前缀(platform/system) + 资源 + 动作;
--   - 权限点必须在 sys_permission 里存在,否则前端拿不到、角色也绑不上。
--
-- 不在此列:
--   POST /api/v1/admin/menu/tree —— 刻意公开(前端动态路由要用),未挂 PermissionMiddleware
-- ============================================================
INSERT INTO sys_permission
    (permission_code, permission_name, permission_type, request_method, api_path, description, is_system)
VALUES
    -- 类目列表:与 /api/v1/admin/category/:id 同类,列表接口供 create/update 场景读取
    ('platform:category:view',        '查看类目',     'api', 'GET', '/api/v1/admin/category',          '所属模块：平台类目管理', TRUE),
    -- 商品列表:同上
    ('platform:product:view',         '查看商品',     'api', 'GET', '/api/v1/admin/products',          '所属模块：平台商品管理', TRUE),
    -- 商品整体更新(PUT /:id/full):与 PUT /:id 是不同接口、不同校验强度,独立权限点
    ('platform:product:full-update',  '整体更新商品', 'api', 'PUT', '/api/v1/admin/products/:id/full', '所属模块：平台商品管理', TRUE),
    -- 操作日志查询
    ('platform:log:view',             '查看操作日志', 'api', 'GET', '/api/v1/admin/log',               '所属模块：平台操作日志', TRUE)
ON CONFLICT (permission_code, request_method, api_path) DO NOTHING;

-- 绑定平台超级管理员(与既有种子一致:该角色持有全部 platform:* 权限)
INSERT INTO sys_role_permission (role_id, permission_id)
SELECT r.role_id, p.permission_id
FROM sys_role r
         JOIN sys_permission p
              ON p.permission_code IN (
                                       'platform:category:view',
                                       'platform:product:view',
                                       'platform:product:full-update',
                                       'platform:log:view'
                  )
WHERE r.role_name = '平台超级管理员'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 说明:sys_menu_permission(菜单-权限绑定)未在本迁移内处理 ——
-- 它只影响"按钮是否在前端显示"(功能可见性),不影响后端鉴权
-- (PermissionMiddleware 直查 sys_role_permission)。若上面 4 个权限点对应的按钮
-- 需要在界面上出现,应在修改菜单种子的迁移里追加绑定。
