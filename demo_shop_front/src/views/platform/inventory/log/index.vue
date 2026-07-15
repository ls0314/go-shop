<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-800 dark:text-white">库存变更日志</h1>
    </div>

    <!-- 搜索筛选栏 -->
    <div class="flex flex-wrap items-center gap-4 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-800">
      <el-form :model="logQuery" inline size="default" class="flex flex-wrap gap-3" @submit.prevent="handleSearch">
        <el-form-item label="SPU ID" class="mb-0">
          <el-input v-model.number="logQuery.spu_id" placeholder="可选" clearable class="w-36" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="SKU ID" class="mb-0">
          <el-input v-model.number="logQuery.sku_id" placeholder="可选" clearable class="w-36" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="变更类型" class="mb-0">
          <el-select v-model="logQuery.change_type" clearable class="w-40">
            <el-option label="手动调整" value="manual_adjust" />
            <el-option label="下单锁定" value="order_lock" />
            <el-option label="支付扣减" value="pay_deduct" />
            <el-option label="取消释放" value="order_release" />
            <el-option label="退款返还" value="refund_release" />
          </el-select>
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

    <!-- 日志表格 -->
    <div class="rounded-xl border border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800">
      <el-table
        :data="stockLogList" border stripe :loading="loading" row-key="log_id" class="w-full"
        :header-cell-style="{ background: '#f8fafc' }"
      >
        <el-table-column prop="log_id" label="日志ID" width="90px" align="center" />
        <el-table-column prop="spu_name" label="所属商品" width="150px" show-overflow-tooltip />
        <el-table-column prop="sku_name" label="SKU名称" width="150px" show-overflow-tooltip />
        <el-table-column label="变更类型" width="110px" align="center">
          <template #default="scope">
            <el-tag :type="changeTypeTag(scope.row.change_type)" size="small" class="rounded-lg">
              {{ changeTypeLabel(scope.row.change_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="变更量" width="90px" align="center">
          <template #default="scope">
            <span :class="scope.row.change_qty > 0 ? 'text-emerald-500' : 'text-rose-500'">
              {{ scope.row.change_qty > 0 ? '+' : '' }}{{ scope.row.change_qty }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="库存变化" width="180px" align="center">
          <template #default="scope">
            <span class="text-sm text-slate-500">
              可用 {{ scope.row.before_stock }} → {{ scope.row.after_stock }}
            </span>
            <span class="mx-1 text-slate-300">|</span>
            <span class="text-sm text-slate-500">
              锁定 {{ scope.row.before_lock }} → {{ scope.row.after_lock }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" width="160px" show-overflow-tooltip />
        <el-table-column prop="order_id" label="关联订单" width="100px" align="center">
          <template #default="scope">
            <span v-if="scope.row.order_id">{{ scope.row.order_id }}</span>
            <span v-else class="text-slate-300">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="操作时间" width="175px" align="center" />
      </el-table>

      <!-- 分页 -->
      <div class="flex justify-end p-4">
        <el-pagination
          v-model:current-page="logQuery.page"
          v-model:page-size="logQuery.page_size"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          small
          @change="handleSearch"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Search, Refresh } from '@element-plus/icons-vue'
import { useInventoryStore } from '@/pinia/modules/inventory'

const inventoryStore = useInventoryStore()
const loading = ref(false)

const logQuery = ref<Record<string, any>>({
  page: 1,
  page_size: 10,
  spu_id: undefined,
  sku_id: undefined,
  change_type: '',
})

const stockLogList = computed(() => inventoryStore.stockLogList)
const total = computed(() => inventoryStore.stockLogTotal)

async function handleSearch() {
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: logQuery.value.page,
      page_size: logQuery.value.page_size,
    }
    if (logQuery.value.sku_id) params.sku_id = logQuery.value.sku_id
    if (logQuery.value.spu_id) params.spu_id = logQuery.value.spu_id
    if (logQuery.value.change_type) params.change_type = logQuery.value.change_type
    await inventoryStore.GetStockLog(params)
  } finally {
    loading.value = false
  }
}

function handleReset() {
  logQuery.value = { page: 1, page_size: 10, spu_id: undefined, sku_id: undefined, change_type: '' }
  handleSearch()
}

const changeTypeMap: Record<string, string> = {
  manual_adjust: '手动调整',
  order_lock: '下单锁定',
  pay_deduct: '支付扣减',
  order_release: '取消释放',
  refund_release: '退款返还',
}

function changeTypeLabel(type: string): string {
  return changeTypeMap[type] || type
}

function changeTypeTag(type: string): string {
  switch (type) {
    case 'manual_adjust': return 'info'
    case 'order_lock': return 'warning'
    case 'pay_deduct': return 'danger'
    case 'order_release': return 'success'
    case 'refund_release': return 'success'
    default: return ''
  }
}

onMounted(() => {
  handleSearch()
})
</script>
