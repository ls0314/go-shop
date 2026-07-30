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
	ProductNotExist            = errors.New("商品不存在")
	ErrSpuTemplate             = errors.New("规格值与规格模板不匹配")
	ErrSkuNum                  = errors.New("SKU数量错误")
	ErrSpecValues              = errors.New("SKU规格值不匹配")
	ErrSkuCodeNotOnly          = errors.New("SKU编码需唯一")
	ErrSpecValuesNotOnly       = errors.New("SKU规格组合唯一")
	ErrNoActiveSku             = errors.New("至少需要一个启用状态的SKU")
	ErrNoAvailableStock        = errors.New("启用SKU的总库存必须大于0")
	ErrInvalidPrice            = errors.New("所有启用SKU的售价必须大于0")
	ErrCategoryNotUsed         = errors.New("类目不可用")
	ErrPublishedCantChangeSpec = errors.New("已上架商品不允许修改规格模板")
	ErrSkuListEmpty            = errors.New("修改后SKU列表为空")
	ErrInvalidStatusTransition = errors.New("商品状态转换不合法")
)

var (
	AddressNotExist   = errors.New("地址不存在")
	AddressListIsNull = errors.New("地址列表为空")
	AddressNumsIsFull = errors.New("地址数量超出限制")
	UserNotSetAddress = errors.New("用户越权修改地址")
	ReceiverNotNull   = errors.New("收货人姓名或手机号码不能为空")
)
var (
	StatusIdNotExist          = "ID不存在"
	StatusInternalServerError = "服务器错误"
	StatusBadRequest          = "请求参数错误"
	StatusNotExistRequest     = "请求内容不存在"
)

// 库存模块错误码 7001-7005
var (
	ErrSkuNotExist    = errors.New("SKU不存在")     // 7001
	ErrStockNegative  = errors.New("调整后库存不能为负数") // 7002
	ErrRemarkEmpty    = errors.New("调整原因不能为空")   // 7003
	ErrStockNotEnough = errors.New("库存不足")       // 7004
	ErrSkuDisabled    = errors.New("SKU已禁用或已删除") // 7005
)

// 购物车模块错误码
var (
	ErrSpuDisabled   = errors.New("SKU已禁用或已删除")
	CartItemMax      = errors.New("购物车数量已达上限（100条）")
	QuantityMax      = errors.New("购买数量超出限制（单SKU购买范围为0~999）")
	CartItemNotExist = errors.New("购物车项不存在")
	NotAuthority     = errors.New("无权操作该购物车项")
)

// 订单模块错误码
var (
	ErrAddressNotExist       = errors.New("收货地址不存在")
	ErrCartNoSettlementItems = errors.New("购物车无可结算商品")
	ErrDuplicateSubmit       = errors.New("重复提交（幂等键已使用）")
	ErrOrderNotExist         = errors.New("订单不存在")
	ErrOrderNoPermission     = errors.New("无权查看/操作该订单")
	ErrOrderCannotCancel     = errors.New("订单状态不允许取消（仅待支付可取消）")
	ErrOrderCannotShip       = errors.New("订单状态不允许发货（仅已支付可发货）")
	ErrExpressIncomplete     = errors.New("快递信息不完整")
	ErrOrderCannotConfirm    = errors.New("订单状态不允许确认收货（仅已发货可确认）")
)

// 支付模块错误码
var (
	ErrPayNoPermission         = errors.New("订单不属于当前用户")
	ErrOrderCannotPay          = errors.New("订单状态不允许支付（仅待支付可支付）")
	ErrPayRecordExisting       = errors.New("已有进行中的支付记录")
	ErrPayRecordNoExist        = errors.New("支付记录不存在")
	ErrPayRecordNoNoPermission = errors.New("无权查看该支付记录")
	ErrPayStatusMisTake        = errors.New("支付状态不正确（非pending，回调幂等）")
	ErrPayAmountMisTake        = errors.New("支付金额与订单金额不匹配")
)
