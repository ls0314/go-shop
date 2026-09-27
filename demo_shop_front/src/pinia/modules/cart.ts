import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { CartItem, AddCartReq, UpdateCartReq, CartPayPreview } from '@/types/cart'
import {
    AddCartApi,
    GetCartListApi,
    GetCartCountApi,
    GetCartPreviewApi,
    UpdateCartApi,
    SelectAllCartApi,
    DeleteCartApi,
} from '@/api/cart'

export const useCartStore = defineStore('cart', () => {
    // ==================== 状态 ====================

    const cartList = ref<CartItem[]>([])
    const cartCount = ref(0)

    // ==================== 加入购物车 ====================

    async function AddCart(data: AddCartReq): Promise<boolean> {
        try {
            await AddCartApi({ sku_id: data.sku_id, quantity: data.quantity || 1 })
            return true
        } catch (error) {
            console.error('加入购物车失败', error)
            return false
        }
    }

    // ==================== 获取列表 ====================

    async function GetCartList(): Promise<CartItem[]> {
        try {
            const res = await GetCartListApi()
            cartList.value = res.data.data || []
            return cartList.value
        } catch (error) {
            console.error('获取购物车列表失败', error)
            return []
        }
    }

    // ==================== 获取角标数量 ====================

    async function GetCartCount(): Promise<number> {
        try {
            const res = await GetCartCountApi()
            cartCount.value = res.data.data?.count || 0
            return cartCount.value
        } catch (error) {
            console.error('获取购物车数量失败', error)
            return 0
        }
    }

    // ==================== 结算预览 ====================

    async function GetCartPreview(): Promise<CartPayPreview | null> {
        try {
            const res = await GetCartPreviewApi()
            return res.data.data as CartPayPreview
        } catch (error) {
            console.error('结算预览失败', error)
            return null
        }
    }

    // ==================== 更新 ====================

    async function UpdateCart(id: number, data: UpdateCartReq): Promise<boolean> {
        try {
            await UpdateCartApi(id, data)
            // 刷新列表
            await GetCartList()
            await GetCartCount()
            return true
        } catch (error) {
            console.error('更新购物车失败', error)
            return false
        }
    }

    // ==================== 全选/取消全选 ====================

    async function SelectAllCart(isSelected: boolean): Promise<boolean> {
        try {
            await SelectAllCartApi({ is_selected: isSelected })
            await GetCartList()
            return true
        } catch (error) {
            console.error('全选操作失败', error)
            return false
        }
    }

    // ==================== 删除 ====================

    async function DeleteCart(id: number): Promise<boolean> {
        try {
            await DeleteCartApi(id)
            await GetCartList()
            await GetCartCount()
            return true
        } catch (error) {
            console.error('删除购物车项失败', error)
            return false
        }
    }

    return {
        cartList,
        cartCount,

        AddCart,
        GetCartList,
        GetCartCount,
        GetCartPreview,
        UpdateCart,
        SelectAllCart,
        DeleteCart,
    }
})
