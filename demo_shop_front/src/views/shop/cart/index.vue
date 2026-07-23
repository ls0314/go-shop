<template>
  <div class="flex h-full flex-col">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-800 dark:text-white">
        购物车
        <span class="ml-2 text-base font-normal text-slate-400">({{ cartList.length }} 种商品)</span>
      </h1>
    </div>

    <!-- 空购物车 -->
    <div
      v-if="!loading && cartList.length === 0"
      class="flex flex-col items-center justify-center rounded-xl bg-white py-24 dark:bg-slate-800"
    >
      <span class="text-5xl">🛒</span>
      <p class="mt-4 text-slate-400">购物车还是空的</p>
      <el-button class="mt-4 h-9 rounded-xl" @click="$router.push('/shop/product/list')">去逛逛</el-button>
    </div>

    <template v-else>
      <!-- 全选操作栏 -->
      <div class="flex shrink-0 items-center justify-between rounded-xl bg-white px-5 py-3 dark:bg-slate-800">
        <label class="flex cursor-pointer items-center gap-2 text-sm text-slate-600 dark:text-slate-300">
          <el-checkbox
            :model-value="isAllSelected"
            :indeterminate="isIndeterminate"
            size="large"
            @change="handleSelectAll"
          />
          全选
        </label>
        <el-button link type="danger" class="text-sm" @click="handleDeleteSelected">
          <el-icon class="mr-1"><Delete /></el-icon>
          删除选中
        </el-button>
      </div>

      <!-- 购物车列表（可滚动区域） -->
      <div class="min-h-0 flex-1 overflow-auto py-4">
        <div
          v-for="item in cartList"
          :key="item.cart_item_id"
          class="rounded-xl bg-white p-4 transition-shadow hover:shadow-sm dark:bg-slate-800"
        >
          <div class="flex items-center gap-4">
            <!-- 选中框 -->
            <el-checkbox
              :model-value="item.is_selected"
              size="large"
              @change="handleToggleSelect(item)"
            />

            <!-- 商品图片 -->
            <div
              class="h-20 w-20 flex-shrink-0 cursor-pointer overflow-hidden rounded-lg bg-slate-50 dark:bg-slate-700"
              @click="$router.push(`/shop/product/${item.spu_id}`)"
            >
              <img
                v-if="item.sku_image || item.main_image"
                :src="item.sku_image || item.main_image"
                class="h-full w-full object-cover"
              />
              <div v-else class="flex h-full w-full items-center justify-center text-2xl text-slate-300">
                <el-icon><Picture /></el-icon>
              </div>
            </div>

            <!-- 商品信息 -->
            <div class="min-w-0 flex-1 cursor-pointer" @click="$router.push(`/shop/product/${item.spu_id}`)">
              <h3 class="truncate text-sm font-medium text-slate-800 dark:text-white">{{ item.spu_name }}</h3>
              <p class="mt-1 truncate text-xs text-slate-400">{{ item.sku_name }}</p>

              <div class="mt-2 flex items-center gap-3">
                <span class="text-base font-bold" style="color: #ff6700">¥{{ item.price }}</span>
                <span v-if="!item.is_available" class="text-xs text-rose-500">{{ item.unavailable_reason }}</span>
              </div>
            </div>

            <!-- 数量调整 -->
            <div class="flex items-center gap-1">
              <el-button
                :disabled="item.quantity <= 1"
                class="h-7 w-7 min-w-0 rounded-full p-0"
                @click="handleQuantityChange(item, -1)"
              >
                <el-icon><Minus /></el-icon>
              </el-button>
              <span class="w-10 text-center text-sm font-medium">{{ item.quantity }}</span>
              <el-button
                :disabled="item.quantity >= item.stock"
                class="h-7 w-7 min-w-0 rounded-full p-0"
                @click="handleQuantityChange(item, 1)"
              >
                <el-icon><Plus /></el-icon>
              </el-button>
            </div>

            <!-- 小计 -->
            <div class="hidden w-24 text-right sm:block">
              <span class="text-sm font-bold" style="color: #ff6700">¥{{ (item.price * item.quantity).toFixed(2) }}</span>
            </div>

            <!-- 删除 -->
            <el-button link type="danger" class="text-sm" @click="handleDelete(item.cart_item_id)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
        </div>
      </div>

      <!-- 底部结算栏（固定底部） -->
      <div
        class="shrink-0 flex items-center justify-between rounded-xl bg-white px-6 py-4 shadow-lg shadow-slate-200 dark:bg-slate-800 dark:shadow-none"
      >
        <div class="flex items-center gap-4 text-sm text-slate-600 dark:text-slate-300">
          <span>已选 <b class="text-slate-800 dark:text-white">{{ selectedCount }}</b> 件</span>
          <span class="hidden sm:inline">共 <b class="text-slate-800 dark:text-white">{{ selectedQuantity }}</b> 个</span>
        </div>
        <div class="flex items-center gap-4">
          <span class="text-sm text-slate-600 dark:text-slate-300">合计：</span>
          <span class="text-2xl font-bold" style="color: #ff6700">¥{{ totalAmount.toFixed(2) }}</span>
          <el-button
            type="primary"
            size="large"
            class="ml-2 h-11 rounded-xl px-8 text-base"
            style="background: #ff6700; border-color: #ff6700"
            :disabled="selectedCount === 0"
            @click="handleCheckout"
          >
            结算
          </el-button>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Delete, Minus, Plus, Picture } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useCartStore } from '@/pinia/modules/cart'
import type { CartItem } from '@/types/cart'

const cartStore = useCartStore()
const router = useRouter()
const loading = ref(false)

const cartList = computed(() => cartStore.cartList)

// 全选/半选状态
const selectedCount = computed(() => cartList.value.filter(i => i.is_selected && i.is_available).length)
const selectedQuantity = computed(() =>
  cartList.value.filter(i => i.is_selected && i.is_available).reduce((s, i) => s + i.quantity, 0)
)
const totalAmount = computed(() =>
  cartList.value.filter(i => i.is_selected && i.is_available).reduce((s, i) => s + i.price * i.quantity, 0)
)
const isAllSelected = computed(() => cartList.value.length > 0 && cartList.value.every(i => i.is_selected))
const isIndeterminate = computed(() => !isAllSelected.value && cartList.value.some(i => i.is_selected))

async function fetchData() {
  loading.value = true
  try { await cartStore.GetCartList() }
  finally { loading.value = false }
}

async function handleToggleSelect(item: CartItem) {
  await cartStore.UpdateCart(item.cart_item_id, { is_selected: !item.is_selected })
}

async function handleSelectAll(val: boolean) {
  await cartStore.SelectAllCart(val)
}

async function handleQuantityChange(item: CartItem, delta: number) {
  const newQty = item.quantity + delta
  if (newQty < 1 || newQty > item.stock) return
  await cartStore.UpdateCart(item.cart_item_id, { quantity: newQty })
}

async function handleDelete(id: number) {
  await cartStore.DeleteCart(id)
  ElMessage.success('已移除')
}

async function handleDeleteSelected() {
  const selectedIds = cartList.value.filter(i => i.is_selected).map(i => i.cart_item_id)
  if (selectedIds.length === 0) { ElMessage.warning('请先选中商品'); return }
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedIds.length} 件商品吗？`, '删除确认', { type: 'warning' })
    for (const id of selectedIds) {
      await cartStore.DeleteCart(id)
    }
    ElMessage.success('已删除')
  } catch { /* 取消 */ }
}

function handleCheckout() {
  router.push('/shop/checkout')
}

onMounted(async () => {
  await fetchData()
  await cartStore.GetCartCount()
})
</script>
