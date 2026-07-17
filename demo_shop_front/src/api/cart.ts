import service from '@/utils/request'
import type { AddCartReq, UpdateCartReq, SelectAllReq } from '@/types/cart'

// 加入购物车
export const AddCartApi = (data: AddCartReq) => {
    return service({
        url: '/api/v1/user/cart',
        method: 'post',
        data,
    })
}

// 获取购物车列表
export const GetCartListApi = () => {
    return service({
        url: '/api/v1/user/cart',
        method: 'get',
    })
}

// 获取购物车总数量（角标用）
export const GetCartCountApi = () => {
    return service({
        url: '/api/v1/user/cart/count',
        method: 'get',
    })
}

// 选中项结算预览
export const GetCartPreviewApi = () => {
    return service({
        url: '/api/v1/user/cart/preview',
        method: 'get',
    })
}

// 更新购物车项（数量/选中）
export const UpdateCartApi = (id: number, data: UpdateCartReq) => {
    return service({
        url: `/api/v1/user/cart/${id}`,
        method: 'put',
        data,
    })
}

// 全选/取消全选
export const SelectAllCartApi = (data: SelectAllReq) => {
    return service({
        url: '/api/v1/user/cart/select-all',
        method: 'put',
        data,
    })
}

// 删除购物车项
export const DeleteCartApi = (id: number) => {
    return service({
        url: `/api/v1/user/cart/${id}`,
        method: 'delete',
    })
}
