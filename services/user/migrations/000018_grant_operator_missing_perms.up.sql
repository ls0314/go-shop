-- ============================================================
-- 000018: 补齐平台运营人员(role_id=2)的两条授权
-- 模块:类目管理 / 库存管理
-- 幂等:ON CONFLICT DO NOTHING
-- ============================================================
--
-- 背景:000011 只把 system:* 绑给超级管理员,platform:* 的绑定不完整。
-- 000017 补齐了 platform:* 权限点本身并把它们绑给三个系统角色,
-- 但按 role_name 精确匹配路径时漏了这两条:
--   - platform:category:view 详情路径(GET /api/v1/admin/category/:id)
--   - platform:inventory:log             (GET /api/v1/admin/inventory/log)
-- 二者在 demo_shop 里都授权给运营人员,本迁移补齐。
--
-- 用 permission_code + api_path 关联而非硬编码 permission_id。
-- ------------------------------------------------------------

INSERT INTO sys_role_permission (role_id, permission_id)
SELECT r.role_id, p.permission_id
FROM sys_role r
         JOIN sys_permission p ON (p.permission_code, p.request_method, p.api_path) IN (
             ('platform:category:view', 'GET', '/api/v1/admin/category/:id'),
             ('platform:inventory:log', 'GET', '/api/v1/admin/inventory/log')
         )
WHERE r.role_name = '平台运营人员'
ON CONFLICT (role_id, permission_id) DO NOTHING;
