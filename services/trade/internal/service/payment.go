package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"demo-shop/services/trade/internal/model"
	"demo-shop/services/trade/internal/repository"
	"demo-shop/services/trade/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// ============================================================
// 支付领域操作
// ============================================================

// PayNoPrefix 支付单号前缀,便于人工区分支付流水与订单号
const PayNoPrefix = "PAY"

// CreatePayment 发起支付。
func (p *PaymentService) CreatePayment(ctx context.Context, orderId, userId int64, payMethod string) (*model.UserPaymentRecord, error) {
	order, err := p.orderRepo.GetOrderById(orderId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrOrderNotExist
		}
		return nil, err
	}
	// 归属校验:防替他人订单发起支付
	if order.UserId != userId {
		return nil, model.ErrPayNoPermission
	}
	if order.OrderStatus != model.OrderPendingPay {
		return nil, model.ErrOrderCannotPay
	}

	// 已有进行中的支付记录就复用,不再新建
	if existing, err := p.paymentRepo.GetPendingByOrderId(orderId); err == nil {
		return existing, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 支付单号由本服务生成:它要写进渠道请求、也要用于回调反查,
	// 必须是"发起时就确定"的值
	payNo := PayNoPrefix + utils.FormatOrderNo(p.idGen.NextId())

	record := &model.UserPaymentRecord{
		PayNo:     payNo,
		OrderId:   orderId,
		UserId:    userId,
		PayMethod: payMethod,
		PayAmount: order.PayAmount,
		PayStatus: model.PayPending,
		// 支付过期时间从订单的 expire_at 派生
		ExpireAt: order.ExpireAt,
	}
	if err := p.paymentRepo.CreatePayment(record); err != nil {
		return nil, err
	}
	return record, nil
}

// GetPayment 按支付单号查流水(校验归属)
func (p *PaymentService) GetPayment(userId int64, payNo string) (*model.UserPaymentRecord, error) {
	record, err := p.paymentRepo.GetPaymentByNo(payNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrPayRecordNoExist
		}
		return nil, err
	}
	if record.UserId != userId {
		return nil, model.ErrPayRecordNoNoPermission
	}
	return record, nil
}

// HandleCallbackResult 回调处理结果
type HandleCallbackResult struct {
	// Accepted 回调是否被受理
	Accepted bool
	// Idempotent true 表示本次未产生新变更(重复回调)
	Idempotent bool
}

// HandleCallback 渠道回调。
func (p *PaymentService) HandleCallback(ctx context.Context, payNo, tradeNo string) (*HandleCallbackResult, error) {
	record, err := p.paymentRepo.GetPaymentByNo(payNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrPayRecordNoExist
		}
		return nil, err
	}

	order, err := p.orderRepo.GetOrderById(record.OrderId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrOrderNotExist
		}
		return nil, err
	}

	// ---- 本地事务:支付成功 + 订单推进 + 日志 ----
	var (
		orderDetails []*model.UserOrderDetail
		applied      bool
	)
	err = p.paymentRepo.DB.Transaction(func(tx *gorm.DB) error {
		paymentTx := p.paymentRepo.WithTx(tx)
		orderTx := p.orderRepo.WithTx(tx)

		// 条件更新:仅 pending 可被标记成功。重复回调时 rows=0
		rows, err := paymentTx.MarkPaid(record.PaymentId, tradeNo, time.Now(), notifyLog(payNo, tradeNo))
		if err != nil {
			return err
		}
		if rows == 0 {
			// 重复回调:幂等命中,不产生新变更
			return nil
		}
		applied = true

		// 订单推进:仅 pending_pay → paid。若订单已被取消(超时扫描抢先),
		// rows=0 —— 这时不能把支付丢掉:钱已经收了。
		// 单体在这里返回 ErrOrderCannotPay 让渠道重试,行为一致
		oRows, err := orderTx.UpdateOrderPaid(order.OrderId, record.PayMethod, time.Now())
		if err != nil {
			return err
		}
		if oRows == 0 {
			return model.ErrOrderCannotPay
		}

		if err := orderTx.CreateOrderLog(&model.UserOrderLog{
			OrderId:     order.OrderId,
			OrderStatus: model.OrderPaid,
			Action:      model.OrderActionPay,
			Operator:    "system",
			Detail:      tradeNo,
		}); err != nil {
			return err
		}

		list, err := orderTx.GetOrderDetailList(order.OrderId)
		if err != nil {
			return err
		}
		orderDetails = list
		return nil
	})
	if err != nil {
		return nil, err
	}
	if !applied {
		// 重复回调:受理且幂等,不重复扣库存
		return &HandleCallbackResult{Accepted: true, Idempotent: true}, nil
	}

	// ---- 事务外:把锁定库存转为实际扣减 ----
	//
	// 用订单上的幂等键:它与下单锁库存、取消释放用的是同一个键,
	// 三种操作的 change_type 不同,product 侧的索引把它们分开,互不遮挡
	deducted := true
	for _, d := range orderDetails {
		if err := p.inventoryRPC.DeductStock(d.SkuId, d.Quantity, order.IdempotentKey); err != nil {
			// 不反向补偿(不能退钱),只告警 —— 由对账按
			// "pay_status=success 但库存未扣"收敛
			logx.Errorf("支付后扣减库存失败(等待对账收敛): orderId=%d skuId=%d qty=%d err=%v",
				order.OrderId, d.SkuId, d.Quantity, err)
			deducted = false
		}
	}
	if !deducted {
		// 返回错误让渠道重试回调 —— 这是**向前补偿**:
		// 重放会被 pay_status 的幂等挡住(不会重复扣款),
		// 但会把没扣成功的库存再试一次
		return nil, model.ErrCannotDeductAfterPay
	}

	return &HandleCallbackResult{Accepted: true, Idempotent: false}, nil
}

// ListPayments 管理端支付流水分页查询
func (p *PaymentService) ListPayments(page, pageSize int, payStatus, payMethod, orderNo string,
	startTime, endTime *time.Time) ([]*model.PaymentView, int64, int, int, error) {
	page, pageSize = normalizePage(page, pageSize)

	items, total, err := p.paymentRepo.GetPaymentList(repository.PaymentListQuery{
		Page:      page,
		PageSize:  pageSize,
		PayStatus: payStatus,
		PayMethod: payMethod,
		OrderNo:   orderNo,
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		return nil, 0, 0, 0, err
	}
	return items, total, page, pageSize, nil
}

// notifyLog 回调原始数据(落 notify_log 列)。
func notifyLog(payNo, tradeNo string) []byte {
	b, err := json.Marshal(map[string]string{
		"pay_no":      payNo,
		"trade_no":    tradeNo,
		"received_at": time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return nil
	}
	return b
}
