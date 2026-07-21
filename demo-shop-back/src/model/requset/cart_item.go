package requset

type UpdateCartItemReq struct {
	Quantity   int64 `json:"quantity"`
	IsSelected bool  `json:"is_selected"`
}

type CreateCartItemReq struct {
	SkuId    int64 `json:"sku_id"`
	Quantity int64 `json:"quantity"`
}

type SelectCartItemReq struct {
	IsSelected bool `json:"is_selected"`
}
