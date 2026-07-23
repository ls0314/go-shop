<template>
  <div class="flex h-full flex-col">
    <h1 class="text-2xl font-bold text-slate-800 dark:text-white">我的订单</h1>

    <!-- Tab 栏 -->
    <div class="mt-6 flex gap-0 border-b border-slate-200 dark:border-slate-700">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        class="relative px-6 pb-3 text-sm font-medium transition"
        :class="activeTab === tab.key
          ? 'text-slate-800 dark:text-white'
          : 'text-slate-400 hover:text-slate-600 dark:hover:text-slate-300'"
        @click="switchTab(tab.key)"
      >
        {{ tab.label }}
        <span
          v-if="activeTab === tab.key"
          class="absolute bottom-0 left-1/2 h-0.5 w-8 -translate-x-1/2 rounded-full bg-slate-800 dark:bg-white"
        />
      </button>
    </div>

    <!-- 加载 -->
    <div v-if="loading" class="flex flex-1 items-center justify-center py-20">
      <el-icon class="animate-spin text-3xl text-slate-400"><Loading /></el-icon>
    </div>

    <!-- 空状态 -->
    <div v-else-if="orderList.length === 0" class="flex flex-1 items-center justify-center py-20">
      <div class="text-center">
        <p class="text-5xl mb-3">📦</p>
        <p class="text-slate-400">暂无订单</p>
        <el-button
          class="mt-4 h-9 rounded-xl"
          style="background:#ff6700;border-color:#ff6700"
          type="primary"
          @click="$router.push('/shop/home')"
        >
          去逛逛
        </el-button>
      </div>
    </div>

    <!-- 订单列表 -->
    <div v-else class="mt-4 flex-1 space-y-4 overflow-auto pb-4">
      <div
        v-for="order in orderList"
        :key="order.order_id"
        class="rounded-xl border border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800"
      >
        <!-- 订单头 -->
        <div class="flex items-center justify-between border-b border-slate-100 px-5 py-3 dark:border-slate-700">
          <span class="text-xs text-slate-400">{{ order.created_at }}</span>
          <span class="text-xs" :class="statusColor(order.order_status)">{{ statusLabel(order.order_status) }}</span>
        </div>

        <!-- 商品概览 -->
        <div class="flex items-center gap-4 p-4" @click="$router.push(`/shop/order/${order.order_id}`)">
          <div class="h-16 w-16 flex-shrink-0 overflow-hidden rounded-lg bg-slate-100 dark:bg-slate-700">
            <img v-if="order.first_image" :src="order.first_image" class="h-full w-full object-cover" />
            <div v-else class="flex h-full w-full items-center justify-center text-slate-400">
              <el-icon :size="18"><Picture /></el-icon>
            </div>
          </div>
          <div class="flex-1 min-w-0">
            <p class="text-sm text-slate-500 dark:text-slate-400">共 {{ order.detail_count }} 件商品</p>
          </div>
          <div class="text-right">
            <p class="text-sm font-bold text-slate-800 dark:text-white">¥{{ order.pay_amount?.toFixed(2) }}</p>
          </div>
        </div>

        <!-- 操作栏 -->
        <div class="flex justify-end gap-2 border-t border-slate-100 px-5 py-3 dark:border-slate-700">
          <el-button
            size="small"
            class="h-8 rounded-lg text-xs"
            @click="$router.push(`/shop/order/${order.order_id}`)"
          >
            查看详情
          </el-button>
          <el-button
            v-if="order.order_status === 'pending_pay'"
            size="small"
            type="danger"
            class="h-8 rounded-lg text-xs"
            @click.stop="handleCancel(order.order_id)"
          >
            取消订单
          </el-button>
          <el-button
            v-if="order.order_status === 'shipped'"
            size="small"
            class="h-8 rounded-lg text-xs"
            style="background:#ff6700;border-color:#ff6700;color:#fff"
            @click.stop="handleConfirm(order.order_id)"
          >
            确认收货
          </el-button>
        </div>
      </div>
    </div>

    <!-- 分页 -->
    <div v-if="total > 0" class="flex justify-center py-4">
      <el-pagination
        v-model:current-page="queryParams.page"
        v-model:page-size="queryParams.page_size"
        :total="total"
        :page-sizes="[10, 20]"
        layout="total, prev, pager, next"
        background
        small
        @size-change="fetchList"
        @current-change="fetchList"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Picture, Loading } from '@element-plus/icons-vue'
import { useOrderStore } from '@/pinia/modules/order'
import type { OrderListItem } from '@/types/order'

const orderStore = useOrderStore()

const loading = ref(false)
const activeTab = ref('')
const orderList = ref<OrderListItem[]>([])
const total = ref(0)
const queryParams = ref({ page: 1, page_size: 10 })

const tabs = [
  { key: '', label: '全部订单' },
  { key: 'pending_pay', label: '待支付' },
  { key: 'shipped', label: '待收货' },
]

function statusLabel(status: string) {
  const map: Record<string, string> = {
    pending_pay: '待支付', paid: '待发货', shipped: '待收货', completed: '已完成', cancelled: '已取消',
  }
  return map[status] || status
}

function statusColor(status: string) {
  const map: Record<string, string> = { pending_pay: 'text-amber-500', paid: 'text-blue-500', shipped: 'text-green-500', completed: 'text-slate-400', cancelled: 'text-slate-400' }
  return map[status] || ''
}

async function fetchList() {
  loading.value = true
  try {
    const res = await orderStore.GetUserOrderList({
      page: queryParams.value.page,
      page_size: queryParams.value.page_size,
      order_status: activeTab.value || undefined,
    })
    if (res) { orderList.value = res.list || []; total.value = res.total || 0 }
  } finally { loading.value = false }
}

function switchTab(key: string) {
  activeTab.value = key
  queryParams.value.page = 1
  fetchList()
}

async function handleCancel(orderId: number) {
  try {
    await ElMessageBox.confirm('确定取消该订单？', '提示', { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' })
    const ok = await orderStore.CancelOrder(orderId)
    ok ? (ElMessage.success('已取消'), fetchList()) : ElMessage.error('取消失败')
  } catch { /* 取消 */ }
}

async function handleConfirm(orderId: number) {
  try {
    await ElMessageBox.confirm('确认已收到商品？', '提示', { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' })
    const ok = await orderStore.ConfirmOrder(orderId)
    ok ? (ElMessage.success('已确认收货'), fetchList()) : ElMessage.error('确认失败')
  } catch { /* 取消 */ }
}

onMounted(() => fetchList())
</script>
