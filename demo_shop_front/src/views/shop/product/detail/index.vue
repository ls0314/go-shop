<template>
  <div class="space-y-6">
    <!-- 返回按钮 -->
    <el-button link type="primary" @click="$router.back()" class="text-sm">
      <el-icon class="mr-1"><ArrowLeft /></el-icon>
      返回列表
    </el-button>

    <div v-if="loading" class="flex justify-center py-20">
      <el-icon class="is-loading text-3xl text-slate-300"><Loading /></el-icon>
    </div>

    <template v-else-if="product">
      <!-- 商品主信息 -->
      <div class="grid gap-8 rounded-xl border border-slate-200 bg-white p-6 dark:border-slate-700 dark:bg-slate-800 lg:grid-cols-2">
        <!-- 左侧：商品图片 -->
        <div>
          <!-- 主图 -->
          <div class="aspect-square overflow-hidden rounded-xl bg-slate-100 dark:bg-slate-700">
            <img
              v-if="currentImage"
              :src="currentImage"
              :alt="product.spu_name"
              class="h-full w-full object-cover"
            />
            <div v-else class="flex h-full w-full items-center justify-center text-6xl text-slate-300">
              <el-icon><Picture /></el-icon>
            </div>
          </div>
          <!-- 缩略图列表 -->
          <div v-if="product.image_list && product.image_list.length > 0" class="mt-4 flex gap-2 overflow-auto">
            <div
              v-for="(img, idx) in product.image_list"
              :key="idx"
              class="h-16 w-16 flex-shrink-0 cursor-pointer overflow-hidden rounded-lg border-2"
              :class="currentImage === img.image_url ? 'border-indigo-500' : 'border-slate-200 dark:border-slate-600'"
              @click="currentImage = img.image_url"
            >
              <img :src="img.image_url" class="h-full w-full object-cover" />
            </div>
          </div>
        </div>

        <!-- 右侧：商品信息 + SKU 选择 -->
        <div class="flex flex-col">
          <h1 class="text-2xl font-bold text-slate-800 dark:text-white">{{ product.spu_name }}</h1>
          <p class="mt-2 text-sm text-slate-500">{{ product.category_name }}{{ product.brand ? ' · ' + product.brand : '' }}</p>

          <!-- 价格区 -->
          <div class="mt-6 flex items-baseline gap-2">
            <span class="text-3xl font-bold text-rose-500">¥{{ currentPrice }}</span>
          </div>

          <!-- 规格选择 -->
          <div v-if="product.spec_template && product.spec_template.length > 0" class="mt-6 space-y-4">
            <div v-for="spec in product.spec_template" :key="spec.name">
              <p class="mb-2 text-sm font-medium text-slate-600 dark:text-slate-300">{{ spec.name }}</p>
              <div class="flex flex-wrap gap-2">
                <button
                  v-for="val in spec.values"
                  :key="val"
                  class="rounded-lg border px-4 py-2 text-sm transition"
                  :class="selectedSpec[spec.name] === val
                    ? 'border-indigo-500 bg-indigo-50 text-indigo-600 dark:bg-indigo-900/20 dark:text-indigo-400'
                    : 'border-slate-200 bg-white text-slate-600 hover:border-indigo-300 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-300'"
                  @click="selectSpec(spec.name, val)"
                >
                  {{ val }}
                </button>
              </div>
            </div>
          </div>

          <!-- 选中 SKU 信息 -->
          <div v-if="currentSku" class="mt-6 rounded-xl bg-slate-50 p-4 dark:bg-slate-700/50">
            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-500">已选规格：{{ currentSku.sku_name }}</span>
<!--              <span class="text-sm text-slate-500">库存：{{ currentSku.stock }}</span>-->
            </div>
          </div>

          <!-- 操作按钮 -->
          <div class="mt-6 flex gap-4">
            <el-button
              size="large"
              class="h-12 flex-1 rounded-xl text-base"
              style="background: #ff6700; border-color: #ff6700; color: #fff"
              :disabled="!currentSku || currentSku.stock <= 0"
              :loading="addCartLoading"
              @click="handleAddCart"
            >
              {{ currentSku && currentSku.stock > 0 ? '加入购物车' : '已售罄' }}
            </el-button>
            <el-button
              size="large"
              class="h-12 flex-1 rounded-xl text-base"
              type="danger"
              :disabled="!currentSku || currentSku.stock <= 0"
              @click="handleBuyNow"
            >
              立即购买
            </el-button>
          </div>

          <!-- 商品描述 -->
          <div v-if="product.description" class="mt-6">
            <p class="text-sm font-medium text-slate-600 dark:text-slate-300">商品描述</p>
            <p class="mt-2 text-sm leading-relaxed text-slate-500">{{ product.description }}</p>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Loading, Picture } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useShopStore } from '@/pinia/modules/shop'
import { useCartStore } from '@/pinia/modules/cart'
import { useUserStore } from '@/pinia/modules/user'
import type { UserSkuItem } from '@/types/product'

const route = useRoute()
const router = useRouter()
const shopStore = useShopStore()
const cartStore = useCartStore()
const userStore = useUserStore()

const loading = ref(false)
const addCartLoading = ref(false)
const currentImage = ref('')

const product = computed(() => shopStore.currentProduct)

// 规格选择状态
const selectedSpec = ref<Record<string, string>>({})

// 根据已选规格匹配 SKU
const currentSku = computed<UserSkuItem | null>(() => {
  if (!product.value || !product.value.sku_list) return null
  const selectedEntries = Object.entries(selectedSpec.value)
  if (selectedEntries.length === 0) return product.value.sku_list[0] ?? null

  return product.value.sku_list.find(sku => {
    const specValues = sku.spec_value || ({} as Record<string, string>)
    return selectedEntries.every(([key, val]) => specValues[key] === val)
  }) ?? null
})

const currentPrice = computed(() => {
  if (currentSku.value) return currentSku.value.price
  if (product.value && product.value.sku_list && product.value.sku_list.length > 0) {
    return product.value.sku_list[0].price
  }
  return 0
})

function selectSpec(specName: string, specValue: string) {
  selectedSpec.value = { ...selectedSpec.value, [specName]: specValue }
}

async function handleAddCart() {
  if (!currentSku.value) return

  // 未登录 → 提示并跳转登录页
  if (!userStore.userToken.access_token) {
    ElMessage.warning('请先登录')
    router.push('/login')
    return
  }

  addCartLoading.value = true
  try {
    const ok = await cartStore.AddCart({ sku_id: currentSku.value.sku_id, quantity: 1 })
    if (ok) {
      ElMessage.success('已加入购物车')
      await cartStore.GetCartCount()
    } else {
      ElMessage.error('加入购物车失败')
    }
  } finally {
    addCartLoading.value = false
  }
}

function handleBuyNow() {
  if (!currentSku.value) return
  if (!userStore.userToken.access_token) {
    ElMessage.warning('请先登录')
    router.push('/login')
    return
  }
  ElMessage.info('支付功能开发中')
}

onMounted(async () => {
  const id = Number(route.params.id)
  if (!id) return

  loading.value = true
  try {
    await shopStore.GetProduct(id)
    if (product.value) {
      currentImage.value = product.value.main_image
      // 初始化选中第一个规格值
      if (product.value.spec_template) {
        const init: Record<string, string> = {}
        for (const spec of product.value.spec_template) {
          if (spec.values && spec.values.length > 0) {
            init[spec.name] = spec.values[0]
          }
        }
        selectedSpec.value = init
      }
    }
  } finally {
    loading.value = false
  }
})
</script>
