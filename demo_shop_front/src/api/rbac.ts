import service from '@/utils/request'
import type {
    PageQuery,
    CreateUserReq,
    UpdateUserReq,
    CreateRoleReq,
    CreatePermissionReq,
    CreateMenuReq,
    CreateDeptReq,
    CreateScopeReq,
} from '@/types/rbac'

// ==================== 工具:分页拉全量 ====================
// 后端列表接口 pageSize 硬上限 100(>100 会被压成 10),
// 树形管理页需要全量数据,故循环翻页直到拉完 total
export async function fetchAll<T>(
    fetcher: (params: PageQuery) => Promise<any>,
    extra: PageQuery = {},
    pageSize = 100,
): Promise<T[]> {
    const list: T[] = []
    const first = await fetcher({ page: 1, pageSize, ...extra })
    const payload = first?.data?.data || {}
    list.push(...(payload.list || []))
    const total = payload.total || list.length
    const pages = Math.ceil(total / pageSize)
    for (let p = 2; p <= pages; p++) {
        const res = await fetcher({ page: p, pageSize, ...extra })
        list.push(...((res?.data?.data?.list) || []))
    }
    return list
}

// ==================== 用户管理 ====================

// 用户列表
export const GetUserListApi = (params: PageQuery = {}) => {
    return service({ url: '/admin/user', method: 'get', params })
}

// 用户详情
export const GetUserApi = (id: number) => {
    return service({ url: `/admin/user/${id}`, method: 'get' })
}

// 新增用户
export const CreateUserApi = (data: CreateUserReq) => {
    return service({ url: '/admin/user', method: 'post', data })
}

// 编辑用户
export const UpdateUserApi = (id: number, data: UpdateUserReq) => {
    return service({ url: `/admin/user/${id}`, method: 'put', data })
}

// 删除用户
export const DeleteUserApi = (id: number) => {
    return service({ url: `/admin/user/${id}`, method: 'delete' })
}

// 获取当前登录用户权限码(按钮级权限)
export const GetUserPermsApi = () => {
    return service({ url: '/user/perms', method: 'get' })
}

// ==================== 用户-角色关联 ====================

// 分配用户角色
export const AssignUserRolesApi = (user_id: number, role_ids: number[]) => {
    return service({ url: '/admin/user/assign-role', method: 'post', data: { user_id, role_ids } })
}

// 查询用户已有角色ID列表
export const GetUserRoleIdsApi = (userId: number) => {
    return service({ url: `/admin/user/${userId}/role`, method: 'get' })
}

// 清空用户角色
export const ClearUserRolesApi = (userId: number) => {
    return service({ url: `/admin/user/${userId}/clear-role`, method: 'delete' })
}

// ==================== 角色管理 ====================

export const GetRoleListApi = (params: PageQuery = {}) => {
    return service({ url: '/admin/role', method: 'get', params })
}

export const GetRoleApi = (id: number) => {
    return service({ url: `/admin/role/${id}`, method: 'get' })
}

export const CreateRoleApi = (data: CreateRoleReq) => {
    return service({ url: '/admin/role', method: 'post', data })
}

export const UpdateRoleApi = (id: number, data: Partial<CreateRoleReq>) => {
    return service({ url: `/admin/role/${id}`, method: 'put', data })
}

export const DeleteRoleApi = (id: number) => {
    return service({ url: `/admin/role/${id}`, method: 'delete' })
}

// ==================== 角色-菜单/权限关联 ====================

// 分配角色菜单
export const AssignRoleMenusApi = (role_id: number, menu_ids: number[]) => {
    return service({ url: '/admin/role/assign-menu', method: 'post', data: { role_id, menu_ids } })
}

// 查询角色已有菜单ID列表
export const GetRoleMenuIdsApi = (roleId: number) => {
    return service({ url: `/admin/role/${roleId}/menu`, method: 'get' })
}

// 分配角色权限
export const AssignRolePermsApi = (role_id: number, perm_ids: number[]) => {
    return service({ url: '/admin/role/assign-perm', method: 'post', data: { role_id, perm_ids } })
}

// 查询角色已有权限ID列表
export const GetRolePermIdsApi = (roleId: number) => {
    return service({ url: `/admin/role/${roleId}/perm`, method: 'get' })
}

// ==================== 权限管理 ====================

export const GetPermissionListApi = (params: PageQuery = {}) => {
    return service({ url: '/admin/permissions', method: 'get', params })
}

export const GetPermissionApi = (id: number) => {
    return service({ url: `/admin/permissions/${id}`, method: 'get' })
}

export const CreatePermissionApi = (data: CreatePermissionReq) => {
    return service({ url: '/admin/permissions', method: 'post', data })
}

export const UpdatePermissionApi = (id: number, data: Partial<CreatePermissionReq>) => {
    return service({ url: `/admin/permissions/${id}`, method: 'put', data })
}

export const DeletePermissionApi = (id: number) => {
    return service({ url: `/admin/permissions/${id}`, method: 'delete' })
}

// ==================== 菜单管理 ====================

export const GetMenuListApi = (params: PageQuery = {}) => {
    return service({ url: '/admin/menu', method: 'get', params })
}

export const GetMenuApi = (id: number) => {
    return service({ url: `/admin/menu/${id}`, method: 'get' })
}

export const CreateMenuApi = (data: CreateMenuReq) => {
    return service({ url: '/admin/menu', method: 'post', data })
}

export const UpdateMenuApi = (id: number, data: Partial<CreateMenuReq>) => {
    return service({ url: `/admin/menu/${id}`, method: 'put', data })
}

export const DeleteMenuApi = (id: number) => {
    return service({ url: `/admin/menu/${id}`, method: 'delete' })
}

// ==================== 部门管理 ====================

export const GetDeptListApi = (params: PageQuery = {}) => {
    return service({ url: '/admin/dept', method: 'get', params })
}

export const GetDeptApi = (id: number) => {
    return service({ url: `/admin/dept/${id}`, method: 'get' })
}

export const CreateDeptApi = (data: CreateDeptReq) => {
    return service({ url: '/admin/dept', method: 'post', data })
}

export const UpdateDeptApi = (id: number, data: Partial<CreateDeptReq>) => {
    return service({ url: `/admin/dept/${id}`, method: 'put', data })
}

export const DeleteDeptApi = (id: number) => {
    return service({ url: `/admin/dept/${id}`, method: 'delete' })
}

// ==================== 数据权限 scope ====================

export const GetScopeListApi = (params: PageQuery = {}) => {
    return service({ url: '/admin/scope', method: 'get', params })
}

export const GetScopeApi = (id: number) => {
    return service({ url: `/admin/scope/${id}`, method: 'get' })
}

export const CreateScopeApi = (data: CreateScopeReq) => {
    return service({ url: '/admin/scope', method: 'post', data })
}

export const UpdateScopeApi = (id: number, data: Partial<CreateScopeReq>) => {
    return service({ url: `/admin/scope/${id}`, method: 'put', data })
}

export const DeleteScopeApi = (id: number) => {
    return service({ url: `/admin/scope/${id}`, method: 'delete' })
}
