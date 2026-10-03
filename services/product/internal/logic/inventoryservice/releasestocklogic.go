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

type ReleaseStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReleaseStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReleaseStockLogic {
	return &ReleaseStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ReleaseStockLogic) ReleaseStock(in *v1_productv1.ReleaseStockReq) (*v1_productv1.ReleaseStockResp, error) {
	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		productTx := l.svcCtx.ProductRepo.WithTx(tx)
		logTx := l.svcCtx.InventoryLogRepo.WithTx(tx)

		sku, err := productTx.GetSkuForUpdate(in.SkuId)
		if err != nil {
			return err
		}

		rowsAffected, err := logTx.CreateInventoryLog(&model.SysProductStockLog{
			SkuId:       in.SkuId,
			ChangeType:  model.StockOrderRelease,
			ChangeQty:   in.Qty,
			BeforeStock: sku.Stock,
			AfterStock:  sku.Stock + in.Qty,
			BeforeLock:  sku.LockStock,
			AfterLock:   sku.LockStock - in.Qty,
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
		// 释放时锁定数量不足。不复用 ErrStockNegative(7002) —— 那条的文案是
		// "调整后库存不能为负数",属管理端调整场景,出现在取消订单链路里会误导排查。
		if sku.LockStock < in.Qty {
			return model.ErrLockStockNotEnough
		}

		return productTx.UpdateSkuStockForRelease(in.SkuId, in.Qty)
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
		errors.Is(err, model.ErrLockStockNotEnough) {
		return &v1_productv1.ReleaseStockResp{ErrorMsg: err.Error()}, nil
	}
	return nil, err
}
