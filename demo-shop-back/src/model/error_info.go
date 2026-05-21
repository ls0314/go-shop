package model

import "errors"

//错误信息待整理 存在错误信息重复

var (
	UserNotLogin           = errors.New("用户未登录")
	UserInfoError          = errors.New("用户信息异常")
	HasNotPerm             = errors.New("权限校验失败")
	UserHasNotPerm         = errors.New("用户无操作权限")
	RegPasswordInvalid     = errors.New("密码必须同时包含字母数字标点符号且>=8位")
	LoginPasswordInvalid   = errors.New("用户名或密码错误")
	PhoneMalformed         = errors.New("手机号格式错误")
	PhoneExist             = errors.New("手机号已注册")
	UsernameExist          = errors.New("用户名已存在")
	UserNotExist           = errors.New("用户不存在")
	EmailExist             = errors.New("邮箱已注册")
	UserIdIsSystem         = errors.New("用户ID禁止修改")
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
	RoleExist              = errors.New("此级目录内该角色已存在")
	RoleNotExist           = errors.New("角色不存在")
	RoleHasRel             = errors.New("角色存在用户关联")
	RoleIsSystem           = errors.New("系统角色禁止修改")
	DeptExist              = errors.New("此级目录内该部门已存在")
	DeptNotExist           = errors.New("部门不存在")
	DeptHasRel             = errors.New("部门存在用户关联")
	DeptIsSystem           = errors.New("系统部门禁止修改")
	ScopeExist             = errors.New("数据权限标识已存在")
	ScopeNotExist          = errors.New("数据权限不存在")
	ScopeNotRole           = errors.New("数据权限所关联的角色不存在")
	ScopeIsRole            = errors.New("禁止修改关联角色")
	RelExist               = errors.New("此关联已存在")
	RelNotExist            = errors.New("此关联不存在")
	UserHasRel             = errors.New("存在关联用户")
	CategoryDisable        = errors.New("父级目录已被禁用")
	CategoryUkExist        = errors.New("同级类目名称重复")
	CategoryParentNotExist = errors.New("父类目不存在")
	CategoryLevelDeep      = errors.New("类目层级超过4级")
	CategoryNotExist       = errors.New("类目不存在")
	CategoryHasChildren    = errors.New("该类目下有子类目，不能删除")
	CategoryHasRel         = errors.New("该类目已关联商品/属性，不能删除")
	CategoryParentInvalid  = errors.New("此父节点无效")
)
var (
	StatusIdNotExist          = "ID不存在"
	StatusInternalServerError = "服务器错误"
	StatusBadRequest          = "请求参数错误"
	StatusNotExistRequest     = "请求内容不存在"
)
