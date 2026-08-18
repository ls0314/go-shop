<template>
  <div class="flex h-full flex-col">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-800 dark:text-white">领券中心</h1>
    </div>

    <!-- Tab 切换:可领取 / 我的卡券 -->
    <div class="mt-4 flex shrink-0 gap-1 rounded-xl bg-slate-100 p-1 dark:bg-slate-800">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        type="button"
        class="flex-1 rounded-lg py-2 text-sm font-medium transition"
        :class="activeTab === tab.key
          ? 'bg-white text-slate-800 shadow-sm dark:bg-slate-700 dark:text-white'
          : 'text-slate-500 hover:text-slate-700 dark:hover:text-slate-300'"
        @click="switchTab(tab.key)"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- ==================== 可领取 ==================== -->
    <template v-if="activeTab === 'receive'">
      <!-- 加载 -->
      <div v-if="loading" class="flex flex-1 items-center justify-center">
        <el-icon class="animate-spin text-3xl text-slate-400"><Loading /></el-icon>
      </div>

      <!-- 空态 -->
      <div v-else-if="templateList.length === 0" class="flex flex-1 flex-col items-center justify-center">
        <span class="text-5xl">🎟️</span>
        <p class="mt-4 text-slate-400">暂无可用优惠券</p>
      </div>

      <!-- 卡片列表 -->
      <div v-else class="mt-4 min-h-0 flex-1 space-y-3 overflow-auto pb-4">
        <div
          v-for="tpl in templateList"
          :key="tpl.template_id"
          class="flex items-stretch overflow-hidden rounded-xl border border-dashed border-slate-300 bg-white dark:border-slate-600 dark:bg-slate-800"
        >
          <!-- 左侧面额区 -->
          <div
            class="flex w-28 flex-shrink-0 flex-col items-center justify-center gap-1 border-r border-dashed border-slate-200 py-4 dark:border-slate-700"
            style="background: linear-gradient(135deg, #ff6700 0%, #ff8534 100%)"
          >
            <span class="text-3xl font-bold text-white">
              {{ tpl.coupon_type === 'full_reduction' ? `¥${tpl.discount_amount}` : `${Math.round(tpl.discount_amount * 10)}折` }}
            </span>
            <span class="text-xs text-white/80">
              {{ tpl.coupon_type === 'full_reduction' ? '满减券' : '折扣券' }}
            </span>
          </div>

          <!-- 中间信息区 -->
          <div class="flex flex-1 flex-col justify-center px-4 py-3">
            <p class="text-sm font-semibold text-slate-800 dark:text-white">{{ tpl.coupon_name }}</p>
            <p class="mt-1 text-xs text-slate-500 dark:text-slate-400">
              {{ tpl.threshold_amount > 0 ? `满 ¥${tpl.threshold_amount} 可用` : '无门槛使用' }}
            </p>
            <p class="mt-0.5 text-xs text-slate-400">
              {{ validityText(tpl) }}
            </p>
            <p v-if="tpl.remaining_count > 0 && tpl.remaining_count <= 20" class="mt-0.5 text-xs" style="color: #ff6700">
              仅剩 {{ tpl.remaining_count }} 张
            </p>
          </div>

          <!-- 右侧领取按钮 -->
          <div class="flex w-24 flex-shrink-0 items-center justify-center">
            <el-button
              v-if="tpl.held_count >= tpl.per_user_limit"
              disabled
              class="h-8 rounded-full px-4 text-xs"
            >
              已领完
            </el-button>
            <el-button
              v-else-if="tpl.remaining_count <= 0"
              disabled
              class="h-8 rounded-full px-4 text-xs"
            >
              已抢光
            </el-button>
            <el-button
              v-else
              type="primary"
              class="h-8 rounded-full px-5 text-xs"
              style="background: #ff6700; border-color: #ff6700"
              :loading="receivingId === tpl.template_id"
              @click="handleReceive(tpl.template_id)"
            >
              立即领取
            </el-button>
          </div>
        </div>

        <!-- 分页 -->
        <div v-if="templateTotal > templateList.length" class="flex justify-center pt-2">
          <el-button link type="primary" :loading="loadingMore" @click="loadMoreTemplates">
            加载更多
          </el-button>
        </div>
      </div>
    </template>

    <!-- ==================== 我的卡券 ==================== -->
    <template v-else>
      <!-- 状态筛选 -->
      <div class="mt-4 flex shrink-0 items-center gap-2">
        <el-radio-group v-model="couponStatus" size="small" @change="handleStatusChange">
          <el-radio-button value="">全部</el-radio-button>
          <el-radio-button value="unused">未使用</el-radio-button>
          <el-radio-button value="used">已使用</el-radio-button>
          <el-radio-button value="expired">已过期</el-radio-button>
        </el-radio-group>
      </div>

      <!-- 加载 -->
      <div v-if="loading" class="flex flex-1 items-center justify-center">
        <el-icon class="animate-spin text-3xl text-slate-400"><Loading /></el-icon>
      </div>

      <!-- 空态 -->
      <div v-else-if="myCouponList.length === 0" class="flex flex-1 flex-col items-center justify-center">
        <span class="text-5xl">💳</span>
        <p class="mt-4 text-slate-400">暂无优惠券</p>
        <el-button class="mt-4 h-9 rounded-xl" @click="switchTab('receive')">去领券</el-button>
      </div>

      <!-- 卡券列表 -->
      <div v-else class="mt-4 min-h-0 flex-1 space-y-3 overflow-auto pb-4">
        <div
          v-for="coupon in myCouponList"
          :key="coupon.user_coupon_id"
          class="flex items-stretch overflow-hidden rounded-xl border border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800"
          :class="coupon.status === 'expired' ? 'opacity-60' : ''"
        >
          <!-- 左侧面额区 -->
          <div
            class="flex w-28 flex-shrink-0 flex-col items-center justify-center gap-1 border-r border-dashed border-slate-200 py-4 dark:border-slate-700"
            :style="coupon.status === 'used'
              ? { background: '#94a3b8' }
              : coupon.status === 'expired'
                ? { background: '#cbd5e1' }
                : { background: 'linear-gradient(135deg, #ff6700 0%, #ff8534 100%)' }"
          >
            <span class="text-3xl font-bold text-white">
              {{ coupon.coupon_type === 'full_reduction' ? `¥${coupon.discount_amount}` : `${Math.round(coupon.discount_amount * 10)}折` }}
            </span>
            <span class="text-xs text-white/80">
              {{ coupon.coupon_type === 'full_reduction' ? '满减券' : '折扣券' }}
            </span>
          </div>

          <!-- 中间信息区 -->
          <div class="flex flex-1 flex-col justify-center px-4 py-3">
            <p class="text-sm font-semibold text-slate-800 dark:text-white">{{ coupon.coupon_name }}</p>
            <p class="mt-1 text-xs text-slate-500 dark:text-slate-400">
              {{ coupon.threshold_amount > 0 ? `满 ¥${coupon.threshold_amount} 可用` : '无门槛使用' }}
            </p>
            <p class="mt-0.5 text-xs text-slate-400">
              <template v-if="coupon.status === 'used'">已使用于订单 {{ coupon.order_no }}</template>
              <template v-else>有效期至 {{ formatExpire(coupon.expire_at) }}</template>
            </p>
          </div>

          <!-- 右侧状态 -->
          <div class="flex w-20 flex-shrink-0 items-center justify-center">
            <el-tag :type="statusTagType(coupon.status)" size="small" class="rounded-lg">
              {{ statusText(coupon.status) }}
            </el-tag>
          </div>
        </div>

        <!-- 分页 -->
        <div v-if="myCouponTotal > myCouponList.length" class="flex justify-center pt-2">
          <el-button link type="primary" :loading="loadingMore" @click="loadMoreMyCoupons">
            加载更多
          </el-button>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Loading } from '@element-plus/icons-vue'
import { useUserStore } from '@/pinia/modules/user'
import { GetReceiveCouponListApi, GetUserCouponListApi, ReceiveCouponApi } from '@/api/coupon'
import type { UserCouponTemplate, UserCouponItem } from '@/types/coupon'

const userStore = useUserStore()
const route = useRoute()

const tabs: { key: 'receive' | 'mine'; label: string }[] = [
  { key: 'receive', label: '可领取' },
  { key: 'mine', label: '我的卡券' },
]
const activeTab = ref<'receive' | 'mine'>('receive')

const loading = ref(false)
const loadingMore = ref(false)

// ==================== 可领取 ====================
const templateList = ref<UserCouponTemplate[]>([])
const templateTotal = ref(0)
const templatePage = ref(1)
const pageSize = 10
const receivingId = ref<number | null>(null)

// ==================== 我的卡券 ====================
const couponStatus = ref('')
const myCouponList = ref<UserCouponItem[]>([])
const myCouponTotal = ref(0)
const myCouponPage = ref(1)

// 未登录提示后跳转登录页
const ensureLogin = (): boolean => {
  if (userStore.userToken.access_token) return true
  ElMessage.warning('请先登录')
  // 记录回跳地址
  sessionStorage.setItem('redirect_after_login', '/shop/coupon')
  // 跳登录页(路由守卫处理 /login 已登录跳商城首页)
  window.location.href = '/login'
  return false
}

const fetchTemplates = async (page = 1) => {
  const res = await GetReceiveCouponListApi({ page, page_size: pageSize })
  return res.data.data
}

const loadTemplates = async () => {
  loading.value = true
  try {
    const payload = await fetchTemplates(1)
    templateList.value = payload?.list || []
    templateTotal.value = payload?.total || 0
    templatePage.value = 1
  } catch (error: any) {
    // 401 由 axios 拦截器处理跳登录
    if (error?.response?.status !== 401) {
      console.error('获取可领优惠券失败:', error)
      ElMessage.error('获取优惠券失败')
    }
  } finally {
    loading.value = false
  }
}

const loadMoreTemplates = async () => {
  loadingMore.value = true
  try {
    const next = templatePage.value + 1
    const payload = await fetchTemplates(next)
    templateList.value = [...templateList.value, ...(payload?.list || [])]
    templatePage.value = next
  } catch (error) {
    console.error('加载更多失败:', error)
  } finally {
    loadingMore.value = false
  }
}

const handleReceive = async (templateId: number) => {
  if (!ensureLogin()) return
  receivingId.value = templateId
  try {
    await ReceiveCouponApi(templateId)
    ElMessage.success('领取成功')
    // 刷新列表(领完后 held_count 变化,按钮状态随之更新)
    await loadTemplates()
  } catch (error: any) {
    console.error('领取失败:', error)
    ElMessage.error(error?.response?.data?.message || '领取失败')
  } finally {
    receivingId.value = null
  }
}

// ==================== 我的卡券 ====================
const fetchMyCoupons = async (page = 1, status = '') => {
  const res = await GetUserCouponListApi({ page, page_size: pageSize, status: status || undefined })
  return res.data.data
}

const loadMyCoupons = async () => {
  loading.value = true
  try {
    const payload = await fetchMyCoupons(1, couponStatus.value)
    myCouponList.value = payload?.list || []
    myCouponTotal.value = payload?.total || 0
    myCouponPage.value = 1
  } catch (error: any) {
    if (error?.response?.status !== 401) {
      console.error('获取我的卡券失败:', error)
      ElMessage.error('获取卡券失败')
    }
  } finally {
    loading.value = false
  }
}

const loadMoreMyCoupons = async () => {
  loadingMore.value = true
  try {
    const next = myCouponPage.value + 1
    const payload = await fetchMyCoupons(next, couponStatus.value)
    myCouponList.value = [...myCouponList.value, ...(payload?.list || [])]
    myCouponPage.value = next
  } catch (error) {
    console.error('加载更多失败:', error)
  } finally {
    loadingMore.value = false
  }
}

const handleStatusChange = () => {
  loadMyCoupons()
}

const switchTab = (key: 'receive' | 'mine') => {
  activeTab.value = key
  if (key === 'receive' && templateList.value.length === 0) {
    loadTemplates()
  } else if (key === 'mine' && myCouponList.value.length === 0) {
    loadMyCoupons()
  }
}

// ==================== 工具函数 ====================
const validityText = (tpl: UserCouponTemplate) => {
  if (tpl.usable_days > 0) return `领取后 ${tpl.usable_days} 天内有效`
  const end = new Date(tpl.end_time)
  if (isNaN(end.getTime())) return ''
  return `有效期至 ${end.getFullYear()}/${end.getMonth() + 1}/${end.getDate()}`
}

const formatExpire = (t: string) => {
  const d = new Date(t)
  if (isNaN(d.getTime())) return '-'
  return `${d.getFullYear()}/${d.getMonth() + 1}/${d.getDate()}`
}

const statusText = (s: string) => {
  return { unused: '未使用', used: '已使用', expired: '已过期' }[s] || s
}

const statusTagType = (s: string): 'success' | 'info' => {
  return s === 'unused' ? 'success' : 'info'
}

onMounted(() => {
  // 支持 ?tab=mine 直达"我的卡券"页签(头像菜单"我的优惠券"入口)
  if (route.query.tab === 'mine') {
    activeTab.value = 'mine'
    loadMyCoupons()
  } else {
    loadTemplates()
  }
})
</script>
