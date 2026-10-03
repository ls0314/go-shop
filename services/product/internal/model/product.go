package model

import (
	"time"

	"gorm.io/datatypes"
)

// ============================================================
// 商品域实体与聚合结构
// (商品/类目域从单体平移,SKU 与库存流水见 inventory.go)
// ============================================================

// SysCategory 类目,对应数据表 sys_category
type SysCategory struct {
	CategoryId    int64     `gorm:"primaryKey;autoIncrement;column:category_id" json:"category_id"`
	ParentId      int64     `gorm:"column:parent_id" json:"parent_id"`
	CategoryName  string    `gorm:"column:category_name" json:"category_name"`
	CategoryLevel int64     `gorm:"column:category_level" json:"category_level"`
	CategoryPath  string    `gorm:"column:category_path" json:"category_path"`
	SortOrder     int64     `gorm:"column:sort_order" json:"sort_order"`
	IconUrl       string    `gorm:"column:icon_url" json:"icon_url"`
	IsLeaf        bool      `gorm:"column:is_leaf" json:"is_leaf"`
	IsVisible     bool      `gorm:"column:is_visible" json:"is_visible"`
	Status        string    `gorm:"column:status" json:"status"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
	CreateBy      int64     `gorm:"column:create_by" json:"create_by"`
	UpdateBy      int64     `gorm:"column:update_by" json:"update_by"`
}

func (SysCategory) TableName() string { return "sys_category" }

// SysProductSpu 商品SPU,对应数据表 sys_product_spu
type SysProductSpu struct {
	SpuId        int64                 `gorm:"primaryKey;autoIncrement;column:spu_id" json:"spu_id"`
	SpuName      string                `gorm:"column:spu_name" json:"spu_name"`
	CategoryId   int64                 `gorm:"column:category_id" json:"category_id"`
	Brand        string                `gorm:"column:brand" json:"brand"`
	Description  string                `gorm:"column:description" json:"description"`
	MainImage    string                `gorm:"column:main_image" json:"main_image"`
	SpuStatus    string                `gorm:"column:spu_status" json:"spu_status"`
	SpecTemplate datatypes.JSON        `gorm:"column:spec_template" json:"spec_template"`
	IsDeleted    bool                  `gorm:"column:is_deleted" json:"is_deleted"`
	Priority     int64                 `gorm:"column:priority" json:"priority"`
	CreatedAt    time.Time             `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time             `gorm:"column:updated_at" json:"updated_at"`
	CreateBy     int64                 `gorm:"column:create_by" json:"create_by"`
	UpdateBy     int64                 `gorm:"column:update_by" json:"update_by"`
	SkuList      *[]SysProductSku      `gorm:"-" json:"sku_list"`
	ImageList    *[]SysProductSpuImage `gorm:"-" json:"image_list"`
}

func (SysProductSpu) TableName() string { return "sys_product_spu" }

// SysProductSpuImage 商品图片,对应数据表 sys_product_spu_image
type SysProductSpuImage struct {
	ImageId   int64     `gorm:"primaryKey;autoIncrement;column:image_id" json:"image_id"`
	SpuId     int64     `gorm:"column:spu_id" json:"spu_id"`
	ImageUrl  string    `gorm:"column:image_url" json:"image_url"`
	SortOrder int64     `gorm:"column:sort_order" json:"sort_order"`
	IsMain    bool      `gorm:"column:is_main" json:"is_main"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (SysProductSpuImage) TableName() string { return "sys_product_spu_image" }

// SpuWithAgg 商品SPU聚合查询结果,含SKU价格、库存、销量聚合字段
type SpuWithAgg struct {
	SysProductSpu
	MinPrice   float64 `gorm:"column:min_price"`
	MaxPrice   float64 `gorm:"column:max_price"`
	TotalStock int64   `gorm:"column:total_stock"`
	TotalSold  int64   `gorm:"column:total_sold"`
}

// SkuValidateResult SKU校验结果,用于上架前校验SKU数量、库存和价格
type SkuValidateResult struct {
	SkuCount int64   `gorm:"column:sku_count"`
	MaxStock int64   `gorm:"column:max_stock"`
	MinPrice float64 `gorm:"column:min_price"`
}

// SpuESDoc 商品搜索文档,由 sys_product_spu 联 sys_category 与 SKU 聚合生成
type SpuESDoc struct {
	SpuId        int64     `gorm:"column:spu_id"`
	SpuName      string    `gorm:"column:spu_name"`
	Brand        string    `gorm:"column:brand"`
	Description  string    `gorm:"column:description"`
	CategoryId   int64     `gorm:"column:category_id"`
	CategoryName string    `gorm:"column:category_name"`
	MainImage    string    `gorm:"column:main_image"`
	TotalStock   int64     `gorm:"column:total_stock"`
	TotalSold    int64     `gorm:"column:total_sold"`
	Priority     int64     `gorm:"column:priority"`
	SpuStatus    string    `gorm:"column:spu_status"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
	MinPrice     float64   `gorm:"column:min_price"`
	MaxPrice     float64   `gorm:"column:max_price"`
}
