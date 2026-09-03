package service

import (
	"demo-shop-back/db"
	"demo-shop-back/src/infra"
	"demo-shop-back/src/infra/mq"
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/model/response"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/utils"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// OrderService 订单模块服务层实例
type OrderService struct {
	OrderRepo   *repository.OrderRepo
	AddressRepo *repository.AddressRepo
	ProductRepo *repository.ProductRepo
	CouponRepo  *repository.CouponRepo
	db          *gorm.DB
	*CartItemService
	*InventoryService
}

// NewOrderService 创建订单服务层实例
// 接收值：使用全局repository初始化，故无接收值
// 返回值：*OrderService - 订单服务层实例指针
func NewOrderService() *OrderService {
	order := &OrderService{
		OrderRepo:        repository.NewOrderRepo(),
		AddressRepo:      repository.NewAddressRepo(),
		CouponRepo:       repository.NewCouponRepo(),
		db:               db.DB,
		CartItemService:  NewCartItemService(),
		InventoryService: NewInventoryService(),
	}
	mq.RegisterCanceller(order)
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
			PayExpireAt: existOrder.CreatedAt.Add(15 * time.Minute), // 订单创建15分钟后支付截至
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

		// 构建响应
		resp = response.CreateOrderResp{
			OrderId:     orderId,
			OrderNo:     orderNo,
			TotalAmount: totalAmount,
			PayAmount:   order.PayAmount,
			OrderStatus: model.OrderPendingPay,
			PayExpireAt: time.Now().Add(15 * time.Minute),
			CreatedAt:   time.Now(),
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	// 创建订单成功后将订单ID传入消息队列,检查订单超时
	if MQ := infra.GetMQ(); MQ != nil {
		_ = MQ.PublishOrderDelay(resp.OrderId)
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

// CancelOrder 取消订单
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

	var resp response.OrderStatusResp

	// 开启事务
	err = o.db.Transaction(func(tx *gorm.DB) error {
		orderTx := o.OrderRepo.WithTx(tx)
		couponTx := o.CouponRepo.WithTx(tx)

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

		// 调用数据层事务取消订单
		if err := orderTx.CancelOrder(orderId); err != nil {
			return err
		}

		// 操作写入订单日志
		err := orderTx.CreateOrderLog(model.UserOrderLog{
			OrderId:     orderId,
			OrderStatus: model.OrderCancelled,
			Action:      model.OrderCancel,
			Operator:    userName,
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
