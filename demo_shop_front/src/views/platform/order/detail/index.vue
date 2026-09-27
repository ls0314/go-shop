<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center gap-4">
      <el-button class="h-9 rounded-xl" @click="$router.back()">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <h1 class="text-2xl font-bold text-slate-800 dark:text-white">订单详情</h1>
    </div>

    <!-- 加载 -->
    <div v-if="loading" class="flex items-center justify-center py-20">
      <el-icon class="animate-spin text-3xl text-slate-400"><Loading /></el-icon>
    </div>

    <template v-else-if="order">
      <!-- 订单概览 -->
      <div class="grid gap-4 md:grid-cols-2">
        <div class="rounded-xl border border-slate-200 bg-white p-5 dark:border-slate-700 dark:bg-slate-800">
          <h3 class="text-base font-semibold text-slate-800 dark:text-white">订单信息</h3>
          <div class="mt-3 space-y-2 text-sm text-slate-600 dark:text-slate-300">
            <p>订单号：{{ order.order_no }}</p>
            <p>下单用户：{{ order.username }}（ID: {{ order.user_id }}）</p>
            <p>
              订单状态：
              <el-tag :type="statusTagType(order.order_status)" size="small" class="rounded-lg">
                {{ statusLabel(order.order_status) }}
              </el-tag>
            </p>
            <p>下单时间：{{ order.created_at }}</p>
            <p>支付方式：{{ order.pay_method || '未支付' }}</p>
            <p>支付时间：{{ order.pay_time || '未支付' }}</p>
            <p v-if="order.buyer_remark">买家备注：{{ order.buyer_remark }}</p>
          </div>
        </div>

        <div class="rounded-xl border border-slate-200 bg-white p-5 dark:border-slate-700 dark:bg-slate-800">
          <h3 class="text-base font-semibold text-slate-800 dark:text-white">收货信息</h3>
          <div class="mt-3 text-sm text-slate-600 dark:text-slate-300">
            <p>{{ order.address_snapshot?.receiver_name }} {{ order.address_snapshot?.receiver_phone }}</p>
            <p class="mt-1">
              {{ order.address_snapshot?.province }}
              {{ order.address_snapshot?.city }}
              {{ order.address_snapshot?.district }}
              {{ order.address_snapshot?.detail_address }}
            </p>
            <p v-if="order.address_snapshot?.postal_code" class="mt-1 text-slate-400">
              邮编：{{ order.address_snapshot.postal_code }}
            </p>
          </div>

          <h3 class="mt-5 text-base font-semibold text-slate-800 dark:text-white">金额信息</h3>
          <div class="mt-3 text-sm text-slate-600 dark:text-slate-300">
            <p>商品总额：<span class="font-bold text-slate-800 dark:text-white">¥{{ order.total_amount?.toFixed(2) }}</span></p>
            <p>实付金额：<span class="text-lg font-bold" style="color: #ff6700">¥{{ order.pay_amount?.toFixed(2) }}</span></p>
          </div>

          <!-- 发货按钮 -->
          <div class="mt-4">
            <el-button
              v-if="order.order_status === 'paid'"
              type="primary"
              class="h-9 rounded-xl"
              @click="openShipDialog"
            >
              立即发货
            </el-button>
          </div>
        </div>
      </div>

      <!-- 商品明细 -->
      <div class="rounded-xl border border-slate-200 bg-white p-5 dark:border-slate-700 dark:bg-slate-800">
        <h3 class="text-base font-semibold text-slate-800 dark:text-white">商品明细</h3>
        <el-table
          :data="order.detail_list"
          border
          stripe
          class="mt-3"
          :header-cell-style="{ background: '#f8fafc' }"
        >
          <el-table-column label="商品图片" width="100">
            <template #default="{ row }">
              <div class="h-14 w-14 overflow-hidden rounded-lg bg-slate-100 dark:bg-slate-700">
                <img v-if="row.main_image" :src="row.main_image" class="h-full w-full object-cover" />
                <div v-else class="flex h-full w-full items-center justify-center text-slate-400">
                  <el-icon :size="18"><Picture /></el-icon>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="spu_name" label="商品名称" min-width="150" show-overflow-tooltip />
          <el-table-column prop="sku_name" label="SKU规格" min-width="150" show-overflow-tooltip />
          <el-table-column prop="unit_price" label="单价" width="100" align="right">
            <template #default="{ row }">¥{{ row.unit_price?.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column prop="quantity" label="数量" width="80" align="center" />
          <el-table-column prop="total_price" label="小计" width="110" align="right">
            <template #default="{ row }">¥{{ row.total_price?.toFixed(2) }}</template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 订单日志 -->
      <div class="rounded-xl border border-slate-200 bg-white p-5 dark:border-slate-700 dark:bg-slate-800">
        <h3 class="text-base font-semibold text-slate-800 dark:text-white">操作日志</h3>
        <el-table
          :data="order.log_list"
          border
          stripe
          class="mt-3"
          :header-cell-style="{ background: '#f8fafc' }"
        >
          <el-table-column prop="created_at" label="时间" width="180" />
          <el-table-column prop="operator" label="操作人" width="120" />
          <el-table-column label="操作" width="120">
            <template #default="{ row }">{{ logActionLabel(row.action) }}</template>
          </el-table-column>
          <el-table-column prop="order_status" label="订单状态" width="110">
            <template #default="{ row }">
              <el-tag :type="statusTagType(row.order_status)" size="small" class="rounded-lg">
                {{ statusLabel(row.order_status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="detail" label="详情" min-width="200" show-overflow-tooltip />
        </el-table>
      </div>
    </template>

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
          <el-input v-model="shipForm.express_company" placeholder="请输入快递公司名称" />
        </el-form-item>
        <el-form-item label="快递单号" required>
          <el-input v-model="shipForm.tracking_no" placeholder="请输入快递单号" />
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
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Picture, Loading } from '@element-plus/icons-vue'
import { useOrderStore } from '@/pinia/modules/order'
import type { AdminOrderDetailResp } from '@/types/order'

const route = useRoute()
const orderStore = useOrderStore()

const loading = ref(false)
const order = ref<AdminOrderDetailResp | null>(null)

const shipDialogVisible = ref(false)
const shipLoading = ref(false)
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

function logActionLabel(action: string) {
  const map: Record<string, string> = {
    create: '创建订单',
    pay: '支付成功',
    cancel: '取消订单',
    ship: '商家发货',
    confirm: '确认收货',
    auto_cancel: '系统自动取消',
  }
  return map[action] || action
}

async function fetchDetail() {
  loading.value = true
  try {
    const id = Number(route.params.id)
    order.value = await orderStore.GetAdminOrderDetail(id)
  } finally {
    loading.value = false
  }
}

function openShipDialog() {
  shipForm.value = { express_company: '', tracking_no: '' }
  shipDialogVisible.value = true
}

async function handleShip() {
  if (!order.value) return
  if (!shipForm.value.express_company || !shipForm.value.tracking_no) {
    ElMessage.warning('请填写完整的快递信息')
    return
  }
  shipLoading.value = true
  try {
    const ok = await orderStore.ShipOrder(order.value.order_id, shipForm.value)
    if (ok) {
      ElMessage.success('发货成功')
      shipDialogVisible.value = false
      fetchDetail()
    } else {
      ElMessage.error('发货失败')
    }
  } finally {
    shipLoading.value = false
  }
}

onMounted(() => fetchDetail())
</script>
