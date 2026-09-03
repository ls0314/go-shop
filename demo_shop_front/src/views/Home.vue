<template>
  <div class="space-y-6">
    <!-- ===== 欢迎标题 ===== -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-slate-800 dark:text-white">
          {{ greeting }}，{{ displayName }} 👋
        </h1>
        <p class="mt-1 text-sm text-slate-400">
          欢迎回来，这里是运营看板示例 —— 后续可接入真实统计接口
        </p>
      </div>
      <el-button type="primary" class="h-9 rounded-xl" style="background:#ff6700;border-color:#ff6700">
        <el-icon class="mr-1"><Refresh /></el-icon>
        刷新数据
      </el-button>
    </div>

    <!-- ===== 顶部统计卡片 ===== -->
    <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <div
        v-for="card in statCards"
        :key="card.label"
        class="rounded-xl border border-slate-200 bg-white p-5 dark:border-slate-700 dark:bg-slate-800"
      >
        <div class="flex items-center justify-between">
          <p class="text-sm text-slate-500 dark:text-slate-400">{{ card.label }}</p>
          <span class="text-lg">{{ card.icon }}</span>
        </div>
        <p class="mt-2 text-3xl font-bold text-slate-800 dark:text-white">{{ card.value }}</p>
        <p class="mt-1 text-xs" :style="{ color: card.trend >= 0 ? '#10b981' : '#ef4444' }">
          {{ card.trend >= 0 ? '↑' : '↓' }} {{ Math.abs(card.trend) }}% 较上周
        </p>
      </div>
    </div>

    <!-- ===== 主体两栏 ===== -->
    <div class="grid gap-4 xl:grid-cols-3">
      <!-- 左:趋势图占位 -->
      <div class="rounded-xl border border-slate-200 bg-white p-5 dark:border-slate-700 dark:bg-slate-800 xl:col-span-2">
        <div class="flex items-center justify-between">
          <h2 class="text-base font-semibold text-slate-800 dark:text-white">访问趋势</h2>
          <el-tag size="small" class="rounded-lg">近 7 日</el-tag>
        </div>
        <!-- TODO(待实现):用 echarts / 自定义 SVG 绘制趋势折线图 -->
        <div class="mt-4 flex h-64 items-center justify-center rounded-lg border border-dashed border-slate-200 bg-slate-50 text-sm text-slate-400 dark:border-slate-600 dark:bg-slate-700/50">
          趋势图区域 —— 待接入图表库
        </div>
      </div>

      <!-- 右:快捷入口 / 待办 -->
      <div class="rounded-xl border border-slate-200 bg-white p-5 dark:border-slate-700 dark:bg-slate-800">
        <h2 class="text-base font-semibold text-slate-800 dark:text-white">快捷操作</h2>
        <div class="mt-4 space-y-2">
          <el-button v-for="a in quickActions" :key="a.title" class="h-10 w-full justify-start rounded-xl" @click="router.push(a.path)">
            <span class="mr-2">{{ a.icon }}</span>
            <span>{{ a.title }}</span>
          </el-button>
        </div>
      </div>
    </div>

    <!-- ===== 底部列表占位 ===== -->
    <div class="rounded-xl border border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800">
      <div class="flex items-center justify-between border-b border-slate-100 p-5 dark:border-slate-700">
        <h2 class="text-base font-semibold text-slate-800 dark:text-white">最近动态</h2>
        <el-button link type="primary" class="text-sm">查看全部</el-button>
      </div>
      <!-- TODO(待实现):接入真实列表接口 -->
      <div class="flex flex-col items-center justify-center py-16 text-slate-400">
        <span class="text-4xl">📭</span>
        <p class="mt-3 text-sm">暂无数据 —— 该区域待实现</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { Refresh } from '@element-plus/icons-vue'
import { useUserStore } from '@/pinia/modules/user'

const router = useRouter()
const userStore = useUserStore()

const displayName = computed(() => userStore.userInfo?.username || '管理员')

// 简单问候语:按小时切换
const greeting = computed(() => {
  const h = new Date().getHours()
  if (h < 6) return '夜深了'
  if (h < 12) return '早上好'
  if (h < 18) return '下午好'
  return '晚上好'
})

// ===== 统计卡片(示例数据,后续接入统计接口) =====
const statCards = [
  { label: '今日订单', value: '128', trend: 12.5, icon: '🧾' },
  { label: '今日销售额(元)', value: '¥24,680', trend: 8.2, icon: '💰' },
  { label: '在售商品', value: '1,024', trend: -3.1, icon: '📦' },
  { label: '注册用户', value: '8,956', trend: 15.0, icon: '👥' },
]

// ===== 快捷操作入口 =====
const quickActions = [
  { icon: '📦', title: '商品管理', path: '/platform/product/list' },
  { icon: '🧾', title: '订单管理', path: '/platform/order/list' },
  { icon: '🎟️', title: '优惠券管理', path: '/platform/coupon/list' },
  { icon: '🔧', title: '系统设置', path: '/admin/system/role' },
]
</script>
