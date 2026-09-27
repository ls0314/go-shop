<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-800 dark:text-white">支付管理</h1>
    </div>

    <!-- 搜索与筛选栏 -->
    <div class="flex flex-wrap items-center gap-4 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-800">
      <el-form :model="query" inline size="default" class="flex flex-wrap gap-3" @submit.prevent="handleSearch">
        <el-form-item label="支付状态" class="mb-0">
          <el-select v-model="query.pay_status" clearable class="w-32">
            <el-option label="全部" value="" />
            <el-option label="待支付" value="pending" />
            <el-option label="已支付" value="success" />
            <el-option label="已失败" value="failed" />
            <el-option label="已关闭" value="closed" />
          </el-select>
        </el-form-item>
        <el-form-item label="支付方式" class="mb-0">
          <el-select v-model="query.pay_method" clearable class="w-32">
            <el-option label="全部" value="" />
            <el-option label="模拟支付" value="mock" />
            <el-option label="支付宝" value="alipay" />
            <el-option label="微信支付" value="wechat" />
          </el-select>
        </el-form-item>
        <el-form-item label="订单号" class="mb-0">
          <el-input v-model="query.order_no" placeholder="精确搜索" clearable class="w-52" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="支付时间" class="mb-0">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            class="w-56"
          />
        </el-form-item>
        <el-form-item class="mb-0">
          <el-button type="primary" @click="handleSearch" :loading="loading" class="h-9 rounded-xl">
            <el-icon class="mr-1"><Search /></el-icon>
            搜索
          </el-button>
          <el-button @click="handleReset" class="h-9 rounded-xl border-slate-200 bg-white text-slate-700 hover:bg-slate-50 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-200 dark:hover:bg-slate-600">
            <el-icon class="mr-1"><Refresh /></el-icon>
            重置
          </el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 数据表格 -->
    <div class="rounded-xl border border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800">
      <el-table :data="payList" border stripe :loading="loading" row-key="payment_id" class="w-full" :header-cell-style="{ background: '#f8fafc' }">
        <el-table-column prop="payment_id" label="支付ID" width="85px" align="center" />
        <el-table-column prop="pay_no" label="支付流水号" width="220px" show-overflow-tooltip />
        <el-table-column prop="order_no" label="订单号" width="200px" show-overflow-tooltip />
        <el-table-column prop="username" label="支付用户" width="110px" />
        <el-table-column label="支付方式" width="100px" align="center">
          <template #default="scope">
            <el-tag :type="methodTag(scope.row.pay_method)" size="small" class="rounded-lg">{{ methodLabel(scope.row.pay_method) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="支付金额" width="110px" align="right">
          <template #default="scope">¥{{ scope.row.pay_amount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="支付状态" width="90px" align="center">
          <template #default="scope">
            <el-tag :type="statusTag(scope.row.pay_status)" size="small" class="rounded-lg">{{ statusLabel(scope.row.pay_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="pay_time" label="支付时间" width="175px" align="center">
          <template #default="scope">
            <span v-if="scope.row.pay_time">{{ scope.row.pay_time }}</span>
            <span v-else class="text-slate-300">-</span>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="flex justify-end p-4">
        <el-pagination
          v-model:current-page="query.page"
          v-model:page-size="query.page_size"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          small
          @change="fetchData"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Search, Refresh } from '@element-plus/icons-vue'
import { GetPaymentListApi } from '@/api/payment'
import type { AdminPaymentItem } from '@/types/payment'

const loading = ref(false)
const payList = ref<AdminPaymentItem[]>([])
const total = ref(0)
const dateRange = ref<string[] | null>(null)

const query = ref<Record<string, any>>({
  page: 1,
  page_size: 20,
  pay_status: '',
  pay_method: '',
  order_no: '',
})

async function fetchData() {
  loading.value = true
  try {
    const res = await GetPaymentListApi({
      page: query.value.page,
      page_size: query.value.page_size,
      pay_status: query.value.pay_status || undefined,
      pay_method: query.value.pay_method || undefined,
      order_no: query.value.order_no || undefined,
      start_time: dateRange.value?.[0] || undefined,
      end_time: dateRange.value?.[1] || undefined,
    })
    const data = res.data.data
    payList.value = data?.list || []
    total.value = data?.total || 0
  } finally {
    loading.value = false
  }
}

function handleSearch() { query.value.page = 1; fetchData() }
function handleReset() {
  query.value = { page: 1, page_size: 20, pay_status: '', pay_method: '', order_no: '' }
  dateRange.value = null
  fetchData()
}

// ==================== 工具函数 ====================
const statusMap: Record<string, string> = { pending: '待支付', success: '已支付', failed: '已失败', closed: '已关闭' }
const statusTagMap: Record<string, string> = { pending: 'warning', success: 'success', failed: 'danger', closed: 'info' }
const methodMap: Record<string, string> = { mock: '模拟支付', alipay: '支付宝', wechat: '微信支付' }
const methodTagMap: Record<string, string> = { mock: 'info', alipay: 'primary', wechat: 'success' }

function statusLabel(s: string) { return statusMap[s] || s }
function statusTag(s: string) { return statusTagMap[s] || '' }
function methodLabel(s: string) { return methodMap[s] || s }
function methodTag(s: string) { return methodTagMap[s] || '' }

onMounted(() => { fetchData() })
</script>
