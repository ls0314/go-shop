-- 000016 回滚:把 api_path 还原为统一命名之前的写法,并删除本次补齐的 4 个权限点。
-- 与 up 对称、逆序执行,保证 down 后库状态与 000015 结束时一致。

-- ---------- 0. 先删本次补齐的权限点(必须在路径还原之前做 —— 此时 api_path 还是新值) ----------
-- 用 EXISTS 关联而非 IN 子查询:后者在子查询出现 NULL 时会静默不匹配。
-- 顺序:先解角色绑定,再删权限点(sys_role_permission 对 sys_permission 有外键)。
DELETE FROM sys_role_permission rp
WHERE EXISTS (
    SELECT 1 FROM sys_permission p
    WHERE p.permission_id = rp.permission_id
      AND (p.permission_code, p.request_method, p.api_path) IN (
          ('platform:category:view',       'GET', '/api/v1/admin/category'),
          ('platform:product:view',        'GET', '/api/v1/admin/products'),
          ('platform:product:full-update', 'PUT', '/api/v1/admin/products/:id/full'),
          ('platform:log:view',            'GET', '/api/v1/admin/log')
      )
);

DELETE FROM sys_permission
WHERE (permission_code, request_method, api_path) IN (
    ('platform:category:view',       'GET', '/api/v1/admin/category'),
    ('platform:product:view',        'GET', '/api/v1/admin/products'),
    ('platform:product:full-update', 'PUT', '/api/v1/admin/products/:id/full'),
    ('platform:log:view',            'GET', '/api/v1/admin/log')
);

-- ---------- 3. 用户 CRUD 与关联:admin 前缀 → 裸前缀 ----------
UPDATE sys_permission SET api_path = '/api/v1/user'                WHERE api_path = '/api/v1/admin/user';
UPDATE sys_permission SET api_path = '/api/v1/user/:id'            WHERE api_path = '/api/v1/admin/user/:id';
UPDATE sys_permission SET api_path = '/api/v1/user/assign-role'    WHERE api_path = '/api/v1/admin/user/assign-role';
UPDATE sys_permission SET api_path = '/api/v1/user/:id/role'       WHERE api_path = '/api/v1/admin/user/:id/role';
UPDATE sys_permission SET api_path = '/api/v1/user/:id/clear-role' WHERE api_path = '/api/v1/admin/user/:id/clear-role';
UPDATE sys_permission SET api_path = '/api/v1/user/assign-dept'    WHERE api_path = '/api/v1/admin/user/assign-dept';
UPDATE sys_permission SET api_path = '/api/v1/user/:id/dept'       WHERE api_path = '/api/v1/admin/user/:id/dept';
UPDATE sys_permission SET api_path = '/api/v1/user/:id/clear-dept' WHERE api_path = '/api/v1/admin/user/:id/clear-dept';

-- ---------- 2. RBAC 与管理资源:admin 前缀 → 裸前缀 ----------
UPDATE sys_permission SET api_path = '/api/v1/menu'                WHERE api_path = '/api/v1/admin/menu';
UPDATE sys_permission SET api_path = '/api/v1/menu/:id'            WHERE api_path = '/api/v1/admin/menu/:id';
UPDATE sys_permission SET api_path = '/api/v1/menu/:id/tree'       WHERE api_path = '/api/v1/admin/menu/:id/tree';
UPDATE sys_permission SET api_path = '/api/v1/menu/assign-perm'    WHERE api_path = '/api/v1/admin/menu/assign-perm';
UPDATE sys_permission SET api_path = '/api/v1/menu/:id/perm'       WHERE api_path = '/api/v1/admin/menu/:id/perm';
UPDATE sys_permission SET api_path = '/api/v1/menu/:id/clear-perm' WHERE api_path = '/api/v1/admin/menu/:id/clear-perm';

UPDATE sys_permission SET api_path = '/api/v1/role'                WHERE api_path = '/api/v1/admin/role';
UPDATE sys_permission SET api_path = '/api/v1/role/:id'            WHERE api_path = '/api/v1/admin/role/:id';
UPDATE sys_permission SET api_path = '/api/v1/role/assign-perm'    WHERE api_path = '/api/v1/admin/role/assign-perm';
UPDATE sys_permission SET api_path = '/api/v1/role/assign-menu'    WHERE api_path = '/api/v1/admin/role/assign-menu';
UPDATE sys_permission SET api_path = '/api/v1/role/:id/perm'       WHERE api_path = '/api/v1/admin/role/:id/perm';
UPDATE sys_permission SET api_path = '/api/v1/role/:id/menu'       WHERE api_path = '/api/v1/admin/role/:id/menu';
UPDATE sys_permission SET api_path = '/api/v1/role/:id/clear-perm' WHERE api_path = '/api/v1/admin/role/:id/clear-perm';
UPDATE sys_permission SET api_path = '/api/v1/role/:id/clear-menu' WHERE api_path = '/api/v1/admin/role/:id/clear-menu';

UPDATE sys_permission SET api_path = '/api/v1/permissions'     WHERE api_path = '/api/v1/admin/permissions';
UPDATE sys_permission SET api_path = '/api/v1/permissions/:id' WHERE api_path = '/api/v1/admin/permissions/:id';

UPDATE sys_permission SET api_path = '/api/v1/scope'     WHERE api_path = '/api/v1/admin/scope';
UPDATE sys_permission SET api_path = '/api/v1/scope/:id' WHERE api_path = '/api/v1/admin/scope/:id';

UPDATE sys_permission SET api_path = '/api/v1/dept'             WHERE api_path = '/api/v1/admin/dept';
UPDATE sys_permission SET api_path = '/api/v1/dept/:id'         WHERE api_path = '/api/v1/admin/dept/:id';
UPDATE sys_permission SET api_path = '/api/v1/dept/tree/:userId' WHERE api_path = '/api/v1/admin/dept/tree/:userId';

UPDATE sys_permission SET api_path = '/api/v1/admin/platform/log'     WHERE api_path = '/api/v1/admin/log';
UPDATE sys_permission SET api_path = '/api/v1/admin/platform/coupons' WHERE api_path = '/api/v1/admin/coupons';

-- ---------- 1. 平台资源:admin → platform(并复原历史坏路径) ----------
UPDATE sys_permission SET api_path = '/api/v1/platform/orders/:id/ship' WHERE api_path = '/api/v1/admin/orders/:id/ship';
UPDATE sys_permission SET api_path = '/api/v1/platform/orders/:id'      WHERE api_path = '/api/v1/admin/orders/:id';
UPDATE sys_permission SET api_path = '/api/v1/platform/orders'          WHERE api_path = '/api/v1/admin/orders';

UPDATE sys_permission SET api_path = '/api/v1/platform/product/withdraw/:id' WHERE api_path = '/api/v1/admin/products/:id/withdraw';
UPDATE sys_permission SET api_path = '/api/v1/platform/product/publish/:id'  WHERE api_path = '/api/v1/admin/products/:id/publish';
UPDATE sys_permission SET api_path = '/api/v1/platform/product/:id'          WHERE api_path = '/api/v1/admin/products/:id';
UPDATE sys_permission SET api_path = '/api/v1/platform/products'             WHERE api_path = '/api/v1/admin/products';

UPDATE sys_permission SET api_path = '/api/v1/platform/category/children/:id' WHERE api_path = '/api/v1/admin/category/children/:id';
UPDATE sys_permission SET api_path = '/api/v1/platform/category/tree'         WHERE api_path = '/api/v1/admin/category/tree';
UPDATE sys_permission SET api_path = '/api/v1/platform/category/:id'          WHERE api_path = '/api/v1/admin/category/:id';
UPDATE sys_permission SET api_path = '/api/v1/platform/category'              WHERE api_path = '/api/v1/admin/category';
