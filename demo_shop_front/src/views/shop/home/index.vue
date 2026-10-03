<template>
  <div>
    <!-- ==================== 顶部轮播 Hero ==================== -->
    <div class="relative overflow-hidden rounded-2xl">
      <div class="relative aspect-[2.8/1] max-h-[420px] min-h-[240px] overflow-hidden rounded-2xl bg-slate-700">
        <div
          v-for="(banner, idx) in banners"
          :key="idx"
          class="absolute inset-0 transition-opacity duration-500"
          :class="idx === currentBanner ? 'opacity-100' : 'opacity-0'"
        >
          <div
            class="flex h-full w-full flex-col items-center justify-center px-8 text-center text-white"
            :style="{ background: banner.bg }"
          >
            <h2 class="text-3xl font-bold md:text-5xl">{{ banner.title }}</h2>
            <p class="mt-3 text-lg text-white/70">{{ banner.subtitle }}</p>
            <el-button
              class="mt-6 h-10 rounded-full border-white/30 bg-white/10 text-white hover:bg-white/20"
              @click="$router.push('/shop/product/list')"
            >
              立即选购
            </el-button>
          </div>
        </div>
      </div>

      <!-- 轮播指示器 -->
      <div class="absolute bottom-4 right-6 flex gap-2">
        <button
          v-for="(_, idx) in banners"
          :key="idx"
          class="h-2 w-6 rounded-full border-0 transition-all"
          :class="idx === currentBanner ? 'bg-white/80' : 'bg-white/30'"
          @click="currentBanner = idx"
        />
      </div>
    </div>

    <!-- ==================== 分类快捷入口 ==================== -->
    <div class="mt-8 grid grid-cols-5 gap-1 rounded-2xl bg-white py-6 dark:bg-slate-800 md:grid-cols-10">
      <div
        v-for="cat in categoryEntries"
        :key="cat.name"
        class="flex cursor-pointer flex-col items-center gap-2 py-3 transition-opacity hover:opacity-70"
        @click="$router.push('/shop/product/list')"
      >
        <span class="text-2xl">{{ cat.icon }}</span>
        <span class="text-xs text-slate-600 dark:text-slate-300">{{ cat.name }}</span>
      </div>
    </div>

    <!-- ==================== 商品展示区 ==================== -->
    <section
      v-for="section in productSections"
      :key="section.title"
      class="mt-10"
    >
      <!-- 区块标题 -->
      <div class="mb-5 flex items-end justify-between">
        <div>
          <h2 class="text-2xl font-bold text-slate-800 dark:text-white">{{ section.title }}</h2>
          <p class="mt-1 text-sm text-slate-400">{{ section.subtitle }}</p>
        </div>
        <el-button link type="primary" class="text-sm" @click="$router.push('/shop/product/list')">
          浏览全部 <el-icon class="ml-1"><ArrowRight /></el-icon>
        </el-button>
      </div>

      <!-- 商品横滚 -->
      <div class="-mx-1 flex gap-3 overflow-x-auto pb-2 scrollbar-hide">
        <div
          v-for="item in productList"
          :key="item.spu_id"
          class="w-[234px] flex-shrink-0 cursor-pointer overflow-hidden rounded-xl bg-white transition-all hover:-translate-y-1 hover:shadow-xl dark:bg-slate-800"
          @click="$router.push(`/shop/product/${item.spu_id}`)"
        >
          <!-- 商品图片 -->
          <div class="aspect-square overflow-hidden bg-slate-50 dark:bg-slate-700">
            <img
              v-if="item.main_image"
              :src="item.main_image"
              :alt="item.spu_name"
              class="h-full w-full object-cover transition-transform duration-300 hover:scale-105"
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
            <p class="mt-1 truncate text-xs text-slate-400">{{ item.brand || item.category_name }}</p>
            <div class="mt-3 flex items-baseline gap-1">
              <span class="text-lg font-bold" style="color: #ff6700">¥{{ item.min_price }}</span>
              <span v-if="item.min_price !== item.max_price" class="text-xs text-slate-300 line-through">
                ¥{{ item.max_price }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ==================== 加载状态 ==================== -->
    <div v-if="loading" class="flex justify-center py-20">
      <el-icon class="is-loading text-3xl text-slate-300"><Loading /></el-icon>
    </div>

    <!-- ==================== 底部服务承诺 ==================== -->
    <div class="mt-14 grid gap-4 rounded-2xl bg-white py-8 dark:bg-slate-800 sm:grid-cols-4">
      <div
        v-for="svc in services"
        :key="svc.title"
        class="flex flex-col items-center gap-2 border-slate-100 px-4 text-center sm:border-r dark:border-slate-700 last:border-r-0"
      >
        <span class="text-2xl">{{ svc.icon }}</span>
        <span class="text-sm font-medium text-slate-700 dark:text-slate-200">{{ svc.title }}</span>
        <span class="text-xs text-slate-400">{{ svc.desc }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ArrowRight, Loading, Picture } from '@element-plus/icons-vue'
import { useShopStore } from '@/pinia/modules/shop'

const shopStore = useShopStore()
const loading = ref(false)

// ==================== 轮播 Banner ====================
const currentBanner = ref(0)
let bannerTimer: ReturnType<typeof setInterval> | null = null

const banners = [
  { title: 'Ds_demo 商城', subtitle: '探索智能生活新方式', bg: 'linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%)' },
  { title: '新品上市', subtitle: '品质好物，限时特惠', bg: 'linear-gradient(135deg, #2d1b69 0%, #614385 50%, #516395 100%)' },
  { title: '每日精选', subtitle: '发现属于你的好设计', bg: 'linear-gradient(135deg, #0f2027 0%, #203a43 50%, #2c5364 100%)' },
]

// ==================== 分类入口 ====================
const categoryEntries = [
  { icon: '📱', name: '手机' },
  { icon: '💻', name: '笔记本' },
  { icon: '📺', name: '电视' },
  { icon: '🎧', name: '耳机' },
  { icon: '⌚', name: '手表' },
  { icon: '🏠', name: '智能家居' },
  { icon: '🔌', name: '充电配件' },
  { icon: '🎮', name: '游戏' },
  { icon: '📦', name: '生活周边' },
  { icon: '🛒', name: '全部商品' },
]

// ==================== 商品区块 ====================
const productSections = [
  { title: '热门推荐', subtitle: '大家都在买' },
  { title: '新品上架', subtitle: '第一时间尝鲜' },
]

// ==================== 服务承诺 ====================
const services = [
  { icon: '🚚', title: '满99包邮', desc: '全国范围送货上门' },
  { icon: '🔄', title: '7天无理由', desc: '安心退换无忧' },
  { icon: '🛡️', title: '正品保障', desc: '品质严选认证' },
  { icon: '💬', title: '在线客服', desc: '全天候贴心服务' },
]

const productList = computed(() => shopStore.productList)

onMounted(async () => {
  loading.value = true
  try {
    await shopStore.GetProductList({ page: 1, page_size: 10, spu_status: 'published' })
  } finally {
    loading.value = false
  }
  // 自动轮播
  bannerTimer = setInterval(() => {
    currentBanner.value = (currentBanner.value + 1) % banners.length
  }, 4000)
})

onUnmounted(() => {
  if (bannerTimer) clearInterval(bannerTimer)
})
</script>

<style scoped>
.scrollbar-hide::-webkit-scrollbar {
  display: none;
}
.scrollbar-hide {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
</style>
