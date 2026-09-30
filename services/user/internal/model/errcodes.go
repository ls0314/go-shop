package model

import "errors"

var (
	// 权限域
	PermissionExist        = errors.New("权限标识已存在")
	PermissionNotExist     = errors.New("权限不存在")
	PermissionIsSystem     = errors.New("系统权限禁止修改")
	PermissionCodeNotAlter = errors.New("权限代码禁止修改")
	PermissionHasRel       = errors.New("权限存在角色关联")

	// 角色域
	RoleExist    = errors.New("此级目录内该角色已存在")
	RoleNotExist = errors.New("角色不存在")
	RoleIsSystem = errors.New("系统角色禁止修改")

	// 角色的关联检查(删角色时)。
	RoleHasUserRel = errors.New("存在关联用户")
	RoleHasMenuRel = errors.New("菜单存在角色关联")
	RoleHasPermRel = errors.New("权限存在角色关联")

	// 菜单域
	MenuExist    = errors.New("此级目录内菜单已存在")
	MenuNotExist = errors.New("菜单不存在")
	MenuHasRel   = errors.New("菜单存在角色关联")

	// 部门域
	DeptExist    = errors.New("此级目录内该部门已存在")
	DeptNotExist = errors.New("部门不存在")
	DeptHasRel   = errors.New("部门存在用户关联")

	// 数据权限域
	ScopeExist    = errors.New("数据权限标识已存在")
	ScopeNotExist = errors.New("数据权限不存在")
	ScopeNotRole  = errors.New("数据权限所关联的角色不存在")
	ScopeIsRole   = errors.New("禁止修改关联角色")

	// 用户域(绑定域校验用)
	UserNotExist = errors.New("用户不存在")
)
