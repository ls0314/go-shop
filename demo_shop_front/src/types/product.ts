// ============================================================
// 规格模板
// ============================================================

/** 规格模板中的单条规格 */
export interface SpecTemplateItem {
  name: string
  values: string[]
}

// ============================================================
// 图片（创建 / 管理端 / 用户端 共用）
// ============================================================

/** 商品图片 */
export interface ProductImage {
  image_id?: number
  image_url: string
  sort_order?: number
  is_main?: boolean
}

// ============================================================
// SKU —— 创建
// ============================================================

/** 创建商品时的 SKU 字段（不传 sku_id） */
export interface CreateSkuItem {
  sku_name?: string
  spec_values: Record<string, string>
  price: number
  cost_price?: number
  stock?: number
  sku_code?: string
  sku_image?: string
  sku_status?: string
}

// ============================================================
// SKU —— 管理端详情（含内部字段）
// ============================================================

/** 管理端 SKU 详情（含 cost_price / lock_stock） */
export interface SkuItem {
  sku_id: number
  spu_id: number
  sku_name: string
  spec_values: Record<string, string>
  price: number
  cost_price: number
  stock: number
  lock_stock: number
  sold_count: number
  sku_code: string
  sku_image: string
  sku_status: string
}

// ============================================================
// SKU —— 用户端详情（不含内部字段）
// ============================================================

/** 用户端 SKU 详情（不含 cost_price / lock_stock） */
export interface UserSkuItem {
  sku_id: number
  spu_id: number
  sku_name: string
  spec_values: Record<string, string>
  price: number
  stock: number
  sold_count: number
  sku_code: string
  sku_image: string
  sku_status: string
}

// ============================================================
// 管理端列表
// ============================================================

/** 管理端商品列表项 */
export interface SpuListItem {
  spu_id: number
  spu_name: string
  category_id: number
  category_name: string
  brand: string
  main_image: string
  spu_status: string
  priority: number
  min_price: number
  max_price: number
  total_stock: number
  total_sold: number
  created_at: string
  updated_at: string
}

/** 管理端商品列表响应 */
export interface GetSpuListResp {
  list: SpuListItem[]
  total: number
  page: number
  pageSize: number
}

// ============================================================
// 用户端列表
// ============================================================

/** 用户端商品列表项 */
export interface UserSpuListItem {
  spu_id: number
  spu_name: string
  category_name: string
  brand: string
  main_image: string
  min_price: number
  max_price: number
  total_sold: number
}

/** 用户端商品列表响应 */
export interface UserGetSpuListResp {
  list: UserSpuListItem[]
  total: number
  page: number
  pageSize: number
}

// ============================================================
// 管理端详情
// ============================================================

/** 管理端商品详情 */
export interface GetProductResp {
  spu_id: number
  spu_name: string
  category_id: number
  category_name: string
  brand: string
  description: string
  main_image: string
  spec_template: SpecTemplateItem[]
  spu_status: string
  priority: number
  sku_list: SkuItem[]
  image_list: ProductImage[]
  created_at: string
  updated_at: string
}

// ============================================================
// 用户端详情
// ============================================================

/** 用户端商品详情 */
export interface UserGetProductResp {
  spu_id: number
  spu_name: string
  category_id: number
  category_name: string
  brand: string
  description: string
  main_image: string
  spec_template: SpecTemplateItem[]
  sku_list: UserSkuItem[]
  image_list: ProductImage[]
  created_at: string
  updated_at: string
}

// ============================================================
// 请求参数
// ============================================================

/** 创建商品请求 */
export interface CreateProductReq {
  spu_name: string
  category_id: number
  brand?: string
  description?: string
  main_image?: string
  spec_template: SpecTemplateItem[]
  priority?: number
  sku_list: CreateSkuItem[]
  image_list?: ProductImage[]
}

/** 商品列表查询参数（管理端 + 用户端共用）
 *
 * 注意 `page_size` 是**下划线**命名:后端结构体的 form 标签就是
 * `page_size`(全项目统一,见 coupon / order / inventoryLog 等 requset),
 * 写成 pageSize 会被 ShouldBindQuery 静默忽略、永远只返回默认 10 条。
 */
export interface SpuQueryReq {
  page?: number
  page_size?: number
  spu_name?: string
  category_id?: number
  spu_status?: string
  brand?: string
  sort?: string
}

/** 部分更新商品请求 */
export interface UpdateProductReq {
  spu_name?: string
  category_id?: number
  brand?: string
  description?: string
  main_image?: string
  priority?: number
  spu_status?: string
}

/** 全量更新商品请求（含 SKU + 图片） */
export interface UpdateProductFullReq {
  spu_name?: string
  category_id?: number
  brand?: string
  description?: string
  main_image?: string
  spec_template?: SpecTemplateItem[]
  priority?: number
  sku_list?: SkuItem[]
  image_list?: ProductImage[]
  delete_image_ids?: number[]
}
