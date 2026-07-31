import service from '@/utils/request'
import type { CreatePaymentReq, AdminPaymentQuery } from '@/types/payment'

// 发起支付
export const CreatePaymentApi = (orderId: number, data: CreatePaymentReq) => {
    return service({
        url: `/api/v1/users/pay/order/${orderId}`,
        method: 'post',
        data,
    })
}

// 查询支付状态
export const GetPaymentApi = (payNo: string) => {
    return service({
        url: `/api/v1/users/pay/${payNo}`,
        method: 'get',
    })
}

// mock 支付回调
export const MockPayCallbackApi = (payNo: string) => {
    return service({
        url: '/api/v1/pay/callback/mock',
        method: 'post',
        data: { pay_no: payNo },
    })
}

// ==================== 管理端 ====================

// 管理端支付列表
export const GetPaymentListApi = (params: AdminPaymentQuery = {}) => {
    return service({
        url: '/api/v1/admin/pay/list',
        method: 'get',
        params,
    })
}
