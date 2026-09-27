import service from '@/utils/request'
import type { AddressFormData } from '@/types/address'

// 新增收货地址
export const CreateAddressApi = (data: AddressFormData) => {
    return service({
        url: '/addresses',
        method: 'post',
        data,
    })
}

// 获取地址列表
export const GetAddressListApi = () => {
    return service({
        url: '/addresses',
        method: 'get',
    })
}

// 获取单个地址详情
export const GetAddressApi = (id: number) => {
    return service({
        url: `/addresses/${id}`,
        method: 'get',
    })
}

// 更新收货地址
export const UpdateAddressApi = (id: number, data: Partial<AddressFormData>) => {
    return service({
        url: `/addresses/${id}`,
        method: 'put',
        data,
    })
}

// 删除收货地址
export const DeleteAddressApi = (id: number) => {
    return service({
        url: `/addresses/${id}`,
        method: 'delete',
    })
}

// 设为默认地址
export const SetDefaultAddressApi = (id: number) => {
    return service({
        url: `/addresses/${id}/default`,
        method: 'put',
    })
}
