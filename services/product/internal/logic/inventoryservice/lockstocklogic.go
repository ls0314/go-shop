package inventoryservicelogic

import (
	"context"
	"demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/model"
	"demo-shop/services/product/internal/svc"
	"demo-shop/services/product/internal/utils"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type LockStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLockStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LockStockLogic {
	return &LockStockLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *LockStockLogic) LockStock(in *v1_productv1.LockStockReq) (*v1_productv1.LockStockResp, error) {
	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		productTx := l.svcCtx.ProductRepo.WithTx(tx)
		logTx := l.svcCtx.InventoryLogRepo.WithTx(tx)

		sku, err := productTx.GetSkuForUpdate(in.SkuId)
		if err != nil {
			return err
		}

		rowsAffected, err := logTx.CreateInventoryLog(&model.SysProductStockLog{
			SkuId:       in.SkuId,
			ChangeType:  model.StockOrderLock,
			ChangeQty:   in.Qty,
			BeforeStock: sku.Stock,
			AfterStock:  sku.Stock - in.Qty,
			BeforeLock:  sku.LockStock,
			AfterLock:   sku.LockStock + in.Qty,
			OrderId:     in.OrderId,
		})
		if err != nil {
			return err
		}
		// 流水已存在 = 本次操作已生效过,幂等成功
		if rowsAffected == 0 {
			return nil
		}

		if sku.SkuStatus != model.SkuStatusActive || sku.IsDeleted {
			return model.ErrSkuDisabled
		}
		if sku.Stock < in.Qty {
			return model.ErrStockNotEnough
		}

		return productTx.UpdateSkuStockForLock(in.SkuId, in.Qty)
	})

	// 库存变更后失效缓存(短 TTL 兜底,失效失败最多 30 秒旧值)
	if l.svcCtx.Redis != nil {
		if _, err := l.svcCtx.Redis.Del(utils.StockGateKey(in.SkuId)); err != nil {
			l.Errorf("失效库存缓存失败: skuId=%d err=%v", in.SkuId, err)
		}
	}

	// 业务失败 → gRPC error 为 nil + error_msg;基础设施失败 → gRPC error
	if errors.Is(err, model.ErrSkuDisabled) ||
		errors.Is(err, model.ErrSkuNotExist) ||
		errors.Is(err, model.ErrStockNotEnough) {
		return &v1_productv1.LockStockResp{ErrorMsg: err.Error()}, nil
	}
	return nil, err
}
