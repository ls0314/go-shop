// ==================== RBAC 系统管理相关类型 ====================

// ==================== 用户管理 ====================

export interface SysUser {
    user_id: number
    username: string
    email: string
    phone: string
    status: string          // active / disabled
    failed_attempts: number
    lock_until: string | null
    created_at: string
}

export interface CreateUserReq {
    username: string
    password: string
    email?: string
    phone?: string
    status?: string
}

export interface UpdateUserReq {
    email?: string
    phone?: string
    status?: string
}

export interface UserListResp {
    list: SysUser[]
    total: number
    page: number
    pageSize: number
}

// ==================== 角色管理 ====================

export interface SysRole {
    role_id: number
    role_name: string
    role_type: string       // platform / custom
    description: string
    is_system: boolean
    is_default: boolean
    data_scope: string      // all / dept / self
    created_at: string
    updated_at: string
}

export interface CreateRoleReq {
    role_name: string
    role_type?: string
    description?: string
    data_scope?: string
}

export interface RoleListResp {
    list: SysRole[]
    total: number
    page: number
    pageSize: number
}

// ==================== 权限管理 ====================

export interface SysPermission {
    permission_id: number
    permission_code: string
    permission_name: string
    permission_type: string    // api / menu / button
    request_method: string     // GET / POST / PUT / DELETE
    api_path: string
    description: string
    is_system: boolean
    created_at: string
}

export interface CreatePermissionReq {
    permission_code: string
    permission_name: string
    permission_type?: string
    request_method?: string
    api_path: string
    description?: string
}

export interface PermissionListResp {
    list: SysPermission[]
    total: number
    page: number
    pageSize: number
}

// ==================== 菜单管理 ====================

export interface SysMenu {
    menu_id: number
    parent_id: number
    menu_name: string
    menu_type: string      // M(目录) / C(菜单) / F(按钮)
    icon?: string
    route_path?: string
    component?: string
    is_visible: boolean
    is_cache: boolean
    sort_order: number
    meta_info?: Record<string, any>
    created_at: string
    children?: SysMenu[] | null
}

export interface CreateMenuReq {
    parent_id: number
    menu_name: string
    menu_type: string
    icon?: string
    route_path?: string
    component?: string
    is_visible?: boolean
    is_cache?: boolean
    sort_order?: number
    meta_info?: Record<string, any>
}

// ==================== 部门管理 ====================

export interface SysDept {
    dept_id: number
    dept_name: string
    parent_id: number
    dept_type: string
    leader_id: number
    sort_order: number
    status: string
    created_at: string
    children?: SysDept[] | null
}

export interface CreateDeptReq {
    dept_name: string
    parent_id?: number
    dept_type?: string
    leader_id?: number
    sort_order?: number
    status?: string
}

// ==================== 数据权限(scope) ====================

export interface SysScope {
    scope_id: number
    role_id: number
    role_name?: string        // JOIN 角色名
    resource_type: string
    field_name: string
    condition_type: string
    condition_value: string
    description: string
    created_at: string
}

export interface CreateScopeReq {
    role_id: number
    resource_type?: string
    field_name?: string
    condition_type?: string
    condition_value?: string
    description?: string
}

// ==================== 关联接口通用响应 ====================

export interface AssignResp {
    id?: number
    role_id?: number
    user_id?: number
    menu_id?: number
    permission_id?: number
}

// ==================== 分页查询参数 ====================

export interface PageQuery {
    page?: number
    pageSize?: number
    [key: string]: any
}
