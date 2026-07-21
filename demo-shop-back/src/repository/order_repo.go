package repository

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/model/response"
	"errors"
	"time"

	"gorm.io/gorm"
)

// ============================================================
// 订单模块数据层定义及实例化部分
// ============================================================

// OrderRepo 订单模块数据层实例
type OrderRepo struct {
	db *gorm.DB
}

// NewOrderRepo 新建订单模块数据层实例
// 接收值：全局数据库操作 无接收值
// 返回值：*OrderRepo - 商品模块数据层实例指针
func NewOrderRepo() *OrderRepo {
	return &OrderRepo{
		db: db.DB,
	}
}

// WithTx 商品表事务实例
// 接收值：db - 数据库事务实例
// 返回值：*OrderRepo - 绑定事务的订单模块数据层指针
func (o *OrderRepo) WithTx(tx *gorm.DB) *OrderRepo {
	return &OrderRepo{db: tx}
}

// ============================================================
// 创建订单模块相关表
// ============================================================

// CreateOrder 创建订单信息主表
// 接收值： orderId - 订单唯一标识
// 返回值：
//
//	int64 - 订单Id
//	error - 错误信息
func (o *OrderRepo) CreateOrder(order *model.UserOrder) (int64, error) {
	err := o.db.Create(order).Error
	if err != nil {
		return 0, err
	}
	return order.OrderId, nil
}

// CreateOrderDetail 创建订单明细表
// 接收值： orderId - 订单唯一标识
// 返回值： error - 错误信息
func (o *OrderRepo) CreateOrderDetail(orderDetail model.UserOrderDetail) error {
	return o.db.Create(&orderDetail).Error
}

// CreateOrderLog 创建订单日志表
// 接收值： orderId - 订单唯一标识
// 返回值： error - 错误信息
func (o *OrderRepo) CreateOrderLog(orderLog model.UserOrderLog) error {
	return o.db.Create(&orderLog).Error
}

// ============================================================
// 查询订单相关表
// ============================================================

// GetOrder 查询订单信息主表（通过订单Id）
// 接收值： orderId - 订单唯一标识
// 返回值：
//
//	*model.UserOrder - 查询获得的订单信息
//	error - 错误信息
func (o *OrderRepo) GetOrder(orderId int64) (*model.UserOrder, error) {
	var order model.UserOrder
	err := o.db.Where("order_id = ? AND is_deleted = ?", orderId, false).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrOrderNotExist
		}
		return nil, err
	}
	return &order, nil
}

// GetOrderWithUserName 联表查询订单信息（联表获取订单对应的用名）
// 接收值： orderId - 订单唯一标识
// 返回值：
//
//	*response.GetOrderResp - 查询获得的订单信息
//	error - 错误信息
func (o *OrderRepo) GetOrderWithUserName(orderId int64) (*response.GetOrderResp, error) {
	var order response.GetOrderResp
	userTable := model.SysUser{}.TableName()
	orderTable := model.UserOrder{}.TableName()

	// 联表查询用户名
	err := o.db.Model(model.UserOrder{}).
		Select(orderTable+".*",
			userTable+".username").
		Joins("LEFT JOIN "+userTable+" ON "+userTable+".user_id = "+orderTable+".user_id").
		Where(orderTable+".order_id = ? AND "+orderTable+".is_deleted = ?", orderId, false).
		Find(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetOrderDetail 查询订单明细表列表（通过订单Id）
// 接收值： orderId - 订单唯一标识
// 返回值：
//
//	[]response.UserGetOrderDetail - 查询获得的订单明细表响应信息列表
//	error - 错误信息
func (o *OrderRepo) GetOrderDetail(orderId int64) ([]response.UserGetOrderDetail, error) {
	var orderDetailList []response.UserGetOrderDetail
	err := o.db.Model(model.UserOrderDetail{}).Where("order_id = ?", orderId).Find(&orderDetailList).Error
	if err != nil {
		return nil, err
	}
	return orderDetailList, nil
}

// GetOrderLog 查询订单日志表列表（通过订单Id）
// 接收值： orderId - 订单唯一标识
// 返回值：
//
//	[]response.UserGetOrderLog - 查询获得的订单日志表响应信息列表
//	error - 错误信息
func (o *OrderRepo) GetOrderLog(orderId int64) ([]response.UserGetOrderLog, error) {
	var orderLogList []response.UserGetOrderLog
	err := o.db.Model(model.UserOrderLog{}).Where("order_id = ?", orderId).Find(&orderLogList).Error
	if err != nil {
		return nil, err
	}
	return orderLogList, nil
}

// GetOrderIdempotentKey 查询订单信息（通过幂等键）
// 接收值： key - 幂等键
// 返回值：
//
//	*model.UserOrder - 幂等键对应的订单信息表（若存在则订单不予创建返回此幂等键对应的信息）
//	error - 错误信息
func (o *OrderRepo) GetOrderIdempotentKey(key string) (*model.UserOrder, error) {
	var order model.UserOrder
	err := o.db.Where("idempotent_key = ?", key).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// GetUserOrderList 查询用户所有订单信息列表（通过用户Id）
// 接收值：
//
//	userId - 用户Id
//	req - 筛选条件（分页，分页大小和订单状态）
//
// 返回值：
//
//	[]response.UserGetOrderList - 符合条件的用户所有的订单信息列表
//	int64 - 总数
//	error - 错误信息
func (o *OrderRepo) GetUserOrderList(userId int64, req requset.UserGetOrderListReq) ([]response.UserGetOrderList, int64, error) {
	var respList []response.UserGetOrderList

	// 构建基础查询链，筛选需要返回的字段
	baseQuery := o.db.Model(&model.UserOrder{}).
		Select("order_id, user_id, order_no, order_status, total_amount, pay_amount, detail_count, first_image, created_at").
		Where("is_deleted = ?", false)

	// 构建筛选条件
	if userId != 0 {
		baseQuery = baseQuery.Where("user_id = ?", userId)
	}

	if req.OrderStatus != "" {
		baseQuery = baseQuery.Where("order_status = ?", req.OrderStatus)
	}

	// 获取总条数
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := req.Page
	pageSize := req.PageSize
	offset := (page - 1) * pageSize

	// 调用查询链
	err := baseQuery.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&respList).Error

	return respList, total, err
}

// GetOrderList 查询全部用户的所有订单信息列表
// 接收值： req - 筛选条件（分页、分页大小、订单状态、订单号和订单时间范围）
//
// 返回值：
//
//	[]response.GetOrderList - 所有符合条件的订单信息列表
//	int64 - 总数
//	error - 错误信息
func (o *OrderRepo) GetOrderList(req requset.GetOrderListReq) ([]response.GetOrderList, int64, error) {
	userTable := model.SysUser{}.TableName()
	orderTable := model.UserOrder{}.TableName()
	var respList []response.GetOrderList

	// 联表获得订单对应用户名并筛选字段构建基础查询链
	baseQuery := o.db.Model(&model.UserOrder{}).
		Select(userTable+".username, "+
			orderTable+".order_id, "+
			orderTable+".order_no, "+
			orderTable+".order_status, "+
			orderTable+".total_amount, "+
			orderTable+".pay_amount, "+
			orderTable+".pay_method, "+
			orderTable+".address_snapshot->> 'receiver_name' AS receiver_name, "+
			orderTable+".address_snapshot->> 'receiver_phone' AS receiver_phone, "+
			orderTable+".created_at").
		Joins("LEFT JOIN "+userTable+" ON "+userTable+".user_id = "+orderTable+".user_id").
		Where(orderTable+".is_deleted = ?", false)

	// 构建筛选条件
	if req.OrderNo != "" {
		baseQuery = baseQuery.Where(".order_no = ?", req.OrderNo)
	}

	if req.OrderStatus != "" {
		baseQuery = baseQuery.Where(".order_status = ?", req.OrderStatus)
	}

	if req.StartTime != nil {
		baseQuery = baseQuery.Where(".created_at >= ?", req.StartTime)
	}

	if req.EndTime != nil {
		endOfDay := req.EndTime.Truncate(24 * time.Hour).Add(24*time.Hour - time.Second)
		baseQuery = baseQuery.Where(".created_at <= ?", endOfDay)
	}

	// 获取总条数
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := req.Page
	pageSize := req.PageSize
	offset := (page - 1) * pageSize
	// 调用查询链
	err := baseQuery.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&respList).Error

	return respList, total, err

}

// ============================================================
// 更新订单相关表
// ============================================================

// CancelOrder 取消订单
// 接收值： orderId - 订单Id
// 返回值： error - 错误信息
func (o *OrderRepo) CancelOrder(orderId int64) error {
	return o.db.Model(model.UserOrder{}).
		Where("order_id = ? AND order_status = ?", orderId, "pending_pay").
		Updates(map[string]interface{}{
			"order_status": "cancelled",
			"updated_at":   time.Now(), // 取消订单后更新操作时间
		}).Error
}

// ShipOrder 订单发货
// 接收值： orderId - 订单Id
// 返回值： error - 错误信息
func (o *OrderRepo) ShipOrder(orderId int64) error {
	return o.db.Model(model.UserOrder{}).
		Where("order_id = ? AND order_status = ?", orderId, "paid").
		Updates(map[string]interface{}{
			"order_status": "shipped",
			"updated_at":   time.Now(), // 取消订单后更新操作时间
		}).Error
}

// ConfirmOrder 确认收货
// 接收值： orderId - 订单Id
// 返回值： error - 错误信息
func (o *OrderRepo) ConfirmOrder(orderId int64) error {
	return o.db.Model(model.UserOrder{}).
		Where("order_id = ? AND order_status = ?", orderId, "shipped").
		Updates(map[string]interface{}{
			"order_status": "completed",
			"updated_at":   time.Now(), // 取消订单后更新操作时间
		}).Error
}
