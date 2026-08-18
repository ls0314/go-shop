<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-800 dark:text-white">优惠券管理</h1>
    </div>

    <!-- 搜索与操作栏 -->
    <div class="flex flex-wrap items-center justify-between gap-4 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-800">
      <el-form :model="queryParams" inline class="flex flex-wrap gap-3" @submit.prevent="handleSearch">
        <el-form-item label="券名称" class="mb-0">
          <el-input
            v-model="queryParams.coupon_name"
            placeholder="请输入优惠券名称"
            clearable
            class="w-56"
            @keyup.enter="handleSearch"
            @clear="handleSearch"
          />
        </el-form-item>
        <el-form-item class="mb-0">
          <el-button type="primary" :loading="loading" class="h-9 rounded-xl" @click="handleSearch">
            <el-icon class="mr-1"><Search /></el-icon>
            搜索
          </el-button>
          <el-button class="h-9 rounded-xl" @click="handleReset">
            <el-icon class="mr-1"><Refresh /></el-icon>
            重置
          </el-button>
        </el-form-item>
      </el-form>

      <el-button type="primary" class="h-9 rounded-xl" @click="handleAdd">
        <el-icon class="mr-1"><Plus /></el-icon>
        新增优惠券
      </el-button>
    </div>

    <!-- 表格主区域 -->
    <div class="rounded-xl border border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800">
      <el-table
        :data="couponList"
        v-loading="loading"
        border
        stripe
        class="w-full"
        :header-cell-style="{ background: 'transparent' }"
        :cell-style="{ background: 'transparent' }"
      >
        <el-table-column prop="template_id" label="ID" width="70px" align="center" />
        <el-table-column prop="coupon_name" label="券名称" min-width="160px" />
        <el-table-column label="类型" width="90px" align="center">
          <template #default="scope">
            <el-tag :type="scope.row.coupon_type === 'full_reduction' ? 'warning' : 'primary'" size="small" class="rounded-lg">
              {{ scope.row.coupon_type === 'full_reduction' ? '满减' : '直减' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="使用门槛" width="110px" align="center">
          <template #default="scope">
            {{ scope.row.threshold_amount > 0 ? `满 ¥${scope.row.threshold_amount}` : '无门槛' }}
          </template>
        </el-table-column>
        <el-table-column label="优惠" width="110px" align="center">
          <template #default="scope">
            <span v-if="scope.row.coupon_type === 'full_reduction'" style="color: #ff6700">-¥{{ scope.row.discount_amount }}</span>
            <span v-else style="color: #ff6700">{{ Math.round(scope.row.discount_amount * 10) }}折</span>
          </template>
        </el-table-column>
        <el-table-column label="已领/总量" width="110px" align="center">
          <template #default="scope">
            {{ scope.row.received_count }}/{{ scope.row.total_count }}
          </template>
        </el-table-column>
        <el-table-column prop="per_user_limit" label="每人限领" width="90px" align="center" />
        <el-table-column label="有效期" min-width="180px">
          <template #default="scope">
            <span v-if="scope.row.usable_days > 0">领取后 {{ scope.row.usable_days }} 天内有效</span>
            <span v-else>{{ formatDate(scope.row.start_time) }} ~ {{ formatDate(scope.row.end_time) }}</span>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="flex justify-end p-4">
        <el-pagination
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.page_size"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @change="fetchData"
        />
      </div>
    </div>

    <!-- 新增优惠券弹窗 -->
    <CouponForm v-model:visible="dialogVisible" @success="handleSuccess" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Search, Refresh, Plus } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import CouponForm from './components/CouponForm.vue'
import { GetCouponListApi } from '@/api/coupon'
import type { CouponTemplateItem } from '@/types/coupon'

const loading = ref(false)
const dialogVisible = ref(false)
const couponList = ref<CouponTemplateItem[]>([])
const total = ref(0)

const queryParams = ref({
  page: 1,
  page_size: 10,
  coupon_name: '',
  coupon_type: '',
})

const fetchData = async () => {
  loading.value = true
  try {
    const res = await GetCouponListApi({
      page: queryParams.value.page,
      page_size: queryParams.value.page_size,
      coupon_name: queryParams.value.coupon_name || undefined,
      coupon_type: queryParams.value.coupon_type || undefined,
    })
    const payload = res.data.data
    couponList.value = payload?.list || []
    total.value = payload?.total || 0
  } catch (error) {
    console.error('获取优惠券列表失败:', error)
    ElMessage.error('获取优惠券列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  queryParams.value.page = 1
  fetchData()
}

const handleReset = () => {
  queryParams.value = { page: 1, page_size: 10, coupon_name: '', coupon_type: '' }
  fetchData()
}

const handleAdd = () => {
  dialogVisible.value = true
}

const handleSuccess = () => {
  fetchData()
}

const formatDate = (t: string) => {
  if (!t) return '-'
  const d = new Date(t)
  if (isNaN(d.getTime())) return '-'
  return `${d.getFullYear()}/${d.getMonth() + 1}/${d.getDate()}`
}

onMounted(() => {
  fetchData()
})
</script>
