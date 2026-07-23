<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-800 dark:text-white">订单管理</h1>
    </div>

    <!-- 筛选栏 -->
    <div class="rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-800">
      <el-form :model="queryParams" inline size="default" class="flex flex-wrap gap-3" @submit.prevent="handleSearch">
        <el-form-item label="订单号" class="mb-0">
          <el-input v-model="queryParams.order_no" placeholder="请输入订单号" clearable class="w-56" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="订单状态" class="mb-0">
          <el-select v-model="queryParams.order_status" placeholder="全部" clearable class="w-36">
            <el-option label="待支付" value="pending_pay" />
            <el-option label="待发货" value="paid" />
            <el-option label="待收货" value="shipped" />
            <el-option label="已完成" value="completed" />
            <el-option label="已取消" value="cancelled" />
          </el-select>
        </el-form-item>
        <el-form-item label="开始时间" class="mb-0">
          <el-date-picker
            v-model="queryParams.start_time"
            type="date"
            placeholder="开始日期"
            value-format="YYYY-MM-DD"
            class="w-40"
          />
        </el-form-item>
        <el-form-item label="结束时间" class="mb-0">
          <el-date-picker
            v-model="queryParams.end_time"
            type="date"
            placeholder="结束日期"
            value-format="YYYY-MM-DD"
            class="w-40"
          />
        </el-form-item>
        <el-form-item class="mb-0">
          <el-button type="primary" :loading="loading" @click="handleSearch" class="h-9 rounded-xl">
            <el-icon class="mr-1"><Search /></el-icon>
            搜索
          </el-button>
          <el-button @click="handleReset" class="h-9 rounded-xl">
            <el-icon class="mr-1"><Refresh /></el-icon>
            重置
          </el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 订单表格 -->
    <div class="rounded-xl border border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800">
      <el-table
        :data="orderList"
        border
        stripe
        :loading="loading"
        :header-cell-style="{ background: '#f8fafc' }"
        class="dark:bg-slate-800"
      >
        <el-table-column prop="order_id" label="订单ID" width="80" align="center" />
        <el-table-column prop="order_no" label="订单号" min-width="180" show-overflow-tooltip />
        <el-table-column prop="user_name" label="下单用户" width="120" />
        <el-table-column label="订单状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.order_status)" size="small" class="rounded-lg">
              {{ statusLabel(row.order_status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="total_amount" label="总金额" width="110" align="right">
          <template #default="{ row }">¥{{ row.total_amount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="pay_amount" label="实付金额" width="110" align="right">
          <template #default="{ row }">¥{{ row.pay_amount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="receiver_name" label="收货人" width="100" show-overflow-tooltip />
        <el-table-column prop="receiver_phone" label="收货电话" width="130" show-overflow-tooltip />
        <el-table-column prop="created_at" label="下单时间" width="180" />
        <el-table-column label="操作" width="160" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="$router.push(`/platform/order/list/detail/${row.order_id}`)">
              详情
            </el-button>
            <el-button
              v-if="row.order_status === 'paid'"
              type="success"
              link
              size="small"
              @click="openShipDialog(row)"
            >
              发货
            </el-button>
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
          layout="total, sizes, prev, pager, next"
          background
          @size-change="fetchList"
          @current-change="fetchList"
        />
      </div>
    </div>

    <!-- 发货弹窗 -->
    <el-dialog
      v-model="shipDialogVisible"
      title="订单发货"
      draggable
      custom-class="rounded-xl"
      :close-on-click-modal="false"
      width="480px"
    >
      <el-form :model="shipForm" label-width="100px">
        <el-form-item label="快递公司" required>
          <el-input v-model="shipForm.express_company" placeholder="请输入快递公司名称" class="w-full" />
        </el-form-item>
        <el-form-item label="快递单号" required>
          <el-input v-model="shipForm.tracking_no" placeholder="请输入快递单号" class="w-full" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="shipDialogVisible = false" class="h-9 rounded-xl">取消</el-button>
        <el-button type="primary" :loading="shipLoading" @click="handleShip" class="h-9 rounded-xl">
          确认发货
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Refresh } from '@element-plus/icons-vue'
import { useOrderStore } from '@/pinia/modules/order'
import type { AdminOrderItem } from '@/types/order'

const orderStore = useOrderStore()

const loading = ref(false)
const orderList = ref<AdminOrderItem[]>([])
const total = ref(0)
const queryParams = ref({
  page: 1,
  page_size: 20,
  order_status: '',
  order_no: '',
  start_time: '',
  end_time: '',
})

// 发货相关
const shipDialogVisible = ref(false)
const shipLoading = ref(false)
const currentShipOrder = ref<AdminOrderItem | null>(null)
const shipForm = ref({ express_company: '', tracking_no: '' })

function statusTagType(status: string) {
  const map: Record<string, string> = {
    pending_pay: 'warning',
    paid: 'primary',
    shipped: 'success',
    completed: 'info',
    cancelled: 'danger',
  }
  return map[status] || 'info'
}

function statusLabel(status: string) {
  const map: Record<string, string> = {
    pending_pay: '待支付',
    paid: '待发货',
    shipped: '待收货',
    completed: '已完成',
    cancelled: '已取消',
  }
  return map[status] || status
}

async function fetchList() {
  loading.value = true
  try {
    const res = await orderStore.GetAdminOrderList({
      page: queryParams.value.page,
      page_size: queryParams.value.page_size,
      order_status: queryParams.value.order_status || undefined,
      order_no: queryParams.value.order_no || undefined,
      start_time: queryParams.value.start_time || undefined,
      end_time: queryParams.value.end_time || undefined,
    })
    if (res) {
      orderList.value = res.list || []
      total.value = res.total || 0
    }
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  queryParams.value.page = 1
  fetchList()
}

function handleReset() {
  queryParams.value = { page: 1, page_size: 20, order_status: '', order_no: '', start_time: '', end_time: '' }
  fetchList()
}

function openShipDialog(row: AdminOrderItem) {
  currentShipOrder.value = row
  shipForm.value = { express_company: '', tracking_no: '' }
  shipDialogVisible.value = true
}

async function handleShip() {
  if (!currentShipOrder.value) return
  if (!shipForm.value.express_company || !shipForm.value.tracking_no) {
    ElMessage.warning('请填写完整的快递信息')
    return
  }
  shipLoading.value = true
  try {
    const ok = await orderStore.ShipOrder(currentShipOrder.value.order_id, shipForm.value)
    if (ok) {
      ElMessage.success('发货成功')
      shipDialogVisible.value = false
      fetchList()
    } else {
      ElMessage.error('发货失败')
    }
  } finally {
    shipLoading.value = false
  }
}

onMounted(() => fetchList())
</script>
