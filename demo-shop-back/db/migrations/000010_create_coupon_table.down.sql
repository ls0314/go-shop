-- =============================================
-- 优惠券模块全量回滚SQL
-- 执行顺序：关联数据 → 菜单/权限主数据 → 业务表
-- =============================================

-- ------------------------------
-- 1. 回滚角色-菜单关联数据
-- ------------------------------
DELETE FROM sys_role_menu
WHERE menu_id IN (
    SELECT menu_id FROM sys_menu WHERE route_path LIKE '/platform/coupon%'
);

-- ------------------------------
-- 2. 回滚角色-权限关联数据
-- ------------------------------
DELETE FROM sys_role_permission
WHERE permission_id IN (
    SELECT permission_id FROM sys_permission WHERE permission_code LIKE 'platform:coupon:%'
);

-- ------------------------------
-- 3. 回滚菜单-权限绑定数据
-- ------------------------------
DELETE FROM sys_menu_permission
WHERE permission_id IN (
    SELECT permission_id FROM sys_permission WHERE permission_code LIKE 'platform:coupon:%'
);

-- ------------------------------
-- 4. 回滚系统菜单数据
-- 注意：自引用外键需从子节点向父节点逐层删除
-- ------------------------------
-- 4.1 删除优惠券列表下的按钮菜单
DELETE FROM sys_menu
WHERE parent_id IN (SELECT menu_id FROM sys_menu WHERE route_path = '/platform/coupon/list')
  AND menu_type = 'F';

-- 4.2 删除优惠券列表菜单
DELETE FROM sys_menu WHERE route_path = '/platform/coupon/list';

-- 4.3 删除优惠券管理父级菜单
DELETE FROM sys_menu WHERE route_path = '/platform/coupon';

-- ------------------------------
-- 5. 回滚系统权限数据
-- ------------------------------
DELETE FROM sys_permission WHERE permission_code LIKE 'platform:coupon:%';

-- ------------------------------
-- 6. 回滚业务表（约束随表级联删除）
-- 依赖顺序：user_coupon → coupon_template
-- ------------------------------
-- 6.1 删除用户优惠券表
DROP TABLE IF EXISTS user_coupon;

-- 6.2 删除优惠券模板表
DROP TABLE IF EXISTS coupon_template;
