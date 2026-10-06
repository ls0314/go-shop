package tradeclient

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"demo-shop-back/src/model/response"
	v1_tradev1 "demo-shop/api/gen/trade/v1"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/datatypes"
)

// ErrUnavailable 客户端未建连(etcd 连不上 / 服务未注册)时的统一错误。
//
// 文案是契约:HTTP 层据此回 503,接线冒烟测试据此跳过该路由。
// 与 productclient / couponclient 的同名变量同一约定。
var ErrUnavailable = errors.New("trade-service 不可用")

// callTimeout 单次订单域 RPC 的超时。
//
// 比商品域长:**下单是 Saga 编排**(核销券 + 逐 SKU 锁库存 + 本地事务),
// 中间有多次跨服务往返。给 5s 是因为它可能真的在排队拿库存行锁,
// 超时太短会把"正在排队"误报成失败,进而触发不必要的补偿。
const callTimeout = 5 * time.Second

// TradeClient 订单域(trade-service)的 RPC 客户端。
//
// 覆盖该服务对外暴露的三个 gRPC service:购物车、订单、支付。
// 三者同库(trade_db),且下单流程跨越三者(建单时要删购物车行),
// 拆成多个客户端只会让调用方各自建连。
type TradeClient struct {
	cart    v1_tradev1.CartServiceClient
	order   v1_tradev1.OrderServiceClient
	payment v1_tradev1.PaymentServiceClient
	conn    *grpc.ClientConn
}

// NewTradeClient 建连 trade-service(etcd 服务发现)。
//
// **只建一条连接**:三个 gRPC service 由同一个进程提供、同一个 etcd key 发现。
func NewTradeClient(etcdHosts []string, etcdKey string) (*TradeClient, error) {
	client, err := zrpc.NewClient(zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{Hosts: etcdHosts, Key: etcdKey},
	})
	if err != nil {
		return nil, err
	}
	conn := client.Conn()
	return &TradeClient{
		cart:    v1_tradev1.NewCartServiceClient(conn),
		order:   v1_tradev1.NewOrderServiceClient(conn),
		payment: v1_tradev1.NewPaymentServiceClient(conn),
		conn:    conn,
	}, nil
}

func (c *TradeClient) Close() error { return c.conn.Close() }

// RestoreError 把服务端 error_msg 还原成普通错误。
//
// trade 侧的错误文案与单体逐字一致(都按单体 error_info.go 对齐),
// 故不需要映射表 —— 直接用文案构造即可,调用方按文案判等。
// 若需要 errors.Is 判等(如超时扫描要区分"已被处理"与"真故障"),
// 调用方可用 errors.Is(err, model.ErrOrderAlreadyCancelled) ——
// 那要求文案完全相同,这一点由两侧的 errcodes 保证。
func RestoreError(errorMsg string) error {
	if errorMsg == "" {
		return nil
	}
	if errorMsg == ErrUnavailable.Error() {
		return ErrUnavailable
	}
	return errors.New(errorMsg)
}

// ============================================================
// 购物车
// ============================================================

// CreateCartItem 加购,返回 (cartItemId, quantity, errMsg, err)
func (c *TradeClient) CreateCartItem(userId, skuId, quantity int64) (int64, int64, string, error) {
	if c == nil {
		return 0, 0, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.cart.CreateCartItem(ctx, &v1_tradev1.CreateCartItemReq{
		UserId:   userId,
		SkuId:    skuId,
		Quantity: quantity,
	})
	if err != nil {
		return 0, 0, "", err
	}
	if resp.ErrorMsg != "" {
		return 0, 0, resp.ErrorMsg, nil
	}
	return resp.GetItem().GetCartItemId(), resp.GetItem().GetQuantity(), "", nil
}

// GetCartItemList 购物车列表(不分页)
func (c *TradeClient) GetCartItemList(userId int64) ([]response.CartItemListResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.cart.ListCartItems(ctx, &v1_tradev1.ListCartItemsReq{UserId: userId})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelCartItems(resp.Items), "", nil
}

// UpdateCartItem 改数量 / 改选中态
func (c *TradeClient) UpdateCartItem(cartItemId, userId, quantity int64, isSelected bool) (*response.CartItemListResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.cart.UpdateCartItem(ctx, &v1_tradev1.UpdateCartItemReq{
		CartItemId: cartItemId,
		UserId:     userId,
		Quantity:   quantity,
		IsSelected: isSelected,
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	item := toModelCartItem(resp.Item)
	return &item, "", nil
}

// DeleteCartItem 删除购物车行
func (c *TradeClient) DeleteCartItem(cartItemId, userId int64) (string, error) {
	if c == nil {
		return "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.cart.DeleteCartItem(ctx, &v1_tradev1.DeleteCartItemReq{
		CartItemId: cartItemId,
		UserId:     userId,
	})
	if err != nil {
		return "", err
	}
	return resp.ErrorMsg, nil
}

// SelectAllCartItems 全选 / 取消全选,返回受影响行数
func (c *TradeClient) SelectAllCartItems(userId int64, isSelected bool) (int64, string, error) {
	if c == nil {
		return 0, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.cart.SelectAllCartItems(ctx, &v1_tradev1.SelectAllCartItemsReq{
		UserId:     userId,
		IsSelected: isSelected,
	})
	if err != nil {
		return 0, "", err
	}
	if resp.ErrorMsg != "" {
		return 0, resp.ErrorMsg, nil
	}
	return resp.Affected, "", nil
}

// GetCartItemCount 购物车行数(角标)
func (c *TradeClient) GetCartItemCount(userId int64) (int64, string, error) {
	if c == nil {
		return 0, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.cart.GetCartItemCount(ctx, &v1_tradev1.GetCartItemCountReq{UserId: userId})
	if err != nil {
		return 0, "", err
	}
	if resp.ErrorMsg != "" {
		return 0, resp.ErrorMsg, nil
	}
	return resp.Total, "", nil
}

// GetCartPayPreview 结算预览
func (c *TradeClient) GetCartPayPreview(userId int64) (*response.CartItemPayResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.cart.GetCartPayPreview(ctx, &v1_tradev1.GetCartPayPreviewReq{UserId: userId})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelPayPreview(resp.Preview), "", nil
}

// ============================================================
// 订单
// ============================================================

// CreateOrderReq 下单入参(HTTP 层已绑定的形状)
//
// 与单体的 requset.CreatOrderReq 字段一致,但**多带一个 AddressSnapshot**:
// 地址表在 user_db,trade 侧跨库取不到,故由调用方传快照 ——
// 这也更贴合语义(订单存的是下单那一刻的地址,用户随后改地址不该影响它)。
type CreateOrderReq struct {
	UserId          int64
	UserName        string
	AddressSnapshot AddressSnapshot
	IdempotentKey   string
	BuyerRemark     string
	UserCouponId    int64
}

// AddressSnapshot 收货地址快照(与 trade 侧 model.AddressSnap 字段一致)
type AddressSnapshot struct {
	ReceiverName  string
	ReceiverPhone string
	Province      string
	City          string
	District      string
	DetailAddress string
	PostalCode    string
}

// CreateOrder 下单(Saga 编排在 trade 侧)
func (c *TradeClient) CreateOrder(r *CreateOrderReq) (*response.CreateOrderResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.order.CreateOrder(ctx, &v1_tradev1.CreateOrderReq{
		UserId:        r.UserId,
		Username:      r.UserName,
		IdempotentKey: r.IdempotentKey,
		BuyerRemark:   r.BuyerRemark,
		UserCouponId:  r.UserCouponId,
		Address: &v1_tradev1.AddressSnapshot{
			ReceiverName:  r.AddressSnapshot.ReceiverName,
			ReceiverPhone: r.AddressSnapshot.ReceiverPhone,
			Province:      r.AddressSnapshot.Province,
			City:          r.AddressSnapshot.City,
			District:      r.AddressSnapshot.District,
			DetailAddress: r.AddressSnapshot.DetailAddress,
			PostalCode:    r.AddressSnapshot.PostalCode,
		},
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return &response.CreateOrderResp{
		OrderId:     resp.OrderId,
		OrderNo:     resp.OrderNo,
		PayAmount:   resp.PayAmount,
		TotalAmount: resp.TotalAmount,
		OrderStatus: resp.OrderStatus,
		PayExpireAt: asTime(resp.PayExpireAt),
		CreatedAt:   asTime(resp.CreatedAt),
	}, "", nil
}

// GetUserOrderList 用户端订单列表
func (c *TradeClient) GetUserOrderList(userId int64, page, pageSize int, orderStatus string) (*response.UserGetOrderListResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.order.ListUserOrders(ctx, &v1_tradev1.ListUserOrdersReq{
		UserId:      userId,
		Page:        int32(page),
		PageSize:    int32(pageSize),
		OrderStatus: orderStatus,
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}

	list := make([]response.UserGetOrderList, 0, len(resp.Items))
	for _, it := range resp.Items {
		list = append(list, response.UserGetOrderList{
			OrderId:     it.OrderId,
			OrderNo:     it.OrderNo,
			OrderStatus: it.OrderStatus,
			TotalAmount: it.TotalAmount,
			PayAmount:   it.PayAmount,
			DetailCount: it.DetailCount,
			FirstImage:  it.FirstImage,
			CreatedAt:   asTime(it.CreatedAt),
		})
	}
	return &response.UserGetOrderListResp{
		List:     list,
		Total:    resp.Total,
		Page:     int(resp.Page),
		PageSize: int(resp.PageSize),
	}, "", nil
}

// GetUserOrder 用户端订单详情
func (c *TradeClient) GetUserOrder(orderId, userId int64) (*response.UserGetOrderResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.order.GetUserOrder(ctx, &v1_tradev1.GetUserOrderReq{
		OrderId: orderId,
		UserId:  userId,
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	o := resp.Order
	out := &response.UserGetOrderResp{
		OrderId:         o.GetOrderId(),
		OrderNo:         o.GetOrderNo(),
		OrderStatus:     o.GetOrderStatus(),
		TotalAmount:     o.GetTotalAmount(),
		PayAmount:       o.GetPayAmount(),
		PayMethod:       o.GetPayMethod(),
		PayTime:         asTime(o.GetPayTime()),
		AddressSnapshot: datatypes.JSON(o.GetAddressSnapshot()),
		BuyerRemark:     o.GetBuyerRemark(),
		CreatedAt:       asTime(o.GetCreatedAt()),
		DetailList:      toModelOrderDetails(resp.Details),
		LogList:         toModelOrderLogs(resp.Logs),
	}
	return out, "", nil
}

// CancelOrder 取消订单。userId 传 0 表示系统取消(不校验归属)
func (c *TradeClient) CancelOrder(orderId, userId int64, operator string) (*response.OrderStatusResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.order.CancelOrder(ctx, &v1_tradev1.CancelOrderReq{
		OrderId:  orderId,
		UserId:   userId,
		Operator: operator,
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		// **注意**:补偿未完成(Compensated=false)不算业务失败 ——
		// 订单已取消是事实,补偿由对账收敛。调用方若需要这个信号,
		// 用 CancelOrderWithCompensation
		return nil, resp.ErrorMsg, nil
	}
	o := resp.Order
	return &response.OrderStatusResp{
		OrderId:     o.GetOrderId(),
		OrderNo:     o.GetOrderNo(),
		OrderStatus: o.GetOrderStatus(),
	}, "", nil
}

// CancelOrderWithCompensation 取消订单并返回补偿是否完成。
//
// 单独一个方法而不是给 CancelOrder 加返回值:绝大多数调用方不关心补偿,
// 加返回值会让它们都多一个用不上的变量。超时扫描需要它 ——
// 补偿失败意味着"库存没还、券没退",那需要告警而不只是记一笔取消。
func (c *TradeClient) CancelOrderWithCompensation(orderId, userId int64, operator string) (*response.OrderStatusResp, bool, string, error) {
	if c == nil {
		return nil, false, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.order.CancelOrder(ctx, &v1_tradev1.CancelOrderReq{
		OrderId:  orderId,
		UserId:   userId,
		Operator: operator,
	})
	if err != nil {
		return nil, false, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, false, resp.ErrorMsg, nil
	}
	o := resp.Order
	return &response.OrderStatusResp{
		OrderId:     o.GetOrderId(),
		OrderNo:     o.GetOrderNo(),
		OrderStatus: o.GetOrderStatus(),
	}, resp.Compensated, "", nil
}

// ConfirmOrder 确认收货
func (c *TradeClient) ConfirmOrder(orderId, userId int64, userName string) (*response.OrderStatusResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.order.ConfirmOrder(ctx, &v1_tradev1.ConfirmOrderReq{
		OrderId:  orderId,
		UserId:   userId,
		Username: userName,
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	o := resp.Order
	return &response.OrderStatusResp{
		OrderId:     o.GetOrderId(),
		OrderNo:     o.GetOrderNo(),
		OrderStatus: o.GetOrderStatus(),
	}, "", nil
}

// GetOrderList 管理端订单列表
func (c *TradeClient) GetOrderList(page, pageSize int, orderStatus, orderNo string,
	startTime, endTime *time.Time) (*response.GetOrderListResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	req := &v1_tradev1.ListOrdersReq{
		Page:        int32(page),
		PageSize:    int32(pageSize),
		OrderStatus: orderStatus,
		OrderNo:     orderNo,
	}
	if startTime != nil {
		req.StartTime = timestamppb.New(*startTime)
	}
	if endTime != nil {
		req.EndTime = timestamppb.New(*endTime)
	}

	resp, err := c.order.ListOrders(ctx, req)
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}

	list := make([]response.GetOrderList, 0, len(resp.Items))
	for _, it := range resp.Items {
		list = append(list, response.GetOrderList{
			OrderId:     it.OrderId,
			OrderNo:     it.OrderNo,
			OrderStatus: it.OrderStatus,
			TotalAmount: it.TotalAmount,
			PayAmount:   it.PayAmount,
			PayMethod:   it.PayMethod,
			CreatedAt:   asTime(it.CreatedAt),
			// UserName / ReceiverName / ReceiverPhone 由调用方回填:
			// 用户名在 user_db、收货人信息在 address_snapshot 的 JSON 里,
			// trade 侧不解析也不跨库取 —— 契约里留了字段,此处不假装能填
		})
	}
	return &response.GetOrderListResp{
		List:     list,
		Total:    resp.Total,
		Page:     int(resp.Page),
		PageSize: int(resp.PageSize),
	}, "", nil
}

// GetOrder 管理端订单详情
func (c *TradeClient) GetOrder(orderId int64) (*response.GetOrderResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.order.GetOrder(ctx, &v1_tradev1.GetOrderReq{OrderId: orderId})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	o := resp.Order
	return &response.GetOrderResp{
		OrderId:         o.GetOrderId(),
		OrderNo:         o.GetOrderNo(),
		OrderStatus:     o.GetOrderStatus(),
		TotalAmount:     o.GetTotalAmount(),
		PayAmount:       o.GetPayAmount(),
		PayMethod:       o.GetPayMethod(),
		PayTime:         asTime(o.GetPayTime()),
		AddressSnapshot: datatypes.JSON(o.GetAddressSnapshot()),
		BuyerRemark:     o.GetBuyerRemark(),
		DetailList:      toModelOrderDetails(resp.Details),
		LogList:         toModelOrderLogs(resp.Logs),
		CreatedAt:       asTime(o.GetCreatedAt()),
		Username:        o.GetUsername(),
		UserId:          o.GetUserId(),
	}, "", nil
}

// ShipOrder 发货
func (c *TradeClient) ShipOrder(orderId int64, userName, expressCompany, trackingNo string) (*response.OrderShipResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.order.ShipOrder(ctx, &v1_tradev1.ShipOrderReq{
		OrderId:        orderId,
		Username:       userName,
		ExpressCompany: expressCompany,
		TrackingNo:     trackingNo,
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	o := resp.Order
	return &response.OrderShipResp{
		OrderId:        o.GetOrderId(),
		OrderNo:        o.GetOrderNo(),
		OrderStatus:    o.GetOrderStatus(),
		ExpressCompany: resp.ExpressCompany,
		TrackingNO:     resp.TrackingNo,
	}, "", nil
}

// ============================================================
// 支付
// ============================================================

// CreatePayment 发起支付
func (c *TradeClient) CreatePayment(orderId, userId int64, payMethod string) (*response.CreatePaymentResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.payment.CreatePayment(ctx, &v1_tradev1.CreatePaymentReq{
		OrderId:   orderId,
		UserId:    userId,
		PayMethod: payMethod,
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return &response.CreatePaymentResp{
		PaymentId: resp.PaymentId,
		PayNo:     resp.PayNo,
		PayAmount: resp.PayAmount,
		PayStatus: resp.PayStatus,
		PayUrl:    resp.PayUrl,
		QrCode:    resp.QrCode,
	}, "", nil
}

// GetPayment 查支付流水
func (c *TradeClient) GetPayment(userId int64, payNo string) (*response.GetPaymentResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.payment.GetPayment(ctx, &v1_tradev1.GetPaymentReq{
		UserId: userId,
		PayNo:  payNo,
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	p := resp.Payment
	return &response.GetPaymentResp{
		PaymentId: p.GetPaymentId(),
		PayNo:     p.GetPayNo(),
		OrderId:   p.GetOrderId(),
		OrderNo:   p.GetOrderNo(),
		PayMethod: p.GetPayMethod(),
		PayAmount: p.GetPayAmount(),
		PayStatus: p.GetPayStatus(),
		PayTime:   asTime(p.GetPayTime()),
		TradeNo:   p.GetTradeNo(),
	}, "", nil
}

// HandleCallback 渠道回调。
//
// **错误语义与其它方法不同**:err != nil 表示"支付已成功但后续步骤失败,
// 请渠道重试回调",不是"回调无效"。见 payment.proto 的说明。
func (c *TradeClient) HandleCallback(payNo, tradeNo string) (bool, string, error) {
	if c == nil {
		return false, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.payment.HandlePaymentCallback(ctx, &v1_tradev1.HandlePaymentCallbackReq{
		PayNo:   payNo,
		TradeNo: tradeNo,
	})
	if err != nil {
		return false, "", err
	}
	return resp.Accepted, resp.ErrorMsg, nil
}

// GetPaymentList 管理端支付流水列表
func (c *TradeClient) GetPaymentList(page, pageSize int, payStatus, payMethod, orderNo string,
	startTime, endTime *time.Time) (*response.GetPaymentListResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	req := &v1_tradev1.ListPaymentsReq{
		Page:      int32(page),
		PageSize:  int32(pageSize),
		PayStatus: payStatus,
		PayMethod: payMethod,
		OrderNo:   orderNo,
	}
	if startTime != nil {
		req.StartTime = timestamppb.New(*startTime)
	}
	if endTime != nil {
		req.EndTime = timestamppb.New(*endTime)
	}

	resp, err := c.payment.ListPayments(ctx, req)
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}

	list := make([]response.GetPaymentList, 0, len(resp.Items))
	for _, it := range resp.Items {
		list = append(list, response.GetPaymentList{
			PaymentId: it.PaymentId,
			PayNo:     it.PayNo,
			OrderNo:   it.OrderNo,
			PayMethod: it.PayMethod,
			PayAmount: it.PayAmount,
			PayStatus: it.PayStatus,
			PayTime:   asTime(it.PayTime),
			// Username 由调用方回填(用户名在 user_db)
		})
	}
	return &response.GetPaymentListResp{
		List:     list,
		Total:    resp.Total,
		Page:     int(resp.Page),
		PageSize: int(resp.PageSize),
	}, "", nil
}

// ============================================================
// proto → 单体 response 的转换
// ============================================================
//
// 为什么保留单体原有的 response 结构体而不直接把 proto 交给 handler:
// HTTP 响应的 JSON 字段名是对前端的契约,已在用的字段不能因为换实现而改名。
// 单体 response.* 带 json tag,proto 带的是 protobuf 的 json_name,两者不总一致。

// toModelCartItem 单个购物车项
func toModelCartItem(p *v1_tradev1.CartItem) response.CartItemListResp {
	if p == nil {
		return response.CartItemListResp{}
	}
	return response.CartItemListResp{
		CartItemId: p.CartItemId,
		SkuId:      p.SkuId,
		SpuId:      p.SpuId,
		SpuName:    p.SpuName,
		MainImage:  p.MainImage,
		SkuName:    p.SkuName,
		// spec_values 在 proto 里是 JSON 文本、在单体 response 里是 JSONMap ——
		// 转换在此完成,不让 handler 碰这件事
		SpecValues:        jsonTextToMap(p.SpecValues),
		SkuImage:          p.SkuImage,
		Price:             p.Price,
		Stock:             p.Stock,
		Quantity:          p.Quantity,
		IsSelected:        p.IsSelected,
		Subtotal:          p.TotalPrice,
		IsAvailable:       p.Available,
		UnavailableReason: p.UnavailableReason,
	}
}

// toModelCartItems 批量转换
func toModelCartItems(items []*v1_tradev1.CartItem) []response.CartItemListResp {
	out := make([]response.CartItemListResp, 0, len(items))
	for _, it := range items {
		out = append(out, toModelCartItem(it))
	}
	return out
}

// toModelPayPreview 结算预览
func toModelPayPreview(p *v1_tradev1.CartPayPreview) *response.CartItemPayResp {
	if p == nil {
		return &response.CartItemPayResp{}
	}
	items := toModelCartItems(p.SelectedItems)

	// 不可购买的行单独分出来:前端要用它们渲染"以下商品无法购买"区块。
	// 这个切分在**调用方**做而不是服务端:服务端只描述"每行是否可购买",
	// 怎么展示是前端的决定
	available := make([]response.CartItemListResp, 0, len(items))
	unavailable := make([]response.CartItemListResp, 0)
	for _, it := range items {
		if it.IsAvailable {
			available = append(available, it)
		} else {
			unavailable = append(unavailable, it)
		}
	}
	return &response.CartItemPayResp{
		Items:            available,
		TotalCount:       int64(len(items)),
		TotalQuantity:    p.TotalQuantity,
		TotalAmount:      p.TotalAmount,
		HasUnavailable:   len(unavailable) > 0,
		UnavailableItems: unavailable,
	}
}

// toModelOrderDetails 订单明细
func toModelOrderDetails(items []*v1_tradev1.OrderDetail) []response.UserGetOrderDetail {
	out := make([]response.UserGetOrderDetail, 0, len(items))
	for _, d := range items {
		if d == nil {
			continue
		}
		out = append(out, response.UserGetOrderDetail{
			DetailId:   d.DetailId,
			SkuId:      d.SkuId,
			SpuName:    d.SpuName,
			SkuName:    d.SkuName,
			SpecValues: d.SpecValues,
			MainImage:  d.MainImage,
			Quantity:   d.Quantity,
			UnitPrice:  d.UnitPrice,
			TotalPrice: d.TotalPrice,
		})
	}
	return out
}

// toModelOrderLogs 订单日志
func toModelOrderLogs(items []*v1_tradev1.OrderLog) []response.UserGetOrderLog {
	out := make([]response.UserGetOrderLog, 0, len(items))
	for _, l := range items {
		if l == nil {
			continue
		}
		out = append(out, response.UserGetOrderLog{
			LogId:       l.LogId,
			OrderId:     l.OrderId,
			OrderStatus: l.OrderStatus,
			Action:      l.Action,
			Operator:    l.Operator,
			Detail:      l.Detail,
			CreatedAt:   asTime(l.CreatedAt),
		})
	}
	return out
}

// jsonTextToMap JSON 文本 → datatypes.JSONMap。
//
// 解析失败按空对象处理:规格是**展示用**信息,不该因为它阻断整个购物车列表。
func jsonTextToMap(text string) datatypes.JSONMap {
	if text == "" {
		return datatypes.JSONMap{}
	}
	out := datatypes.JSONMap{}
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		return datatypes.JSONMap{}
	}
	return out
}

// asTime 空 timestamp → 零值时间。
//
// 为什么不用 AsTime() 直接调:proto3 的可选时间用 nil 表达"没有",
// 而单体的 response 结构体用的是值类型 time.Time —— 直接调会 panic
// (nil 上没有方法), 或用 GetXxx() 会得到 1970,都不是"零值"该有的样子。
func asTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}
