import service from '@/utils/request'
import type { CreateCouponReq } from '@/types/coupon'

// ==================== 管理端 ====================

// 创建优惠券模板
export const CreateCouponApi = (data: CreateCouponReq) => {
    return service({
        url: '/admin/platform/coupons',
        method: 'post',
        data,
    })
}

// 优惠券模板列表（管理端）
export const GetCouponListApi = (params?: { page?: number; page_size?: number; coupon_name?: string; coupon_type?: string }) => {
    return service({
        url: '/admin/platform/coupons',
        method: 'get',
        params,
    })
}

// ==================== 用户端 ====================

// 领券中心模板列表（可领取的券）
export const GetReceiveCouponListApi = (params?: { page?: number; page_size?: number }) => {
    return service({
        url: '/users/platform/coupons/templates',
        method: 'get',
        params,
    })
}

// 我的卡券列表
export const GetUserCouponListApi = (params?: { page?: number; page_size?: number; status?: string }) => {
    return service({
        url: '/users/platform/coupons',
        method: 'get',
        params,
    })
}

// 领取优惠券
export const ReceiveCouponApi = (templateId: number) => {
    return service({
        url: `/users/platform/coupons/receive/${templateId}`,
        method: 'post',
    })
}

// 结算可用券列表
export const GetAvailableCouponApi = (orderAmount: number) => {
    return service({
        url: '/users/platform/coupons/available',
        method: 'get',
        params: { order_amount: orderAmount },
    })
}
