import service from '@/utils/request'
import type {
    CreateOrderReq,
    OrderQueryReq,
    OrderShipReq,
    AdminOrderQueryReq,
} from '@/types/order'

// ==================== 用户端 ====================

// 创建订单
export const CreateOrderApi = (data: CreateOrderReq) => {
    return service({
        url: '/orders',
        method: 'post',
        data,
    })
}

// 用户订单列表
export const GetUserOrderListApi = (params: OrderQueryReq = {}) => {
    return service({
        url: '/orders',
        method: 'get',
        params,
    })
}

// 用户订单详情
export const GetUserOrderDetailApi = (id: number) => {
    return service({
        url: `/orders/${id}`,
        method: 'get',
    })
}

// 取消订单
export const CancelOrderApi = (id: number) => {
    return service({
        url: `/orders/${id}/cancel`,
        method: 'put',
    })
}

// 确认收货
export const ConfirmOrderApi = (id: number) => {
    return service({
        url: `/orders/${id}/confirm`,
        method: 'put',
    })
}

// ==================== 管理端 ====================

// 管理端订单列表
export const GetAdminOrderListApi = (params: AdminOrderQueryReq = {}) => {
    return service({
        url: '/admin/orders',
        method: 'get',
        params,
    })
}

// 管理端订单详情
export const GetAdminOrderDetailApi = (id: number) => {
    return service({
        url: `/admin/orders/${id}`,
        method: 'get',
    })
}

// 发货
export const ShipOrderApi = (id: number, data: OrderShipReq) => {
    return service({
        url: `/admin/orders/${id}/ship`,
        method: 'put',
        data,
    })
}
