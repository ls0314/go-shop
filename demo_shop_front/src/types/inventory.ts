// 库存管理相关类型定义

// 单个SKU库存信息
export interface SkuInventory {
    sku_id: number
    sku_name: string
    spu_id: number
    spu_name: string
    spec_value: Record<string, any>
    stock: number
    lock_stock: number
    total_stock: number
    sold_count: number
    sku_status: string
}

// SPU下所有SKU库存汇总
export interface SkuInventoryList {
    list: SkuInventory[]
    total_stock: number
    total_lock: number
    total_sold: number
}

// 库存调整请求
export interface InventoryAdjustReq {
    sku_id: number
    change_qty: number
    remark: string
}

// 库存调整响应
export interface InventoryAdjustResp {
    before_stock: number
    after_stock: number
}

// 库存变更日志查询参数
export interface InventoryLogQuery {
    page?: number
    page_size?: number
    sku_id?: number
    spu_id?: number
    change_type?: string
    start_time?: string
    end_time?: string
}

// 单条库存变更日志
export interface InventoryLogItem {
    log_id: number
    sku_id: number
    sku_name: string
    spu_name: string
    change_type: string
    change_qty: number
    before_stock: number
    after_stock: number
    before_lock: number
    after_lock: number
    order_id: number
    remark: string
    created_at: string
    create_by: number
}

// 库存变更日志分页响应
export interface InventoryLogListResp {
    list: InventoryLogItem[]
    total: number
    page: number
    page_size: number
}

// 低库存预警查询参数
export interface InventoryWarnQuery {
    threshold?: number
    spu_status?: string
}

// 低库存预警条目
export interface InventoryWarnItem {
    sku_id: number
    spu_name: string
    sku_name: string
    stock: number
    lock_stock: number
    sold_count: number
    sku_status: string
}
