// 优惠券相关类型定义

// ==================== 优惠券模板（管理端） ====================

// 创建优惠券模板请求
export interface CreateCouponReq {
    coupon_name: string
    coupon_type: string          // full_reduction(满减) / direct_discount(直减)
    threshold_amount: number     // 使用门槛,0 表示无门槛
    discount_amount: number      // 优惠金额(满减)或折扣率(直减)
    total_count: number          // 发放总量
    per_user_limit: number       // 每人限领
    usable_days: number          // 领取后有效天数(>0 表示相对有效期)
    start_time?: string          // 固定有效期开始(usable_days=0 时必填)
    end_time?: string            // 固定有效期结束(usable_days=0 时必填)
}

// 优惠券模板列表项（管理端）
export interface CouponTemplateItem {
    template_id: number
    coupon_name: string
    coupon_type: string
    threshold_amount: number
    discount_amount: number
    total_count: number
    received_count: number
    per_user_limit: number
    usable_days: number
    start_time: string
    end_time: string
    status?: string
}

export interface CouponTemplateListResp {
    list: CouponTemplateItem[]
    page: number
    page_size: number
    total: number
}

// ==================== 领券中心模板（用户端） ====================

// 领券中心模板列表项
export interface UserCouponTemplate {
    template_id: number
    coupon_name: string
    coupon_type: string
    threshold_amount: number
    discount_amount: number
    per_user_limit: number
    held_count: number           // 当前用户已领数量(服务端计算)
    remaining_count: number      // 剩余可领数量
    usable_days: number
    start_time: string
    end_time: string
}

export interface UserCouponTemplateListResp {
    list: UserCouponTemplate[]
    page: number
    page_size: number
    total: number
}

// ==================== 我的卡券（用户端） ====================

// 我的卡券列表项
export interface UserCouponItem {
    user_coupon_id: number
    coupon_name: string
    coupon_type: string
    threshold_amount: number
    discount_amount: number
    status: string                // unused / used / expired
    expire_at: string
    order_no: string
    used_at: string
}

export interface UserCouponListResp {
    list: UserCouponItem[]
    page: number
    page_size: number
    total: number
}

// 领取响应
export interface ReceiveCouponResp {
    user_coupon_id: number
    expire_time: string
}

// ==================== 结算可用券 ====================

// 结算可用券列表项（PayAfter 已按升序排列,第一张最优）
export interface AvailableCouponItem {
    user_coupon_id: number
    coupon_name: string
    coupon_type: string
    threshold_amount: number
    discount_amount: number
    pay_after: number             // 使用该券后的实付金额
}

export interface AvailableCouponResp {
    list: AvailableCouponItem[]
}
