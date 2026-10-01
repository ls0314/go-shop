-- ============================================================
-- 000018 down:回滚平台运营人员的两条授权
-- ============================================================

DELETE FROM sys_role_permission
WHERE role_id = (SELECT role_id FROM sys_role WHERE role_name = '平台运营人员')
  AND permission_id IN (
      SELECT permission_id
      FROM sys_permission
      WHERE (permission_code, request_method, api_path) IN (
          ('platform:category:view', 'GET', '/api/v1/admin/category/:id'),
          ('platform:inventory:log', 'GET', '/api/v1/admin/inventory/log')
      )
  );
