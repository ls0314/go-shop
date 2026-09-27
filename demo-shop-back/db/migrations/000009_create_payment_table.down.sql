-- =============================================
-- 支付模块全量回滚SQL
-- 执行顺序：关联数据 → 菜单/权限主数据 → 业务表
-- =============================================

-- ------------------------------
-- 1. 回滚角色-菜单关联数据
-- ------------------------------
DELETE FROM sys_role_menu
WHERE menu_id IN (
    SELECT menu_id FROM sys_menu WHERE route_path LIKE '/platform/pay%'
);

-- ------------------------------
-- 2. 回滚角色-权限关联数据
-- ------------------------------
DELETE FROM sys_role_permission
WHERE permission_id IN (
    SELECT permission_id FROM sys_permission WHERE permission_code = 'platform:pay:view'
);

-- ------------------------------
-- 3. 回滚菜单-权限绑定数据
-- ------------------------------
DELETE FROM sys_menu_permission
WHERE permission_id IN (
    SELECT permission_id FROM sys_permission WHERE permission_code = 'platform:pay:view'
);

-- ------------------------------
-- 4. 回滚系统菜单数据
-- 注意：自引用外键需从子节点向父节点逐层删除
-- ------------------------------
-- 4.1 删除支付列表菜单
DELETE FROM sys_menu WHERE route_path = '/platform/pay/list';

-- 4.2 删除支付管理父级菜单
DELETE FROM sys_menu WHERE route_path = '/platform/pay';

-- ------------------------------
-- 5. 回滚系统权限数据
-- ------------------------------
DELETE FROM sys_permission WHERE permission_code = 'platform:pay:view';

-- ------------------------------
-- 6. 回滚业务表
-- ------------------------------
DROP TABLE IF EXISTS user_payment_record;
