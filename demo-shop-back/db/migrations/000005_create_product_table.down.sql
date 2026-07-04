-- =============================================
-- 商品模块全量回滚SQL
-- 执行顺序：关联数据 → 菜单/权限主数据 → 业务表（子表→父表）
-- =============================================

-- ------------------------------
-- 1. 回滚角色-菜单关联数据
-- ------------------------------
DELETE FROM sys_role_menu
WHERE menu_id IN (
    SELECT menu_id FROM sys_menu WHERE route_path LIKE '/platform/product%'
);

-- ------------------------------
-- 2. 回滚角色-权限关联数据
-- ------------------------------
DELETE FROM sys_role_permission
WHERE permission_id IN (
    SELECT permission_id FROM sys_permission WHERE permission_code LIKE 'platform:product:%'
);

-- ------------------------------
-- 3. 回滚菜单-权限绑定数据
-- ------------------------------
DELETE FROM sys_menu_permission
WHERE permission_id IN (
    SELECT permission_id FROM sys_permission WHERE permission_code LIKE 'platform:product:%'
);

-- ------------------------------
-- 4. 回滚系统菜单数据
-- 注意：自引用外键需从子节点向父节点逐层删除
-- ------------------------------
-- 4.1 删除商品列表下的按钮菜单
DELETE FROM sys_menu
WHERE parent_id IN (SELECT menu_id FROM sys_menu WHERE route_path = '/platform/product/list')
  AND menu_type = 'F';

-- 4.2 删除商品列表菜单
DELETE FROM sys_menu WHERE route_path = '/platform/product/list';

-- 4.3 删除商品管理父级菜单
DELETE FROM sys_menu WHERE route_path = '/platform/product';

-- ------------------------------
-- 5. 回滚系统权限数据
-- ------------------------------
DELETE FROM sys_permission WHERE permission_code LIKE 'platform:product:%';

-- ------------------------------
-- 6. 回滚业务表（按外键依赖逆序删除）
-- 依赖顺序：product_stock_log → product_sku → product_spu
--          product_spu_image → product_spu
-- ------------------------------
-- 6.1 删除库存变更日志表
DROP TABLE IF EXISTS sys_product_stock_log;

-- 6.2 删除SPU图片表
DROP TABLE IF EXISTS sys_product_spu_image;

-- 6.3 删除SKU表
DROP TABLE IF EXISTS sys_product_sku;

-- 6.4 删除SPU主表
DROP TABLE IF EXISTS sys_product_spu;