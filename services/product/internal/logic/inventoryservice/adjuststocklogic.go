package inventoryservicelogic

import (
	"context"
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/model"
	"demo-shop/services/product/internal/svc"
	"demo-shop/services/product/internal/utils"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type AdjustStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdjustStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdjustStockLogic {
	return &AdjustStockLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// AdjustStock 手动调整库存(管理端)。
//
// 从单体 src/service/inventory_service.go 的 AdjustStock 平移。
// 为什么必须是本服务的 RPC:实现是"SELECT ... FOR UPDATE 行锁 + 写库存流水 +
// 改库存"三步同一事务,拆库后调用方无法在自己的事务里操作 sys_product_sku。
//
// 幂等:沿用三元组唯一索引 (order_id, sku_id, change_type) 作闸 ——
// 手动调整的 order_id 为 NULL,PG 的唯一索引对 NULL 不去重,故这里的幂等
// **不来自索引**,而是来自"先读再写"的行锁串行化:并发两次调整会依次生效。
// 这与库存四操作(order_id 非空、靠索引去重)的语义不同,不要混淆。
func (l *AdjustStockLogic) AdjustStock(in *v1_productv1.AdjustStockReq) (*v1_productv1.AdjustStockResp, error) {
	// 调整原因不能为空:这是审计要求,不是可选的备注
	if in.Remark == "" {
		return &v1_productv1.AdjustStockResp{ErrorMsg: model.ErrRemarkEmpty.Error()}, nil
	}

	var beforeStock, afterStock int64
	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		productTx := l.svcCtx.ProductRepo.WithTx(tx)
		logTx := l.svcCtx.InventoryLogRepo.WithTx(tx)

		// 并发安全查询:行锁串行化同一 SKU 的并发调整
		sku, err := productTx.GetSkuForUpdate(in.SkuId)
		if err != nil {
			return err
		}

		// 先前插流水(与原实现同序):入库失败即整体回滚,不留"改了库存没流水"
		rowsAffected, err := logTx.CreateInventoryLog(&model.SysProductStockLog{
			SkuId:       sku.SkuId,
			ChangeType:  model.StockManualAdjust,
			ChangeQty:   in.ChangeQty,
			BeforeStock: sku.Stock,
			AfterStock:  sku.Stock + in.ChangeQty,
			BeforeLock:  sku.LockStock,
			AfterLock:   sku.LockStock,
			Remark:      in.Remark,
			CreateBy:    in.UserId,
		})
		if err != nil {
			return err
		}
		// 与原实现一致:流水未插入(理论上仅当 order_id 非空时才会发生)视为幂等命中,
		// 不计库存;此处 order_id 恒为 NULL,正常路径 rowsAffected 恒为 1。
		if rowsAffected == 0 {
			beforeStock, afterStock = sku.Stock, sku.Stock
			return nil
		}

		// 调整后不能为负
		if sku.Stock+in.ChangeQty < 0 {
			return model.ErrStockNegative
		}

		beforeStock = sku.Stock
		afterStock = sku.Stock + in.ChangeQty

		return productTx.UpdateStock(in.SkuId, afterStock)
	})

	// 库存变更后失效闸门键(短 TTL 兜底,失效失败最多 30 秒旧值)。
	// 放在事务之外:回滚时库存没变,但清一次缓存无害;成功时这才是必须的动作。
	if l.svcCtx.Redis != nil {
		if _, err := l.svcCtx.Redis.Del(utils.StockGateKey(in.SkuId)); err != nil {
			l.Errorf("失效库存缓存失败: skuId=%d err=%v", in.SkuId, err)
		}
	}

	if err != nil {
		// 业务失败 → error_msg;基础设施故障 → gRPC error
		if errors.Is(err, model.ErrRemarkEmpty) ||
			errors.Is(err, model.ErrSkuNotExist) ||
			errors.Is(err, model.ErrStockNegative) {
			return &v1_productv1.AdjustStockResp{ErrorMsg: err.Error()}, nil
		}
		return nil, err
	}

	return &v1_productv1.AdjustStockResp{
		BeforeStock: beforeStock,
		AfterStock:  afterStock,
	}, nil
}
