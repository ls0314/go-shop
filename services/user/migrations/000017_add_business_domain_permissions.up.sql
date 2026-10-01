-- ============================================================
-- 000017: 补齐业务域权限点 + 角色授权
-- 模块:类目管理 / 商品管理 / 优惠券管理 / 订单管理 / 支付管理
-- 权限码命名:platform:<module>:<action>
-- 幂等:全部使用 ON CONFLICT DO NOTHING
-- ============================================================
--
-- 背景:判权已收口到 user-service(PermissionMiddleware 注入 UserRPC,
-- 按 api_path + request_method 精确匹配权限码)。000011 只 seed 了
-- system:* 权限并把它绑给平台超级管理员,platform:* 中属于业务域的部分
-- 从未迁入本库,导致对应管理端接口一律 403。
--
-- 本次补齐 18 行(16 个权限码,/api/v1/admin/products/:id 与
-- /api/v1/admin/products 这两个 product:view 路径各占一行)。
-- api_path 必须与后端 c.FullPath() 完全一致(含 :id 占位符)。
-- ------------------------------------------------------------

-- ------------------------------------------------------------
-- 一、权限点 seed(sys_permission)
-- ------------------------------------------------------------

-- platform:category:view 的详情路径(列表路径已存在)
INSERT INTO sys_permission (permission_code, permission_name, permission_type, request_method, api_path, description, is_system) VALUES
    ('platform:category:view',     '查看类目',     'api', 'GET',    '/api/v1/admin/category/:id',          '所属模块：平台类目管理', TRUE)
ON CONFLICT (permission_code, request_method, api_path) DO NOTHING;

-- 类目管理:创建/更新/删除/树/子类目
INSERT INTO sys_permission (permission_code, permission_name, permission_type, request_method, api_path, description, is_system) VALUES
    ('platform:category:create',   '创建类目',     'api', 'POST',   '/api/v1/admin/category',              '所属模块：平台类目管理', TRUE),
    ('platform:category:update',   '更新类目',     'api', 'PUT',    '/api/v1/admin/category/:id',          '所属模块：平台类目管理', TRUE),
    ('platform:category:delete',   '删除类目',     'api', 'DELETE', '/api/v1/admin/category/:id',          '所属模块：平台类目管理', TRUE),
    ('platform:category:tree',     '查看类目树',   'api', 'GET',    '/api/v1/admin/category/tree',         '所属模块：平台类目管理', TRUE),
    ('platform:category:children', '查看子类目',   'api', 'GET',    '/api/v1/admin/category/children/:id', '所属模块：平台类目管理', TRUE)
ON CONFLICT (permission_code, request_method, api_path) DO NOTHING;

-- 商品管理:查看详情/查看列表/创建/更新/删除/上架/下架
-- 注:platform:product:view 另有一行落在 /api/v1/admin/inventory/warning,
--    属原有设计(库存预警复用商品查看权限),两边库一致,本迁移不动。
INSERT INTO sys_permission (permission_code, permission_name, permission_type, request_method, api_path, description, is_system) VALUES
    ('platform:product:view',      '查看商品',     'api', 'GET',    '/api/v1/admin/products/:id',          '所属模块：平台商品管理', TRUE),
    ('platform:product:view',      '查看商品',     'api', 'GET',    '/api/v1/admin/products',              '所属模块：平台商品管理', TRUE),
    ('platform:product:create',    '创建商品',     'api', 'POST',   '/api/v1/admin/products',              '所属模块：平台商品管理', TRUE),
    ('platform:product:update',    '更新商品',     'api', 'PUT',    '/api/v1/admin/products/:id',          '所属模块：平台商品管理', TRUE),
    ('platform:product:delete',    '删除商品',     'api', 'DELETE', '/api/v1/admin/products/:id',          '所属模块：平台商品管理', TRUE),
    ('platform:product:publish',   '上架商品',     'api', 'POST',   '/api/v1/admin/products/:id/publish',  '所属模块：平台商品管理', TRUE),
    ('platform:product:withdraw',  '下架商品',     'api', 'POST',   '/api/v1/admin/products/:id/withdraw', '所属模块：平台商品管理', TRUE)
ON CONFLICT (permission_code, request_method, api_path) DO NOTHING;

-- 优惠券管理:查看列表/创建模板
INSERT INTO sys_permission (permission_code, permission_name, permission_type, request_method, api_path, description, is_system) VALUES
    ('platform:coupon:view',       '查看优惠券模板列表', 'api', 'GET',  '/api/v1/admin/coupons',          '所属模块：优惠券管理', TRUE),
    ('platform:coupon:create',     '创建优惠券模板',     'api', 'POST', '/api/v1/admin/coupons',          '所属模块：优惠券管理', TRUE)
ON CONFLICT (permission_code, request_method, api_path) DO NOTHING;

-- 订单管理:查看列表/查看详情/发货
INSERT INTO sys_permission (permission_code, permission_name, permission_type, request_method, api_path, description, is_system) VALUES
    ('platform:order:view',        '查看订单',     'api', 'GET',    '/api/v1/admin/orders',                '所属模块：订单管理', TRUE),
    ('platform:order:view',        '查看订单',     'api', 'GET',    '/api/v1/admin/orders/:id',            '所属模块：订单管理', TRUE),
    ('platform:order:ship',        '订单发货',     'api', 'PUT',    '/api/v1/admin/orders/:id/ship',       '所属模块：订单管理', TRUE)
ON CONFLICT (permission_code, request_method, api_path) DO NOTHING;

-- 支付管理:查看支付记录
INSERT INTO sys_permission (permission_code, permission_name, permission_type, request_method, api_path, description, is_system) VALUES
    ('platform:pay:view',          '查看支付记录', 'api', 'GET',    '/api/v1/admin/pay/list',              '所属模块：支付管理', TRUE)
ON CONFLICT (permission_code, request_method, api_path) DO NOTHING;

-- ------------------------------------------------------------
-- 二、角色-权限分配(sys_role_permission)
-- 000011 只把 system:* 绑给了超级管理员,platform:* 需在此补齐。
-- 用 permission_code 关联而非硬编码 permission_id,保证新建库也正确。
-- ------------------------------------------------------------

-- 平台超级管理员:全部 platform:* 权限
INSERT INTO sys_role_permission (role_id, permission_id)
SELECT r.role_id, p.permission_id
FROM sys_role r
         CROSS JOIN sys_permission p
WHERE r.role_name = '平台超级管理员'
  AND p.permission_code LIKE 'platform:%'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 平台运营人员:类目改/树/子类目、商品改/详情/上架/下架、订单查看/发货、支付查看、优惠券查看
INSERT INTO sys_role_permission (role_id, permission_id)
SELECT r.role_id, p.permission_id
FROM sys_role r
         JOIN sys_permission p ON (p.permission_code, p.request_method, p.api_path) IN (
             ('platform:category:update',   'PUT',  '/api/v1/admin/category/:id'),
             ('platform:category:tree',     'GET',  '/api/v1/admin/category/tree'),
             ('platform:category:children', 'GET',  '/api/v1/admin/category/children/:id'),
             ('platform:product:update',    'PUT',  '/api/v1/admin/products/:id'),
             ('platform:product:view',      'GET',  '/api/v1/admin/products/:id'),
             ('platform:product:publish',   'POST', '/api/v1/admin/products/:id/publish'),
             ('platform:product:withdraw',  'POST', '/api/v1/admin/products/:id/withdraw'),
             ('platform:order:view',        'GET',  '/api/v1/admin/orders'),
             ('platform:order:view',        'GET',  '/api/v1/admin/orders/:id'),
             ('platform:order:ship',        'PUT',  '/api/v1/admin/orders/:id/ship'),
             ('platform:pay:view',          'GET',  '/api/v1/admin/pay/list'),
             ('platform:coupon:view',       'GET',  '/api/v1/admin/coupons')
         )
WHERE r.role_name = '平台运营人员'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 平台审核人员:类目详情/树/子类目、商品详情、订单查看、优惠券查看
INSERT INTO sys_role_permission (role_id, permission_id)
SELECT r.role_id, p.permission_id
FROM sys_role r
         JOIN sys_permission p ON (p.permission_code, p.request_method, p.api_path) IN (
             ('platform:category:view',     'GET', '/api/v1/admin/category/:id'),
             ('platform:category:tree',     'GET', '/api/v1/admin/category/tree'),
             ('platform:category:children', 'GET', '/api/v1/admin/category/children/:id'),
             ('platform:product:view',      'GET', '/api/v1/admin/products/:id'),
             ('platform:order:view',        'GET', '/api/v1/admin/orders'),
             ('platform:order:view',        'GET', '/api/v1/admin/orders/:id'),
             ('platform:coupon:view',       'GET', '/api/v1/admin/coupons')
         )
WHERE r.role_name = '平台审核人员'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- ------------------------------------------------------------
-- 三、序列修正:上面未显式指定 permission_id,由序列分配,
-- 显式 setval 以防后续手工插入 ID 冲突(与 000011 末尾一致)。
-- ------------------------------------------------------------
SELECT setval(
               pg_get_serial_sequence('sys_permission', 'permission_id'),
               GREATEST((SELECT MAX(permission_id) FROM sys_permission), 1),
               TRUE
       );
