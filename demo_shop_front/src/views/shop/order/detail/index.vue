<template>
  <div class="flex h-full flex-col">
    <!-- 返回 + 标题 -->
    <div class="flex items-center gap-4">
      <el-button class="h-9 rounded-xl" @click="$router.back()">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <h1 class="text-2xl font-bold text-slate-800 dark:text-white">订单详情</h1>
    </div>

    <!-- 加载 -->
    <div v-if="loading" class="flex flex-1 items-center justify-center py-20">
      <el-icon class="animate-spin text-3xl text-slate-400"><Loading /></el-icon>
    </div>

    <template v-else-if="order">
      <!-- 订单状态卡片 -->
      <div class="mt-6 rounded-xl border border-slate-200 bg-white p-6 dark:border-slate-700 dark:bg-slate-800">
        <div class="flex items-center justify-between">
          <div>
            <span class="text-sm text-slate-500 dark:text-slate-400">订单状态</span>
            <el-tag
              :type="statusTagType(order.order_status)"
              size="large"
              class="ml-3 rounded-lg"
            >
              {{ statusLabel(order.order_status) }}
            </el-tag>
          </div>
          <div class="flex gap-2">
            <el-button
              v-if="order.order_status === 'pending_pay'"
              type="danger"
              class="h-9 rounded-xl"
              @click="handleCancel"
            >
              取消订单
            </el-button>
            <el-button
              v-if="order.order_status === 'shipped'"
              type="primary"
              class="h-9 rounded-xl"
              style="background-color: #ff6700; border-color: #ff6700"
              @click="handleConfirm"
            >
              确认收货
            </el-button>
          </div>
        </div>
        <p class="mt-2 text-sm text-slate-500 dark:text-slate-400">
          订单号：{{ order.order_no }} ｜ 下单时间：{{ order.created_at }}
        </p>
      </div>

      <!-- 收货地址 -->
      <div class="mt-4 rounded-xl border border-slate-200 bg-white p-5 dark:border-slate-700 dark:bg-slate-800">
        <h3 class="text-base font-semibold text-slate-800 dark:text-white">收货信息</h3>
        <div class="mt-2 text-sm text-slate-600 dark:text-slate-300">
          <p>
            {{ order.address_snapshot?.receiver_name }}
            {{ order.address_snapshot?.receiver_phone }}
          </p>
          <p class="mt-1">
            {{ order.address_snapshot?.province }}
            {{ order.address_snapshot?.city }}
            {{ order.address_snapshot?.district }}
            {{ order.address_snapshot?.detail_address }}
          </p>
          <p v-if="order.buyer_remark" class="mt-2 text-slate-400">
            备注：{{ order.buyer_remark }}
          </p>
        </div>
      </div>

      <!-- 商品明细 -->
      <div class="mt-4 rounded-xl border border-slate-200 bg-white p-5 dark:border-slate-700 dark:bg-slate-800">
        <h3 class="text-base font-semibold text-slate-800 dark:text-white">商品明细</h3>
        <div class="mt-3 space-y-3">
          <div
            v-for="item in order.detail_list"
            :key="item.detail_id"
            class="flex items-center gap-4 rounded-lg bg-slate-50 p-3 dark:bg-slate-700/50"
          >
            <div class="h-16 w-16 flex-shrink-0 overflow-hidden rounded-lg bg-slate-200 dark:bg-slate-600">
              <img v-if="item.main_image" :src="item.main_image" class="h-full w-full object-cover" />
              <div v-else class="flex h-full w-full items-center justify-center text-slate-400">
                <el-icon :size="20"><Picture /></el-icon>
              </div>
            </div>
            <div class="flex-1">
              <p class="text-sm font-medium text-slate-800 dark:text-white">{{ item.spu_name }}</p>
              <p class="text-xs text-slate-500 dark:text-slate-400">{{ item.sku_name }}</p>
              <p class="mt-1 text-sm font-bold" style="color: #ff6700">¥{{ item.unit_price.toFixed(2) }}</p>
            </div>
            <div class="text-right">
              <p class="text-sm text-slate-600 dark:text-slate-300">×{{ item.quantity }}</p>
              <p class="text-sm font-bold text-slate-800 dark:text-white">¥{{ item.total_price.toFixed(2) }}</p>
            </div>
          </div>
        </div>
        <div class="mt-4 border-t border-slate-100 pt-3 text-right dark:border-slate-700">
          <p class="text-sm text-slate-500 dark:text-slate-400">
            共 {{ order.detail_list?.length || 0 }} 件商品，合计
            <span class="text-xl font-bold" style="color: #ff6700">¥{{ order.pay_amount.toFixed(2) }}</span>
          </p>
        </div>
      </div>

      <!-- 订单日志 -->
      <div class="mt-4 rounded-xl border border-slate-200 bg-white p-5 dark:border-slate-700 dark:bg-slate-800">
        <h3 class="text-base font-semibold text-slate-800 dark:text-white">订单日志</h3>
        <div class="mt-3 space-y-2">
          <div
            v-for="log in order.log_list"
            :key="log.log_id"
            class="flex items-center gap-3 text-sm text-slate-600 dark:text-slate-300"
          >
            <span class="text-slate-400">{{ log.created_at }}</span>
            <span>｜</span>
            <span>{{ logActionLabel(log.action) }}</span>
            <span v-if="log.detail" class="text-slate-400">（{{ log.detail }}）</span>
          </div>
          <div v-if="!order.log_list?.length" class="text-sm text-slate-400">暂无日志</div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Picture, Loading } from '@element-plus/icons-vue'
import { useOrderStore } from '@/pinia/modules/order'
import type { OrderDetailResp } from '@/types/order'

const route = useRoute()
const orderStore = useOrderStore()

const loading = ref(false)
const order = ref<OrderDetailResp | null>(null)

function statusTagType(status: string) {
  const map: Record<string, string> = {
    pending_pay: 'warning',
    paid: 'primary',
    shipped: 'success',
    completed: 'info',
    cancelled: 'danger',
  }
  return map[status] || 'info'
}

function statusLabel(status: string) {
  const map: Record<string, string> = {
    pending_pay: '待支付',
    paid: '待发货',
    shipped: '待收货',
    completed: '已完成',
    cancelled: '已取消',
  }
  return map[status] || status
}

function logActionLabel(action: string) {
  const map: Record<string, string> = {
    create: '创建订单',
    pay: '支付成功',
    cancel: '取消订单',
    ship: '商家发货',
    confirm: '确认收货',
    auto_cancel: '系统自动取消',
  }
  return map[action] || action
}

async function fetchDetail() {
  loading.value = true
  try {
    const id = Number(route.params.id)
    order.value = await orderStore.GetUserOrderDetail(id)
  } finally {
    loading.value = false
  }
}

async function handleCancel() {
  if (!order.value) return
  try {
    await ElMessageBox.confirm('确定要取消该订单吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    const ok = await orderStore.CancelOrder(order.value.order_id)
    if (ok) {
      ElMessage.success('订单已取消')
      fetchDetail()
    } else {
      ElMessage.error('取消订单失败')
    }
  } catch {
    // 用户取消操作
  }
}

async function handleConfirm() {
  if (!order.value) return
  try {
    await ElMessageBox.confirm('确认已收到商品？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    const ok = await orderStore.ConfirmOrder(order.value.order_id)
    if (ok) {
      ElMessage.success('确认收货成功')
      fetchDetail()
    } else {
      ElMessage.error('确认收货失败')
    }
  } catch {
    // 用户取消操作
  }
}

onMounted(() => fetchDetail())
</script>
