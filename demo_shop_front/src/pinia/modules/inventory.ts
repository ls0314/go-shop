import { defineStore } from 'pinia'
import { ref } from 'vue'
import type {
    SkuInventory,
    SkuInventoryList,
    InventoryAdjustReq,
    InventoryAdjustResp,
    InventoryLogQuery,
    InventoryLogItem,
    InventoryLogListResp,
    InventoryWarnQuery,
    InventoryWarnItem,
} from '@/types/inventory'
import {
    GetSkuStockApi,
    GetSkuListBySpuApi,
    AdjustStockApi,
    GetStockLogApi,
    GetWarnStockApi,
} from '@/api/inventory'

export const useInventoryStore = defineStore('inventory', () => {
    // ==================== 状态 ====================

    /** 单个SKU库存详情 */
    const currentSkuStock = ref<SkuInventory | null>(null)
    /** SPU下SKU库存列表 */
    const skuListBySpu = ref<SkuInventory[]>([])
    /** SPU库存汇总 */
    const spuStockSummary = ref<Omit<SkuInventoryList, 'list'> | null>(null)
    /** 库存变更日志列表 */
    const stockLogList = ref<InventoryLogItem[]>([])
    /** 日志总数 */
    const stockLogTotal = ref(0)
    /** 低库存预警列表 */
    const warnStockList = ref<InventoryWarnItem[]>([])

    // ==================== 查询 ====================

    /** 查询单个SKU库存 */
    async function GetSkuStock(skuId: number): Promise<SkuInventory | null> {
        try {
            const res = await GetSkuStockApi(skuId)
            const data = res.data.data as SkuInventory
            currentSkuStock.value = data
            return data
        } catch (error) {
            console.error('查询SKU库存失败', error)
            return null
        }
    }

    /** 查询SPU下所有SKU库存 */
    async function GetSkuListBySpu(spuId: number): Promise<SkuInventoryList | null> {
        try {
            const res = await GetSkuListBySpuApi(spuId)
            const data = res.data.data as SkuInventoryList
            skuListBySpu.value = data.list
            spuStockSummary.value = {
                total_stock: data.total_stock,
                total_lock: data.total_lock,
                total_sold: data.total_sold,
            }
            return data
        } catch (error) {
            console.error('查询SPU库存列表失败', error)
            return null
        }
    }

    // ==================== 调整 ====================

    /** 手动调整库存 */
    async function AdjustStock(data: InventoryAdjustReq): Promise<InventoryAdjustResp | null> {
        try {
            const res = await AdjustStockApi(data)
            return res.data.data as InventoryAdjustResp
        } catch (error) {
            console.error('调整库存失败', error)
            return null
        }
    }

    // ==================== 日志 ====================

    /** 查询库存变更日志 */
    async function GetStockLog(params?: InventoryLogQuery): Promise<InventoryLogListResp | null> {
        try {
            const res = await GetStockLogApi(params)
            const data = res.data.data as InventoryLogListResp
            stockLogList.value = data.list
            stockLogTotal.value = data.total
            return data
        } catch (error) {
            console.error('查询库存变更日志失败', error)
            return null
        }
    }

    // ==================== 预警 ====================

    /** 查询低库存预警列表 */
    async function GetWarnStock(params?: InventoryWarnQuery): Promise<InventoryWarnItem[] | null> {
        try {
            const res = await GetWarnStockApi(params)
            const data = res.data.data as InventoryWarnItem[]
            warnStockList.value = data
            return data
        } catch (error) {
            console.error('查询低库存预警失败', error)
            return null
        }
    }

    // ==================== 导出 ====================

    return {
        currentSkuStock,
        skuListBySpu,
        spuStockSummary,
        stockLogList,
        stockLogTotal,
        warnStockList,

        GetSkuStock,
        GetSkuListBySpu,
        AdjustStock,
        GetStockLog,
        GetWarnStock,
    }
})
