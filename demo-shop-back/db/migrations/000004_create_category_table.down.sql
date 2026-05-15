BEGIN;

-- 删除按钮菜单对应的权限绑定
WITH category_list_menu AS (
    SELECT menu_id
    FROM sys_menu
    WHERE route_path = '/platform/category/list'
    LIMIT 1
),
     button_menus AS (
         SELECT m.menu_id
         FROM sys_menu m
                  JOIN category_list_menu c
                       ON m.parent_id = c.menu_id
         WHERE m.menu_type = 'F'
           AND m.menu_name IN (
                               '新增类目',
                               '编辑',
                               '删除',
                               '查看详情'
             )
     )
DELETE FROM sys_menu_permission mp
    USING button_menus bm
WHERE mp.menu_id = bm.menu_id;


-- 删除类目菜单对应的权限绑定
DELETE FROM sys_menu_permission mp
    USING sys_menu m
WHERE mp.menu_id = m.menu_id
  AND m.route_path IN (
       '/platform',
       '/platform/category',
       '/platform/category/list'
    );


-- 删除类目权限对应的菜单绑定
DELETE FROM sys_menu_permission mp
    USING sys_permission p
WHERE mp.permission_id = p.permission_id
  AND p.permission_code IN (
        'platform:category:create',
        'platform:category:update',
        'platform:category:delete',
        'platform:category:view',
        'platform:category:tree',
        'platform:category:children'
    );


-- 删除角色权限绑定
DELETE FROM sys_role_permission rp
    USING sys_role r, sys_permission p
WHERE rp.role_id = r.role_id
  AND rp.permission_id = p.permission_id
  AND r.role_name IN (
                      '平台超级管理员',
                      '平台运营人员',
                      '平台审核人员'
    )
  AND p.permission_code IN (
                            'platform:category:create',
                            'platform:category:update',
                            'platform:category:delete',
                            'platform:category:view',
                            'platform:category:tree',
                            'platform:category:children'
    );

-- 删除按钮菜单

WITH category_list_menu AS (
    SELECT menu_id
    FROM sys_menu
    WHERE route_path = '/platform/category/list'
    LIMIT 1
)
DELETE FROM sys_menu m
    USING category_list_menu c
WHERE m.parent_id = c.menu_id
  AND m.menu_type = 'F'
  AND m.menu_name IN (
                      '新增类目',
                      '编辑',
                      '删除',
                      '查看详情'
    );


-- 删除页面菜单

DELETE FROM sys_menu
WHERE route_path = '/platform/category/list';

DELETE FROM sys_menu
WHERE route_path = '/platform/category';

DELETE FROM sys_menu
WHERE route_path = '/platform';


-- 删除系统权限

DELETE FROM sys_permission
WHERE permission_code IN (
          'platform:category:create',
          'platform:category:update',
          'platform:category:delete',
          'platform:category:view',
          'platform:category:tree',
          'platform:category:children'
    );


-- 删除系统角色

DELETE FROM sys_role
WHERE role_name IN (
                    '平台超级管理员',
                    '平台运营人员',
                    '平台审核人员'
    )
  AND role_type = 'platform';


-- 删除类目表

DROP TABLE IF EXISTS sys_category;


COMMIT;