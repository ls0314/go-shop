-- =============================================
-- 订单模块全量回滚SQL
-- 执行顺序：关联数据 → 菜单/权限主数据 → 业务表（子表→父表）
-- =============================================

-- ------------------------------
-- 1. 回滚角色-菜单关联数据
-- ------------------------------
DELETE FROM sys_role_menu
WHERE menu_id IN (
    SELECT menu_id FROM sys_menu WHERE route_path LIKE '/platform/order%'
);

-- ------------------------------
-- 2. 回滚角色-权限关联数据
-- ------------------------------
DELETE FROM sys_role_permission
WHERE permission_id IN (
    SELECT permission_id FROM sys_permission WHERE permission_code LIKE 'platform:order:%'
);

-- ------------------------------
-- 3. 回滚菜单-权限绑定数据
-- ------------------------------
DELETE FROM sys_menu_permission
WHERE permission_id IN (
    SELECT permission_id FROM sys_permission WHERE permission_code LIKE 'platform:order:%'
);

-- ------------------------------
-- 4. 回滚系统菜单数据
-- 注意：自引用外键需从子节点向父节点逐层删除
-- ------------------------------
-- 4.1 删除订单列表下的按钮菜单（查看详情、发货）
DELETE FROM sys_menu
WHERE parent_id IN (SELECT menu_id FROM sys_menu WHERE route_path = '/platform/order/list')
  AND menu_type = 'F';

-- 4.2 删除订单列表菜单
DELETE FROM sys_menu WHERE route_path = '/platform/order/list';

-- 4.3 删除订单管理父级菜单
DELETE FROM sys_menu WHERE route_path = '/platform/order';

-- ------------------------------
-- 5. 回滚系统权限数据
-- ------------------------------
DELETE FROM sys_permission WHERE permission_code LIKE 'platform:order:%';

-- ------------------------------
-- 6. 回滚业务表（按外键依赖逆序删除）
-- 依赖顺序：user_order_log → user_order_detail → user_order_master
-- ------------------------------
-- 6.1 删除订单日志表
DROP TABLE IF EXISTS user_order_log;

-- 6.2 删除订单明细表
DROP TABLE IF EXISTS user_order_detail;

-- 6.3 删除订单主表
DROP TABLE IF EXISTS user_order_master;
