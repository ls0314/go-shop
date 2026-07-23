// 订单相关类型定义

// ==================== 创建订单 ====================

export interface CreateOrderReq {
    address_id: number
    idempotent_key: string
    buyer_remark?: string
}

export interface CreateOrderResp {
    order_id: number
    order_no: string
    pay_amount: number
    total_amount: number
    order_status: string
    pay_expire_at: string
    created_at: string
}

// ==================== 订单列表项 ====================

export interface OrderListItem {
    order_id: number
    order_no: string
    order_status: string
    total_amount: number
    pay_amount: number
    detail_count: number
    first_image: string
    created_at: string
}

export interface OrderListResp {
    list: OrderListItem[]
    total: number
    page: number
    page_size: number
}

// ==================== 订单详情 ====================

export interface OrderDetailItem {
    detail_id: number
    sku_id: number
    spu_name: string
    sku_name: string
    spec_values: Record<string, any>
    main_image: string
    quantity: number
    unit_price: number
    total_price: number
}

export interface OrderLogItem {
    log_id: number
    order_id: number
    order_status: string
    action: string
    operator: string
    detail: string
    created_at: string
}

export interface OrderDetailResp {
    order_id: number
    order_no: string
    order_status: string
    total_amount: number
    pay_amount: number
    pay_method: string
    pay_time: string
    address_snapshot: {
        receiver_name: string
        receiver_phone: string
        province: string
        city: string
        district: string
        detail_address: string
        postal_code: string
    }
    buyer_remark: string
    detail_list: OrderDetailItem[]
    log_list: OrderLogItem[]
    created_at: string
}

// ==================== 管理端列表项（多用户信息） ====================

export interface AdminOrderItem {
    order_id: number
    order_no: string
    order_status: string
    total_amount: number
    pay_amount: number
    pay_method: string
    user_name: string
    receiver_name: string
    receiver_phone: string
    created_at: string
}

export interface AdminOrderListResp {
    list: AdminOrderItem[]
    total: number
    page: number
    page_size: number
}

// ==================== 管理端详情 ====================

export interface AdminOrderDetailResp extends OrderDetailResp {
    username: string
    user_id: number
}

// ==================== 状态操作响应 ====================

export interface OrderStatusResp {
    order_id: number
    order_no: string
    order_status: string
}

// ==================== 发货 ====================

export interface OrderShipReq {
    express_company: string
    tracking_no: string
}

export interface OrderShipResp extends OrderStatusResp {
    express_company: string
    tracking_no: string
}

// ==================== 查询参数 ====================

export interface OrderQueryReq {
    page?: number
    page_size?: number
    order_status?: string
}

export interface AdminOrderQueryReq extends OrderQueryReq {
    order_no?: string
    start_time?: string
    end_time?: string
}
