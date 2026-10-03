package productservicelogic

import (
	"context"
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/converter"
	"demo-shop/services/product/internal/model"
	"demo-shop/services/product/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProductListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductListLogic {
	return &GetProductListLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// GetProductList 管理端商品列表。
// 分页上限 100(管理端需要一次拉全),用户端见 UserGetProductList(上限 50)。
func (l *GetProductListLogic) GetProductList(in *v1_productv1.GetProductListReq) (*v1_productv1.GetProductListResp, error) {
	page, pageSize := normalizePage(int(in.Page), int(in.PageSize), 100)
	req := converter.ToSpuQueryReq(int32(page), int32(pageSize), in.SpuName,
		nullableInt64(in.CategoryId), in.SpuStatus, in.Brand, in.Sort)

	rows, total, err := l.svcCtx.ProductRepo.GetSpuList(req)
	if err != nil {
		return nil, err
	}

	list := make([]model.SpuList, 0, len(rows))
	for _, row := range rows {
		// 类目在同一个库内,展示名直接取;取不到说明数据被删或有脏引用
		category, err := getCategory(l.svcCtx, row.CategoryId)
		if err != nil {
			return nil, err
		}
		if category == nil {
			return &v1_productv1.GetProductListResp{ErrorMsg: model.CategoryNotExist.Error()}, nil
		}
		list = append(list, model.SpuList{
			SpuId:        row.SpuId,
			SpuName:      row.SpuName,
			CategoryId:   row.CategoryId,
			CategoryName: category.CategoryName,
			Brand:        row.Brand,
			MainImage:    row.MainImage,
			SpuStatus:    row.SpuStatus,
			Priority:     row.Priority,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
			MinPrice:     int64(row.MinPrice),
			MaxPrice:     int64(row.MaxPrice),
			TotalStock:   row.TotalStock,
			TotalSold:    row.TotalSold,
		})
	}

	return &v1_productv1.GetProductListResp{
		Items:    converter.ToProtoSpuListItem(&list),
		Total:    total,
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	}, nil
}
