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

type PaymentRepo struct {
	db *gorm.DB
}

func NewPaymentRepo() *PaymentRepo {
	return &PaymentRepo{
		db: db.DB,
	}
}

func (p *PaymentRepo) WithTx(tx *gorm.DB) *PaymentRepo {
	return &PaymentRepo{
		db: tx,
	}
}

// ============================================================
// 创建订单相关表
// ============================================================

func (p *PaymentRepo) CreatePayment(payment *model.UserPayment) (int64, error) {
	err := p.db.Create(payment).Error
	if err != nil {
		return 0, err
	}
	return payment.PaymentId, err
}

// ============================================================
// 查询订单相关表
// ============================================================

func (p *PaymentRepo) GetPayment(paymentNo string) (*model.UserPayment, error) {
	var payment model.UserPayment
	err := p.db.Where("pay_no = ?", paymentNo).First(&payment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrPayRecordNoExist
		}
		return nil, err
	}
	return &payment, nil
}

func (p *PaymentRepo) GetPaymentByOrderId(orderId int64) (*model.UserPayment, error) {
	var payment model.UserPayment
	err := p.db.Model(model.UserPayment{}).Where("order_id = ? AND pay_status = ?", orderId, model.PayPending).First(&payment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

func (p *PaymentRepo) GetPaymentWithOrderNo(paymentNo string) (*response.GetPaymentResp, error) {
	paymentTable := model.UserPayment{}.TableName()
	orderTable := model.UserOrder{}.TableName()
	var payment response.GetPaymentResp

	err := p.db.Model(&model.UserPayment{}).
		Select(paymentTable+".payment_id",
			paymentTable+".pay_no",
			paymentTable+".order_id",
			paymentTable+".pay_method",
			paymentTable+".pay_amount",
			paymentTable+".pay_status",
			paymentTable+".pay_time",
			paymentTable+".trade_no",
			orderTable+".order_no").
		Joins("LEFT JOIN "+orderTable+" ON "+paymentTable+".order_id = "+orderTable+".order_id").
		Where(paymentTable+".pay_no = ? AND "+orderTable+".is_deleted = ?", paymentNo, false).
		First(&payment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrPayRecordNoExist
		}
		return nil, err
	}
	return &payment, nil
}

func (p *PaymentRepo) GetPaymentList(req requset.GetPaymentListReq) ([]response.GetPaymentList, int64, error) {
	orderTable := model.UserOrder{}.TableName()
	userTable := model.SysUser{}.TableName()
	paymentTable := model.UserPayment{}.TableName()

	var payments []response.GetPaymentList
	baseQuery := p.db.Model(&model.UserPayment{}).
		Select(paymentTable+".payment_id",
			paymentTable+".pay_no",
			paymentTable+".pay_method",
			paymentTable+".pay_amount",
			paymentTable+".pay_status",
			paymentTable+".pay_time",
			orderTable+".order_no",
			userTable+".username").
		Joins("LEFT JOIN " + orderTable + " ON " + paymentTable + ".order_id = " + orderTable + ".order_id").
		Joins("LEFT JOIN " + userTable + " ON " + paymentTable + ".user_id = " + userTable + ".user_id")

	if req.OrderNo != "" {
		baseQuery = baseQuery.Where(orderTable+".order_no = ?", req.OrderNo)
	}
	if req.PayStatus != "" {
		baseQuery = baseQuery.Where(paymentTable+".pay_status = ?", req.PayStatus)
	}
	if req.PayMethod != "" {
		baseQuery = baseQuery.Where(paymentTable+".pay_method = ?", req.PayMethod)
	}
	if req.StartTime != nil {
		baseQuery = baseQuery.Where(paymentTable+".created_at >= ?", req.StartTime)
	}
	if req.EndTime != nil {
		endOfDay := req.EndTime.Truncate(24 * time.Hour).Add(24*time.Hour - time.Second)
		baseQuery = baseQuery.Where(paymentTable+".created_at <= ?", endOfDay)
	}

	// 获取总条数
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := req.Page
	pageSize := req.PageSize

	offset := (page - 1) * pageSize

	err := baseQuery.Order(paymentTable + ".created_at DESC").Offset(offset).Limit(pageSize).Find(&payments).Error

	return payments, total, err
}

// ============================================================
// 更新订单相关表
// ============================================================

func (p *PaymentRepo) CallbackPayment(req *requset.CallbackPaymentReq) error {
	return p.db.Model(model.UserPayment{}).
		Where("pay_no = ? AND pay_status = ?", req.PayNo, model.PayPending).
		Updates(map[string]interface{}{
			"pay_status": model.PaySuccess,
			"pay_time":   time.Now(),
			"trade_no":   req.TradeNo,
			"updated_at": time.Now(),
		}).Error
}
