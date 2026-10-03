package model

import "errors"

// 库存错误码，编号与文案与单体 model 包保持一致。
// RPC 的 error_msg 传的是文案，调用方按文案还原成本地错误变量，
// 因此文案必须逐字一致（errors.Is 判等才成立）。
var (
	ErrSkuNotExist        = errors.New("SKU不存在")     // 7001
	ErrStockNegative      = errors.New("调整后库存不能为负数") // 7002
	ErrStockNotEnough     = errors.New("库存不足")       // 7004
	ErrSkuDisabled        = errors.New("SKU已禁用或已删除") // 7005
	ErrLockStockNotEnough = errors.New("锁定库存不足")     // 7006
	// ErrRemarkEmpty 手动调整库存必须填原因(7003)
	ErrRemarkEmpty = errors.New("调整原因不能为空")
)

// 类目域错误码。单体未编号,此处沿用原文案。
var (
	CategoryDisable        = errors.New("父级目录已被禁用")
	CategoryUkExist        = errors.New("同级类目名称重复")
	CategoryParentNotExist = errors.New("父类目不存在")
	CategoryLevelDeep      = errors.New("类目层级超过4级")
	CategoryNotExist       = errors.New("类目不存在")
	CategoryHasChildren    = errors.New("该类目下有子类目，不能删除")
	CategoryHasRel         = errors.New("该类目已关联商品/属性，不能删除")
	CategoryParentInvalid  = errors.New("此父节点无效")
)

// 商品域错误码。单体未编号,此处沿用原文案。
var (
	ProductNotExist            = errors.New("商品不存在")
	ErrSpuDisabled             = errors.New("SKU已禁用或已删除")
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
