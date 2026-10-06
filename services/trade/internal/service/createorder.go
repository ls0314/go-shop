package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"demo-shop/services/trade/internal/dto/req"
	"demo-shop/services/trade/internal/dto/resp"
	"demo-shop/services/trade/internal/infra/mq"
	"demo-shop/services/trade/internal/model"
	"demo-shop/services/trade/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ============================================================
// 下单:Saga 编排
// ============================================================
//
// 步骤:
//
//	step_use_coupon    marketing RPC  UseCoupon     补偿:ReturnCoupon
//	step_lock_stock    product RPC    LockStock     补偿:ReleaseStock
//	step_persist_order 本地事务        订单+明细+删购物车+日志+outbox   无补偿

// Saga 步骤名。写进日志,事后复盘"走到哪一步失败"靠它
const (
	stepUseCoupon    = "use_coupon"
	stepLockStock    = "lock_stock"
	stepPersistOrder = "persist_order"
)

// CreateOrder 下单。
func (o *OrderService) CreateOrder(ctx context.Context, r *req.CreateOrderReq) (*resp.CreateOrderResp, error) {
	// ---- 入参形状校验 ----
	if err := r.Validate(); err != nil {
		return nil, err
	}

	scopedKey := scopedIdempotentKey(r.UserId, r.IdempotentKey)

	// ---- 幂等预检:同一键重复提交直接返回已有订单 ----
	if existing, err := o.orderRepo.GetOrderByIdempotentKey(scopedKey); err == nil {
		return toCreateOrderResp(existing), nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// ---- 准备购物车行----
	cartLines, err := o.prepareCartLines(r.UserId)
	if err != nil {
		return nil, err
	}

	// ---- 生成订单号 ----
	orderNo := utils.FormatOrderNo(o.idGen.NextId())

	totalAmount := calcTotalAmount(cartLines)
	payAmount := totalAmount
	// 截止时间只算一次,写进 expire_at;MQ 延迟 TTL 与超时扫描都读那一列
	expireAt := time.Now().Add(model.OrderPayTTL)

	// order 在闭包外声明、闭包内填充:建单事务会通过 GORM 回填自增主键,
	// 事务提交后这里就能读到 order_id。
	order := &model.UserOrder{
		OrderNo: orderNo,
		UserId:  r.UserId,
		// 用户名快照:下单那一刻的值,不随用户改名而变
		Username:        r.UserName,
		OrderStatus:     model.OrderPendingPay,
		TotalAmount:     totalAmount,
		PayAmount:       payAmount,
		AddressSnapshot: marshalAddress(r.AddressSnapshot),
		BuyerRemark:     r.BuyerRemark,
		IdempotentKey:   scopedKey,
		DetailCount:     int64(len(cartLines)),
		FirstImage:      firstImage(cartLines),
		ExpireAt:        expireAt,
	}

	// 逐 SKU 锁库存的进度
	var lockedLines []cartLine

	// ---- Saga ----
	s := newSaga(
		sagaStep{
			name: stepUseCoupon,
			forward: func(ctx context.Context) error {
				if r.UserCouponId == 0 {
					return nil
				}
				paid, uErr := o.couponRPC.UseCoupon(r.UserCouponId, r.IdempotentKey, orderNo, r.UserId, totalAmount)
				if uErr != nil {
					return uErr
				}
				// 实付金额由 marketing 侧算(满减/直减的差别不外泄),
				payAmount = paid
				order.PayAmount = paid
				return nil
			},
			compensate: func(ctx context.Context) error {
				if r.UserCouponId == 0 {
					return nil
				}
				// 幂等且宽容:该键没核销过、或券已归还,marketing 侧都算成功
				return o.couponRPC.ReturnCoupon(r.IdempotentKey, r.UserId)
			},
		},
		sagaStep{
			name: stepLockStock,
			forward: func(ctx context.Context) error {
				for _, line := range cartLines {
					if lErr := o.inventoryRPC.LockStock(line.SkuId, line.Quantity, r.IdempotentKey); lErr != nil {
						return lErr
					}
					lockedLines = append(lockedLines, line)
				}
				return nil
			},
			compensate: func(ctx context.Context) error {
				return o.releaseLocked(lockedLines, r.IdempotentKey)
			},
		},
		sagaStep{
			name: stepPersistOrder,
			forward: func(ctx context.Context) error {
				return o.persistOrder(&persistOrderArgs{
					Order:     order,
					CartLines: cartLines,
					Operator:  r.UserName,
				})
			},
			// 本地事务失败即整体失败:它自己没有副作用需要撤,故不设补偿
			compensate: nil,
		},
	)

	failedStep, err := s.run(ctx)
	if err != nil {
		// 把"走到哪一步失败"记下来
		logx.Errorf("下单 Saga 失败: orderNo=%s idempotentKey=%s failedStep=%s err=%v",
			orderNo, r.IdempotentKey, failedStep, err)

		// 并发重复提交:唯一索引挡住了第二次 INSERT。
		if errors.Is(err, errDuplicateIdempotentKey) {
			if existing, qErr := o.orderRepo.GetOrderByIdempotentKey(scopedKey); qErr == nil {
				return toCreateOrderResp(existing), nil
			}
		}
		return nil, err
	}

	return toCreateOrderResp(order), nil
}

// ============================================================
// 内部辅助
// ============================================================

// cartLine 购物车行在下单流程内的形态:已带商品快照与算好的单价。
type cartLine struct {
	CartItemId int64
	SkuId      int64
	Quantity   int64
	// 以下来自 product-service 的 BatchGetSkus
	SpuName    string
	SkuName    string
	SpecValues string
	MainImage  string
	UnitPrice  float64
}

// prepareCartLines 取"已选中"的购物车行并回填商品快照。
func (o *OrderService) prepareCartLines(userId int64) ([]cartLine, error) {
	items, err := o.cartRepo.GetSelectedCartItemList(userId)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, model.ErrCartNoSettlementItems
	}

	skuIds := make([]int64, 0, len(items))
	for _, it := range items {
		skuIds = append(skuIds, it.SkuId)
	}

	// 一次批量取回,避免 N+1
	snapshots, err := o.productRPC.BatchGetSkus(skuIds)
	if err != nil {
		return nil, err
	}

	lines := make([]cartLine, 0, len(items))
	for _, it := range items {
		snap, ok := snapshots[it.SkuId]
		if !ok || !snap.Available {
			// SKU 已下架/删除:结算时直接拒绝,而不是"静默少买一件"
			return nil, model.ErrSkuDisabled
		}
		lines = append(lines, cartLine{
			CartItemId: it.CartItemId,
			SkuId:      it.SkuId,
			Quantity:   it.Quantity,
			SpuName:    snap.SpuName,
			SkuName:    snap.SkuName,
			SpecValues: snap.SpecValues,
			MainImage:  snap.MainImage,
			// 单价取**下单那一刻**的商品价,并固化进明细快照 ——
			// 订单是历史凭证,商品改价后不能跟着变
			UnitPrice: snap.Price,
		})
	}
	return lines, nil
}

// releaseLocked 逆序释放已锁库存(step_lock_stock 的补偿)。
func (o *OrderService) releaseLocked(locked []cartLine, idempotencyKey string) error {
	var firstErr error
	for i := len(locked) - 1; i >= 0; i-- {
		line := locked[i]
		// ReleaseStock 在 product-service 侧按
		// (idempotency_key, sku_id, change_type) 幂等,重复调用只释放一次
		if err := o.inventoryRPC.ReleaseStock(line.SkuId, line.Quantity, idempotencyKey); err != nil {
			logx.Errorf("Saga 补偿失败(等待对账收敛): 释放库存 skuId=%d qty=%d err=%v",
				line.SkuId, line.Quantity, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// persistOrderArgs persistOrder 的入参(打包传递,避免长参数列表)
type persistOrderArgs struct {
	Order     *model.UserOrder
	CartLines []cartLine
	Operator  string
}

// persistOrder 建单的本地事务(订单 + 明细 + 删购物车 + 日志 + outbox)。
func (o *OrderService) persistOrder(args *persistOrderArgs) error {
	order := args.Order

	return o.orderRepo.DB.Transaction(func(tx *gorm.DB) error {
		orderTx := o.orderRepo.WithTx(tx)
		cartTx := o.cartRepo.WithTx(tx)
		outboxTx := o.outboxRepo.WithTx(tx)

		if err := orderTx.CreateOrder(order); err != nil {
			return translateDuplicateKey(err)
		}

		for _, line := range args.CartLines {
			detail := &model.UserOrderDetail{
				OrderId:    order.OrderId,
				SkuId:      line.SkuId,
				SpuName:    line.SpuName,
				SkuName:    line.SkuName,
				SpecValues: parseSpecValues(line.SpecValues),
				MainImage:  line.MainImage,
				Quantity:   line.Quantity,
				UnitPrice:  line.UnitPrice,
				TotalPrice: line.UnitPrice * float64(line.Quantity),
			}
			if err := orderTx.CreateOrderDetail(detail); err != nil {
				return err
			}
			// 删购物车项。失败不回滚库存 —— 整个事务回滚后 Saga 会统一补偿
			if err := cartTx.DeleteCartItem(line.CartItemId); err != nil {
				return err
			}
		}

		if err := orderTx.CreateOrderLog(&model.UserOrderLog{
			OrderId:     order.OrderId,
			OrderStatus: model.OrderPendingPay,
			Action:      model.OrderActionCreate,
			Operator:    args.Operator,
		}); err != nil {
			return err
		}

		// 延迟取消消息与订单**同事务**写入:事务提交则消息必达。
		// 这正是 outbox 模式要解决的问题 —— 否则"订单建了消息没发"
		// 会让这张单永远不会超时取消。
		//
		// 交换机与路由键取自 mq 包的常量,**不在这里手写字面量**:
		// 投递器、队列声明、消费者绑定三处必须一致,而漂移了不会编译报错,
		// 只表现为"消息发出去了但没人收到"。
		return outboxTx.CreateOutbox(&model.OutBoxMessage{
			MessageId:     fmt.Sprintf("delay:%d", order.OrderId),
			AggregateType: model.OutboxAggregateOrder,
			AggregateId:   strconv.FormatInt(order.OrderId, 10),
			EventType:     model.OutboxEventOrderDelayCancel,
			Exchange:      mq.ExchangeOrderTradeDelay,
			RoutingKey:    mq.RoutingKeyOrderTradeDelay,
			// 负载是 JSON(格式契约见 mq/payload.go)。不用裸 orderId:
			// 消费者还要拿幂等键去调库存与券服务
			PayLoad: mq.BuildOrderDelayPayload(order),
			Status:  model.OutboxPending,
			// **投递时间 = 支付截止时间**。投递器据此算 per-message TTL,
			// 于是消息恰好在订单该超时的时刻转入死信队列。
			// 若这里写 now(),消息会被立刻投进延迟队列并等满 15 分钟 ——
			// 那些 3 分钟后才下单的订单就会晚取消 12 分钟
			NextRetryAt: order.ExpireAt,
		})
	})
}

// errDuplicateIdempotentKey 并发重复提交的哨兵错误。
var errDuplicateIdempotentKey = errors.New("重复提交")

// translateDuplicateKey 把 PostgreSQL 的唯一键冲突翻译成哨兵错误。
func translateDuplicateKey(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "23505") ||
		strings.Contains(err.Error(), "duplicate key") {
		return errDuplicateIdempotentKey
	}
	return err
}

// scopedIdempotentKey 给幂等键加用户维度,作为 user_order_master.idempotent_key。
func scopedIdempotentKey(userId int64, key string) string {
	return strconv.FormatInt(userId, 10) + ":" + key
}

// calcTotalAmount 合计金额(未抵扣券)
func calcTotalAmount(lines []cartLine) float64 {
	var total float64
	for _, l := range lines {
		total += l.UnitPrice * float64(l.Quantity)
	}
	return total
}

// firstImage 首图(冗余列,列表页展示用)
func firstImage(lines []cartLine) string {
	if len(lines) == 0 {
		return ""
	}
	return lines[0].MainImage
}

// marshalAddress 地址快照转 jsonb
func marshalAddress(snap model.AddressSnap) datatypes.JSON {
	b, err := json.Marshal(snap)
	if err != nil {
		// 快照是自己定义的结构体,序列化失败说明代码有问题而非输入问题
		logx.Errorf("地址快照序列化失败: %v", err)
		return datatypes.JSON([]byte(`{}`))
	}
	return datatypes.JSON(b)
}

// parseSpecValues 商品规格 JSON 文本转 jsonb。
func parseSpecValues(raw string) datatypes.JSONMap {
	out := datatypes.JSONMap{}
	if raw == "" {
		return out
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		logx.Errorf("规格快照解析失败(按空处理): %v", err)
		return datatypes.JSONMap{}
	}
	return out
}

// toCreateOrderResp 实体 → 下单响应
func toCreateOrderResp(order *model.UserOrder) *resp.CreateOrderResp {
	if order == nil {
		return nil
	}
	return &resp.CreateOrderResp{
		OrderId:     order.OrderId,
		OrderNo:     order.OrderNo,
		TotalAmount: order.TotalAmount,
		PayAmount:   order.PayAmount,
		OrderStatus: order.OrderStatus,
		PayExpireAt: order.ExpireAt,
		CreatedAt:   order.CreatedAt,
	}
}
