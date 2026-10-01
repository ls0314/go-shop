-- ============================================================
-- 000017 down:回滚业务域权限点与角色授权
-- ============================================================
-- 只删除本迁移新增的 18 行权限点。
-- platform:inventory:{adjust,log,view}、platform:product:view(库存预警)
-- 与 platform:category:view(列表)在 000011 及更早的迁移里已存在,
-- 且被角色 2/3 使用,不在回滚范围内。
--
-- 先删角色绑定再删权限点(外键 fk_role_permission_permission)。
-- ------------------------------------------------------------

-- ------------------------------------------------------------
-- 一、解除角色-权限绑定
-- ------------------------------------------------------------
DELETE FROM sys_role_permission
WHERE permission_id IN (
    SELECT permission_id
    FROM sys_permission
    WHERE (permission_code, request_method, api_path) IN (
        ('platform:category:view',     'GET',    '/api/v1/admin/category/:id'),
        ('platform:category:create',   'POST',   '/api/v1/admin/category'),
        ('platform:category:update',   'PUT',    '/api/v1/admin/category/:id'),
        ('platform:category:delete',   'DELETE', '/api/v1/admin/category/:id'),
        ('platform:category:tree',     'GET',    '/api/v1/admin/category/tree'),
        ('platform:category:children', 'GET',    '/api/v1/admin/category/children/:id'),
        ('platform:product:view',      'GET',    '/api/v1/admin/products/:id'),
        ('platform:product:view',      'GET',    '/api/v1/admin/products'),
        ('platform:product:create',    'POST',   '/api/v1/admin/products'),
        ('platform:product:update',    'PUT',    '/api/v1/admin/products/:id'),
        ('platform:product:delete',    'DELETE', '/api/v1/admin/products/:id'),
        ('platform:product:publish',   'POST',   '/api/v1/admin/products/:id/publish'),
        ('platform:product:withdraw',  'POST',   '/api/v1/admin/products/:id/withdraw'),
        ('platform:coupon:view',       'GET',    '/api/v1/admin/coupons'),
        ('platform:coupon:create',     'POST',   '/api/v1/admin/coupons'),
        ('platform:order:view',        'GET',    '/api/v1/admin/orders'),
        ('platform:order:view',        'GET',    '/api/v1/admin/orders/:id'),
        ('platform:order:ship',        'PUT',    '/api/v1/admin/orders/:id/ship'),
        ('platform:pay:view',          'GET',    '/api/v1/admin/pay/list')
    )
);

-- ------------------------------------------------------------
-- 二、删除本迁移新增的权限点
-- ------------------------------------------------------------
DELETE FROM sys_permission
WHERE (permission_code, request_method, api_path) IN (
    ('platform:category:view',     'GET',    '/api/v1/admin/category/:id'),
    ('platform:category:create',   'POST',   '/api/v1/admin/category'),
    ('platform:category:update',   'PUT',    '/api/v1/admin/category/:id'),
    ('platform:category:delete',   'DELETE', '/api/v1/admin/category/:id'),
    ('platform:category:tree',     'GET',    '/api/v1/admin/category/tree'),
    ('platform:category:children', 'GET',    '/api/v1/admin/category/children/:id'),
    ('platform:product:view',      'GET',    '/api/v1/admin/products/:id'),
    ('platform:product:view',      'GET',    '/api/v1/admin/products'),
    ('platform:product:create',    'POST',   '/api/v1/admin/products'),
    ('platform:product:update',    'PUT',    '/api/v1/admin/products/:id'),
    ('platform:product:delete',    'DELETE', '/api/v1/admin/products/:id'),
    ('platform:product:publish',   'POST',   '/api/v1/admin/products/:id/publish'),
    ('platform:product:withdraw',  'POST',   '/api/v1/admin/products/:id/withdraw'),
    ('platform:coupon:view',       'GET',    '/api/v1/admin/coupons'),
    ('platform:coupon:create',     'POST',   '/api/v1/admin/coupons'),
    ('platform:order:view',        'GET',    '/api/v1/admin/orders'),
    ('platform:order:view',        'GET',    '/api/v1/admin/orders/:id'),
    ('platform:order:ship',        'PUT',    '/api/v1/admin/orders/:id/ship'),
    ('platform:pay:view',          'GET',    '/api/v1/admin/pay/list')
);
