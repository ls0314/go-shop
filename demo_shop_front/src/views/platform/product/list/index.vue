<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-800 dark:text-white">商品管理</h1>
      <el-button type="primary" @click="handleCreate" class="h-9 rounded-xl">
        <el-icon class="mr-1"><Plus /></el-icon>
        新增商品
      </el-button>
    </div>

    <!-- 搜索与筛选栏 -->
    <div class="flex flex-wrap items-center gap-4 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-800">
      <el-form :model="queryParams" inline size="default" class="flex flex-wrap gap-3" @submit.prevent="handleSearch">
        <el-form-item label="商品名称" class="mb-0">
          <el-input v-model="queryParams.spu_name" placeholder="模糊搜索" clearable class="w-48" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="商品状态" class="mb-0">
          <el-select v-model="queryParams.spu_status" class="w-44">
            <el-option label="全部" value="" />
            <el-option label="草稿" value="draft" />
            <el-option label="已上架" value="published" />
            <el-option label="已下架" value="withdrawn" />
          </el-select>
        </el-form-item>
        <el-form-item label="品牌" class="mb-0">
          <el-input v-model="queryParams.brand" placeholder="品牌筛选" clearable class="w-40" @keyup.enter="handleSearch" />
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
      <el-table
        :data="productList"
        border
        stripe
        :loading="loading"
        class="w-full"
        :header-cell-style="{ background: '#f8fafc' }"
      >
        <el-table-column prop="spu_id" label="ID" width="80px" align="center" />
        <el-table-column prop="spu_name" label="商品名称" min-width="180px" show-overflow-tooltip />
        <el-table-column prop="category_name" label="所属类目" width="140px" show-overflow-tooltip />
        <el-table-column prop="brand" label="品牌" width="100px" />
        <el-table-column label="主图" width="90px" align="center">
          <template #default="scope">
            <el-image
              v-if="scope.row.main_image"
              :src="scope.row.main_image"
              :preview-src-list="[scope.row.main_image]"
              preview-teleported
              fit="cover"
              class="h-10 w-10 rounded-lg"
            />
            <span v-else class="text-slate-400 text-xs">无</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100px" align="center">
          <template #default="scope">
            <el-tag
              :type="statusTagType(scope.row.spu_status)"
              size="small"
              class="rounded-lg"
            >
              {{ statusLabel(scope.row.spu_status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="最低价" width="100px" align="right">
          <template #default="scope">¥{{ scope.row.min_price }}</template>
        </el-table-column>
        <el-table-column label="最高价" width="100px" align="right">
          <template #default="scope">¥{{ scope.row.max_price }}</template>
        </el-table-column>
        <el-table-column prop="total_stock" label="总库存" width="90px" align="center" />
        <el-table-column prop="total_sold" label="总销量" width="90px" align="center" />
        <el-table-column prop="priority" label="排序" width="80px" align="center" />
        <el-table-column prop="created_at" label="创建时间" width="170px" align="center" />
        <el-table-column prop="updated_at" label="更新时间" width="170px" align="center" />
        <el-table-column label="操作" width="280px" align="center" fixed="right">
          <template #default="scope">
            <el-button type="info" link size="small" @click="handleViewDetail(scope.row)">详情</el-button>
            <el-button type="primary" link size="small" @click="handleEdit(scope.row)">编辑</el-button>
            <el-button
              v-if="scope.row.spu_status === 'draft' || scope.row.spu_status === 'withdrawn'"
              type="success"
              link
              size="small"
              @click="handlePublish(scope.row)"
            >上架</el-button>
            <el-button
              v-if="scope.row.spu_status === 'published'"
              type="warning"
              link
              size="small"
              @click="handleWithdraw(scope.row)"
            >下架</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="flex justify-end p-4">
        <el-pagination
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.pageSize"
          :total="productTotal"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          background
          @size-change="handleSearch"
          @current-change="handleSearch"
        />
      </div>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog
      v-model="detailVisible"
      title="商品详情"
      width="800px"
      draggable
      custom-class="rounded-xl"
      :close-on-click-modal="false"
    >
      <div v-if="currentProduct" class="space-y-4">
        <div class="grid grid-cols-3 gap-4">
          <div class="space-y-1">
            <div class="text-sm text-slate-500 dark:text-slate-400">商品ID</div>
            <div class="font-medium">{{ currentProduct.spu_id }}</div>
          </div>
          <div class="space-y-1">
            <div class="text-sm text-slate-500 dark:text-slate-400">商品名称</div>
            <div class="font-medium">{{ currentProduct.spu_name }}</div>
          </div>
          <div class="space-y-1">
            <div class="text-sm text-slate-500 dark:text-slate-400">品牌</div>
            <div class="font-medium">{{ currentProduct.brand || '-' }}</div>
          </div>
          <div class="space-y-1">
            <div class="text-sm text-slate-500 dark:text-slate-400">所属类目</div>
            <div class="font-medium">{{ currentProduct.category_name }}</div>
          </div>
          <div class="space-y-1">
            <div class="text-sm text-slate-500 dark:text-slate-400">状态</div>
            <div class="font-medium">
              <el-tag :type="statusTagType(currentProduct.spu_status)" size="small">{{ statusLabel(currentProduct.spu_status) }}</el-tag>
            </div>
          </div>
          <div class="space-y-1">
            <div class="text-sm text-slate-500 dark:text-slate-400">排序权重</div>
            <div class="font-medium">{{ currentProduct.priority }}</div>
          </div>
        </div>

        <div v-if="currentProduct.description" class="space-y-1">
          <div class="text-sm text-slate-500 dark:text-slate-400">商品描述</div>
          <div class="text-sm" v-html="currentProduct.description" />
        </div>

        <div class="space-y-2">
          <div class="text-sm font-medium text-slate-700 dark:text-slate-300">SKU 列表（{{ currentProduct.sku_list?.length || 0 }} 个）</div>
          <el-table :data="currentProduct.sku_list" border size="small" max-height="300">
            <el-table-column prop="sku_name" label="SKU名称" min-width="140" show-overflow-tooltip />
            <el-table-column label="规格值" width="160">
              <template #default="{ row }">
                <span class="text-xs">{{ JSON.stringify(row.spec_values) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="price" label="售价" width="90" align="right" />
            <el-table-column prop="cost_price" label="成本价" width="90" align="right" />
            <el-table-column prop="stock" label="库存" width="70" align="center" />
            <el-table-column prop="sold_count" label="销量" width="70" align="center" />
            <el-table-column prop="sku_status" label="状态" width="80" align="center">
              <template #default="{ row }">
                <el-tag :type="row.sku_status === 'active' ? 'success' : 'info'" size="small">{{ row.sku_status === 'active' ? '启用' : '禁用' }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <div v-if="currentProduct.image_list?.length" class="space-y-2">
          <div class="text-sm font-medium text-slate-700 dark:text-slate-300">商品图片</div>
          <div class="flex flex-wrap gap-2">
            <div v-for="img in currentProduct.image_list" :key="img.image_id" class="relative">
              <el-image :src="img.image_url" fit="cover" class="h-20 w-20 rounded-lg" />
              <span v-if="img.is_main" class="absolute bottom-0 left-0 right-0 bg-indigo-500 text-white text-xs text-center rounded-b-lg">主图</span>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Search, Refresh, Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useProductStore } from '@/pinia/modules/product'
import type { SpuListItem } from '@/types/product'

const router = useRouter()

const productStore = useProductStore()

// 从 store 取状态
const productList = ref<SpuListItem[]>([])
const productTotal = ref(0)
const loading = ref(false)
const detailVisible = ref(false)
const currentProduct = ref<any>(null)

// 查询参数
const queryParams = ref({
  page: 1,
  pageSize: 10,
  spu_name: '',
  spu_status: '',
  brand: '',
})

// 状态标签
const statusTagType = (status: string) => {
  const map: Record<string, string> = { draft: 'info', published: 'success', withdrawn: 'warning' }
  return map[status] || 'info'
}
const statusLabel = (status: string) => {
  const map: Record<string, string> = { draft: '草稿', published: '已上架', withdrawn: '已下架' }
  return map[status] || status
}

// 获取列表
const fetchList = async () => {
  loading.value = true
  try {
    const data = await productStore.GetProductList({
      page: queryParams.value.page,
      // 下划线命名:后端 SpuQueryReq 的 form 标签是 page_size
      page_size: queryParams.value.pageSize,
      spu_name: queryParams.value.spu_name || undefined,
      spu_status: queryParams.value.spu_status || undefined,
      brand: queryParams.value.brand || undefined,
    })
    if (data) {
      productList.value = data.list
      productTotal.value = data.total
    }
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  queryParams.value.page = 1
  fetchList()
}

// 重置
const handleReset = () => {
  queryParams.value = { page: 1, pageSize: 10, spu_name: '', spu_status: '', brand: '' }
  fetchList()
}

// 查看详情
const handleViewDetail = async (row: SpuListItem) => {
  loading.value = true
  try {
    const detail = await productStore.GetProduct(row.spu_id)
    if (detail) {
      currentProduct.value = detail
      detailVisible.value = true
    } else {
      ElMessage.error('获取商品详情失败')
    }
  } finally {
    loading.value = false
  }
}

// 新增
const handleCreate = () => {
  router.push('/platform/product/create')
}

// 编辑
const handleEdit = (row: SpuListItem) => {
  router.push(`/platform/product/create/${row.spu_id}`)
}

// 上架
const handlePublish = async (row: SpuListItem) => {
  const ok = await productStore.PublishProduct(row.spu_id)
  if (ok) {
    ElMessage.success('上架成功')
    fetchList()
  } else {
    ElMessage.error('上架失败')
  }
}

// 下架
const handleWithdraw = async (row: SpuListItem) => {
  const ok = await productStore.WithdrawProduct(row.spu_id)
  if (ok) {
    ElMessage.success('下架成功')
    fetchList()
  } else {
    ElMessage.error('下架失败')
  }
}

// 删除
const handleDelete = async (row: SpuListItem) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除商品【${row.spu_name}】吗？删除后可在数据库中恢复。`,
      '删除确认',
      { confirmButtonText: '确认删除', cancelButtonText: '取消', type: 'warning', draggable: true }
    )
    const ok = await productStore.DeleteProduct(row.spu_id)
    if (ok) {
      ElMessage.success('删除成功')
      fetchList()
    } else {
      ElMessage.error('删除失败')
    }
  } catch {
    // 取消
  }
}

onMounted(() => {
  fetchList()
})
</script>
