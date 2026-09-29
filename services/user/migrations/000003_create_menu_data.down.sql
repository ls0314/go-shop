BEGIN;

-- 1. 删除角色菜单关联
DELETE FROM sys_role_menu
WHERE role_id IN (1, 2) AND menu_id IN (1,2,3,4);

-- 2. 删除用户角色关联
DELETE FROM sys_user_role
WHERE user_id = 1 AND role_id IN (1,2);

-- 3. 删除菜单
DELETE FROM sys_menu
WHERE menu_id IN (1,2,3,4);

-- 4. 删除角色
DELETE FROM sys_role
WHERE role_id IN (1,2);

-- 5. 删除用户
DELETE FROM sys_user
WHERE user_id = 1;

-- 6. 重置自增序列（可选，恢复干净状态）
SELECT setval(pg_get_serial_sequence('sys_user', 'user_id'), 1, false);
SELECT setval(pg_get_serial_sequence('sys_role', 'role_id'), 1, false);
SELECT setval(pg_get_serial_sequence('sys_menu', 'menu_id'), 1, false);

COMMIT;