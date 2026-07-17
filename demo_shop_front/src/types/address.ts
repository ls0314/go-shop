// 收货地址相关类型定义

// 地址信息
export interface UserAddress {
    address_id: number
    receiver_name: string
    receiver_phone: string
    province: string
    city: string
    district: string
    detail_address: string
    postal_code: string
    is_default: boolean
    address_tag: string
    created_at: string
}

// 新增/编辑地址表单
export interface AddressFormData {
    receiver_name: string
    receiver_phone: string
    province: string
    city: string
    district: string
    detail_address: string
    postal_code?: string
    is_default?: boolean
    address_tag?: string
}
