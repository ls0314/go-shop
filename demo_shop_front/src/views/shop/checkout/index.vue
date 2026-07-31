<template>
  <div class="flex h-full flex-col">
    <h1 class="text-2xl font-bold text-slate-800 dark:text-white">确认订单</h1>

    <div v-if="loading" class="flex flex-1 items-center justify-center">
      <el-icon class="animate-spin text-3xl text-slate-400"><Loading /></el-icon>
    </div>

    <template v-else-if="items.length > 0">
      <div class="mt-6 flex-1 space-y-6 overflow-auto pb-6">
        <!-- 收货地址 -->
        <div class="rounded-xl border border-slate-200 bg-white p-5 dark:border-slate-700 dark:bg-slate-800">
          <h3 class="text-base font-semibold text-slate-800 dark:text-white">收货地址</h3>
          <div v-if="addresses.length === 0" class="mt-3 text-sm text-slate-400">
            暂无地址，请先
            <el-button link type="primary" @click="$router.push('/shop/address')">添加地址</el-button>
          </div>
          <div v-else class="mt-3 space-y-2">
            <div
              v-for="addr in addresses"
              :key="addr.address_id"
              class="cursor-pointer rounded-lg border p-3 transition"
              :class="selectedAddressId === addr.address_id
                ? 'border-orange-400 bg-orange-50 dark:border-orange-500 dark:bg-orange-900/20'
                : 'border-slate-100 hover:border-slate-300 dark:border-slate-700'"
              @click="selectedAddressId = addr.address_id"
            >
              <div class="flex items-center justify-between">
                <span class="font-medium text-slate-800 dark:text-white">{{ addr.receiver_name }} {{ addr.receiver_phone }}</span>
                <el-tag v-if="addr.is_default" size="small" type="warning" class="rounded-lg">默认</el-tag>
              </div>
              <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">
                {{ addr.province }}{{ addr.city }}{{ addr.district }} {{ addr.detail_address }}
              </p>
            </div>
          </div>
        </div>

        <!-- 商品列表 -->
        <div class="rounded-xl border border-slate-200 bg-white p-5 dark:border-slate-700 dark:bg-slate-800">
          <h3 class="text-base font-semibold text-slate-800 dark:text-white">商品明细</h3>
          <div class="mt-3 space-y-3">
            <div
              v-for="item in items"
              :key="item.cart_item_id"
              class="flex items-center gap-4 rounded-lg bg-slate-50 p-3 dark:bg-slate-700/50"
            >
              <div class="h-16 w-16 flex-shrink-0 overflow-hidden rounded-lg bg-slate-200">
                <img v-if="item.sku_image || item.main_image" :src="item.sku_image || item.main_image" class="h-full w-full object-cover" />
                <div v-else class="flex h-full w-full items-center justify-center text-slate-400"><el-icon :size="20"><Picture /></el-icon></div>
              </div>
              <div class="flex-1 min-w-0">
                <p class="truncate text-sm font-medium text-slate-800 dark:text-white">{{ item.spu_name }}</p>
                <p class="truncate text-xs text-slate-500 dark:text-slate-400">{{ item.sku_name }}</p>
              </div>
              <span class="text-sm font-bold" style="color: #ff6700">¥{{ item.price.toFixed(2) }}</span>
              <span class="text-sm text-slate-500">×{{ item.quantity }}</span>
              <span class="text-sm font-bold text-slate-800 dark:text-white">¥{{ (item.price * item.quantity).toFixed(2) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 底部结算 -->
      <div class="shrink-0 flex items-center justify-between rounded-xl bg-white px-6 py-4 shadow-lg dark:bg-slate-800">
        <div class="text-sm text-slate-500 dark:text-slate-400">
          共 <b class="text-slate-800 dark:text-white">{{ items.length }}</b> 种商品，
          <b class="text-slate-800 dark:text-white">{{ totalQuantity }}</b> 件
        </div>
        <div class="flex items-center gap-3">
          <span class="text-sm text-slate-600 dark:text-slate-300">合计：</span>
          <span class="text-2xl font-bold" style="color: #ff6700">¥{{ totalAmount.toFixed(2) }}</span>
          <el-button
            type="primary"
            size="large"
            :loading="submitting"
            :disabled="!selectedAddressId"
            class="ml-2 h-11 rounded-xl px-10 text-base"
            style="background: #ff6700; border-color: #ff6700"
            @click="handleSubmit"
          >
            提交订单
          </el-button>
        </div>
      </div>
    </template>

    <!-- 空 -->
    <div v-else class="flex flex-1 items-center justify-center">
      <div class="text-center">
        <p class="text-5xl">📦</p>
        <p class="mt-4 text-slate-400">暂无可结算商品</p>
        <el-button class="mt-4 h-9 rounded-xl" @click="$router.push('/shop/cart')">返回购物车</el-button>
      </div>
    </div>

    <!-- 下单成功弹窗 -->
    <el-dialog v-model="successVisible" title="下单成功" width="400px" custom-class="rounded-xl" :close-on-click-modal="false" :show-close="false">
      <div class="text-center py-4">
        <p class="text-5xl">🎉</p>
        <p class="mt-3 text-slate-600 dark:text-slate-300">订单已创建</p>
	        <p class="mt-3 text-sm text-slate-400">请在 15 分钟内完成支付，超时自动取消</p>
        <p class="text-sm text-slate-400">订单号：{{ createdOrderNo }}</p>
      </div>
      <template #footer>
        <el-button type="success" size="large" class="h-9 rounded-xl" @click="handleGoPay">去支付</el-button>
        <el-button class="h-9 rounded-xl" @click="$router.push('/shop/order/list')">查看订单</el-button>
        <el-button type="primary" class="h-9 rounded-xl" style="background:#ff6700;border-color:#ff6700" @click="$router.push('/shop/home')">继续购物</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Picture, Loading } from '@element-plus/icons-vue'
import { useCartStore } from '@/pinia/modules/cart'
import { useAddressStore } from '@/pinia/modules/address'
import { useOrderStore } from '@/pinia/modules/order'

const cartStore = useCartStore()
const addressStore = useAddressStore()
const orderStore = useOrderStore()
const router = useRouter()

const loading = ref(false)
const submitting = ref(false)
const successVisible = ref(false)
const createdOrderId = ref(0)
const createdOrderNo = ref('')
const selectedAddressId = ref<number | null>(null)

const items = ref<any[]>([])

const totalQuantity = computed(() => items.value.reduce((s, i) => s + i.quantity, 0))
const totalAmount = computed(() => items.value.reduce((s, i) => s + i.price * i.quantity, 0))

const addresses = computed(() => addressStore.addressList)

async function fetchData() {
  loading.value = true
  try {
    // 获取购物车列表
    await cartStore.GetCartList()
    // 获取地址列表
    await addressStore.GetAddressList()

    // 只取选中且可购买的
    items.value = cartStore.cartList.filter(i => i.is_selected && i.is_available)

    // 默认选中默认地址
    const defaultAddr = addresses.value.find(a => a.is_default)
    if (defaultAddr) selectedAddressId.value = defaultAddr.address_id
    else if (addresses.value.length > 0) selectedAddressId.value = addresses.value[0].address_id
  } finally {
    loading.value = false
  }
}

function generateUUID() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, c => {
    const r = Math.random() * 16 | 0
    return (c === 'x' ? r : (r & 0x3 | 0x8)).toString(16)
  })
}

function handleGoPay() {
  successVisible.value = false
  router.push(`/shop/pay/${createdOrderId.value}`)
}

async function handleSubmit() {
  if (!selectedAddressId.value) { ElMessage.warning('请选择收货地址'); return }

  submitting.value = true
  try {
    const result = await orderStore.CreateOrder({
      address_id: selectedAddressId.value,
      idempotent_key: generateUUID(),
    })
    if (result) {
      createdOrderId.value = result.order_id
      createdOrderNo.value = result.order_no
      successVisible.value = true
      // 刷新购物车
      await cartStore.GetCartList()
      await cartStore.GetCartCount()
    } else {
      ElMessage.error('下单失败')
    }
  } catch {
    ElMessage.error('下单失败')
  } finally {
    submitting.value = false
  }
}

onMounted(() => fetchData())
</script>
