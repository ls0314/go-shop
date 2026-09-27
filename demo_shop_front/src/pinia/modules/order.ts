import { defineStore } from 'pinia'
import { ref } from 'vue'
import type {
    CreateOrderReq,
    OrderListItem,
    OrderDetailResp,
    AdminOrderItem,
    AdminOrderDetailResp,
    OrderShipReq,
} from '@/types/order'
import {
    CreateOrderApi,
    GetUserOrderListApi,
    GetUserOrderDetailApi,
    CancelOrderApi,
    ConfirmOrderApi,
    GetAdminOrderListApi,
    GetAdminOrderDetailApi,
    ShipOrderApi,
} from '@/api/order'

export const useOrderStore = defineStore('order', () => {
    // ==================== 用户端状态 ====================

    const userOrderList = ref<OrderListItem[]>([])
    const userOrderTotal = ref(0)
    const userOrderDetail = ref<OrderDetailResp | null>(null)

    // ==================== 管理端状态 ====================

    const adminOrderList = ref<AdminOrderItem[]>([])
    const adminOrderTotal = ref(0)
    const adminOrderDetail = ref<AdminOrderDetailResp | null>(null)

    // ==================== 创建订单 ====================

    async function CreateOrder(data: CreateOrderReq) {
        const res = await CreateOrderApi(data)
        return res.data.data
    }

    // ==================== 用户订单列表 ====================

    async function GetUserOrderList(params?: { page?: number; page_size?: number; order_status?: string }) {
        try {
            const res = await GetUserOrderListApi(params)
            const payload = res.data.data
            userOrderList.value = payload?.list || []
            userOrderTotal.value = payload?.total || 0
            return payload
        } catch (error) {
            console.error('获取用户订单列表失败', error)
            return null
        }
    }

    // ==================== 用户订单详情 ====================

    async function GetUserOrderDetail(id: number): Promise<OrderDetailResp | null> {
        try {
            const res = await GetUserOrderDetailApi(id)
            userOrderDetail.value = res.data.data as OrderDetailResp
            return userOrderDetail.value
        } catch (error) {
            console.error('获取订单详情失败', error)
            return null
        }
    }

    // ==================== 取消订单 ====================

    async function CancelOrder(id: number): Promise<boolean> {
        try {
            await CancelOrderApi(id)
            return true
        } catch (error) {
            console.error('取消订单失败', error)
            return false
        }
    }

    // ==================== 确认收货 ====================

    async function ConfirmOrder(id: number): Promise<boolean> {
        try {
            await ConfirmOrderApi(id)
            return true
        } catch (error) {
            console.error('确认收货失败', error)
            return false
        }
    }

    // ==================== 管理端订单列表 ====================

    async function GetAdminOrderList(params?: {
        page?: number
        page_size?: number
        order_status?: string
        order_no?: string
        start_time?: string
        end_time?: string
    }) {
        try {
            const res = await GetAdminOrderListApi(params)
            const payload = res.data.data
            adminOrderList.value = payload?.list || []
            adminOrderTotal.value = payload?.total || 0
            return payload
        } catch (error) {
            console.error('获取管理端订单列表失败', error)
            return null
        }
    }

    // ==================== 管理端订单详情 ====================

    async function GetAdminOrderDetail(id: number): Promise<AdminOrderDetailResp | null> {
        try {
            const res = await GetAdminOrderDetailApi(id)
            adminOrderDetail.value = res.data.data as AdminOrderDetailResp
            return adminOrderDetail.value
        } catch (error) {
            console.error('获取管理端订单详情失败', error)
            return null
        }
    }

    // ==================== 发货 ====================

    async function ShipOrder(id: number, data: OrderShipReq): Promise<boolean> {
        try {
            await ShipOrderApi(id, data)
            return true
        } catch (error) {
            console.error('发货失败', error)
            return false
        }
    }

    return {
        userOrderList,
        userOrderTotal,
        userOrderDetail,
        adminOrderList,
        adminOrderTotal,
        adminOrderDetail,

        CreateOrder,
        GetUserOrderList,
        GetUserOrderDetail,
        CancelOrder,
        ConfirmOrder,
        GetAdminOrderList,
        GetAdminOrderDetail,
        ShipOrder,
    }
})
