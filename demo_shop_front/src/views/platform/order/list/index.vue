<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-800 dark:text-white">订单管理</h1>
    </div>

    <!-- 筛选栏 -->
    <div class="rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-800">
      <el-form :model="queryParams" inline size="default" class="flex flex-wrap gap-3" @submit.prevent="handleSearch">
        <el-form-item label="订单号" class="mb-0">
          <el-input v-model="queryParams.order_no" placeholder="精确搜索" clearable class="w-52" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="订单状态" class="mb-0">
          <el-select v-model="queryParams.order_status" placeholder="全部" clearable class="w-32">
            <el-option label="待支付" value="pending_pay" />
            <el-option label="已支付" value="paid" />
            <el-option label="已发货" value="shipped" />
            <el-option label="已完成" value="completed" />
            <el-option label="已取消" value="cancelled" />
          </el-select>
        </el-form-item>
        <el-form-item label="下单时间" class="mb-0">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始"
            end-placeholder="结束"
            value-format="YYYY-MM-DD"
            class="w-56"
          />
        </el-form-item>
        <el-form-item class="mb-0">
          <el-button type="primary" :loading="loading" @click="handleSearch" class="h-9 rounded-xl">
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

    <!-- 订单表格 -->
    <div class="rounded-xl border border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800">
      <el-table :data="orderList" border stripe :loading="loading" :header-cell-style="{ background: '#f8fafc' }">
        <el-table-column prop="order_id" label="ID" width="80" align="center" />
        <el-table-column prop="order_no" label="订单号" min-width="190" show-overflow-tooltip />
        <el-table-column prop="user_name" label="下单用户" width="110" />
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.order_status)" size="small" class="rounded-lg">{{ statusLabel(row.order_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="total_amount" label="总金额" width="105" align="right">
          <template #default="{ row }">¥{{ row.total_amount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="pay_amount" label="实付" width="105" align="right">
          <template #default="{ row }">¥{{ row.pay_amount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="支付方式" width="100" align="center">
          <template #default="{ row }">
            <span v-if="row.pay_method">{{ payMethodLabel(row.pay_method) }}</span>
            <span v-else class="text-slate-300">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="receiver_name" label="收货人" width="90" />
        <el-table-column prop="receiver_phone" label="收货电话" width="125" />
        <el-table-column prop="created_at" label="下单时间" width="170" />
        <el-table-column label="操作" width="150" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="$router.push(`/platform/order/list/detail/${row.order_id}`)">详情</el-button>
            <el-button v-if="row.order_status === 'paid'" type="success" link size="small" @click="openShipDialog(row)">发货</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="flex justify-end p-4">
        <el-pagination
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.page_size"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          small
          @change="fetchList"
        />
      </div>
    </div>

    <!-- 发货弹窗 -->
    <el-dialog v-model="shipDialogVisible" title="订单发货" width="460px" draggable custom-class="rounded-xl" :close-on-click-modal="false">
      <el-form ref="shipFormRef" :model="shipForm" :rules="shipRules" label-width="100px">
        <el-form-item label="订单号"><span class="text-sm text-slate-700">{{ currentShipOrder?.order_no }}</span></el-form-item>
        <el-form-item label="收货人"><span class="text-sm text-slate-500">{{ currentShipOrder?.receiver_name }} {{ currentShipOrder?.receiver_phone }}</span></el-form-item>
        <el-form-item label="快递公司" prop="express_company">
          <el-select v-model="shipForm.express_company" class="w-full" filterable allow-create placeholder="选择或输入">
            <el-option v-for="c in expressCompanies" :key="c" :label="c" :value="c" />
          </el-select>
        </el-form-item>
        <el-form-item label="快递单号" prop="tracking_no">
          <el-input v-model="shipForm.tracking_no" placeholder="请输入快递单号" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="shipDialogVisible = false" class="h-9 rounded-xl">取消</el-button>
        <el-button type="primary" :loading="shipLoading" @click="handleShip" class="h-9 rounded-xl">确认发货</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Refresh } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'
import { useOrderStore } from '@/pinia/modules/order'
import type { AdminOrderItem } from '@/types/order'

const store = useOrderStore()
const loading = ref(false)
const dateRange = ref<string[] | null>(null)
const queryParams = ref({ page: 1, page_size: 20, order_status: '', order_no: '' })

// 响应式从 store 取数据
const orderList = computed(() => store.adminOrderList)
const total = computed(() => store.adminOrderTotal)

// ==================== 发货 ====================
const shipDialogVisible = ref(false)
const shipLoading = ref(false)
const currentShipOrder = ref<AdminOrderItem | null>(null)
const shipFormRef = ref<FormInstance>()
const shipForm = ref({ express_company: '', tracking_no: '' })
const shipRules: FormRules = {
  express_company: [{ required: true, message: '请输入快递公司', trigger: 'blur' }],
  tracking_no: [{ required: true, message: '请输入快递单号', trigger: 'blur' }],
}
const expressCompanies = ['顺丰速运', '中通快递', '圆通速递', '申通快递', '韵达快递', '京东物流', 'EMS']

// ==================== 数据获取 ====================
async function fetchList() {
  loading.value = true
  try {
    await store.GetAdminOrderList({
      page: queryParams.value.page,
      page_size: queryParams.value.page_size,
      order_status: queryParams.value.order_status || undefined,
      order_no: queryParams.value.order_no || undefined,
      start_time: dateRange.value?.[0] || undefined,
      end_time: dateRange.value?.[1] || undefined,
    })
  } finally { loading.value = false }
}

function handleSearch() { queryParams.value.page = 1; fetchList() }
function handleReset() { queryParams.value = { page: 1, page_size: 20, order_status: '', order_no: '' }; dateRange.value = null; fetchList() }

// ==================== 发货 ====================
function openShipDialog(row: AdminOrderItem) {
  currentShipOrder.value = row
  shipForm.value = { express_company: '', tracking_no: '' }
  shipFormRef.value?.clearValidate()
  shipDialogVisible.value = true
}

async function handleShip() {
  if (!shipFormRef.value || !currentShipOrder.value) return
  await shipFormRef.value.validate(async (valid) => {
    if (!valid) return
    shipLoading.value = true
    try {
      const ok = await store.ShipOrder(currentShipOrder.value!.order_id, shipForm.value)
      ok ? (ElMessage.success('发货成功'), shipDialogVisible.value = false, fetchList()) : ElMessage.error('发货失败')
    } finally { shipLoading.value = false }
  })
}

// ==================== 工具 ====================
const statusMap: Record<string, string> = { pending_pay: '待支付', paid: '已支付', shipped: '已发货', completed: '已完成', cancelled: '已取消' }
const tagMap: Record<string, string> = { pending_pay: 'warning', paid: 'success', shipped: 'primary', completed: 'info', cancelled: 'danger' }
const payMap: Record<string, string> = { mock: '模拟支付', alipay: '支付宝', wechat: '微信支付' }
function statusLabel(s: string) { return statusMap[s] || s }
function statusTag(s: string) { return tagMap[s] || '' }
function payMethodLabel(s: string) { return payMap[s] || s }

onMounted(() => fetchList())
</script>
