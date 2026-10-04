package service

import (
	"demo-shop/services/trade/internal/repository"
	"demo-shop/services/trade/internal/utils"
)

// OrderService 订单服务层。
type OrderService struct {
	orderRepo  *repository.OrderRepo
	cartRepo   *repository.CartItemRepo
	outboxRepo *repository.OutboxRepo

	inventoryRPC InventoryRPC
	couponRPC    CouponRPC
	// productRPC 下单要读 SKU 快照(取价并固化进明细),与 inventoryRPC
	// 是同一个服务的两个 gRPC service(一条连接两个域)
	productRPC ProductReadRPC
	// idGen 订单号生成器
	idGen *utils.Snowflake
}

// NewOrderService 创建订单服务层实例
// 接收值：repo - 本服务数据层；*RPC - 下游客户端；idGen - 订单号生成器
// 返回值：*OrderService - 订单服务层实例指针
func NewOrderService(orderRepo *repository.OrderRepo, cartRepo *repository.CartItemRepo,
	outboxRepo *repository.OutboxRepo, inventoryRPC InventoryRPC, couponRPC CouponRPC,
	productRPC ProductReadRPC, idGen *utils.Snowflake) *OrderService {
	return &OrderService{
		orderRepo:    orderRepo,
		cartRepo:     cartRepo,
		outboxRepo:   outboxRepo,
		inventoryRPC: inventoryRPC,
		couponRPC:    couponRPC,
		productRPC:   productRPC,
		idGen:        idGen,
	}
}

// CartService 购物车服务层。
type CartService struct {
	cartRepo   *repository.CartItemRepo
	productRPC ProductReadRPC
}

// NewCartService 创建购物车服务层实例
func NewCartService(cartRepo *repository.CartItemRepo, productRPC ProductReadRPC) *CartService {
	return &CartService{
		cartRepo:   cartRepo,
		productRPC: productRPC,
	}
}

// PaymentService 支付服务层。
//
// 与订单域同库同进程,故它能直接调 OrderService 推进订单状态 ——
// 不需要为"支付成功要改订单状态"再造一个 RPC。
type PaymentService struct {
	orderRepo   *repository.OrderRepo
	paymentRepo *repository.PaymentRepo
	// inventoryRPC 支付成功后把锁定库存转为实际扣减。
	inventoryRPC InventoryRPC
	orderSvc     *OrderService
	// idGen 支付单号生成器(与订单号同一个雪花实例)
	idGen *utils.Snowflake
}

// NewPaymentService 创建支付服务层实例
func NewPaymentService(orderRepo *repository.OrderRepo, paymentRepo *repository.PaymentRepo,
	inventoryRPC InventoryRPC, orderSvc *OrderService, idGen *utils.Snowflake) *PaymentService {
	return &PaymentService{
		orderRepo:    orderRepo,
		paymentRepo:  paymentRepo,
		inventoryRPC: inventoryRPC,
		orderSvc:     orderSvc,
		idGen:        idGen,
	}
}
