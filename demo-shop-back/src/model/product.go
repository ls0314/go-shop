package model

import (
	"time"

	"gorm.io/datatypes"
)

// SysProductSpu 商品SPU结构体, 对应数据表sys_product_spu
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

// TableName 返回商品SPU表名
// 接收值：无接收值
// 返回值：string - 数据库表名
func (SysProductSpu) TableName() string {
	return "sys_product_spu"
}

// SysProductSku 商品SKU结构体, 对应数据表sys_product_sku
type SysProductSku struct {
	SkuId      int64             `gorm:"primaryKey;autoIncrement;column:sku_id" json:"sku_id"`
	SpuId      int64             `gorm:"column:spu_id" json:"spu_id"`
	SkuName    string            `gorm:"column:sku_name" json:"sku_name"`
	SpecValues datatypes.JSONMap `gorm:"column:spec_values" json:"spec_values"`
	Price      float64           `gorm:"column:price" json:"price"`
	CostPrice  float64           `gorm:"column:cost_price" json:"cost_price"`
	Stock      int64             `gorm:"column:stock" json:"stock"`
	LockStock  int64             `gorm:"column:lock_stock" json:"lock_stock"`
	SoldCount  int64             `gorm:"column:sold_count" json:"sold_count"`
	SkuCode    string            `gorm:"column:sku_code" json:"sku_code"`
	SkuImage   string            `gorm:"column:sku_image" json:"sku_image"`
	SkuStatus  string            `gorm:"column:sku_status" json:"sku_status"`
	IsDeleted  bool              `gorm:"column:is_deleted" json:"is_deleted"`
	CreatedAt  time.Time         `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time         `gorm:"column:updated_at" json:"updated_at"`
	CreateBy   int64             `gorm:"column:create_by" json:"create_by"`
	UpdateBy   int64             `gorm:"column:update_by" json:"update_by"`
}

// TableName 返回商品SKU表名
// 接收值：无接收值
// 返回值：string - 数据库表名
func (SysProductSku) TableName() string {
	return "sys_product_sku"
}

// SysProductSpuImage 商品图片结构体, 对应数据表sys_product_spu_image
type SysProductSpuImage struct {
	ImageId   int64     `gorm:"primaryKey;autoIncrement;column:image_id" json:"image_id"`
	SpuId     int64     `gorm:"column:spu_id" json:"spu_id"`
	ImageUrl  string    `gorm:"column:image_url" json:"image_url"`
	SortOrder int64     `gorm:"column:sort_order" json:"sort_order"`
	IsMain    bool      `gorm:"column:is_main" json:"is_main"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName 返回商品图片表名
// 接收值：无接收值
// 返回值：string - 数据库表名
func (SysProductSpuImage) TableName() string {
	return "sys_product_spu_image"
}

// SysProductStockLog 库存变更日志结构体, 对应数据表sys_product_stock_log
type SysProductStockLog struct {
	LogId       int64     `gorm:"primaryKey;column:log_id" json:"log_id"`
	SkuId       int64     `gorm:"column:sku_id" json:"sku_id"`
	ChangeType  string    `gorm:"column:change_type" json:"change_type"`
	ChangeQty   int64     `gorm:"column:change_qty" json:"change_qty"`
	BeforeStock int64     `gorm:"column:before_stock" json:"before_stock"`
	AfterStock  int64     `gorm:"column:after_stock" json:"after_stock"`
	BeforeLock  int64     `gorm:"column:before_lock" json:"before_lock"`
	AfterLock   int64     `gorm:"column:after_lock" json:"after_lock"`
	OrderId     int64     `gorm:"column:order_id" json:"order_id"`
	Remark      string    `gorm:"column:remark" json:"remark"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	CreateBy    int64     `gorm:"column:create_by" json:"create_by"`
}

// TableName 返回库存变更日志表名
// 接收值：无接收值
// 返回值：string - 数据库表名
func (SysProductStockLog) TableName() string {
	return "sys_product_stock_log"
}

// SpuWithAgg 商品SPU聚合查询结果结构体, 包含SKU价格、库存、销量聚合字段
type SpuWithAgg struct {
	SysProductSpu
	MinPrice   float64 `gorm:"column:min_price"`
	MaxPrice   float64 `gorm:"column:max_price"`
	TotalStock int64   `gorm:"column:total_stock"`
	TotalSold  int64   `gorm:"column:total_sold"`
}

// SkuValidateResult SKU校验结果结构体, 用于上架前校验SKU数量、库存和价格
type SkuValidateResult struct {
	SkuCount int64   `gorm:"column:sku_count"`
	MaxStock int64   `gorm:"column:max_stock"`
	MinPrice float64 `gorm:"column:min_price"`
}

type SkuListWithAgg struct {
	TotalStock int64 `gorm:"column:total_stock"`
	TotalLock  int64 `gorm:"column:total_lock"`
	TotalSold  int64 `gorm:"column:total_sold"`
}

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
