package repository

import (
	"time"

	"demo-shop/services/trade/internal/model"

	"gorm.io/gorm"
)

// OrderRepo 订单三表数据层实例。
type OrderRepo struct {
	DB *gorm.DB
}

// NewOrderRepo 创建订单表数据层实例
func NewOrderRepo(conn *gorm.DB) *OrderRepo {
	return &OrderRepo{DB: conn}
}

// WithTx 切换数据库事务实例
func (o *OrderRepo) WithTx(tx *gorm.DB) *OrderRepo {
	return &OrderRepo{DB: tx}
}

// CreateOrder 写入订单主表,回填自增主键
func (o *OrderRepo) CreateOrder(order *model.UserOrder) error {
	return o.DB.Create(order).Error
}

// CreateOrderDetail 写入订单明细
func (o *OrderRepo) CreateOrderDetail(detail *model.UserOrderDetail) error {
	return o.DB.Create(detail).Error
}

// CreateOrderLog 写入订单日志(Saga 每步补偿都会落一条)
func (o *OrderRepo) CreateOrderLog(log *model.UserOrderLog) error {
	return o.DB.Create(log).Error
}

// GetOrderById 根据ID查订单。查不到透传 gorm.ErrRecordNotFound,由调用方判语义。
func (o *OrderRepo) GetOrderById(orderId int64) (*model.UserOrder, error) {
	var order model.UserOrder
	if err := o.DB.Where("order_id = ?", orderId).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// GetOrderByNo 根据订单号查订单
func (o *OrderRepo) GetOrderByNo(orderNo string) (*model.UserOrder, error) {
	var order model.UserOrder
	if err := o.DB.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// GetOrderByIdempotentKey 按幂等键查订单(重复下单检测)。
// 查不到返回 gorm.ErrRecordNotFound —— 那是正常路径(首次下单)。
func (o *OrderRepo) GetOrderByIdempotentKey(key string) (*model.UserOrder, error) {
	var order model.UserOrder
	if err := o.DB.Where("idempotent_key = ?", key).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// GetOrderDetailList 查订单明细
func (o *OrderRepo) GetOrderDetailList(orderId int64) ([]*model.UserOrderDetail, error) {
	var list []*model.UserOrderDetail
	err := o.DB.Where("order_id = ?", orderId).Find(&list).Error
	return list, err
}

// GetOrderLogList 查订单日志(按时间正序,便于阅读"这单经历了什么")
func (o *OrderRepo) GetOrderLogList(orderId int64) ([]*model.UserOrderLog, error) {
	var list []*model.UserOrderLog
	err := o.DB.Where("order_id = ?", orderId).Order("created_at ASC").Find(&list).Error
	return list, err
}

// OrderListQuery 订单列表查询条件。
type OrderListQuery struct {
	Page        int
	PageSize    int
	OrderStatus string
	OrderNo     string
	StartTime   *time.Time
	EndTime     *time.Time
}

// GetOrderList 分页查询订单。userId > 0 时只查该用户的(用户端);为 0 查全部(管理端)。
func (o *OrderRepo) GetOrderList(userId int64, q OrderListQuery) ([]*model.UserOrder, int64, error) {
	query := o.DB.Model(&model.UserOrder{})
	if userId > 0 {
		query = query.Where("user_id = ?", userId)
	}
	if q.OrderStatus != "" {
		query = query.Where("order_status = ?", q.OrderStatus)
	}
	if q.OrderNo != "" {
		query = query.Where("order_no = ?", q.OrderNo)
	}
	if q.StartTime != nil {
		query = query.Where("created_at >= ?", *q.StartTime)
	}
	if q.EndTime != nil {
		query = query.Where("created_at <= ?", *q.EndTime)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*model.UserOrder
	offset := (q.Page - 1) * q.PageSize
	err := query.Order("created_at DESC").Limit(q.PageSize).Offset(offset).Find(&list).Error
	return list, total, err
}

// ListExpirePendingPay 查已过支付截止时间、仍待支付的订单(超时扫描用)。
func (o *OrderRepo) ListExpirePendingPay(now time.Time, limit int) ([]*model.UserOrder, error) {
	var list []*model.UserOrder
	err := o.DB.Where("order_status = ? AND expire_at <= ?", model.OrderPendingPay, now).
		Order("expire_at ASC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

// CancelOrder 取消订单(条件更新:仅 pending_pay 可取消)。
func (o *OrderRepo) CancelOrder(orderId int64) (int64, error) {
	res := o.DB.Model(&model.UserOrder{}).
		Where("order_id = ? AND order_status = ?", orderId, model.OrderPendingPay).
		Update("order_status", model.OrderCancelled)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

// UpdateOrderStatus 条件推进订单状态(支付回调/发货/确认收货共用)。
func (o *OrderRepo) UpdateOrderStatus(orderId int64, fromStatus, toStatus string) (int64, error) {
	res := o.DB.Model(&model.UserOrder{}).
		Where("order_id = ? AND order_status = ?", orderId, fromStatus).
		Update("order_status", toStatus)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

// UpdateOrderPaid 支付成功:推进状态并写入支付方式与支付时间。
func (o *OrderRepo) UpdateOrderPaid(orderId int64, payMethod string, payTime time.Time) (int64, error) {
	res := o.DB.Model(&model.UserOrder{}).
		Where("order_id = ? AND order_status = ?", orderId, model.OrderPendingPay).
		Updates(map[string]interface{}{
			"order_status": model.OrderPaid,
			"pay_method":   payMethod,
			"pay_time":     payTime,
		})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}
