-- ============================================================
-- 000011 回滚:撤销 RBAC 系统管理模块的 权限点 + 菜单 + 分配
-- 顺序:先删关联绑定,再删按钮/菜单/权限点
-- 注意:只删除本迁移新增的权限码(system:*)与菜单节点
-- ============================================================

-- 1. 删除角色-菜单分配(仅 /admin 下本次由 000011 分配给平台超管的)
DELETE FROM sys_role_menu
WHERE role_id = (SELECT role_id FROM sys_role WHERE role_name = '平台超级管理员')
  AND menu_id IN (SELECT menu_id FROM sys_menu WHERE route_path LIKE '/admin%');

-- 2. 删除角色-权限分配(system:* 权限码)
DELETE FROM sys_role_permission
WHERE permission_id IN (
    SELECT permission_id FROM sys_permission WHERE permission_code LIKE 'system:%'
);

-- 3. 删除菜单-权限绑定(system:* 权限码相关)
DELETE FROM sys_menu_permission
WHERE permission_id IN (
    SELECT permission_id FROM sys_permission WHERE permission_code LIKE 'system:%'
);

-- 4. 删除按钮(F)菜单:挂在权限/菜单/部门管理列表下的
DELETE FROM sys_menu
WHERE menu_type = 'F'
  AND parent_id IN (
      SELECT menu_id FROM sys_menu
      WHERE route_path IN ('/admin/system/permission', '/admin/system/menu', '/admin/system/dept')
  );

-- 5. 删除本次新增的列表菜单节点
DELETE FROM sys_menu
WHERE route_path IN ('/admin/system/permission', '/admin/system/menu', '/admin/system/dept');

-- 6. 删除 system:* 权限点
DELETE FROM sys_permission WHERE permission_code LIKE 'system:%';

-- 7. 修正序列
SELECT setval(
               pg_get_serial_sequence('sys_menu', 'menu_id'),
               GREATEST((SELECT MAX(menu_id) FROM sys_menu), 1),
               TRUE
       );
SELECT setval(
               pg_get_serial_sequence('sys_permission', 'permission_id'),
               GREATEST((SELECT MAX(permission_id) FROM sys_permission), 1),
               TRUE
       );
