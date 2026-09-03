-- ============================================================
-- 000012 回滚:撤销数据权限菜单(不删 system:scope:* 权限码,由 000011 down 处理)
-- ============================================================

-- 1. 删除平台超管对该菜单的分配
DELETE FROM sys_role_menu
WHERE role_id = (SELECT role_id FROM sys_role WHERE role_name = '平台超级管理员')
  AND menu_id IN (SELECT menu_id FROM sys_menu WHERE route_path LIKE '/admin/system/scope%');

-- 2. 删除菜单-权限绑定
DELETE FROM sys_menu_permission
WHERE menu_id IN (SELECT menu_id FROM sys_menu WHERE route_path LIKE '/admin/system/scope%');

-- 3. 删除按钮与列表菜单
DELETE FROM sys_menu WHERE route_path LIKE '/admin/system/scope%';
DELETE FROM sys_menu
WHERE parent_id IN (SELECT menu_id FROM sys_menu WHERE route_path = '/admin/system/scope')
  AND menu_type = 'F';

-- 4. 修正序列
SELECT setval(
               pg_get_serial_sequence('sys_menu', 'menu_id'),
               GREATEST((SELECT MAX(menu_id) FROM sys_menu), 1),
               TRUE
       );
