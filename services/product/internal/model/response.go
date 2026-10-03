package model

import (
	"time"

	"gorm.io/datatypes"
)

// ============================================================
// 商品域响应结构(从单体 model/response 平移)
// repo 层的 Scan 直接依赖这些结构,故随域一起搬。
// ============================================================

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
	Stock        int64  `json:"stock"`
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
	SkuId      int64             `json:"sku_id"`
	SpuId      int64             `json:"spu_id"`
	SkuName    string            `json:"sku_name"`
	SpecValues datatypes.JSONMap `json:"spec_values"`
	Price      float64           `json:"price"`
	CostPrice  float64           `json:"cost_price"`
	Stock      int64             `json:"stock"`
	LockStock  int64             `json:"lock_stock"`
	SoldCount  int64             `json:"sold_count"`
	SkuCode    string            `json:"sku_code"`
	SkuImage   string            `json:"sku_image"`
	SkuStatus  string            `json:"sku_status"`
}

// UserSkuList 用户端商品详情响应体Sku子列表
type UserSkuList struct {
	SkuId      int64             `json:"sku_id"`
	SpuId      int64             `json:"spu_id"`
	SkuName    string            `json:"sku_name"`
	SpecValues datatypes.JSONMap `json:"spec_values"`
	Price      float64           `json:"price"`
	Stock      int64             `json:"stock"`
	SoldCount  int64             `json:"sold_count"`
	SkuCode    string            `json:"sku_code"`
	SkuImage   string            `json:"sku_image"`
	SkuStatus  string            `json:"sku_status"`
}

// ============================================================
// 库存域响应结构(管理端库存看板)
// ============================================================

// StockWarnItem 低库存预警条目。查询时联 spu 取 spu_name,
// 故 spu_name 不在 SysProductSku 上,单独建结构承接 Scan 结果。
type StockWarnItem struct {
	SkuId     int64  `gorm:"column:sku_id" json:"sku_id"`
	SpuName   string `gorm:"column:spu_name" json:"spu_name"`
	SkuName   string `gorm:"column:sku_name" json:"sku_name"`
	Stock     int64  `gorm:"column:stock" json:"stock"`
	LockStock int64  `gorm:"column:lock_stock" json:"lock_stock"`
	SoldCount int64  `gorm:"column:sold_count" json:"sold_count"`
	SkuStatus string `gorm:"column:sku_status" json:"sku_status"`
}

// StockLogItem 库存流水条目:流水行 + 联表取到的 sku/spu 名称。
// 名称不在 SysProductStockLog 上,故单独建结构承接 Scan 结果。
type StockLogItem struct {
	SysProductStockLog
	// 以下三个字段由 LEFT JOIN sys_product_sku / sys_product_spu 得到
	SkuName string `gorm:"column:sku_name" json:"sku_name"`
	SpuId   int64  `gorm:"column:spu_id" json:"spu_id"`
	SpuName string `gorm:"column:spu_name" json:"spu_name"`
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

// ============================================================
// 类目域响应结构
// ============================================================

// CreateCategoryResp 创建类目接口响应结构体
type CreateCategoryResp struct {
	CategoryId   int64  `json:"category_id"`
	CategoryPath string `json:"category_path"`
}

// GetTreeCategoryResp 获取类目树响应结构体
type GetTreeCategoryResp struct {
	CategoryId    int64                  `json:"category_id"`
	CategoryName  string                 `json:"category_name"`
	CategoryLevel int64                  `json:"category_level"`
	IsVisible     bool                   `json:"is_visible"`
	Status        string                 `json:"status"`
	Children      []*GetTreeCategoryResp `json:"children"`
	ParentId      int64                  `json:"-"`
}

// GetListCategoryResp 获取子类目列表响应结构体
type GetListCategoryResp struct {
	CategoryId    int64  `json:"category_id"`
	ParentId      int64  `json:"parent_id"`
	CategoryName  string `json:"category_name"`
	CategoryLevel int64  `json:"category_level"`
	SortOrder     int64  `json:"sort_order"`
	IsLeaf        bool   `json:"is_leaf"`
	IsVisible     bool   `json:"is_visible"`
	Status        string `json:"status"`
}
