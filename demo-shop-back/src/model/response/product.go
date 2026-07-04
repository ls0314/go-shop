package response

import (
	"time"

	"gorm.io/datatypes"
)

// SpuList 管理端商品列表响应Spu子列表
type SpuList struct {
	SpuId        int64     `json:"spu_id"`
	SpuName      string    `json:"spu_name"`
	CategoryId   int64     `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Brand        string    `json:"brand"`
	MainImage    string    `json:"main_image"`
	SpuStatus    string    `json:"spu_status"`
	Priority     int64     `json:"priority"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	MinPrice     int64     `json:"min_price"`
	MaxPrice     int64     `json:"max_price"`
	TotalStock   int64     `json:"total_stock"`
	TotalSold    int64     `json:"total_sold"`
}

// UserSpuList 用户端商品列表响应Spu子列表
type UserSpuList struct {
	SpuId        int64  `json:"spu_id"`
	SpuName      string `json:"spu_name"`
	CategoryName string `json:"category_name"`
	Brand        string `json:"brand"`
	MainImage    string `json:"main_image"`
	MinPrice     int64  `json:"min_price"`
	MaxPrice     int64  `json:"max_price"`
	TotalSold    int64  `json:"total_sold"`
}

// GetProductListResp 管理端商品列表响应体
type GetProductListResp struct {
	List     *[]SpuList `json:"list"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
}

// UserGetProductListResp 用户端商品列表响应体
type UserGetProductListResp struct {
	List     *[]UserSpuList `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
}

// ImageList 管理端商品详情响应体图像子列表
type ImageList struct {
	ImageId   int64  `json:"image_id"`
	ImageUrl  string `json:"image_url"`
	SortOrder int64  `json:"sort_order"`
	IsMain    bool   `json:"is_main"`
}

// SkuList 管理端商品详情响应体Sku子列表
type SkuList struct {
	SkuId     int64             `json:"sku_id"`
	SpuId     int64             `json:"spu_id"`
	SkuName   string            `json:"sku_name"`
	SpecValue datatypes.JSONMap `json:"spec_value"`
	Price     float64           `json:"price"`
	CostPrice float64           `json:"cost_price"`
	Stock     int64             `json:"stock"`
	LockStock int64             `json:"lock_stock"`
	SoldCount int64             `json:"sold_count"`
	SkuCode   string            `json:"sku_code"`
	SkuImage  string            `json:"sku_image"`
	SkuStatus string            `json:"sku_status"`
}

// UserSkuList 用户端商品详情响应体Sku子列表
type UserSkuList struct {
	SkuId     int64             `json:"sku_id"`
	SpuId     int64             `json:"spu_id"`
	SkuName   string            `json:"sku_name"`
	SpecValue datatypes.JSONMap `json:"spec_value"`
	Price     float64           `json:"price"`
	Stock     int64             `json:"stock"`
	SoldCount int64             `json:"sold_count"`
	SkuCode   string            `json:"sku_code"`
	SkuImage  string            `json:"sku_image"`
	SkuStatus string            `json:"sku_status"`
}

// GetProductResp 管理端商品详情响应体
type GetProductResp struct {
	SpuId        int64          `json:"spu_id"`
	SpuName      string         `json:"spu_name"`
	CategoryId   int64          `json:"category_id"`
	CategoryName string         `json:"category_name"`
	Brand        string         `json:"brand"`
	Description  string         `json:"description"`
	MainImage    string         `json:"main_image"`
	SpecTemplate datatypes.JSON `json:"spec_template"`
	SpuStatus    string         `json:"spu_status"`
	Priority     int64          `json:"priority"`
	SkuList      *[]SkuList     `json:"sku_list"`
	ImageList    *[]ImageList   `json:"image_list"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// UserGetProductResp 用户端商品详情响应体
type UserGetProductResp struct {
	SpuId        int64          `json:"spu_id"`
	SpuName      string         `json:"spu_name"`
	CategoryId   int64          `json:"category_id"`
	CategoryName string         `json:"category_name"`
	Brand        string         `json:"brand"`
	Description  string         `json:"description"`
	MainImage    string         `json:"main_image"`
	SpecTemplate datatypes.JSON `json:"spec_template"`
	SpuStatus    string         `json:"spu_status"`
	Priority     int64          `json:"priority"`
	SkuList      *[]UserSkuList `json:"sku_list"`
	ImageList    *[]ImageList   `json:"image_list"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}
