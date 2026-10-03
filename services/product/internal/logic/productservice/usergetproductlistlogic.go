package productservicelogic

import (
	"context"
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/converter"
	"demo-shop/services/product/internal/infra/es"
	"demo-shop/services/product/internal/model"
	"demo-shop/services/product/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserGetProductListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserGetProductListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserGetProductListLogic {
	return &UserGetProductListLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// UserGetProductList 用户端商品列表(仅已上架)。
//
// [FIX-10] 单体在校验上不对称:入口不校验翻页,于是 ES 路径原样透传 Page/PageSize,
// 而 DB 路径自己做了默认值与封顶(前者不封顶、后者封顶 50),同一接口两条路径行为不同。
// 改为入口统一校验一次,两条路径共用同一组参数(上限 50,用户端不需要更大的页)。
func (l *UserGetProductListLogic) UserGetProductList(in *v1_productv1.UserGetProductListReq) (*v1_productv1.UserGetProductListResp, error) {
	page, pageSize := normalizePage(int(in.Page), int(in.PageSize), 50)
	req := converter.ToSpuQueryReq(int32(page), int32(pageSize), in.SpuName,
		nullableInt64(in.CategoryId), "", in.Brand, in.Sort)
	// 用户端只展示已上架商品
	req.SpuStatus = model.SpuStatusPublished

	// 有关键词且 ES 可用时走搜索;ES 报错则回落 DB,不让搜索故障影响下单前的浏览
	if req.SpuName != "" && l.svcCtx.ES != nil {
		resp, err := l.getProductListByES(req)
		if err == nil {
			return resp, nil
		}
		l.Errorf("ES 搜索失败,降级直查数据库: %v", err)
	}
	return l.getProductSpuListByDB(req)
}

// getProductListByES 走 ES 搜索
func (l *UserGetProductListLogic) getProductListByES(req model.SpuQueryReq) (*v1_productv1.UserGetProductListResp, error) {
	var categoryId int64
	if req.CategoryId != nil {
		categoryId = *req.CategoryId
	}
	searchProduct, err := l.svcCtx.ES.Search(context.Background(), &es.SearchRequest{
		Keyword:    req.SpuName,
		CategoryId: categoryId,
		Brand:      req.Brand,
		Sort:       req.Sort,
		Page:       req.Page,
		PageSize:   req.PageSize,
	})
	if err != nil {
		return nil, err
	}

	spuList := make([]model.UserSpuList, 0, len(searchProduct.Products))
	for _, product := range searchProduct.Products {
		spuList = append(spuList, model.UserSpuList{
			SpuId:        product.SpuId,
			SpuName:      product.SpuName,
			CategoryName: product.CategoryName,
			Brand:        product.Brand,
			MainImage:    product.MainImage,
			MaxPrice:     int64(product.MaxPrice),
			MinPrice:     int64(product.MinPrice),
			TotalSold:    product.TotalSold,
		})
	}

	return &v1_productv1.UserGetProductListResp{
		Items:    converter.ToProtoUserSpuListItem(&spuList),
		Total:    searchProduct.Total,
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	}, nil
}

// getProductSpuListByDB 直连 DB 获取用户端商品列表
func (l *UserGetProductListLogic) getProductSpuListByDB(req model.SpuQueryReq) (*v1_productv1.UserGetProductListResp, error) {
	rows, total, err := l.svcCtx.ProductRepo.GetSpuList(req)
	if err != nil {
		return nil, err
	}

	list := make([]model.UserSpuList, 0, len(rows))
	for _, row := range rows {
		category, err := getCategory(l.svcCtx, row.CategoryId)
		if err != nil {
			return nil, err
		}
		if category == nil {
			return &v1_productv1.UserGetProductListResp{ErrorMsg: model.CategoryNotExist.Error()}, nil
		}
		list = append(list, model.UserSpuList{
			SpuId:        row.SpuId,
			SpuName:      row.SpuName,
			CategoryName: category.CategoryName,
			Brand:        row.Brand,
			MainImage:    row.MainImage,
			MinPrice:     int64(row.MinPrice),
			MaxPrice:     int64(row.MaxPrice),
			TotalSold:    row.TotalSold,
			// [FIX-9] 单体此处漏赋 Stock,用户端列表库存恒为 0;
			// 而同文件的 GetProduct 路径专门用 fillSkuStock 覆盖库存,可见是漏写。
			Stock: row.TotalStock,
		})
	}

	return &v1_productv1.UserGetProductListResp{
		Items:    converter.ToProtoUserSpuListItem(&list),
		Total:    total,
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	}, nil
}
