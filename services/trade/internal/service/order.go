package service

import (
	"context"
	"errors"

	"demo-shop/services/trade/internal/dto/req"
	"demo-shop/services/trade/internal/dto/resp"
	"demo-shop/services/trade/internal/model"
	"demo-shop/services/trade/internal/repository"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// ============================================================
// 订单状态机与查询
// ============================================================

// CancelOrderBySystem 系统取消订单(超时扫描用)。
func (o *OrderService) CancelOrderBySystem(ctx context.Context, orderId int64, operator string) (*resp.CancelOrderResp, error) {
	return o.CancelOrder(ctx, &req.CancelOrderReq{
		OrderId:  orderId,
		UserId:   0,
		Operator: operator,
	})
}

// CancelOrder 取消订单(用户主动或系统触发)。
func (o *OrderService) CancelOrder(ctx context.Context, r *req.CancelOrderReq) (*resp.CancelOrderResp, error) {
	order, err := o.orderRepo.GetOrderById(r.OrderId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrOrderNotExist
		}
		return nil, err
	}

	// 归属校验:系统取消(UserId = 0)跳过
	if r.UserId != 0 && order.UserId != r.UserId {
		return nil, model.ErrOrderNoPermission
	}
	// 仅待支付可取消。已取消/已支付都要给出明确的"不必重复处理"错误 ——
	// 超时扫描据此把这一类算作 skipped 而不是 failed
	if order.OrderStatus != model.OrderPendingPay {
		if order.OrderStatus == model.OrderCancelled {
			return nil, model.ErrOrderAlreadyCancelled
		}
		return nil, model.ErrOrderCannotCancel
	}

	// ---- 本地事务:条件改状态 + 写日志 ----
	// details 在事务内读出,供提交后释放库存用
	var details []*model.UserOrderDetail
	rows, err := o.cancelLocal(order, r.Operator, &details)
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		// 并发:另一个取消者先提交了(条件更新 status = pending_pay 不再匹配)。
		// 这不是故障 —— 目标已达成
		return nil, model.ErrOrderAlreadyCancelled
	}

	// ---- 事务已提交:补偿步骤,失败不回滚取消结果 ----
	compensated := o.releaseOnCancel(ctx, order, details)
	if !o.returnCouponOnCancel(ctx, order) {
		compensated = false
	}

	order.OrderStatus = model.OrderCancelled
	return &resp.CancelOrderResp{Order: order, Compensated: compensated}, nil
}

// cancelLocal 取消的本地事务部分:条件改状态 + 写订单日志。
func (o *OrderService) cancelLocal(order *model.UserOrder, operator string, details *[]*model.UserOrderDetail) (int64, error) {
	var rows int64
	err := o.orderRepo.DB.Transaction(func(tx *gorm.DB) error {
		orderTx := o.orderRepo.WithTx(tx)

		affected, err := orderTx.CancelOrder(order.OrderId)
		if err != nil {
			return err
		}
		rows = affected
		if affected == 0 {
			return nil // 让调用方判"已被别人处理",不必写日志
		}

		list, err := orderTx.GetOrderDetailList(order.OrderId)
		if err != nil {
			return err
		}
		*details = list

		return orderTx.CreateOrderLog(&model.UserOrderLog{
			OrderId:     order.OrderId,
			OrderStatus: model.OrderCancelled,
			Action:      model.OrderActionCancel,
			Operator:    operator,
		})
	})
	return rows, err
}

// releaseOnCancel 提交后释放库存。返回是否全部成功。
func (o *OrderService) releaseOnCancel(ctx context.Context, order *model.UserOrder, details []*model.UserOrderDetail) bool {
	allOK := true
	for _, d := range details {
		// 幂等键用订单上的 idempotent_key:它与下单锁库存时用的是同一个键
		// (product 侧 lock 与 release 的 change_type 不同,故不会互相挡住)
		if err := o.inventoryRPC.ReleaseStock(d.SkuId, d.Quantity, order.IdempotentKey); err != nil {
			logx.Errorf("取消订单释放库存失败(等待对账收敛): orderId=%d skuId=%d qty=%d err=%v",
				order.OrderId, d.SkuId, d.Quantity, err)
			allOK = false
		}
	}
	return allOK
}

// returnCouponOnCancel 提交后退还券。返回是否成功。
func (o *OrderService) returnCouponOnCancel(ctx context.Context, order *model.UserOrder) bool {
	if order.PayAmount == order.TotalAmount {
		return true // 没抵扣过,无券可退
	}
	if err := o.couponRPC.ReturnCoupon(order.IdempotentKey, order.UserId); err != nil {
		logx.Errorf("取消订单退还券失败(等待对账收敛): orderId=%d err=%v", order.OrderId, err)
		return false
	}
	return true
}

// ============================================================
// 查询
// ============================================================

// ListUserOrders 用户端订单分页列表
func (o *OrderService) ListUserOrders(userId int64, r *req.ListUserOrdersReq) (*resp.OrderPage, error) {
	page, pageSize := normalizePage(r.Page, r.PageSize)

	items, total, err := o.orderRepo.GetOrderList(userId, repository.OrderListQuery{
		Page:        page,
		PageSize:    pageSize,
		OrderStatus: r.OrderStatus,
	})
	if err != nil {
		return nil, err
	}
	return &resp.OrderPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

// GetUserOrder 用户端订单详情(订单 + 明细 + 日志)
func (o *OrderService) GetUserOrder(orderId, userId int64) (*resp.OrderDetailView, error) {
	order, err := o.orderRepo.GetOrderById(orderId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrOrderNotExist
		}
		return nil, err
	}
	// 归属校验放表的所有权方这一侧:调用方可能传错 userId,不能替它兜底
	if order.UserId != userId {
		return nil, model.ErrOrderNoPermission
	}
	return o.loadOrderDetail(order)
}

// ConfirmOrder 确认收货(仅已发货可确认)
func (o *OrderService) ConfirmOrder(orderId, userId int64, userName string) (*model.UserOrder, error) {
	order, err := o.orderRepo.GetOrderById(orderId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrOrderNotExist
		}
		return nil, err
	}
	if order.UserId != userId {
		return nil, model.ErrOrderNoPermission
	}
	if err := o.transit(order, model.OrderShipped, model.OrderCompleted,
		model.OrderActionConfirm, userName, model.ErrOrderCannotConfirm); err != nil {
		return nil, err
	}
	order.OrderStatus = model.OrderCompleted
	return order, nil
}

// ============================================================
// 管理端
// ============================================================

// ListOrders 管理端订单分页列表(userId 传 0 表示不限用户)
func (o *OrderService) ListOrders(r *req.ListOrdersReq) (*resp.OrderPage, error) {
	page, pageSize := normalizePage(r.Page, r.PageSize)

	items, total, err := o.orderRepo.GetOrderList(0, repository.OrderListQuery{
		Page:        page,
		PageSize:    pageSize,
		OrderStatus: r.OrderStatus,
		OrderNo:     r.OrderNo,
		StartTime:   r.StartTime,
		EndTime:     r.EndTime,
	})
	if err != nil {
		return nil, err
	}
	return &resp.OrderPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

// GetOrder 管理端订单详情。**不校验归属** —— 管理端可以看任何订单
func (o *OrderService) GetOrder(orderId int64) (*resp.OrderDetailView, error) {
	order, err := o.orderRepo.GetOrderById(orderId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrOrderNotExist
		}
		return nil, err
	}
	return o.loadOrderDetail(order)
}

// ShipOrder 发货(仅已支付可发货)
func (o *OrderService) ShipOrder(r *req.ShipOrderReq) (*resp.ShipOrderResp, error) {
	order, err := o.orderRepo.GetOrderById(r.OrderId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrOrderNotExist
		}
		return nil, err
	}
	if err := o.transit(order, model.OrderPaid, model.OrderShipped,
		model.OrderActionShip, r.UserName, model.ErrOrderCannotShip); err != nil {
		return nil, err
	}
	order.OrderStatus = model.OrderShipped
	return &resp.ShipOrderResp{
		Order:          order,
		ExpressCompany: r.ExpressCompany,
		TrackingNo:     r.TrackingNo,
	}, nil
}

// ============================================================
// 共用私有辅助
// ============================================================

// transit 状态推进(带起点条件的条件更新 + 写日志)。
func (o *OrderService) transit(order *model.UserOrder, from, to, action, operator string, statusErr error) error {
	return o.orderRepo.DB.Transaction(func(tx *gorm.DB) error {
		orderTx := o.orderRepo.WithTx(tx)

		rows, err := orderTx.UpdateOrderStatus(order.OrderId, from, to)
		if err != nil {
			return err
		}
		if rows == 0 {
			// 起点不匹配:状态已被别的路径改过(或并发)。返回领域错误,
			// 由调用方决定 HTTP 码 —— 不是故障
			return statusErr
		}
		return orderTx.CreateOrderLog(&model.UserOrderLog{
			OrderId:     order.OrderId,
			OrderStatus: to,
			Action:      action,
			Operator:    operator,
		})
	})
}

// loadOrderDetail 组装订单详情(主表 + 明细 + 日志)
func (o *OrderService) loadOrderDetail(order *model.UserOrder) (*resp.OrderDetailView, error) {
	details, err := o.orderRepo.GetOrderDetailList(order.OrderId)
	if err != nil {
		return nil, err
	}
	logs, err := o.orderRepo.GetOrderLogList(order.OrderId)
	if err != nil {
		return nil, err
	}
	return &resp.OrderDetailView{Order: order, Details: details, Logs: logs}, nil
}

// normalizePage 分页归一化:page<=0 → 1;pageSize<=0 → 10;>100 封顶 100。
func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
