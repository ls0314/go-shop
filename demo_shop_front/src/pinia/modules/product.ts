import { defineStore } from 'pinia'
import { ref } from 'vue'
import type {
    CreateProductReq,
    SpuQueryReq,
    UpdateProductReq,
    UpdateProductFullReq,
    GetSpuListResp,
    GetProductResp,
    SpuListItem,
} from "@/types/product"
import {
    CreateProductApi,
    GetProductListApi,
    GetProductApi,
    UpdateProductApi,
    UpdateProductFullApi,
    DeleteProductApi,
    PublishProductApi,
    WithdrawProductApi,
} from '@/api/product'

export const useProductStore = defineStore('product', () => {
    // ==================== 状态 ====================

    /** 商品列表 */
    const productList = ref<SpuListItem[]>([])
    /** 列表总数 */
    const productTotal = ref(0)
    /** 当前查看的商品详情 */
    const currentProduct = ref<GetProductResp | null>(null)

    // ==================== 创建 ====================

    /** 创建商品（含 SKU + 图片） */
    async function CreateProduct(data: CreateProductReq): Promise<{ spu_id: number } | null> {
        try {
            const res = await CreateProductApi(data)
            return res.data.data as { spu_id: number }
        } catch (error) {
            console.error('创建商品失败', error)
            return null
        }
    }

    // ==================== 查询列表 ====================

    /** 获取管理端商品列表（分页 + 筛选） */
    async function GetProductList(params?: SpuQueryReq): Promise<GetSpuListResp | null> {
        try {
            const res = await GetProductListApi(params)
            const data = res.data.data as GetSpuListResp
            productList.value = data.list
            productTotal.value = data.total
            return data
        } catch (error) {
            console.error('获取商品列表失败', error)
            return null
        }
    }

    // ==================== 查询详情 ====================

    /** 获取管理端商品详情 */
    async function GetProduct(id: number): Promise<GetProductResp | null> {
        try {
            const res = await GetProductApi(id)
            const data = res.data.data as GetProductResp
            currentProduct.value = data
            return data
        } catch (error) {
            console.error('获取商品详情失败', error)
            return null
        }
    }

    // ==================== 更新 ====================

    /** 部分更新商品基本信息 */
    async function UpdateProduct(id: number, data: UpdateProductReq): Promise<GetProductResp | null> {
        try {
            const res = await UpdateProductApi(id, data)
            const updated = res.data.data as GetProductResp
            currentProduct.value = updated
            return updated
        } catch (error) {
            console.error('更新商品失败', error)
            return null
        }
    }

    /** 全量更新商品（含 SKU + 图片增改删） */
    async function UpdateProductFull(id: number, data: UpdateProductFullReq): Promise<GetProductResp | null> {
        try {
            const res = await UpdateProductFullApi(id, data)
            const updated = res.data.data as GetProductResp
            currentProduct.value = updated
            return updated
        } catch (error) {
            console.error('全量更新商品失败', error)
            return null
        }
    }

    // ==================== 删除 ====================

    /** 软删除商品 */
    async function DeleteProduct(id: number): Promise<boolean> {
        try {
            await DeleteProductApi(id)
            if (currentProduct.value && currentProduct.value.spu_id === id) {
                currentProduct.value = null
            }
            return true
        } catch (error) {
            console.error('删除商品失败', error)
            return false
        }
    }

    // ==================== 上下架 ====================

    /** 上架商品 */
    async function PublishProduct(id: number): Promise<boolean> {
        try {
            await PublishProductApi(id)
            return true
        } catch (error) {
            console.error('上架商品失败', error)
            return false
        }
    }

    /** 下架商品 */
    async function WithdrawProduct(id: number): Promise<boolean> {
        try {
            await WithdrawProductApi(id)
            return true
        } catch (error) {
            console.error('下架商品失败', error)
            return false
        }
    }

    // ==================== 工具 ====================

    /** 清空当前商品 */
    function resetCurrentProduct() {
        currentProduct.value = null
    }

    // ==================== 导出 ====================

    return {
        // 状态
        productList,
        productTotal,
        currentProduct,

        // 方法
        CreateProduct,
        GetProductList,
        GetProduct,
        UpdateProduct,
        UpdateProductFull,
        DeleteProduct,
        PublishProduct,
        WithdrawProduct,
        resetCurrentProduct,
    }
})
