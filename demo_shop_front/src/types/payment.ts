// 支付相关类型定义

export interface CreatePaymentReq {
    pay_method: string
}

export interface CreatePaymentResp {
    payment_id: number
    pay_no: string
    pay_amount: number
    pay_status: string
    pay_url?: string
    qr_code?: string
}

export interface PaymentResult {
    payment_id: number
    pay_no: string
    order_id: number
    order_no: string
    pay_method: string
    pay_amount: number
    pay_status: string
    pay_time: string
    trade_no: string
}

// ==================== 管理端支付列表 ====================

export interface AdminPaymentItem {
    payment_id: number
    pay_no: string
    order_no: string
    username: string
    pay_method: string
    pay_amount: number
    pay_status: string
    pay_time: string
}

export interface AdminPaymentListResp {
    list: AdminPaymentItem[]
    total: number
    page: number
    page_size: number
}

export interface AdminPaymentQuery {
    page?: number
    page_size?: number
    pay_status?: string
    pay_method?: string
    order_no?: string
    start_time?: string
    end_time?: string
}
