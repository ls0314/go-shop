package repository

import (
	"time"

	"demo-shop/services/trade/internal/model"

	"gorm.io/gorm"
)

// PaymentRepo 支付流水表数据层实例
type PaymentRepo struct {
	DB *gorm.DB
}

// NewPaymentRepo 创建支付流水表数据层实例
func NewPaymentRepo(conn *gorm.DB) *PaymentRepo {
	return &PaymentRepo{DB: conn}
}

// WithTx 切换数据库事务实例
func (p *PaymentRepo) WithTx(tx *gorm.DB) *PaymentRepo {
	return &PaymentRepo{DB: tx}
}

// CreatePayment 写入支付流水,回填自增主键
func (p *PaymentRepo) CreatePayment(record *model.UserPaymentRecord) error {
	return p.DB.Create(record).Error
}

// GetPaymentByNo 按支付流水号查。查不到透传 gorm.ErrRecordNotFound。
func (p *PaymentRepo) GetPaymentByNo(payNo string) (*model.UserPaymentRecord, error) {
	var record model.UserPaymentRecord
	if err := p.DB.Where("pay_no = ?", payNo).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// GetPaymentByTradeNo 按渠道交易号查
func (p *PaymentRepo) GetPaymentByTradeNo(tradeNo string) (*model.UserPaymentRecord, error) {
	var record model.UserPaymentRecord
	if err := p.DB.Where("trade_no = ?", tradeNo).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// GetPendingByOrderId 查该订单进行中的支付流水。
func (p *PaymentRepo) GetPendingByOrderId(orderId int64) (*model.UserPaymentRecord, error) {
	var record model.UserPaymentRecord
	err := p.DB.Where("order_id = ? AND pay_status = ?", orderId, model.PayPending).
		Order("created_at DESC").
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// MarkPaid 标记支付成功(条件更新:仅 pending 可被标记)。
func (p *PaymentRepo) MarkPaid(paymentId int64, tradeNo string, payTime time.Time, notifyLog []byte) (int64, error) {
	res := p.DB.Model(&model.UserPaymentRecord{}).
		Where("payment_id = ? AND pay_status = ?", paymentId, model.PayPending).
		Updates(map[string]interface{}{
			"pay_status": model.PaySuccess,
			"trade_no":   tradeNo,
			"pay_time":   payTime,
			"notify_log": notifyLog,
		})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

// PaymentListQuery 支付列表查询条件
type PaymentListQuery struct {
	Page      int
	PageSize  int
	PayStatus string
	PayMethod string
	OrderNo   string
	StartTime *time.Time
	EndTime   *time.Time
}

// GetPaymentList 分页查询支付流水,联订单表取订单号(同库)。
func (p *PaymentRepo) GetPaymentList(q PaymentListQuery) ([]*model.PaymentView, int64, error) {
	query := p.DB.Model(&model.UserPaymentRecord{}).
		Joins("LEFT JOIN user_order_master ON user_order_master.order_id = user_payment_record.order_id")
	if q.PayStatus != "" {
		query = query.Where("user_payment_record.pay_status = ?", q.PayStatus)
	}
	if q.PayMethod != "" {
		query = query.Where("user_payment_record.pay_method = ?", q.PayMethod)
	}
	if q.OrderNo != "" {
		query = query.Where("user_order_master.order_no = ?", q.OrderNo)
	}
	if q.StartTime != nil {
		query = query.Where("user_payment_record.created_at >= ?", *q.StartTime)
	}
	if q.EndTime != nil {
		query = query.Where("user_payment_record.created_at <= ?", *q.EndTime)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*model.PaymentView
	offset := (q.Page - 1) * q.PageSize
	err := query.
		Select("user_payment_record.*, user_order_master.order_no").
		Order("user_payment_record.created_at DESC").
		Limit(q.PageSize).Offset(offset).
		Find(&list).Error
	return list, total, err
}
