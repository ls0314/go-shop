package model

import "errors"

var (
	PermissionExist        = errors.New("权限标识已存在")
	PermissionNotExist     = errors.New("权限不存在")
	PermissionIsSystem     = errors.New("系统权限禁止修改")
	PermissionCodeNotAlter = errors.New("权限代码禁止修改")
	PermissionHasRel       = errors.New("权限存在角色关联")
)
