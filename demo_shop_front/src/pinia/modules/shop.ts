import { defineStore } from 'pinia'
import { ref } from 'vue'
import type {
    UserSpuListItem,
    UserGetSpuListResp,
    UserGetProductResp,
    SpuQueryReq,
} from '@/types/product'
import {
    UserGetProductListApi,
    UserGetProductApi,
} from '@/api/product'

export const useShopStore = defineStore('shop', () => {
    // ==================== 状态 ====================

    /** 商品列表 */
    const productList = ref<UserSpuListItem[]>([])
    /** 列表总数 */
    const productTotal = ref(0)
    /** 当前商品详情 */
    const currentProduct = ref<UserGetProductResp | null>(null)

    // ==================== 商品列表 ====================

    /** 获取用户端商品列表（分页 + 筛选） */
    async function GetProductList(params?: SpuQueryReq): Promise<UserGetSpuListResp | null> {
        try {
            const res = await UserGetProductListApi(params)
            const data = res.data.data as UserGetSpuListResp
            productList.value = data.list
            productTotal.value = data.total
            return data
        } catch (error) {
            console.error('获取商品列表失败', error)
            return null
        }
    }

    // ==================== 商品详情 ====================

    /** 获取用户端商品详情 */
    async function GetProduct(id: number): Promise<UserGetProductResp | null> {
        try {
            const res = await UserGetProductApi(id)
            const data = res.data.data as UserGetProductResp
            currentProduct.value = data
            return data
        } catch (error) {
            console.error('获取商品详情失败', error)
            return null
        }
    }

    // ==================== 工具 ====================

    /** 清空当前商品 */
    function resetCurrentProduct() {
        currentProduct.value = null
    }

    return {
        productList,
        productTotal,
        currentProduct,

        GetProductList,
        GetProduct,
        resetCurrentProduct,
    }
})
