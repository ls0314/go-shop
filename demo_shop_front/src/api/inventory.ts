import service from '@/utils/request'
import type { InventoryLogQuery, InventoryAdjustReq, InventoryWarnQuery } from '@/types/inventory'

// 查询单个SKU库存
export const GetSkuStockApi = (skuId: number) => {
    return service({
        url: `/admin/inventory/sku/${skuId}`,
        method: 'get',
    })
}

// 查询SPU下所有SKU库存
export const GetSkuListBySpuApi = (spuId: number) => {
    return service({
        url: `/admin/inventory/spu/${spuId}`,
        method: 'get',
    })
}

// 手动调整库存
export const AdjustStockApi = (data: InventoryAdjustReq) => {
    return service({
        url: '/admin/inventory/adjust',
        method: 'post',
        data,
    })
}

// 查询库存变更日志
export const GetStockLogApi = (params: InventoryLogQuery = {}) => {
    return service({
        url: '/admin/inventory/log',
        method: 'get',
        params,
    })
}

// 查询低库存预警列表
export const GetWarnStockApi = (params: InventoryWarnQuery = {}) => {
    return service({
        url: '/admin/inventory/warning',
        method: 'get',
        params,
    })
}
