package service

import (
	"context"
	"demo-shop-back/src/infra/cache"
	"demo-shop-back/src/infra/mq"
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/model/response"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/utils"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// OrderService 订单模块服务层实例
type OrderService struct {
	OrderRepo         *repository.OrderRepo
	AddressRepo       *repository.AddressRepo
	ProductRepo       *repository.ProductRepo
	CouponRepo        *repository.CouponRepo
	OutboxMessageRepo *repository.OutboxMessageRepo
	db                *gorm.DB
	cache             *cache.RedisService // 库存闸门(deps.GateCache,含熔断语义);nil = 闸门关闭
	*CartItemService
	*InventoryService
}

// NewOrderService 创建订单服务层实例
// 接收值：deps - 服务层依赖（由 composition root 注入）
// 返回值：*OrderService - 订单服务层实例指针
func NewOrderService(deps ServiceDeps) *OrderService {
	order := &OrderService{
		OrderRepo:         repository.NewOrderRepo(deps.DB),
		AddressRepo:       repository.NewAddressRepo(deps.DB),
		CouponRepo:        repository.NewCouponRepo(deps.DB),
		OutboxMessageRepo: repository.NewOutboxMessage(deps.DB),
		db:                deps.DB,
		cache:             deps.GateCache,
		CartItemService:   NewCartItemService(deps),
		InventoryService:  NewInventoryService(deps),
	}
	return order
}

// CreateOrder 创建订单
// 流程：幂等检查 → 校验地址 → 查询购物车选中项 → 逐项校验 → 雪花算法生成订单号 → 事务写入
// 接收值：
//
//	req    - 创建订单请求（含地址ID、幂等键、买家备注）
//	userId - 下单用户ID
//
// 返回值：
//
//	*response.CreateOrderResp - 创建成功返回订单信息
//	error                     - 错误信息
func (o *OrderService) CreateOrder(req *requset.CreatOrderReq, userId int64, userName string) (*response.CreateOrderResp, error) {
	// 幂等检查：idempotent_key 已存在则直接返回已有订单（幂等返回，不报错）
	if existOrder, err := o.OrderRepo.GetOrderIdempotentKey(req.IdempotentKey); err == nil && existOrder != nil {
		return &response.CreateOrderResp{
			OrderId:     existOrder.OrderId,
			OrderNo:     existOrder.OrderNo,
			PayAmount:   existOrder.PayAmount,
			TotalAmount: existOrder.TotalAmount,
			OrderStatus: existOrder.OrderStatus,
			PayExpireAt: existOrder.CreatedAt.Add(model.OrderPayTTL), // 订单创建15分钟后支付截至
			CreatedAt:   existOrder.CreatedAt,
		}, nil
	}

	// 校验收货地址
	address, err := o.AddressRepo.GetAddressSnap(req.AddressId, userId)
	if err != nil {
		return nil, model.ErrAddressNotExist
	}

	// 构建地址快照实例
	addressSnap := model.AddressSnap{
		ReceiverName:  address.ReceiverName,
		ReceiverPhone: address.ReceiverPhone,
		Province:      address.Province,
		City:          address.City,
		District:      address.District,
		DetailAddress: address.DetailAddress,
		PostalCode:    address.PostalCode,
	}

	// 序列化地址快照方便传入数据库
	jsonAddress, err := json.Marshal(addressSnap)
	if err != nil {
		return nil, err
	}

	// 查询购物车选中项
	cartItemList, err := o.CartItemRepo.GetSelectCartItem(userId)
	if err != nil {
		return nil, err
	}
	if len(cartItemList) == 0 {
		return nil, model.ErrCartNoSettlementItems
	}

	// 逐项校验 + 计算总金额
	var totalAmount float64
	for _, cartItem := range cartItemList {
		// 校验购买数量
		if err := o.validateQuantity(cartItem.Quantity); err != nil {
			return nil, err
		}
		// 校验商品状态
		if _, err := o.validateProductAvailable(cartItem.SkuId); err != nil {
			return nil, err
		}
		// 校验库存
		if cartItem.Stock < cartItem.Quantity {
			return nil, model.ErrStockNotEnough
		}
		totalAmount += cartItem.Price * float64(cartItem.Quantity)
	}

	// 雪花算法生成订单号
	orderNo := utils.NextOrderNo()

	// 订单中第一个商品图片
	firstImage := ""
	if len(cartItemList) > 0 {
		if cartItemList[0].SkuImage != "" {
			firstImage = cartItemList[0].SkuImage
		} else {
			firstImage = cartItemList[0].MainImage
		}
	}

	// 库存阀门预扣
	//gateDeducted 收集本请求已预扣的 (skuId, qty):事务失败时逐项补偿
	type skuDeduction struct {
		skuId, qty int64
	}
	var gateDeducted []skuDeduction
	cacheSvc := o.cache
	if cacheSvc != nil {
		ctx := context.Background()
		gateBroken := false // 中途 Redis 异常:还掉已扣的,本单整体降级为无闸门走 DB
		for _, item := range cartItemList {
			code, err := cacheSvc.DeductSkuStock(ctx, item.SkuId, item.Quantity)
			if err != nil {
				gateBroken = true
				log.Printf("[WARN] 库存闸门异常,本单降级直走 DB: skuId=%d err=%v", item.SkuId, err)
				break
			}
			if code == cache.GateBackfill {
				// 回填值直接用购物车查询带出的 item.Stock(与 DB 同源,允许瞬时偏差):
				// 复用既有查询结果,回填路径不新增 DB 查询
				cacheSvc.FillGateCounter(ctx, fmt.Sprintf("sku:stock:%d", item.SkuId), item.Stock, 30*time.Second)
				code, err = cacheSvc.DeductSkuStock(ctx, item.SkuId, item.Quantity)
				if err != nil {
					gateBroken = true
					break
				}
			}
			switch {
			case code > 0:
				gateDeducted = append(gateDeducted, skuDeduction{item.SkuId, item.Quantity})
			case code == cache.GateSoldOut:
				// 任一 SKU 不足:还掉已扣的再快速失败 —— 此时还没触碰 DB 行锁,
				// 拒绝成本是一次 Redis 往返 + N 次 Redis 补偿,对比原方案(打到行锁排队)近乎免费
				for _, d := range gateDeducted {
					_ = cacheSvc.CompensateSkuStock(ctx, d.skuId, d.qty)
				}
				return nil, model.ErrStockNotEnough
			}
		}
		if gateBroken {
			for _, d := range gateDeducted {
				_ = cacheSvc.CompensateSkuStock(ctx, d.skuId, d.qty)
			}
			gateDeducted = nil // 降级:本单不再视为已过闸门,事务失败后也就无需补偿
		}
	}
	// 构建订单实体
	order := &model.UserOrder{
		OrderNo:         orderNo,
		UserId:          userId,
		OrderStatus:     model.OrderPendingPay,
		TotalAmount:     totalAmount,
		PayAmount:       totalAmount,
		AddressSnapshot: jsonAddress,
		BuyerRemark:     req.BuyerRemark,
		IdempotentKey:   req.IdempotentKey,
		DetailCount:     int64(len(cartItemList)),
		FirstImage:      firstImage,
	}

	// 开启事务
	var resp response.CreateOrderResp
	err = o.db.Transaction(func(tx *gorm.DB) error {
		orderTx := o.OrderRepo.WithTx(tx)
		cartItemTx := o.CartItemRepo.WithTx(tx)
		couponTx := o.CouponRepo.WithTx(tx)
		outBoxTx := o.OutboxMessageRepo.WithTx(tx)

		// 使用优惠券
		if req.UserCouponId != 0 {
			coupon, err := couponTx.GetUserCoupon(req.UserCouponId)
			if err != nil {
				return err
			}
			if userId != coupon.UserId {
				return model.ErrUseCouponNoNoPermission
			}
			rowsAffected, err := couponTx.UseCoupon(req.UserCouponId, orderNo)
			if err != nil {
				return err
			}
			if rowsAffected == 0 {
				return model.ErrCouponNotExistOrUsed
			}
			var payAmount float64
			if totalAmount >= coupon.ThresholdAmount {
				if coupon.CouponType == "full_reduction" {
					payAmount = totalAmount - coupon.DiscountAmount
				} else {
					payAmount = totalAmount * coupon.DiscountAmount
				}
			} else {
				return model.ErrCouponThresholdNotMet
			}
			order.PayAmount = payAmount
		}

		// 写入订单主表
		orderId, err := orderTx.CreateOrder(order)
		if err != nil {
			return err
		}

		// 写入订单明细 + 锁定库存 + 清理购物车（全部共享外层事务）
		for _, cartItem := range cartItemList {
			orderDetail := model.UserOrderDetail{
				OrderId:    orderId,
				SkuId:      cartItem.SkuId,
				SpuName:    cartItem.SpuName,
				SkuName:    cartItem.SkuName,
				SpecValues: cartItem.SpecValues,
				MainImage:  cartItem.MainImage,
				Quantity:   cartItem.Quantity,
				UnitPrice:  cartItem.Price,
				TotalPrice: cartItem.Price * float64(cartItem.Quantity),
			}
			// 创建订单明细
			if err := orderTx.CreateOrderDetail(orderDetail); err != nil {
				return err
			}

			// 锁定库存（共享外层事务 tx）
			if err := o.InventoryService.LockStockWithTx(tx, cartItem.SkuId, cartItem.Quantity, orderId); err != nil {
				return err
			}

			// 删除购物车项
			if err := cartItemTx.DeleteCartItem(cartItem.CartItemId); err != nil {
				return err
			}
		}

		// 写入订单日志
		err = orderTx.CreateOrderLog(model.UserOrderLog{
			OrderId:     orderId,
			OrderStatus: model.OrderPendingPay,
			Action:      model.OrderCreate,
			Operator:    userName,
		})
		if err != nil {
			return err
		}

		if err := outBoxTx.CreateOutbox(model.OutBoxMessage{
			MessageId:     fmt.Sprintf("delay:%d", orderId),
			AggregateType: model.OutboxAggregateOrder,
			AggregateId:   strconv.FormatInt(orderId, 10),
			EventType:     model.OutboxEventOrderDelayCancel,
			Exchange:      "",
			RoutingKey:    mq.QueueOrderDelay,
			PayLoad:       strconv.FormatInt(orderId, 10),
			Status:        model.OutboxPending,
		}); err != nil {
			return err
		}

		// 构建响应
		resp = response.CreateOrderResp{
			OrderId:     orderId,
			OrderNo:     orderNo,
			TotalAmount: totalAmount,
			PayAmount:   order.PayAmount,
			OrderStatus: model.OrderPendingPay,
			PayExpireAt: time.Now().Add(model.OrderPayTTL),
			CreatedAt:   time.Now(),
		}

		return nil
	})
	if err != nil {
		// 闸门补偿:下单事务失败 → 逐项归还预扣的库存额度(与领券闸门同一补偿语义:
		// 只还「真的扣过」的;失败不重试,少放行方向安全,对账兜底)
		if cacheSvc != nil {
			for _, d := range gateDeducted {
				if cErr := cacheSvc.CompensateSkuStock(context.Background(), d.skuId, d.qty); cErr != nil {
					log.Printf("[WARN] 库存闸门补偿失败(等待对账收敛): skuId=%d err=%v", d.skuId, cErr)
				}
			}
		}
		return nil, err
	}
	return &resp, nil
}

// GetUserOrderList 用户分页查询自身订单列表
// 接收值：
//
//	req    - 查询订单列表请求（含地址ID、幂等键、买家备注）
//	userId - 下单用户ID
//
// 返回值：
//
//	*response.UserGetOrderListResp - 返回用户自身分页订单列表
//	error                     - 错误信息
func (o *OrderService) GetUserOrderList(userId int64, req requset.UserGetOrderListReq) (*response.UserGetOrderListResp, error) {
	// 防止参数越界
	if req.Page <= 0 {
		req.Page = 1
	}
	// 防参数越界:<=0 用默认 10;>100 封顶 100(而非压成 10,避免大 pageSize 反而返回最少)
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// 调用数据层获得订单分页列表
	orderList, total, err := o.OrderRepo.GetUserOrderList(userId, req)
	if err != nil {
		return nil, err
	}

	// 构建查询响应
	resp := response.UserGetOrderListResp{
		List:     orderList,
		Page:     req.Page,
		PageSize: req.PageSize,
		Total:    total,
	}
	return &resp, nil
}

// GetUserOrder 用户查询自身某一订单详情(返回订单信息,订单明细列表,订单日志列表)
// 接收值：
//
//	orderId - 查询订单ID
//	userId - 下单用户ID
//
// 返回值：
//
//	*response.UserGetOrderResp - 返回用户查询的订单详细信息
//	error                     - 错误信息
func (o *OrderService) GetUserOrder(orderId, userId int64) (*response.UserGetOrderResp, error) {
	// 调用数据层查询订单主表信息
	order, err := o.OrderRepo.GetOrder(orderId)
	if err != nil {
		return nil, err
	}

	// 归属校验
	if userId != order.UserId {
		return nil, model.ErrOrderNoPermission
	}

	// 调用数据层查询订单明细表信息
	orderDetailList, err := o.OrderRepo.GetOrderDetail(orderId)
	if err != nil {
		return nil, err
	}
	// 调用数据层查询订单日志表信息
	orderLogList, err := o.OrderRepo.GetOrderLog(orderId)
	if err != nil {
		return nil, err
	}

	// 构建响应
	resp := response.UserGetOrderResp{
		OrderId:         orderId,
		OrderNo:         order.OrderNo,
		OrderStatus:     order.OrderStatus,
		TotalAmount:     order.TotalAmount,
		PayAmount:       order.PayAmount,
		PayMethod:       order.PayMethod,
		PayTime:         order.PayTime,
		AddressSnapshot: order.AddressSnapshot,
		BuyerRemark:     order.BuyerRemark,
		DetailList:      orderDetailList,
		LogList:         orderLogList,
		CreatedAt:       order.CreatedAt,
	}

	return &resp, nil
}

// CancelOrder 取消订单(用户主动取消路径,带归属校验)
// 接收值：
//
//	orderId - 查询订单ID
//	userId - 下单用户ID
//	userName - 下单用户名
//
// 返回值：
//
//	*response.OrderStatusResp - 返回订单操作后状态信息
//	error                     - 错误信息
func (o *OrderService) CancelOrder(orderId, userId int64, userName string) (*response.OrderStatusResp, error) {
	// 调用数据层查询订单主表信息
	order, err := o.OrderRepo.GetOrder(orderId)
	if err != nil {
		return nil, err
	}
	// 归属校验
	if userId != order.UserId {
		return nil, model.ErrOrderNoPermission
	}
	// 状态更改校验 - 只有pending_pay才能被取消
	if order.OrderStatus != model.OrderPendingPay {
		return nil, model.ErrOrderCannotCancel
	}

	return o.cancel(order, userName)
}

// CancelOrderBySystem 系统自动取消(MQ 超时消费者 / 定时扫描任务)
func (o *OrderService) CancelOrderBySystem(orderId int64, operator string) (*response.OrderStatusResp, error) {
	order, err := o.OrderRepo.GetOrder(orderId)
	if err != nil {
		return nil, err
	}
	if order.OrderStatus != model.OrderPendingPay {
		return nil, model.ErrOrderCannotCancel
	}
	return o.cancel(order, operator)
}

// cancel 取消订单的共享事务体(不做归属校验、不做状态校验,调用方必须先校验)
// 接收值:order - 已加载的订单主表信息;operator - 操作人(用户名或"系统")
func (o *OrderService) cancel(order *model.UserOrder, operator string) (*response.OrderStatusResp, error) {
	orderId := order.OrderId
	var resp response.OrderStatusResp
	err := o.db.Transaction(func(tx *gorm.DB) error {
		orderTx := o.OrderRepo.WithTx(tx)
		couponTx := o.CouponRepo.WithTx(tx)

		// 调用数据层事务取消订单
		rows, err := orderTx.CancelOrder(orderId)
		if err != nil {
			return err
		}
		if rows == 0 {
			return model.ErrOrderAlreadyCancelled
		}

		if order.PayAmount != order.TotalAmount {
			coupon, err := couponTx.GetUserCouponByOrderNo(order.OrderNo)
			if err != nil {
				return err
			}
			rowsAffected, err := couponTx.RefundCoupon(coupon.UserCouponId)
			if err != nil {
				return err
			}
			if rowsAffected == 0 {
				return model.ErrCannotCancelCoupon
			}
		}

		// 操作写入订单日志
		err = orderTx.CreateOrderLog(model.UserOrderLog{
			OrderId:     orderId,
			OrderStatus: model.OrderCancelled,
			Action:      model.OrderCancel,
			Operator:    operator,
		})
		if err != nil {
			return err
		}
		// 调用数据层获取订单明细列表，后逐个释放锁定的库存
		orderDetailList, err := orderTx.GetOrderDetail(orderId)
		if err != nil {
			return err
		}
		for _, orderDetail := range orderDetailList {
			// 调用库存内部接口释放锁定库存（共享外部事务）
			err := o.ReleaseStockWithTx(tx, orderDetail.SkuId, orderDetail.Quantity, orderId)
			if err != nil {
				return err
			}
		}
		// 构建响应
		resp = response.OrderStatusResp{
			OrderId:     orderId,
			OrderNo:     order.OrderNo,
			OrderStatus: model.OrderCancelled,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetOrderList 管理端按条件查询全部订单
// 接收值：
//
//	req - 查询参数（分页，分页大小，订单状态，订单号和时间范围）
//
// 返回值：
//
//	*response.GetOrderListResp - 查询到符合条件的订单列表
//	error                     - 错误信息
func (o *OrderService) GetOrderList(req requset.GetOrderListReq) (*response.GetOrderListResp, error) {
	// 防止参数越界
	if req.Page <= 0 {
		req.Page = 1
	}
	// 防参数越界:<=0 用默认 10;>100 封顶 100(而非压成 10,避免大 pageSize 反而返回最少)
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// 调用数据层获取符合条件的订单主表列表
	orderList, total, err := o.OrderRepo.GetOrderList(req)
	if err != nil {
		return nil, err
	}

	// 构建响应
	resp := response.GetOrderListResp{
		Page:     req.Page,
		PageSize: req.PageSize,
		Total:    total,
		List:     orderList,
	}
	return &resp, nil
}

// GetOrder 管理端查询订单详细信息（较用户端多出订单所属的用户ID和用户名）
// 接收值：
//
//	orderId - 被查询订单ID
//
// 返回值：
//
//	*response.GetOrderResp - 查询到的订单详情
//	error                     - 错误信息
func (o *OrderService) GetOrder(orderId int64) (*response.GetOrderResp, error) {

	// 调用数据层获取出明细表和日志表的其余订单信息
	order, err := o.OrderRepo.GetOrderWithUserName(orderId)
	if err != nil {
		return nil, err
	}

	// 调用数据层获取订单明细列表
	orderDetailList, err := o.OrderRepo.GetOrderDetail(orderId)
	if err != nil {
		return nil, err
	}

	// 调用数据层获取订单日志列表
	orderLogList, err := o.OrderRepo.GetOrderLog(orderId)
	if err != nil {
		return nil, err
	}

	// 构建响应
	resp := response.GetOrderResp{
		OrderId:         orderId,
		OrderNo:         order.OrderNo,
		OrderStatus:     order.OrderStatus,
		TotalAmount:     order.TotalAmount,
		PayAmount:       order.PayAmount,
		PayMethod:       order.PayMethod,
		PayTime:         order.PayTime,
		AddressSnapshot: order.AddressSnapshot,
		BuyerRemark:     order.BuyerRemark,
		DetailList:      orderDetailList,
		LogList:         orderLogList,
		CreatedAt:       order.CreatedAt,
		UserId:          order.UserId,
		Username:        order.Username,
	}
	return &resp, nil
}

// OrderShip 管理端查询发货
// 接收值：
//
//	orderId - 被查询订单ID
//	userName - 发货操作人
//	req - 请求参数（快递公司和快递单号）
//
// 返回值：
//
//	*response.OrderShipResp - 订单状态和快递信息
//	error                     - 错误信息
func (o *OrderService) OrderShip(orderId int64, userName string, req requset.OrderShipReq) (*response.OrderShipResp, error) {
	// 校验参数
	if req.TrackingNO == "" || req.ExpressCompany == "" {
		return nil, model.ErrExpressIncomplete
	}

	// 调用数据层查询订单信息
	order, err := o.OrderRepo.GetOrder(orderId)
	if err != nil {
		return nil, err
	}

	// 校验订单状态 - 只有已支付的订单才能进行发货操作
	if order.OrderStatus != model.OrderPaid {
		return nil, model.ErrOrderCannotShip
	}

	// 开启事务
	err = o.db.Transaction(func(tx *gorm.DB) error {
		orderTx := o.OrderRepo.WithTx(tx)

		// 调用数据层执行修改订单状态为已发货
		err := orderTx.ShipOrder(orderId)
		if err != nil {
			return err
		}

		// 将操作写入订单日志
		err = orderTx.CreateOrderLog(model.UserOrderLog{
			OrderId:     orderId,
			OrderStatus: model.OrderShipped,
			Action:      model.OrderShip,
			Operator:    userName,
			Detail:      req.ExpressCompany + req.TrackingNO,
		})
		if err != nil {
			return err
		}
		return nil
	})

	// 构建响应
	resp := response.OrderShipResp{
		OrderId:        orderId,
		OrderNo:        order.OrderNo,
		OrderStatus:    model.OrderShipped,
		ExpressCompany: req.ExpressCompany,
		TrackingNO:     req.TrackingNO,
	}

	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ConfirmOrder 用户端确认收货
// 接收值：
//
//	orderId - 被查询订单ID
//	userName - 收货人用户名
//	userId - 收货人用户Id
//
// 返回值：
//
//	*response.OrderStatusResp - 订单状态
//	error                     - 错误信息
func (o *OrderService) ConfirmOrder(orderId, userId int64, userName string) (*response.OrderStatusResp, error) {
	// 调用数据层获取订单信息
	order, err := o.OrderRepo.GetOrder(orderId)
	if err != nil {
		return nil, err
	}

	// 归属校验
	if userId != order.UserId {
		return nil, model.ErrOrderNoPermission
	}

	// 订单状态校验 - 只有已发货的订单才能被确认
	if order.OrderStatus != model.OrderShipped {
		return nil, model.ErrOrderCannotConfirm
	}

	// 开启事务
	err = o.db.Transaction(func(tx *gorm.DB) error {
		orderTx := o.OrderRepo.WithTx(tx)

		// 调用数据层更改订单状态为以收货
		if err := orderTx.ConfirmOrder(orderId); err != nil {
			return err
		}

		// 将操作写入订单日志表
		err := orderTx.CreateOrderLog(model.UserOrderLog{
			OrderId:     orderId,
			OrderStatus: model.OrderCompleted,
			Action:      model.OrderConfirm,
			Operator:    userName,
		})
		if err != nil {
			return err
		}
		return nil
	})
	// 构建响应
	resp := response.OrderStatusResp{
		OrderId:     orderId,
		OrderNo:     order.OrderNo,
		OrderStatus: model.OrderCompleted,
	}
	return &resp, nil
}
