<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-800 dark:text-white">库存管理</h1>
    </div>

    <!-- Tab 切换 -->
    <el-tabs v-model="activeTab" class="rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-800">
      <el-tab-pane label="SKU库存" name="sku" />
      <el-tab-pane label="低库存预警" name="warn" />
    </el-tabs>

    <!-- ==================== SKU库存 Tab ==================== -->
    <template v-if="activeTab === 'sku'">
      <div class="flex flex-wrap items-center gap-4 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-800">
        <el-form :model="skuQuery" inline size="default" class="flex flex-wrap gap-3" @submit.prevent="handleSkuSearch">
          <el-form-item label="SPU ID" class="mb-0">
            <el-input v-model.number="skuQuery.spu_id" placeholder="请输入SPU ID" clearable class="w-48" @keyup.enter="handleSkuSearch" />
          </el-form-item>
          <el-form-item class="mb-0">
            <el-button type="primary" @click="handleSkuSearch" :loading="skuLoading" class="h-9 rounded-xl">
              <el-icon class="mr-1"><Search /></el-icon>
              查询
            </el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 库存汇总卡片 -->
      <div v-if="spuStockSummary" class="grid gap-4 md:grid-cols-3">
        <div class="rounded-xl border border-slate-200 bg-white p-5 dark:border-slate-700 dark:bg-slate-800">
          <p class="text-sm text-slate-500 dark:text-slate-400">总库存</p>
          <p class="mt-2 text-3xl font-bold text-blue-500">{{ spuStockSummary.total_stock }}</p>
        </div>
        <div class="rounded-xl border border-slate-200 bg-white p-5 dark:border-slate-700 dark:bg-slate-800">
          <p class="text-sm text-slate-500 dark:text-slate-400">锁定库存</p>
          <p class="mt-2 text-3xl font-bold text-amber-500">{{ spuStockSummary.total_lock }}</p>
        </div>
        <div class="rounded-xl border border-slate-200 bg-white p-5 dark:border-slate-700 dark:bg-slate-800">
          <p class="text-sm text-slate-500 dark:text-slate-400">总销量</p>
          <p class="mt-2 text-3xl font-bold text-emerald-500">{{ spuStockSummary.total_sold }}</p>
        </div>
      </div>

      <!-- SKU库存表格 -->
      <div class="rounded-xl border border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800">
        <el-table
          :data="skuList" border stripe :loading="skuLoading" row-key="sku_id" class="w-full"
          :header-cell-style="{ background: '#f8fafc' }"
        >
          <el-table-column prop="sku_id" label="SKU ID" width="80px" align="center" />
          <el-table-column prop="spu_name" label="所属商品" width="160px" show-overflow-tooltip />
          <el-table-column prop="sku_name" label="SKU名称" min-width="160px" show-overflow-tooltip />
          <el-table-column prop="stock" label="可用库存" width="110px" align="center" />
          <el-table-column prop="lock_stock" label="锁定库存" width="110px" align="center" />
          <el-table-column prop="total_stock" label="总库存" width="110px" align="center" />
          <el-table-column prop="sold_count" label="销量" width="90px" align="center" />
          <el-table-column prop="sku_status" label="SKU状态" width="100px" align="center">
            <template #default="scope">
              <el-tag :type="scope.row.sku_status === 'active' ? 'success' : 'danger'" size="small" class="rounded-lg">
                {{ scope.row.sku_status === 'active' ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120px" align="center" fixed="right">
            <template #default="scope">
              <el-button type="primary" link size="small" @click="openAdjustDialog(scope.row)">调整库存</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div v-if="!skuLoading && skuQuery.spu_id && skuList.length === 0" class="py-12 text-center text-sm text-slate-400">
          暂无数据，请输入有效SPU ID查询
        </div>
      </div>
    </template>

    <!-- ==================== 低库存预警 Tab ==================== -->
    <template v-if="activeTab === 'warn'">
      <div class="flex flex-wrap items-center gap-4 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-800">
        <el-form :model="warnQuery" inline size="default" class="flex flex-wrap gap-3" @submit.prevent="handleWarnSearch">
          <el-form-item label="预警阈值" class="mb-0">
            <el-input-number v-model="warnQuery.threshold" :min="1" class="w-36" />
          </el-form-item>
          <el-form-item label="SPU状态" class="mb-0">
            <el-select v-model="warnQuery.spu_status" class="w-36">
              <el-option label="已上架" value="published" />
              <el-option label="全部" value="" />
            </el-select>
          </el-form-item>
          <el-form-item class="mb-0">
            <el-button type="primary" @click="handleWarnSearch" :loading="warnLoading" class="h-9 rounded-xl">
              <el-icon class="mr-1"><Search /></el-icon>
              查询
            </el-button>
          </el-form-item>
        </el-form>
      </div>

      <div class="rounded-xl border border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800">
        <el-table
          :data="warnStockList" border stripe :loading="warnLoading" row-key="sku_id" class="w-full"
          :header-cell-style="{ background: '#f8fafc' }"
        >
          <el-table-column prop="sku_id" label="SKU ID" width="80px" align="center" />
          <el-table-column prop="spu_name" label="所属商品" width="160px" show-overflow-tooltip />
          <el-table-column prop="sku_name" label="SKU名称" min-width="160px" show-overflow-tooltip />
          <el-table-column prop="stock" label="可用库存" width="110px" align="center">
            <template #default="scope">
              <span class="font-bold text-rose-500">{{ scope.row.stock }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="sold_count" label="销量" width="90px" align="center" />
          <el-table-column prop="sku_status" label="状态" width="100px" align="center">
            <template #default="scope">
              <el-tag :type="scope.row.sku_status === 'active' ? 'success' : 'danger'" size="small" class="rounded-lg">
                {{ scope.row.sku_status === 'active' ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120px" align="center" fixed="right">
            <template #default="scope">
              <el-button type="primary" link size="small" @click="openAdjustDialog(scope.row)">调整库存</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </template>

    <!-- ==================== 库存调整弹窗 ==================== -->
    <el-dialog
      v-model="adjustDialogVisible"
      title="库存调整"
      width="460px"
      draggable
      custom-class="rounded-xl"
      :close-on-click-modal="false"
    >
      <el-form ref="adjustFormRef" :model="adjustForm" :rules="adjustRules" label-width="110px">
        <el-form-item label="SKU名称">
          <span class="font-medium text-slate-700 dark:text-slate-200">{{ adjustTarget?.sku_name }}</span>
        </el-form-item>
        <el-form-item label="所属商品">
          <span class="text-slate-500">{{ adjustTarget?.spu_name }}</span>
        </el-form-item>
        <el-form-item label="当前可用库存">
          <span class="font-bold text-blue-500">{{ adjustTarget?.stock }}</span>
        </el-form-item>
        <el-form-item label="变更数量" prop="change_qty">
          <el-input-number v-model="adjustForm.change_qty" :min="-999999" class="w-full" placeholder="正数增加，负数减少" />
        </el-form-item>
        <el-form-item v-if="adjustResult" label="调整结果">
          <span class="text-sm">
            {{ adjustResult.before_stock }} →
            <span class="font-bold text-emerald-500">{{ adjustResult.after_stock }}</span>
          </span>
        </el-form-item>
        <el-form-item label="调整原因" prop="remark">
          <el-input v-model="adjustForm.remark" type="textarea" :rows="2" placeholder="请输入调整原因（必填）" maxlength="200" show-word-limit />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="adjustDialogVisible = false" class="h-9 rounded-xl">取消</el-button>
        <el-button type="primary" @click="handleAdjustSubmit" :loading="adjustSubmitting" class="h-9 rounded-xl">确认调整</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { useInventoryStore } from '@/pinia/modules/inventory'
import type { SkuInventory, InventoryWarnItem } from '@/types/inventory'

const inventoryStore = useInventoryStore()

const activeTab = ref('sku')

// SKU 查询
const skuQuery = ref({ spu_id: undefined as number | undefined })
const skuLoading = ref(false)

// 低库存预警查询
const warnQuery = ref({ threshold: 10, spu_status: 'published' })
const warnLoading = ref(false)

// 库存调整弹窗
const adjustDialogVisible = ref(false)
const adjustTarget = ref<SkuInventory | InventoryWarnItem | null>(null)
const adjustSubmitting = ref(false)
const adjustFormRef = ref<FormInstance>()
const adjustForm = ref({ change_qty: 0, remark: '' })
const adjustResult = ref<{ before_stock: number; after_stock: number } | null>(null)
const adjustRules: FormRules = {
  remark: [{ required: true, message: '请输入调整原因', trigger: 'blur' }],
}

// 响应式数据
const skuList = computed(() => inventoryStore.skuListBySpu)
const spuStockSummary = computed(() => inventoryStore.spuStockSummary)
const warnStockList = computed(() => inventoryStore.warnStockList)

// SKU 库存查询
async function handleSkuSearch() {
  if (!skuQuery.value.spu_id) { ElMessage.warning('请输入SPU ID'); return }
  skuLoading.value = true
  try { await inventoryStore.GetSkuListBySpu(skuQuery.value.spu_id) }
  finally { skuLoading.value = false }
}

// 低库存预警查询
async function handleWarnSearch() {
  warnLoading.value = true
  try {
    await inventoryStore.GetWarnStock({
      threshold: warnQuery.value.threshold,
      spu_status: warnQuery.value.spu_status || undefined,
    })
  }
  finally { warnLoading.value = false }
}

// 打开调整弹窗
function openAdjustDialog(row: SkuInventory | InventoryWarnItem) {
  adjustTarget.value = row
  adjustForm.value = { change_qty: 0, remark: '' }
  adjustResult.value = null
  adjustDialogVisible.value = true
}

// 提交库存调整
async function handleAdjustSubmit() {
  if (!adjustFormRef.value) return
  await adjustFormRef.value.validate(async (valid) => {
    if (!valid || !adjustTarget.value) return
    adjustSubmitting.value = true
    try {
      const result = await inventoryStore.AdjustStock({
        sku_id: adjustTarget.value.sku_id,
        change_qty: adjustForm.value.change_qty,
        remark: adjustForm.value.remark,
      })
      if (result) {
        adjustResult.value = result
        ElMessage.success('库存调整成功')
        if (activeTab.value === 'sku' && skuQuery.value.spu_id) await handleSkuSearch()
        else if (activeTab.value === 'warn') await handleWarnSearch()
        setTimeout(() => { adjustDialogVisible.value = false }, 1500)
      } else {
        ElMessage.error('库存调整失败')
      }
    }
    finally { adjustSubmitting.value = false }
  })
}
</script>
