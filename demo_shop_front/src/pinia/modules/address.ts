import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { UserAddress, AddressFormData } from '@/types/address'
import {
    CreateAddressApi,
    GetAddressListApi,
    UpdateAddressApi,
    DeleteAddressApi,
    SetDefaultAddressApi,
} from '@/api/address'

export const useAddressStore = defineStore('address', () => {
    const addressList = ref<UserAddress[]>([])

    // 获取地址列表
    async function GetAddressList(): Promise<UserAddress[]> {
        try {
            const res = await GetAddressListApi()
            addressList.value = res.data.data || []
            return addressList.value
        } catch (error) {
            console.error('获取地址列表失败', error)
            return []
        }
    }

    // 新增地址
    async function CreateAddress(data: AddressFormData): Promise<boolean> {
        try {
            await CreateAddressApi(data)
            await GetAddressList()
            return true
        } catch (error) {
            console.error('新增地址失败', error)
            return false
        }
    }

    // 更新地址
    async function UpdateAddress(id: number, data: Partial<AddressFormData>): Promise<boolean> {
        try {
            await UpdateAddressApi(id, data)
            await GetAddressList()
            return true
        } catch (error) {
            console.error('更新地址失败', error)
            return false
        }
    }

    // 删除地址
    async function DeleteAddress(id: number): Promise<boolean> {
        try {
            await DeleteAddressApi(id)
            await GetAddressList()
            return true
        } catch (error) {
            console.error('删除地址失败', error)
            return false
        }
    }

    // 设为默认
    async function SetDefaultAddress(id: number): Promise<boolean> {
        try {
            await SetDefaultAddressApi(id)
            await GetAddressList()
            return true
        } catch (error) {
            console.error('设置默认地址失败', error)
            return false
        }
    }

    return {
        addressList,

        GetAddressList,
        CreateAddress,
        UpdateAddress,
        DeleteAddress,
        SetDefaultAddress,
    }
})
