package converter

import (
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/model"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ============================================================
// SKU / SPU 读接口 与 库存查询的 proto 转换
// ============================================================

// ToProtoGetSkuResp 组装 SKU 快照 + 可用性判定位。
//
// 为什么把三个判定(active / deleted / spu_published)算好再传,而不是让调用方自己拼:
// "能否购买"的规则属于商品域,散到调用方会出现各写一份、口径漂移;
// 单体时代购物车就是这么做的(cart_item_service.go 里手写三条件)。
func ToProtoGetSkuResp(sku *model.SysProductSku, spu *model.SysProductSpu) *v1_productv1.GetSkuResp {
	if sku == nil {
		return nil
	}
	resp := &v1_productv1.GetSkuResp{
		Sku:        ToProtoSku(sku),
		SkuActive:  sku.SkuStatus == model.SkuStatusActive,
		SkuDeleted: sku.IsDeleted,
	}
	if spu != nil {
		resp.SpuPublished = spu.SpuStatus == model.SpuStatusPublished
		resp.SpuDeleted = spu.IsDeleted
		resp.SpuName = spu.SpuName
		resp.CategoryId = spu.CategoryId
		resp.SpuMainImage = spu.MainImage
		resp.SpuStatus = spu.SpuStatus
	}
	return resp
}

// ToProtoSkuStock 由 SKU 实体 + 归属 SPU 名组装库存条目。
// dbStock 单独传入:调用方可选择用实时值(闸门键)或 DB 快照覆盖。
func ToProtoSkuStock(sku *model.SysProductSku, spuName string) *v1_productv1.SkuStock {
	if sku == nil {
		return nil
	}
	return &v1_productv1.SkuStock{
		SkuId:      sku.SkuId,
		SkuName:    sku.SkuName,
		SpuId:      sku.SpuId,
		SpuName:    spuName,
		SpecValues: jsonMapToText(sku.SpecValues),
		Stock:      sku.Stock,
		LockStock:  sku.LockStock,
		SoldCount:  sku.SoldCount,
		SkuStatus:  sku.SkuStatus,
	}
}

// ToProtoSkuStockFromWarnItem 由预警查询结果组装库存条目(已带 spu_name)
func ToProtoSkuStockFromWarnItem(item *model.StockWarnItem) *v1_productv1.SkuStock {
	if item == nil {
		return nil
	}
	return &v1_productv1.SkuStock{
		SkuId:     item.SkuId,
		SkuName:   item.SkuName,
		SpuName:   item.SpuName,
		Stock:     item.Stock,
		LockStock: item.LockStock,
		SoldCount: item.SoldCount,
		SkuStatus: item.SkuStatus,
	}
}

// ToProtoStockLog 由流水查询结果组装 proto(流水行 + 联表取到的名称)
func ToProtoStockLog(item *model.StockLogItem) *v1_productv1.StockLog {
	if item == nil {
		return nil
	}
	return &v1_productv1.StockLog{
		LogId:       item.LogId,
		SkuId:       item.SkuId,
		SkuName:     item.SkuName,
		SpuId:       item.SpuId,
		SpuName:     item.SpuName,
		ChangeType:  item.ChangeType,
		ChangeQty:   item.ChangeQty,
		BeforeStock: item.BeforeStock,
		AfterStock:  item.AfterStock,
		BeforeLock:  item.BeforeLock,
		AfterLock:   item.AfterLock,
		OrderId:     item.OrderId,
		Remark:      item.Remark,
		CreateBy:    item.CreateBy,
		CreatedAt:   timestamppb.New(item.CreatedAt),
	}
}
