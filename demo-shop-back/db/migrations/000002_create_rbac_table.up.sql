-- 数据库未设置完整 在完善RBAC过程中逐步修改
-- 创建权限表
CREATE TABLE IF NOT EXISTS sys_permission (
    permission_id BIGSERIAL PRIMARY KEY,
    permission_code VARCHAR(100) UNIQUE NOT NULL,
    permission_name VARCHAR(100) NOT NULL,
    permission_type VARCHAR(20) NOT NULL ,
    request_method VARCHAR(10),
    api_path VARCHAR(500) NOT NULL ,
    description TEXT ,
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 添加列注释
COMMENT ON COLUMN sys_permission.permission_id IS '权限表唯一标识';
COMMENT ON COLUMN sys_permission.permission_code IS '权限标识';
COMMENT ON COLUMN sys_permission.permission_name IS '权限名称';
COMMENT ON COLUMN sys_permission.permission_type IS '权限类型';
COMMENT ON COLUMN sys_permission.request_method IS 'GET/POST/PUT/DELETE';
COMMENT ON COLUMN sys_permission.api_path IS 'API路径';
COMMENT ON COLUMN sys_permission.description IS '权限描述';
COMMENT ON COLUMN sys_permission.is_system IS '是否系统内置权限';
COMMENT ON COLUMN sys_permission.created_at IS '创建时间';



-- 创建菜单表
CREATE TABLE IF NOT EXISTS sys_menu (
    menu_id BIGSERIAL PRIMARY KEY,
    parent_id BIGINT DEFAULT 0,
    menu_name VARCHAR(100) NOT NULL,
    menu_type VARCHAR(20) DEFAULT 'menu',
    icon VARCHAR(200),
    route_path VARCHAR(200),
    component VARCHAR(200),
    is_visible BOOLEAN DEFAULT TRUE,
    is_cache BOOLEAN DEFAULT TRUE,
    sort_order INT DEFAULT 0,
    meta_info jsonb,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 添加列注释
COMMENT ON COLUMN sys_menu.menu_id IS '菜单表唯一标识';
COMMENT ON COLUMN sys_menu.parent_id IS '上级ID';
COMMENT ON COLUMN sys_menu.menu_name IS '菜单名称';
COMMENT ON COLUMN sys_menu.menu_type IS '菜单类型';
COMMENT ON COLUMN sys_menu.icon IS '图标';
COMMENT ON COLUMN sys_menu.route_path IS '前端路由路径';
COMMENT ON COLUMN sys_menu.component IS '组件路径';
COMMENT ON COLUMN sys_menu.is_visible IS '是否可见';
COMMENT ON COLUMN sys_menu.is_cache IS '是否缓存';
COMMENT ON COLUMN sys_menu.sort_order IS '排序';
COMMENT ON COLUMN sys_menu.meta_info IS '元信息';
COMMENT ON COLUMN sys_menu.created_at IS '创建时间';

-- 同一父节点下菜单名不同
ALTER TABLE sys_menu ADD CONSTRAINT uk_menu_parent_name UNIQUE (parent_id, menu_name);

-- 创建角色表
CREATE TABLE IF NOT EXISTS sys_role (
    role_id BIGSERIAL PRIMARY KEY,
    role_name VARCHAR(100) NOT NULL UNIQUE,
    role_type VARCHAR(20) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE,
    is_default BOOLEAN DEFAULT FALSE,
    data_scope VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_by BIGINT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_by BIGINT
);

-- 添加列注释
COMMENT ON COLUMN sys_role.role_id IS '角色表唯一标识';
COMMENT ON COLUMN sys_role.role_name IS '角色名称';
COMMENT ON COLUMN sys_role.role_type IS '角色类型: platform-平台, seller-商户, system-系统';
COMMENT ON COLUMN sys_role.description IS '描述';
COMMENT ON COLUMN sys_role.is_system IS '是否系统内置角色';
COMMENT ON COLUMN sys_role.is_default IS '是否默认角色（新用户自动分配）';
COMMENT ON COLUMN sys_role.data_scope IS '数据权限范围: all-全部, dept-本部门, self-仅自己';
COMMENT ON COLUMN sys_role.created_at IS '创建时间';
COMMENT ON COLUMN sys_role.create_by IS '创建人 （新用户创建默认System）';
COMMENT ON COLUMN sys_role.updated_at IS '更新时间';
COMMENT ON COLUMN sys_role.update_by IS '更新人';

-- 创建角色权限关联表
CREATE TABLE IF NOT EXISTS sys_role_permission (
   id BIGSERIAL PRIMARY KEY,
   role_id BIGINT NOT NULL,
   permission_id BIGINT NOT NULL,
   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 添加列注释
COMMENT ON COLUMN sys_role_permission.id IS '角色权限表唯一标识';
COMMENT ON COLUMN sys_role_permission.role_id IS '角色表ID(外键)';
COMMENT ON COLUMN sys_role_permission.permission_id IS '权限表ID(外键)';
COMMENT ON COLUMN sys_role_permission.created_at IS '创建时间';

-- 添加唯一约束 + 外键约束
ALTER TABLE sys_role_permission ADD CONSTRAINT uk_role_permission UNIQUE (role_id, permission_id);
ALTER TABLE sys_role_permission ADD CONSTRAINT fk_role_permission_role FOREIGN KEY (role_id) REFERENCES sys_role(role_id);
ALTER TABLE sys_role_permission ADD CONSTRAINT fk_role_permission_permission FOREIGN KEY (permission_id) REFERENCES sys_permission(permission_id);

-- 创建用户角色关联表
CREATE TABLE IF NOT EXISTS sys_user_role (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 添加列注释
COMMENT ON COLUMN sys_user_role.id IS '用户角色表唯一标识';
COMMENT ON COLUMN sys_user_role.user_id IS '用户表ID(外键)';
COMMENT ON COLUMN sys_user_role.role_id IS '角色表ID(外键)';
COMMENT ON COLUMN sys_user_role.created_at IS '创建时间';

-- 添加唯一约束 + 外键约束
ALTER TABLE sys_user_role ADD CONSTRAINT uk_user_role UNIQUE (user_id, role_id);
ALTER TABLE sys_user_role ADD CONSTRAINT fk_user_role_user FOREIGN KEY (user_id) REFERENCES sys_user(user_id);
ALTER TABLE sys_user_role ADD CONSTRAINT fk_user_role_role FOREIGN KEY (role_id) REFERENCES sys_role(role_id);

-- 创建部门表（商户内部组织/平台组织）
CREATE TABLE IF NOT EXISTS sys_department (
    dept_id BIGSERIAL PRIMARY KEY,
    dept_name VARCHAR(100),
    parent_id BIGINT DEFAULT 0,
    dept_type VARCHAR(20) DEFAULT 'platform',
    leader_id BIGINT,
    sort_order INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 添加列注释
COMMENT ON COLUMN sys_department.dept_id IS '部门表唯一标识';
COMMENT ON COLUMN sys_department.dept_name IS '部门名称';
COMMENT ON COLUMN sys_department.parent_id IS '上级部门ID';
COMMENT ON COLUMN sys_department.dept_type IS '部门类型: platform-平台部门, seller-商户部门';
COMMENT ON COLUMN sys_department.leader_id IS '部门负责人';
COMMENT ON COLUMN sys_department.sort_order IS '排序';
COMMENT ON COLUMN sys_department.status IS '部门状态';
COMMENT ON COLUMN sys_department.created_at IS '创建时间';

-- 创建用户部门表
CREATE TABLE IF NOT EXISTS sys_user_dept (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    is_primary BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 添加列注释
COMMENT ON COLUMN sys_user_dept.id IS '用户部门表唯一标识';
COMMENT ON COLUMN sys_user_dept.user_id IS '用户表ID(外键)';
COMMENT ON COLUMN sys_user_dept.role_id IS '部门表ID(外键)';
COMMENT ON COLUMN sys_user_dept.is_primary IS '是否主部门';
COMMENT ON COLUMN sys_user_dept.created_at IS '创建时间';

-- 外键约束
ALTER TABLE sys_user_dept ADD CONSTRAINT fk_user_dept_user FOREIGN KEY (user_id) REFERENCES sys_user(user_id);
ALTER TABLE sys_user_dept ADD CONSTRAINT fk_user_dept_dept FOREIGN KEY (role_id) REFERENCES sys_department(dept_id);

-- 创建数据权限规则表
CREATE TABLE IF NOT EXISTS sys_data_scope (
    scope_id BIGSERIAL PRIMARY KEY,
    role_id BIGINT NOT NULL,
    resource_type VARCHAR(50),
    field_name VARCHAR(50),
    condition_type VARCHAR(50),
    condition_value TEXT,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 添加列注释
COMMENT ON COLUMN sys_data_scope.scope_id IS '唯一标识';
COMMENT ON COLUMN sys_data_scope.role_id IS '角色ID（外键)';
COMMENT ON COLUMN sys_data_scope.resource_type IS '资源类型: order/product/user表等';
COMMENT ON COLUMN sys_data_scope.field_name IS '字段名';
COMMENT ON COLUMN sys_data_scope.condition_type IS '条件类型: eq/ne/gt/lt/like/in';
COMMENT ON COLUMN sys_data_scope.condition_value IS '条件值';
COMMENT ON COLUMN sys_data_scope.description IS '描述';
COMMENT ON COLUMN sys_data_scope.created_at IS '创建时间';

-- 外键约束
ALTER TABLE sys_data_scope ADD CONSTRAINT fk_data_scope_role FOREIGN KEY (role_id) REFERENCES sys_role(role_id);

-- 创建操作日志表
CREATE TABLE IF NOT EXISTS sys_operation_log (
    log_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    username VARCHAR(50) NOT NULL,
    module VARCHAR(50) NOT NULL,
    operation VARCHAR(50) NOT NULL,
    request_method VARCHAR(50),
    request_url VARCHAR(50),
    request_params TEXT,
    ip_address VARCHAR(50),
    user_agent VARCHAR(500),
    execute_time INT,
    status BOOLEAN DEFAULT true,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 添加列注释
COMMENT ON COLUMN sys_operation_log.log_id IS '日志表唯一标识';
COMMENT ON COLUMN sys_operation_log.user_id IS '用户ID（外键）';
COMMENT ON COLUMN sys_operation_log.username IS '用户名';
COMMENT ON COLUMN sys_operation_log.module IS '操作模块';
COMMENT ON COLUMN sys_operation_log.operation IS '操作类型';
COMMENT ON COLUMN sys_operation_log.request_method IS '请求方法';
COMMENT ON COLUMN sys_operation_log.request_url IS '请求URL';
COMMENT ON COLUMN sys_operation_log.request_params IS '请求参数';
COMMENT ON COLUMN sys_operation_log.ip_address IS 'IP地址';
COMMENT ON COLUMN sys_operation_log.user_agent IS '用户代理';
COMMENT ON COLUMN sys_operation_log.execute_time IS '执行时间(ms)';
COMMENT ON COLUMN sys_operation_log.status IS '操作状态';
COMMENT ON COLUMN sys_operation_log.error_message IS '错误信息';
COMMENT ON COLUMN sys_operation_log.created_at IS '创建时间';

-- 外键约束
ALTER TABLE sys_operation_log ADD CONSTRAINT fk_op_log_user FOREIGN KEY (user_id) REFERENCES sys_user(user_id);
