<template>
  <div>
    <!-- ==================== 顶部：标题 + 搜索 ==================== -->
    <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-slate-800 dark:text-white">全部商品</h1>
        <p v-if="!loading" class="mt-1 text-sm text-slate-400">共 {{ total }} 件商品</p>
      </div>
      <div class="w-full sm:w-64">
        <el-input
          v-model="query.spu_name"
          placeholder="搜索商品"
          clearable
          :prefix-icon="Search"
          class="search-input"
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        />
      </div>
    </div>

    <!-- ==================== 排序栏（小米风格 Pill 按钮） ==================== -->
    <div class="mt-5 flex flex-wrap items-center gap-2 rounded-xl bg-white px-5 py-3 dark:bg-slate-800">
      <span class="mr-2 text-xs text-slate-400">排序：</span>
      <button
        v-for="s in sorts"
        :key="s.value"
        class="rounded-lg px-4 py-2 text-sm transition"
        :class="query.sort === s.value
          ? 'bg-slate-900 font-medium text-white dark:bg-white dark:text-slate-900'
          : 'text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-700'"
        @click="changeSort(s.value)"
      >
        {{ s.label }}
      </button>
    </div>

    <!-- ==================== 加载状态 ==================== -->
    <div v-if="loading" class="flex justify-center py-24">
      <el-icon class="is-loading text-3xl text-slate-300"><Loading /></el-icon>
    </div>

    <!-- ==================== 空状态 ==================== -->
    <div
      v-else-if="productList.length === 0"
      class="flex flex-col items-center justify-center rounded-xl bg-white py-24 dark:bg-slate-800"
    >
      <span class="text-5xl">📦</span>
      <p class="mt-4 text-slate-400">暂无商品</p>
    </div>

    <!-- ==================== 商品网格 ==================== -->
    <div v-else class="mt-5 grid gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      <div
        v-for="item in productList"
        :key="item.spu_id"
        class="group cursor-pointer overflow-hidden bg-white transition-all duration-300 hover:-translate-y-1 hover:shadow-lg dark:bg-slate-800"
        @click="$router.push(`/shop/product/${item.spu_id}`)"
      >
        <!-- 商品图片 -->
        <div class="aspect-square overflow-hidden bg-slate-50 dark:bg-slate-700">
          <img
            v-if="item.main_image"
            :src="item.main_image"
            :alt="item.spu_name"
            class="h-full w-full object-cover transition-transform duration-500 group-hover:scale-105"
          />
          <div v-else class="flex h-full w-full items-center justify-center text-5xl text-slate-200">
            <el-icon><Picture /></el-icon>
          </div>
        </div>

        <!-- 商品信息 -->
        <div class="p-4">
          <h3 class="truncate text-sm font-medium text-slate-800 dark:text-white">
            {{ item.spu_name }}
          </h3>
          <p class="mt-1 truncate text-xs text-slate-400">{{ item.category_name }}{{ item.brand ? ' · ' + item.brand : '' }}</p>

          <!-- 价格行 -->
          <div class="mt-3 flex items-baseline gap-2">
            <span class="text-lg font-bold" style="color: #ff6700">¥{{ item.min_price }}</span>
            <span
              v-if="item.min_price !== item.max_price"
              class="text-xs text-slate-300 line-through decoration-slate-300"
            >
              ¥{{ item.max_price }}
            </span>
          </div>

          <!-- 已售 + 评价 -->
          <div class="mt-2 flex items-center gap-3 text-xs text-slate-400">
            <span>已售 {{ formatSold(item.total_sold) }}</span>
            <span class="text-slate-200">|</span>
            <span class="flex items-center gap-1">
              <span class="text-amber-400">★</span>
              <span>{{ (4.5 + Math.random() * 0.5).toFixed(1) }}</span>
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- ==================== 分页 ==================== -->
    <div v-if="total > 0" class="mt-8 flex justify-center">
      <el-pagination
        v-model:current-page="query.page"
        v-model:page-size="query.pageSize"
        :total="total"
        :page-sizes="[12, 24, 48]"
        layout="prev, pager, next, total"
        background
        @change="fetchData"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Search, Loading, Picture } from '@element-plus/icons-vue'
import { useShopStore } from '@/pinia/modules/shop'

const shopStore = useShopStore()
const loading = ref(false)

const query = ref<Record<string, any>>({
  page: 1,
  pageSize: 24,
  spu_name: '',
  sort: '',
})

const sorts = [
  { label: '默认', value: '' },
  { label: '价格从低到高', value: 'price_asc' },
  { label: '价格从高到低', value: 'price_desc' },
  { label: '销量优先', value: 'sold_desc' },
  { label: '最新上架', value: 'newest' },
]

const productList = computed(() => shopStore.productList)
const total = computed(() => shopStore.productTotal)

function changeSort(val: string) {
  query.value.sort = val
  query.value.page = 1
  fetchData()
}

function handleSearch() {
  query.value.page = 1
  fetchData()
}

async function fetchData() {
  loading.value = true
  try {
    await shopStore.GetProductList({
      page: query.value.page,
      // 下划线命名:后端 SpuQueryReq 的 form 标签是 page_size
      page_size: query.value.pageSize,
      spu_name: query.value.spu_name || undefined,
      sort: query.value.sort || undefined,
      spu_status: 'published',
    })
  } finally {
    loading.value = false
  }
}

function formatSold(n: number): string {
  if (n >= 10000) return (n / 10000).toFixed(1) + '万'
  return String(n)
}

fetchData()
</script>

<style scoped>
:deep(.search-input .el-input__wrapper) {
  border-radius: 999px;
}
</style>
