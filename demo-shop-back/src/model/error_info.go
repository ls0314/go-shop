package model

import "errors"

//错误信息待整理 存在错误信息重复

var (
	RegPasswordInvalid     = errors.New("密码必须同时包含字母数字标点符号且>=8位")
	LoginPasswordInvalid   = errors.New("用户名或密码错误")
	PhoneMalformed         = errors.New("手机号格式错误")
	PhoneExist             = errors.New("手机号已注册")
	UsernameExist          = errors.New("用户名已存在")
	EmailExist             = errors.New("邮箱已注册")
	TokenExpired           = errors.New("token已过期")
	TokenNotValidYet       = errors.New("token尚未生效")
	TokenMalformed         = errors.New("token格式错误")
	TokenInvalid           = errors.New("token错误")
	PermissionExist        = errors.New("权限标识已存在")
	PermissionNotExist     = errors.New("权限不存在")
	PermissionIsSystem     = errors.New("系统权限禁止修改")
	PermissionCodeNotAlter = errors.New("权限代码禁止修改")
	PermissionHasRel       = errors.New("权限存在角色关联")
	MenuExist              = errors.New("此级目录内菜单已存在")
	MenuNotExist           = errors.New("菜单不存在")
	MenuHasRel             = errors.New("菜单存在角色关联")
	RoleExist              = errors.New("此级目录内角色已存在")
	RoleNotExist           = errors.New("角色不存在")
	RoleHasRel             = errors.New("角色存在用户关联")
	RoleIsSystem           = errors.New("系统角色禁止修改")
)
var (
	StatusIdNotExist          = "ID不存在"
	StatusInternalServerError = "服务器错误"
	StatusBadRequest          = "请求参数错误"
	StatusNotExistRequest     = "请求内容不存在"
)
