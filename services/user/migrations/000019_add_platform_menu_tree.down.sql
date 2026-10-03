-- ============================================================
-- 000019 down:回滚平台管理菜单树
-- 先删绑定(外键),再按「按钮 → 列表页 → 模块目录 → 平台管理」自底向上删菜单
-- ============================================================
BEGIN;

-- 平台支的全部菜单 id（平台管理自身 + 两级后代）
CREATE TEMP TABLE tmp_platform_menu_ids ON COMMIT DROP AS
SELECT menu_id FROM sys_menu
WHERE route_path = '/platform' OR route_path LIKE '/platform/%'
UNION
SELECT m.menu_id
FROM sys_menu m
         JOIN sys_menu p ON p.menu_id = m.parent_id
WHERE p.route_path = '/platform' OR p.route_path LIKE '/platform/%';

-- 先解绑定
DELETE FROM sys_menu_permission WHERE menu_id IN (SELECT menu_id FROM tmp_platform_menu_ids);
DELETE FROM sys_role_menu WHERE menu_id IN (SELECT menu_id FROM tmp_platform_menu_ids);

-- 再按层级自底向上删菜单（避免 parent_id 悬空）
-- 最深一层是按钮，其父是列表页
DELETE FROM sys_menu
WHERE parent_id IN (SELECT menu_id FROM tmp_platform_menu_ids)
  AND menu_id IN (SELECT menu_id FROM tmp_platform_menu_ids);

-- 列表页
DELETE FROM sys_menu
WHERE parent_id IN (SELECT menu_id FROM sys_menu WHERE route_path = '/platform' OR route_path LIKE '/platform/%')
  AND menu_id IN (SELECT menu_id FROM tmp_platform_menu_ids);

-- 模块目录
DELETE FROM sys_menu
WHERE parent_id = (SELECT menu_id FROM sys_menu WHERE route_path = '/platform')
  AND menu_id IN (SELECT menu_id FROM tmp_platform_menu_ids);

-- 平台管理自身
DELETE FROM sys_menu WHERE route_path = '/platform';

COMMIT;
