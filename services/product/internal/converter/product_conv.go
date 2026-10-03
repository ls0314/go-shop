package converter

import (
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/model"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/datatypes"
)

// jsonToText 把 JSONB 字段转成 proto 承载的 JSON 文本。
// 空值统一为空串 —— proto 侧用 "" 表示 SQL NULL,与 DB 语义一致。
func jsonToText(raw datatypes.JSON) string {
	if len(raw) == 0 {
		return ""
	}
	return string(raw)
}

// textToJSON 把 proto 的 JSON 文本还原成 JSONB 字段。
// 空串还原为 NULL,避免写入空 JSON 触发 DB 的类型错误。
func textToJSON(s string) datatypes.JSON {
	if s == "" {
		return nil
	}
	return datatypes.JSON([]byte(s))
}

// jsonMapToText 把 spec_values 转成 JSON 文本。空 map 视为空串。
func jsonMapToText(m datatypes.JSONMap) string {
	if len(m) == 0 {
		return ""
	}
	b, err := m.MarshalJSON()
	if err != nil {
		return ""
	}
	return string(b)
}

// textToJSONMap 把 JSON 文本还原为 spec_values。空串还原为空 map
// (DB 该列 NOT NULL DEFAULT '{}',不能写 NULL)。
func textToJSONMap(s string) datatypes.JSONMap {
	if s == "" {
		return datatypes.JSONMap{}
	}
	m := datatypes.JSONMap{}
	if err := m.UnmarshalJSON([]byte(s)); err != nil {
		return datatypes.JSONMap{}
	}
	return m
}

// ToProtoSku 把 model 转成 proto。
func ToProtoSku(s *model.SysProductSku) *v1_productv1.Sku {
	if s == nil {
		return nil
	}
	return &v1_productv1.Sku{
		SkuId:      s.SkuId,
		SpuId:      s.SpuId,
		SkuName:    s.SkuName,
		SpecValues: jsonMapToText(s.SpecValues),
		Price:      s.Price,
		CostPrice:  s.CostPrice,
		Stock:      s.Stock,
		LockStock:  s.LockStock,
		SoldCount:  s.SoldCount,
		SkuCode:    s.SkuCode,
		SkuImage:   s.SkuImage,
		SkuStatus:  s.SkuStatus,
	}
}

// FromProtoSku 把 proto 入参转成 model。
// created_at/updated_at 由数据库生成,不从入参取。
func FromProtoSku(p *v1_productv1.Sku) model.SysProductSku {
	if p == nil {
		return model.SysProductSku{}
	}
	return model.SysProductSku{
		SkuId:      p.GetSkuId(),
		SpuId:      p.GetSpuId(),
		SkuName:    p.GetSkuName(),
		SpecValues: textToJSONMap(p.GetSpecValues()),
		Price:      p.GetPrice(),
		CostPrice:  p.GetCostPrice(),
		Stock:      p.GetStock(),
		LockStock:  p.GetLockStock(),
		SoldCount:  p.GetSoldCount(),
		SkuCode:    p.GetSkuCode(),
		SkuImage:   p.GetSkuImage(),
		SkuStatus:  p.GetSkuStatus(),
	}
}

// ToProtoSkuList 转换 SKU 列表。
func ToProtoSkuList(list []model.SysProductSku) []*v1_productv1.Sku {
	out := make([]*v1_productv1.Sku, 0, len(list))
	for i := range list {
		out = append(out, ToProtoSku(&list[i]))
	}
	return out
}

// ToProtoImage 把 model 转成 proto。
func ToProtoImage(i *model.SysProductSpuImage) *v1_productv1.SpuImage {
	if i == nil {
		return nil
	}
	return &v1_productv1.SpuImage{
		ImageId:   i.ImageId,
		SpuId:     i.SpuId,
		ImageUrl:  i.ImageUrl,
		SortOrder: i.SortOrder,
		IsMain:    i.IsMain,
	}
}

// FromProtoImage 把 proto 入参转成 model。
func FromProtoImage(p *v1_productv1.SpuImage) model.SysProductSpuImage {
	if p == nil {
		return model.SysProductSpuImage{}
	}
	return model.SysProductSpuImage{
		ImageId:   p.GetImageId(),
		SpuId:     p.GetSpuId(),
		ImageUrl:  p.GetImageUrl(),
		SortOrder: p.GetSortOrder(),
		IsMain:    p.GetIsMain(),
	}
}

// FromProtoSpu 把 proto 入参转成 model(含 SKU 与图片列表)。
func FromProtoSpu(p *v1_productv1.Spu) *model.SysProductSpu {
	if p == nil {
		return &model.SysProductSpu{}
	}
	spu := &model.SysProductSpu{
		SpuId:        p.GetSpuId(),
		SpuName:      p.GetSpuName(),
		CategoryId:   p.GetCategoryId(),
		Brand:        p.GetBrand(),
		Description:  p.GetDescription(),
		MainImage:    p.GetMainImage(),
		SpecTemplate: textToJSON(p.GetSpecTemplate()),
		SpuStatus:    p.GetSpuStatus(),
		Priority:     p.GetPriority(),
		IsDeleted:    p.GetIsDeleted(),
	}
	if len(p.GetSkuList()) > 0 {
		skuList := make([]model.SysProductSku, 0, len(p.GetSkuList()))
		for _, s := range p.GetSkuList() {
			skuList = append(skuList, FromProtoSku(s))
		}
		spu.SkuList = &skuList
	}
	if len(p.GetImageList()) > 0 {
		imageList := make([]model.SysProductSpuImage, 0, len(p.GetImageList()))
		for _, i := range p.GetImageList() {
			imageList = append(imageList, FromProtoImage(i))
		}
		spu.ImageList = &imageList
	}
	return spu
}

// ToProtoProductDetail 由商品详情响应组装 proto。
// 管理端与用户端共用同一结构:锁库存/成本价等内部字段由调用方按端裁剪。
func ToProtoProductDetail(d *model.GetProductResp) *v1_productv1.ProductDetail {
	if d == nil {
		return nil
	}
	skuList := make([]*v1_productv1.Sku, 0, len(*d.SkuList))
	for i := range *d.SkuList {
		s := &(*d.SkuList)[i]
		skuList = append(skuList, &v1_productv1.Sku{
			SkuId:      s.SkuId,
			SpuId:      s.SpuId,
			SkuName:    s.SkuName,
			SpecValues: jsonMapToText(s.SpecValues),
			Price:      s.Price,
			CostPrice:  s.CostPrice,
			Stock:      s.Stock,
			LockStock:  s.LockStock,
			SoldCount:  s.SoldCount,
			SkuCode:    s.SkuCode,
			SkuImage:   s.SkuImage,
			SkuStatus:  s.SkuStatus,
		})
	}
	imageList := make([]*v1_productv1.SpuImage, 0, len(*d.ImageList))
	for i := range *d.ImageList {
		img := &(*d.ImageList)[i]
		imageList = append(imageList, &v1_productv1.SpuImage{
			ImageId:   img.ImageId,
			ImageUrl:  img.ImageUrl,
			SortOrder: img.SortOrder,
			IsMain:    img.IsMain,
		})
	}
	return &v1_productv1.ProductDetail{
		SpuId:        d.SpuId,
		SpuName:      d.SpuName,
		CategoryId:   d.CategoryId,
		CategoryName: d.CategoryName,
		Brand:        d.Brand,
		Description:  d.Description,
		MainImage:    d.MainImage,
		SpecTemplate: jsonToText(d.SpecTemplate),
		SpuStatus:    d.SpuStatus,
		Priority:     d.Priority,
		SkuList:      skuList,
		ImageList:    imageList,
		CreatedAt:    timestamppb.New(d.CreatedAt),
		UpdatedAt:    timestamppb.New(d.UpdatedAt),
	}
}

// ToProtoProductDetailFromUser 由用户端详情响应组装 proto。
// 用户端不返回 cost_price 与 lock_stock,此处显式留空。
func ToProtoProductDetailFromUser(d *model.UserGetProductResp) *v1_productv1.ProductDetail {
	if d == nil {
		return nil
	}
	skuList := make([]*v1_productv1.Sku, 0, len(*d.SkuList))
	for i := range *d.SkuList {
		s := &(*d.SkuList)[i]
		skuList = append(skuList, &v1_productv1.Sku{
			SkuId:      s.SkuId,
			SpuId:      s.SpuId,
			SkuName:    s.SkuName,
			SpecValues: jsonMapToText(s.SpecValues),
			Price:      s.Price,
			Stock:      s.Stock,
			SoldCount:  s.SoldCount,
			SkuCode:    s.SkuCode,
			SkuImage:   s.SkuImage,
			SkuStatus:  s.SkuStatus,
		})
	}
	imageList := make([]*v1_productv1.SpuImage, 0, len(*d.ImageList))
	for i := range *d.ImageList {
		img := &(*d.ImageList)[i]
		imageList = append(imageList, &v1_productv1.SpuImage{
			ImageId:   img.ImageId,
			ImageUrl:  img.ImageUrl,
			SortOrder: img.SortOrder,
			IsMain:    img.IsMain,
		})
	}
	return &v1_productv1.ProductDetail{
		SpuId:        d.SpuId,
		SpuName:      d.SpuName,
		CategoryId:   d.CategoryId,
		CategoryName: d.CategoryName,
		Brand:        d.Brand,
		Description:  d.Description,
		MainImage:    d.MainImage,
		SpecTemplate: jsonToText(d.SpecTemplate),
		SpuStatus:    d.SpuStatus,
		Priority:     d.Priority,
		SkuList:      skuList,
		ImageList:    imageList,
		CreatedAt:    timestamppb.New(d.CreatedAt),
		UpdatedAt:    timestamppb.New(d.UpdatedAt),
	}
}

// ToProtoSpuListItem 转换管理端列表项。
func ToProtoSpuListItem(items *[]model.SpuList) []*v1_productv1.SpuListItem {
	if items == nil {
		return nil
	}
	out := make([]*v1_productv1.SpuListItem, 0, len(*items))
	for i := range *items {
		it := &(*items)[i]
		out = append(out, &v1_productv1.SpuListItem{
			SpuId:        it.SpuId,
			SpuName:      it.SpuName,
			CategoryId:   it.CategoryId,
			CategoryName: it.CategoryName,
			Brand:        it.Brand,
			MainImage:    it.MainImage,
			SpuStatus:    it.SpuStatus,
			Priority:     it.Priority,
			MinPrice:     it.MinPrice,
			MaxPrice:     it.MaxPrice,
			TotalStock:   it.TotalStock,
			TotalSold:    it.TotalSold,
			CreatedAt:    timestamppb.New(it.CreatedAt),
			UpdatedAt:    timestamppb.New(it.UpdatedAt),
		})
	}
	return out
}

// ToProtoUserSpuListItem 转换用户端列表项。
func ToProtoUserSpuListItem(items *[]model.UserSpuList) []*v1_productv1.UserSpuListItem {
	if items == nil {
		return nil
	}
	out := make([]*v1_productv1.UserSpuListItem, 0, len(*items))
	for i := range *items {
		it := &(*items)[i]
		out = append(out, &v1_productv1.UserSpuListItem{
			SpuId:        it.SpuId,
			SpuName:      it.SpuName,
			CategoryName: it.CategoryName,
			Brand:        it.Brand,
			MainImage:    it.MainImage,
			MinPrice:     it.MinPrice,
			MaxPrice:     it.MaxPrice,
			TotalSold:    it.TotalSold,
			Stock:        it.Stock,
		})
	}
	return out
}

// ToSpuQueryReq 由 proto 查询入参转成领域查询条件。
func ToSpuQueryReq(page, pageSize int32, spuName string, categoryId *int64, spuStatus, brand, sort string) model.SpuQueryReq {
	return model.SpuQueryReq{
		Page:       int(page),
		PageSize:   int(pageSize),
		SpuName:    spuName,
		CategoryId: categoryId,
		SpuStatus:  spuStatus,
		Brand:      brand,
		Sort:       sort,
	}
}

// ToFullUpdateProductReq 由 proto 入参转成全量更新请求。
func ToFullUpdateProductReq(p *v1_productv1.UpdateProductFullReq) model.FullUpdateProductReq {
	req := model.FullUpdateProductReq{
		SpuName:        p.GetSpuName(),
		Brand:          p.GetBrand(),
		Description:    p.GetDescription(),
		MainImage:      p.GetMainImage(),
		DeleteImageIds: p.GetDeleteImageIds(),
	}
	if p.CategoryId != nil {
		v := p.GetCategoryId().GetValue()
		req.CategoryId = &v
	}
	if p.Priority != nil {
		v := p.GetPriority().GetValue()
		req.Priority = &v
	}
	if p.SpecTemplate != "" {
		raw := textToJSON(p.GetSpecTemplate())
		req.SpecTemplate = &raw
	}
	if len(p.GetSkuList()) > 0 {
		skuList := make([]model.SysProductSku, 0, len(p.GetSkuList()))
		for _, s := range p.GetSkuList() {
			skuList = append(skuList, FromProtoSku(s))
		}
		req.SkuList = &skuList
	}
	if len(p.GetImageList()) > 0 {
		imageList := make([]model.SysProductSpuImage, 0, len(p.GetImageList()))
		for _, i := range p.GetImageList() {
			imageList = append(imageList, FromProtoImage(i))
		}
		req.ImageList = &imageList
	}
	return req
}
