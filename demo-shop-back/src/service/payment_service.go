package service

import (
	"context"
	"demo-shop-back/src/infra/pay"
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/model/response"
	"demo-shop-back/src/repository"
	"fmt"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

// ============================================================
// 支付模块服务层定义及实例化
// ============================================================

// PaymentService 支付模块服务层实例
type PaymentService struct {
	PaymentRepo *repository.PaymentRepo
	OrderRepo   *repository.OrderRepo
	db          *gorm.DB
	*InventoryService
}

// NewPaymentService 创建支付模块服务层实例
// 接收值：deps - 服务层依赖（由 composition root 注入）
// 返回值：*PaymentService - 支付模块服务层实例指针
func NewPaymentService(deps ServiceDeps) *PaymentService {
	return &PaymentService{
		PaymentRepo:      repository.NewPaymentRepo(deps.DB),
		OrderRepo:        repository.NewOrderRepo(deps.DB),
		db:               deps.DB,
		InventoryService: NewInventoryService(deps),
	}
}

// ============================================================
// 发起支付
// ============================================================

// CreatePayment 对指定订单发起支付
// 流程：查询订单 → 归属校验 → 状态校验 → 防重 → 创建支付记录 → [调用网关]
//
// 接收值：
//
//	orderId - 订单ID
//	userId  - 当前用户ID
//	req     - 支付请求（含支付方式）
//
// 返回值：
//
//	*response.CreatePaymentResp - 支付参数（mock返回记录信息，真实支付返回支付链接）
//	error                        - 错误信息
func (p *PaymentService) CreatePayment(orderId, userId int64, req *requset.CreatePaymentReq) (*response.CreatePaymentResp, error) {
	// 查询订单
	order, err := p.OrderRepo.GetOrder(orderId)
	if err != nil {
		return nil, err
	}

	// 校验订单归属
	if order.UserId != userId {
		return nil, model.ErrPayNoPermission
	}

	// 校验订单状态：仅 pending_pay 可发起支付
	if order.OrderStatus != model.OrderPendingPay {
		return nil, model.ErrPayStatusMisTake
	}

	// 防重：同订单已有 pending 支付记录则直接返回
	paymentExist, err := p.PaymentRepo.GetPaymentByOrderId(orderId)
	if err != nil {
		return nil, err
	}
	if paymentExist != nil {
		return &response.CreatePaymentResp{
			PaymentId: paymentExist.PaymentId,
			PayNo:     paymentExist.PayNo,
			PayAmount: paymentExist.PayAmount,
			PayStatus: paymentExist.PayStatus,
		}, nil
	}

	// 生成支付流水号：PAY + yyyyMMddHHmmss + 4位随机数
	payNo := "PAY" + time.Now().Format("20060102150405") + fmt.Sprintf("%04d", rand.Intn(10000))

	// 创建支付记录（状态为 pending）
	payment := &model.UserPayment{
		UserId:    userId,
		OrderId:   orderId,
		PayNo:     payNo,
		PayMethod: req.PayMethod,
		PayAmount: order.PayAmount,
		PayStatus: model.PayPending,
		ExpireAt:  time.Now().Add(15 * time.Minute),
	}
	paymentId, err := p.PaymentRepo.CreatePayment(payment)
	if err != nil {
		return nil, err
	}

	// 根据支付方式返回不同结果
	switch req.PayMethod {
	case model.PayMethodMock:
		// mock 模式：返回支付记录信息，前端后续调回调接口完成支付
		return &response.CreatePaymentResp{
			PaymentId: paymentId,
			PayNo:     payment.PayNo,
			PayAmount: payment.PayAmount,
			PayStatus: payment.PayStatus,
		}, nil

	case model.PayMethodWechat, model.PayMethodAlipay:
		// 真实支付：调用网关生成支付链接
		gw, gwErr := pay.Get(req.PayMethod)
		if gwErr != nil {
			return nil, gwErr
		}
		payResp, gwErr := gw.CreatePayment(context.Background(), pay.PayRequest{
			PayNo:    payNo,
			OrderNo:  order.OrderNo,
			Amount:   fmt.Sprintf("%.2f", order.PayAmount),
			Subject:  "商品订单" + order.OrderNo,
			ExpireAt: payment.ExpireAt.Format(time.RFC3339),
		})
		if gwErr != nil {
			return nil, gwErr
		}
		return &response.CreatePaymentResp{
			PaymentId: paymentId,
			PayNo:     payment.PayNo,
			PayAmount: payment.PayAmount,
			PayStatus: payment.PayStatus,
			PayUrl:    payResp.PayUrl,
			QrCode:    payResp.QrCode,
		}, nil

	default:
		return nil, fmt.Errorf("不支持的支付方式: %s", req.PayMethod)
	}
}

// ============================================================
// 查询支付状态
// ============================================================

// GetPayment 根据支付流水号查询支付记录和当前状态
// 流程：查询支付记录 → 归属校验 → JOIN 订单号返回
//
// 接收值：
//
//	userId - 当前用户ID
//	payNo  - 支付流水号
//
// 返回值：
//
//	*response.GetPaymentResp - 支付记录详情（含订单号）
//	error                     - 错误信息
func (p *PaymentService) GetPayment(userId int64, payNo string) (*response.GetPaymentResp, error) {
	// 查询支付记录
	payment, err := p.PaymentRepo.GetPayment(payNo)
	if err != nil {
		return nil, model.ErrPayRecordNoExist
	}

	// 归属校验
	if payment.UserId != userId {
		return nil, model.ErrPayRecordNoNoPermission
	}

	// JOIN 订单表获取订单号
	paymentResp, err := p.PaymentRepo.GetPaymentWithOrderNo(payNo)
	if err != nil {
		return nil, err
	}
	return paymentResp, nil
}

// ============================================================
// 支付回调
// ============================================================

// HandleCallback 处理支付回调（mock / wechat / alipay 统一入口）
// 安全校验链：网关签名验证 → 幂等 → 金额校验 → 订单状态校验 → 事务更新
//
// 接收值：
//
//	payMethod      - 支付方式（mock/wechat/alipay）
//	callbackParams - 回调原始参数（mock 时为前端传入的 pay_no + trade_no）
//
// 返回值：error - 错误信息（nil 表示支付成功处理）
func (p *PaymentService) HandleCallback(payMethod string, callbackParams requset.CallbackPaymentReq) error {
	// 获取对应支付网关
	gw, err := pay.Get(payMethod)
	if err != nil {
		return err
	}

	// 网关解析回调参数（含签名验证，mock 网关直接透传）
	callbackData, err := gw.ParsePayment(context.Background(), callbackParams)
	if err != nil {
		return err
	}

	// 查询支付记录
	payment, err := p.PaymentRepo.GetPayment(callbackData.PayNo)
	if err != nil {
		return model.ErrPayRecordNoExist
	}

	// 幂等检查：已成功的支付记录直接返回（不报错，返回 nil 表示成功）
	if payment.PayStatus == model.PaySuccess {
		return nil
	}

	// 支付状态校验：仅 pending 可转为 success
	if payment.PayStatus != model.PayPending {
		return model.ErrPayStatusMisTake
	}

	// 金额校验：非 mock 模式下比对回调金额与支付记录金额
	if payMethod != model.PayMethodMock && callbackData.Amount > 0 && callbackData.Amount != payment.PayAmount {
		return model.ErrPayAmountMisTake
	}

	// 查询关联订单
	order, err := p.OrderRepo.GetOrder(payment.OrderId)
	if err != nil {
		return err
	}

	// 订单状态校验：仅 pending_pay 可转为 paid
	if order.OrderStatus != model.OrderPendingPay {
		return model.ErrOrderCannotPay
	}

	// 金额复核：支付记录金额与订单金额必须一致
	if payment.PayAmount != order.PayAmount {
		return model.ErrPayAmountMisTake
	}

	// 开启事务：更新支付记录 + 更新订单 + 订单日志 + 扣减库存
	return p.db.Transaction(func(tx *gorm.DB) error {
		paymentTx := p.PaymentRepo.WithTx(tx)
		orderTx := p.OrderRepo.WithTx(tx)

		// 更新支付记录 → success（使用网关解析后的 TradeNo，非原始参数）
		if err := paymentTx.CallbackPayment(&requset.CallbackPaymentReq{
			PayNo:   callbackData.PayNo,
			TradeNo: callbackData.TradeNo,
		}); err != nil {
			return err
		}

		// 更新订单 → paid（含 pay_method 和 pay_time）
		if err := orderTx.PayOrder(order.OrderId, payMethod); err != nil {
			return err
		}

		// 写入订单日志
		if err := orderTx.CreateOrderLog(model.UserOrderLog{
			OrderId:     order.OrderId,
			OrderStatus: model.OrderPaid,
			Action:      model.OrderPay,
			Operator:    "system",
			Detail:      callbackData.TradeNo,
		}); err != nil {
			return err
		}

		// 获取订单明细 → 逐项扣减锁定库存（共享外层事务）
		orderDetails, err := orderTx.GetOrderDetail(payment.OrderId)
		if err != nil {
			return err
		}
		for _, detail := range orderDetails {
			if err := p.InventoryService.DeductStockWithTx(tx, detail.SkuId, detail.Quantity, payment.OrderId); err != nil {
				return err
			}
		}

		return nil
	})
}

// ============================================================
// 管理端支付列表
// ============================================================

// GetPaymentList 管理端分页查看所有支付记录
// 接收值：
//
//	req - 查询参数（分页、状态筛选、方式筛选、订单号、时间范围）
//
// 返回值：
//
//	*response.GetPaymentListResp - 分页支付列表
//	error                         - 错误信息
func (p *PaymentService) GetPaymentList(req requset.GetPaymentListReq) (*response.GetPaymentListResp, error) {
	// 防止参数越界
	if req.Page <= 0 {
		req.Page = 1
	}
	// 防参数越界:<=0 用默认 20;>100 封顶 100(而非压成 20,避免大 pageSize 反而返回最少)
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// 调用数据层查询
	payments, total, err := p.PaymentRepo.GetPaymentList(req)
	if err != nil {
		return nil, err
	}

	return &response.GetPaymentListResp{
		List:     payments,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}
