<template>
  <div class="flex h-full flex-col">
    <h1 class="text-2xl font-bold text-slate-800 dark:text-white">收银台</h1>

    <div v-if="loading" class="flex flex-1 items-center justify-center">
      <el-icon class="animate-spin text-3xl text-slate-400"><Loading /></el-icon>
    </div>

    <template v-else-if="orderInfo">
      <!-- 支付金额 -->
      <div class="mt-6 rounded-xl border border-slate-200 bg-white p-6 text-center dark:border-slate-700 dark:bg-slate-800">
        <p class="text-sm text-slate-500 dark:text-slate-400">应付金额</p>
        <p class="mt-2 text-4xl font-bold" style="color: #ff6700">¥{{ orderInfo.pay_amount?.toFixed(2) }}</p>
        <p class="mt-1 text-xs text-slate-400">订单号：{{ orderInfo.order_no }}</p>
        <p v-if="!paySuccess" class="mt-2 text-sm text-amber-500">
          <el-icon class="align-middle"><Clock /></el-icon>
          {{ countdownText }} 后订单将自动取消
        </p>
      </div>

      <!-- 支付成功 -->
      <div v-if="paySuccess" class="mt-6 rounded-xl border border-green-200 bg-green-50 p-6 text-center dark:border-green-800 dark:bg-green-900/20">
        <p class="text-4xl">✅</p>
        <p class="mt-3 text-lg font-semibold text-green-700 dark:text-green-400">支付成功</p>
        <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">交易号：{{ tradeNo || '已生成' }}</p>
        <div class="mt-4 flex justify-center gap-3">
          <el-button size="large" class="h-10 rounded-xl" @click="$router.push(`/shop/order/${orderId}`)">查看订单</el-button>
          <el-button size="large" class="h-10 rounded-xl" type="primary" style="background:#ff6700;border-color:#ff6700" @click="$router.push('/shop/home')">继续购物</el-button>
        </div>
      </div>

      <!-- 支付方式 -->
      <div v-else class="mt-4 rounded-xl border border-slate-200 bg-white p-5 dark:border-slate-700 dark:bg-slate-800">
        <h3 class="text-base font-semibold text-slate-800 dark:text-white">选择支付方式</h3>

        <div class="mt-3 space-y-2">
          <div
            class="flex cursor-pointer items-center gap-4 rounded-lg border p-4 transition"
            :class="selectedMethod === 'mock'
              ? 'border-orange-400 bg-orange-50 dark:border-orange-500 dark:bg-orange-900/20'
              : 'border-slate-100 hover:border-slate-300 dark:border-slate-700'"
            @click="selectedMethod = 'mock'"
          >
            <div class="flex h-10 w-14 items-center justify-center rounded-lg bg-purple-100 text-purple-600 text-sm font-bold">MOCK</div>
            <div class="flex-1">
              <p class="font-medium text-slate-800 dark:text-white">模拟支付</p>
              <p class="text-xs text-slate-400">开发测试用，立即完成支付</p>
            </div>
            <el-icon v-if="selectedMethod === 'mock'" size="20" color="#ff6700"><CircleCheck /></el-icon>
          </div>
        </div>
      </div>

      <!-- 支付按钮 -->
      <div v-if="!paySuccess" class="mt-6">
        <el-button
          type="primary"
          size="large"
          :loading="paying"
          class="h-12 w-full rounded-xl text-base font-semibold"
          style="background:#ff6700;border-color:#ff6700"
          @click="handlePay"
        >
          {{ paying ? '支付处理中...' : `立即支付 ¥${orderInfo.pay_amount?.toFixed(2)}` }}
        </el-button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Clock, CircleCheck, Loading } from '@element-plus/icons-vue'
import { useOrderStore } from '@/pinia/modules/order'
import { CreatePaymentApi, MockPayCallbackApi } from '@/api/payment'
import type { OrderDetailResp } from '@/types/order'

const route = useRoute()
const orderStore = useOrderStore()

const orderId = Number(route.params.orderId)
const loading = ref(false)
const paying = ref(false)
const paySuccess = ref(false)
const selectedMethod = ref('mock')
const orderInfo = ref<OrderDetailResp | null>(null)
const tradeNo = ref('')
const countdownText = ref('')

let countdownTimer: ReturnType<typeof setInterval> | null = null

async function fetchOrder() {
  loading.value = true
  try { orderInfo.value = await orderStore.GetUserOrderDetail(orderId) }
  finally { loading.value = false }
}

function startCountdown() {
  updateCountdown()
  countdownTimer = setInterval(updateCountdown, 1000)
}

function updateCountdown() {
  const createdAt = new Date(orderInfo.value.created_at).getTime() - 8 * 3600 * 1000
  // 带Z/时区标记的时间会被Date正确解析，无需手动补偿
  // 仅当确定服务端返回北京时间且无时区标记时，才需要修正（注意是减8小时）
  // if (!/[Z]|[+-]\d{2}:\d{2}/.test(orderInfo.value.created_at)) {
  //   createdAt -= 8 * 3600 * 1000
  // }
  const expireAt = createdAt + 15 * 60 * 1000
  const remaining = Math.max(0, expireAt - Date.now())

  if (remaining <= 0) {
    countdownText.value = '已超时'
    clearInterval(countdownTimer) // 超时后停止定时器
    return
  }

  const m = Math.floor(remaining / 60000)
  const s = Math.floor((remaining % 60000) / 1000)
  countdownText.value = `${m}分${String(s).padStart(2, '0')}秒`
}

async function handlePay() {
  if (!orderInfo.value || orderInfo.value.order_status !== 'pending_pay') {
    ElMessage.error('当前订单状态无法支付')
    return
  }
  paying.value = true
  try {
    // 1. 发起支付
    const payRes = await CreatePaymentApi(orderId, { pay_method: selectedMethod.value })
    const payData = payRes.data.data
    if (!payData?.pay_no) { ElMessage.error('发起支付失败'); return }

    // 2. Mock 回调完成支付
    await MockPayCallbackApi(payData.pay_no)
    tradeNo.value = 'MOCK' + Date.now()
    paySuccess.value = true
    ElMessage.success('支付成功')
  } catch {
    ElMessage.error('支付失败，请重试')
  } finally {
    paying.value = false
  }
}

onMounted(async () => {
  await fetchOrder()
  if (orderInfo.value?.order_status === 'pending_pay') startCountdown()
})

onBeforeUnmount(() => { if (countdownTimer) clearInterval(countdownTimer) })
</script>
