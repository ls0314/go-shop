-- ============================================================
-- 000019: 补平台管理菜单树 + 角色/权限绑定
--
-- 背景：user-service 的迁移只 seed 了 /admin 那一支（000003/000011/000012），
--       /platform 父菜单从未创建。于是 000006 想往它下面挂库存菜单时静默失效
--       （platform_menu CTE 取不到父节点 → FROM platform_menu p 得 0 行，不报错也不插入）。
--       后果：超管虽持有全部 platform:* 权限点，但前端菜单树里平台管理整块缺失。
--
-- 蓝图来源：demo_shop 中 /platform 那一支，共 28 行菜单（14 个 M/页面 + 14 个 F/按钮）。
--           该蓝图在 demo_shop 里是历史手工插入的，本迁移把它固化为可复现的 seed。
--
-- 关联键：两库 menu_id 空间已分叉，故一律按 route_path / (parent_id, menu_name) 关联，
--         权限按 permission_code 关联，不硬编码任何 id。
-- 幂等：菜单按 uk_menu_parent_name 做 DO UPDATE（可重复执行并对齐蓝图）；
--       绑定做 DO NOTHING（不覆盖线上手工调整）。
-- ============================================================
BEGIN;

-- ------------------------------------------------------------
-- 一、M 类菜单：平台管理及其页面（14 行）
-- 子菜单的 parent_id 由父菜单的 route_path 反查，避免硬编码 id
-- ------------------------------------------------------------

-- 第一层：平台管理
INSERT INTO sys_menu (
    parent_id, menu_name, menu_type, icon, route_path, component,
    is_visible, is_cache, sort_order, meta_info, created_at
) VALUES (
    0, '平台管理', 'M', 'Setting', '/platform', 'Layout',
    TRUE, TRUE, 1,
    '{"icon": "Setting", "title": "平台管理", "noCache": false}'::jsonb, CURRENT_TIMESTAMP
)
ON CONFLICT ON CONSTRAINT uk_menu_parent_name DO UPDATE SET
    menu_type = EXCLUDED.menu_type,
    icon = EXCLUDED.icon,
    route_path = EXCLUDED.route_path,
    component = EXCLUDED.component,
    is_visible = EXCLUDED.is_visible,
    is_cache = EXCLUDED.is_cache,
    sort_order = EXCLUDED.sort_order,
    meta_info = EXCLUDED.meta_info;

-- 第二层：6 个模块目录
INSERT INTO sys_menu (
    parent_id, menu_name, menu_type, icon, route_path, component,
    is_visible, is_cache, sort_order, meta_info, created_at
)
SELECT p.menu_id, v.menu_name, 'M', v.icon, v.route_path, 'ParentView',
       TRUE, TRUE, v.sort_order, v.meta_info::jsonb, CURRENT_TIMESTAMP
FROM sys_menu p
         CROSS JOIN (VALUES
    ('类目管理',   'List',      '/platform/category',  1, '{"icon": "List", "title": "类目管理", "noCache": false}'),
    ('商品管理',   'product',   '/platform/product',   2, '{"icon": "product", "title": "商品管理", "noCache": false}'),
    ('库存管理',   'inventory', '/platform/inventory', 3, '{"icon": "inventory", "title": "库存管理", "noCache": false}'),
    ('优惠券管理', 'Ticket',    '/platform/coupon',    3, '{"icon": "Ticket", "title": "优惠券管理", "noCache": false}'),
    ('订单管理',   'Document',  '/platform/order',     4, '{"icon": "Document", "title": "订单管理", "noCache": false}'),
    ('支付管理',   'Money',     '/platform/pay',       5, '{"icon": "Money", "title": "支付管理", "noCache": false}')
) AS v(menu_name, icon, route_path, sort_order, meta_info)
WHERE p.route_path = '/platform'
ON CONFLICT ON CONSTRAINT uk_menu_parent_name DO UPDATE SET
    menu_type = EXCLUDED.menu_type,
    icon = EXCLUDED.icon,
    route_path = EXCLUDED.route_path,
    component = EXCLUDED.component,
    is_visible = EXCLUDED.is_visible,
    is_cache = EXCLUDED.is_cache,
    sort_order = EXCLUDED.sort_order,
    meta_info = EXCLUDED.meta_info;

-- 第三层：7 个列表页（parent_id 按父目录的 route_path 反查）
INSERT INTO sys_menu (
    parent_id, menu_name, menu_type, icon, route_path, component,
    is_visible, is_cache, sort_order, meta_info, created_at
)
SELECT p.menu_id, v.menu_name, 'M', '', v.route_path, v.component,
       TRUE, TRUE, v.sort_order, v.meta_info::jsonb, CURRENT_TIMESTAMP
FROM sys_menu p
         JOIN (VALUES
    ('/platform/category',  '类目列表',   '/platform/category/list',  'platform/category/list/index',  1, '{"icon": "", "title": "类目列表", "noCache": false}'),
    ('/platform/product',   '商品列表',   '/platform/product/list',   'platform/product/list/index',   1, '{"icon": "", "title": "商品列表", "noCache": false}'),
    ('/platform/inventory', '库存列表',   '/platform/inventory/list', 'platform/inventory/list/index', 1, '{"icon": "", "title": "库存列表", "noCache": false}'),
    ('/platform/inventory', '库存日志',   '/platform/inventory/log',  'platform/inventory/log/index',  2, '{"icon": "", "title": "库存日志", "noCache": false}'),
    ('/platform/order',     '订单列表',   '/platform/order/list',     'platform/order/list/index',     1, '{"icon": "", "title": "订单列表", "noCache": false}'),
    ('/platform/pay',       '支付列表',   '/platform/pay/list',       'platform/pay/list/index',       1, '{"icon": "", "title": "支付列表", "noCache": false}'),
    ('/platform/coupon',    '优惠券列表', '/platform/coupon/list',    'platform/coupon/list/index',    1, '{"icon": "", "title": "优惠券列表", "noCache": false}')
) AS v(parent_route, menu_name, route_path, component, sort_order, meta_info)
              ON v.parent_route = p.route_path
ON CONFLICT ON CONSTRAINT uk_menu_parent_name DO UPDATE SET
    menu_type = EXCLUDED.menu_type,
    icon = EXCLUDED.icon,
    route_path = EXCLUDED.route_path,
    component = EXCLUDED.component,
    is_visible = EXCLUDED.is_visible,
    is_cache = EXCLUDED.is_cache,
    sort_order = EXCLUDED.sort_order,
    meta_info = EXCLUDED.meta_info;

-- ------------------------------------------------------------
-- 二、F 类按钮（14 行，按父菜单 route_path 挂载）
-- ------------------------------------------------------------
INSERT INTO sys_menu (
    parent_id, menu_name, menu_type, icon, route_path, component,
    is_visible, is_cache, sort_order, meta_info, created_at
)
SELECT p.menu_id, v.menu_name, 'F', '', '', NULL,
       FALSE, FALSE, v.sort_order, v.meta_info::jsonb, CURRENT_TIMESTAMP
FROM sys_menu p
         JOIN (VALUES
    ('/platform/category/list',  '新增类目',   1, '{"title": "新增类目"}'),
    ('/platform/category/list',  '编辑',       2, '{"title": "编辑"}'),
    ('/platform/category/list',  '删除',       3, '{"title": "删除"}'),
    ('/platform/category/list',  '查看详情',   4, '{"title": "查看详情"}'),
    ('/platform/product/list',   '新增商品',   1, '{"title": "新增商品"}'),
    ('/platform/product/list',   '编辑',       2, '{"title": "编辑"}'),
    ('/platform/product/list',   '删除',       3, '{"title": "删除"}'),
    ('/platform/product/list',   '查看详情',   4, '{"title": "查看详情"}'),
    ('/platform/product/list',   '上架',       5, '{"title": "上架"}'),
    ('/platform/product/list',   '下架',       6, '{"title": "下架"}'),
    ('/platform/order/list',     '查看详情',   1, '{"title": "查看详情"}'),
    ('/platform/order/list',     '发货',       2, '{"title": "发货"}'),
    ('/platform/coupon/list',    '新增优惠券', 1, '{"title": "新增优惠券"}'),
    ('/platform/coupon/list',    '查看详情',   2, '{"title": "查看详情"}')
) AS v(parent_route, menu_name, sort_order, meta_info)
              ON v.parent_route = p.route_path
ON CONFLICT ON CONSTRAINT uk_menu_parent_name DO UPDATE SET
    menu_type = EXCLUDED.menu_type,
    icon = EXCLUDED.icon,
    route_path = EXCLUDED.route_path,
    component = EXCLUDED.component,
    is_visible = EXCLUDED.is_visible,
    is_cache = EXCLUDED.is_cache,
    sort_order = EXCLUDED.sort_order,
    meta_info = EXCLUDED.meta_info;

-- ------------------------------------------------------------
-- 三、角色-菜单绑定
-- 按蓝图精确枚举:超管持整个平台支(28 行);运营与审核的差异不是"全部 M 类"——
-- 审查人员不含「库存管理」与「库存日志」两个节点,故逐个列出而不做规则推导。
-- ------------------------------------------------------------
WITH platform_scope AS (
    -- 平台支全部菜单:平台管理自身 + 两级后代
    SELECT menu_id FROM sys_menu
    WHERE route_path = '/platform' OR route_path LIKE '/platform/%'
    UNION
    SELECT m.menu_id
    FROM sys_menu m
             JOIN sys_menu p ON p.menu_id = m.parent_id
    WHERE p.route_path = '/platform' OR p.route_path LIKE '/platform/%'
)
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT r.role_id, ps.menu_id
FROM sys_role r
         CROSS JOIN platform_scope ps
WHERE r.role_name = '平台超级管理员'
ON CONFLICT ON CONSTRAINT uk_role_menu DO NOTHING;

INSERT INTO sys_role_menu (role_id, menu_id)
SELECT r.role_id, m.menu_id
FROM sys_role r
         JOIN (VALUES
    -- 平台运营人员:13 个 M 节点
    ('平台运营人员', '/platform'),
    ('平台运营人员', '/platform/category'),
    ('平台运营人员', '/platform/category/list'),
    ('平台运营人员', '/platform/product'),
    ('平台运营人员', '/platform/product/list'),
    ('平台运营人员', '/platform/inventory'),
    ('平台运营人员', '/platform/inventory/log'),
    ('平台运营人员', '/platform/order'),
    ('平台运营人员', '/platform/order/list'),
    ('平台运营人员', '/platform/pay'),
    ('平台运营人员', '/platform/pay/list'),
    ('平台运营人员', '/platform/coupon'),
    ('平台运营人员', '/platform/coupon/list'),
    -- 平台审核人员:9 个 M 节点(无库存管理/库存日志)
    ('平台审核人员', '/platform'),
    ('平台审核人员', '/platform/category'),
    ('平台审核人员', '/platform/category/list'),
    ('平台审核人员', '/platform/product'),
    ('平台审核人员', '/platform/product/list'),
    ('平台审核人员', '/platform/order'),
    ('平台审核人员', '/platform/order/list'),
    ('平台审核人员', '/platform/coupon'),
    ('平台审核人员', '/platform/coupon/list')
) AS v(role_name, route_path) ON v.role_name = r.role_name
         JOIN sys_menu m ON m.route_path = v.route_path AND m.menu_type = 'M'
ON CONFLICT ON CONSTRAINT uk_role_menu DO NOTHING;

-- ------------------------------------------------------------
-- 四、菜单-权限绑定（决定按钮级权限的显隐）
-- 页面级：M 型列表页绑定各域的查看权限
-- 按钮级：F 型按钮经「父菜单 route_path + 自身 menu_name」定位
--
-- 注：按 demo_shop 蓝图原样搬运。蓝图里「上架」绑的是 platform:product:withdraw、
--     「下架」绑的是 platform:product:update、「新增商品」无绑定 —— 这三处是蓝图
--     自身的错绑，本迁移不擅自改，保持与蓝图一致，另行确认后再修正。
-- ------------------------------------------------------------

-- 页面级
INSERT INTO sys_menu_permission (menu_id, permission_id)
SELECT m.menu_id, p.permission_id
FROM (VALUES
    ('/platform/category/list',  'platform:category:view'),
    ('/platform/product/list',   'platform:product:view'),
    ('/platform/inventory/list', 'platform:inventory:view'),
    ('/platform/inventory/log',  'platform:inventory:log'),
    ('/platform/order/list',     'platform:order:view'),
    ('/platform/pay/list',       'platform:pay:view'),
    ('/platform/coupon/list',    'platform:coupon:view')
) AS v(route_path, perm_code)
         JOIN sys_menu m ON m.route_path = v.route_path AND m.menu_type = 'M'
         JOIN sys_permission p ON p.permission_code = v.perm_code
-- uk_menu_permission 是 CREATE UNIQUE INDEX 建的纯索引(非表约束),
-- 故这里用列推导而不能用 ON CONFLICT ON CONSTRAINT
ON CONFLICT (menu_id, permission_id) DO NOTHING;

-- 按钮级
INSERT INTO sys_menu_permission (menu_id, permission_id)
SELECT m.menu_id, p.permission_id
FROM (VALUES
    ('/platform/category/list', '新增类目',   'platform:category:create'),
    ('/platform/category/list', '编辑',       'platform:category:update'),
    ('/platform/category/list', '删除',       'platform:category:delete'),
    ('/platform/category/list', '查看详情',   'platform:category:view'),
    ('/platform/product/list',  '编辑',       'platform:product:update'),
    ('/platform/product/list',  '删除',       'platform:product:delete'),
    ('/platform/product/list',  '查看详情',   'platform:product:view'),
    ('/platform/product/list',  '上架',       'platform:product:withdraw'),
    ('/platform/product/list',  '下架',       'platform:product:update'),
    ('/platform/order/list',    '查看详情',   'platform:order:view'),
    ('/platform/order/list',    '发货',       'platform:order:ship'),
    ('/platform/coupon/list',   '新增优惠券', 'platform:coupon:create'),
    ('/platform/coupon/list',   '查看详情',   'platform:coupon:view')
) AS v(parent_route, button_name, perm_code)
         JOIN sys_menu parent ON parent.route_path = v.parent_route AND parent.menu_type = 'M'
         JOIN sys_menu m ON m.parent_id = parent.menu_id AND m.menu_name = v.button_name AND m.menu_type = 'F'
         JOIN sys_permission p ON p.permission_code = v.perm_code
ON CONFLICT (menu_id, permission_id) DO NOTHING;

COMMIT;
