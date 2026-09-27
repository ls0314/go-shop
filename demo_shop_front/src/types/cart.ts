// 购物车相关类型定义

// 购物车列表项（含实时联表查询的 SKU/SPU 信息）
export interface CartItem {
    cart_item_id: number
    sku_id: number
    spu_id: number
    spu_name: string
    main_image: string
    sku_name: string
    spec_values: Record<string, any>
    sku_image: string
    price: number
    stock: number
    quantity: number
    is_selected: boolean
    subtotal: number
    is_available: boolean
    unavailable_reason: string
}

// 加入购物车请求
export interface AddCartReq {
    sku_id: number
    quantity?: number
}

// 加入购物车响应
export interface AddCartResp {
    cart_item_id: number
    quantity: number
}

// 更新购物车项请求
export interface UpdateCartReq {
    quantity?: number
    is_selected?: boolean
}

// 全选/取消全选请求
export interface SelectAllReq {
    is_selected: boolean
}

// 结算预览响应
export interface CartPayPreview {
    items: CartItem[]
    total_count: number
    total_quantity: number
    total_amount: number
    has_unavailable: boolean
    unavailable_items: CartItem[]
}
