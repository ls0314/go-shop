<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-800 dark:text-white">类目管理</h1>
    </div>

    <!-- 搜索与操作栏 -->
    <div class="flex flex-wrap items-center justify-between gap-4 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-800">
      <el-form :model="queryParams" inline size="default" class="flex flex-wrap gap-3" @submit.prevent="handleSearch">
        <el-form-item label="搜索类目" class="mb-0">
          <div class="relative" ref="searchContainerRef">
            <el-input
                v-model="queryParams.keyword"
                placeholder="请输入类目名称搜索"
                clearable
                class="w-72"
                @keyup.enter="handleSearch"
                @input="handleSearchInput"
                @clear="handleClear"
            />
            <!-- 搜索结果下拉列表 -->
            <div
                v-if="showSearchResults && searchResults.length > 0"
                class="absolute z-50 mt-1 w-full max-h-60 overflow-auto rounded-lg border border-slate-200 bg-white shadow-lg dark:border-slate-600 dark:bg-slate-800"
            >
              <div
                  v-for="item in searchResults"
                  :key="item.category_id"
                  class="px-3 py-2 text-sm cursor-pointer hover:bg-indigo-50 dark:hover:bg-slate-700"
                  :class="selectedSearchItem?.category_id === item.category_id ? 'bg-indigo-50 text-indigo-600 dark:bg-slate-700 dark:text-indigo-400' : ''"
                  @click.stop="handleSelectSearchResult(item)"
              >
                <div class="font-medium">{{ item.category_name }}</div>
                <div class="text-xs text-slate-500 dark:text-slate-400">{{ item.full_path }}</div>
              </div>
            </div>
          </div>
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

      <el-button type="primary" @click="handleAddRoot" class="h-9 rounded-xl">
        <el-icon class="mr-1"><Plus /></el-icon>
        新增顶级类目
      </el-button>
    </div>

    <!-- 树形表格主区域 -->
    <div class="rounded-xl border border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800">
      <el-table
          :data="categoryTree"
          row-key="category_id"
          border
          stripe
          :loading="loading"
          default-expand-all
          :tree-props="{ children: 'children', hasChildren: 'is_leaf' }"
          class="w-full"
          :header-cell-style="{ background: 'transparent' }"
          :cell-style="{ background: 'transparent' }"
      >
        <el-table-column prop="category_name" label="类目名称" min-width="220px" />
        <el-table-column prop="category_level" label="层级" width="80px" align="center" />
        <el-table-column prop="sort_order" label="排序" width="100px" align="center" />
        <el-table-column prop="status" label="状态" width="100px" align="center">
          <template #default="scope">
            <el-tag
                :type="scope.row.status === 'active' ? 'success' : 'danger'"
                size="small"
                class="rounded-lg"
            >
              {{ scope.row.status === 'active' ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="is_visible" label="可见性" width="100px" align="center">
          <template #default="scope">
            <el-tag
                :type="scope.row.is_visible ? 'primary' : 'info'"
                size="small"
                class="rounded-lg"
            >
              {{ scope.row.is_visible ? '显示' : '隐藏' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180px" align="center" />
        <el-table-column label="操作" width="300px" align="center" fixed="right">
          <template #default="scope">
            <el-button
                type="info"
                link
                size="small"
                @click="handleViewDetail(scope.row)"
                class="mr-1"
            >
              查看详情
            </el-button>
            <el-button
                type="primary"
                link
                size="small"
                @click="handleAddChild(scope.row)"
                class="mr-1"
            >
              新增子类目
            </el-button>
            <el-button
                type="warning"
                link
                size="small"
                @click="handleEdit(scope.row)"
                class="mr-1"
            >
              编辑
            </el-button>
            <el-button
                type="danger"
                link
                size="small"
                @click="handleDelete(scope.row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 新增/编辑类目弹窗 -->
    <CategoryForm
        v-model:visible="dialogVisible"
        :edit-data="currentEditData"
        :category-tree="categoryTree"
        @success="handleSuccess"
    />

    <!-- 类目详情弹窗 -->
    <el-dialog
        v-model="detailDialogVisible"
        title="类目详情"
        width="600px"
        draggable
        custom-class="rounded-xl"
        :close-on-click-modal="false"
    >
      <div v-if="categoryDetail" class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-1">
            <div class="text-sm text-slate-500 dark:text-slate-400">类目ID</div>
            <div class="font-medium">{{ categoryDetail.category_id }}</div>
          </div>
          <div class="space-y-1">
            <div class="text-sm text-slate-500 dark:text-slate-400">父类目ID</div>
            <div class="font-medium">{{ categoryDetail.parent_id }}</div>
          </div>
          <div class="space-y-1">
            <div class="text-sm text-slate-500 dark:text-slate-400">类目名称</div>
            <div class="font-medium">{{ categoryDetail.category_name }}</div>
          </div>
          <div class="space-y-1">
            <div class="text-sm text-slate-500 dark:text-slate-400">类目层级</div>
            <div class="font-medium">{{ categoryDetail.category_level }}</div>
          </div>
          <div class="space-y-1">
            <div class="text-sm text-slate-500 dark:text-slate-400">排序</div>
            <div class="font-medium">{{ categoryDetail.sort_order }}</div>
          </div>
          <div class="space-y-1">
            <div class="text-sm text-slate-500 dark:text-slate-400">是否叶子节点</div>
            <div class="font-medium">{{ categoryDetail.is_leaf ? '是' : '否' }}</div>
          </div>
          <div class="space-y-1">
            <div class="text-sm text-slate-500 dark:text-slate-400">可见性</div>
            <div class="font-medium">{{ categoryDetail.is_visible ? '显示' : '隐藏' }}</div>
          </div>
          <div class="space-y-1">
            <div class="text-sm text-slate-500 dark:text-slate-400">状态</div>
            <div class="font-medium">
              <el-tag :type="categoryDetail.status === 'active' ? 'success' : 'danger'" size="small">
                {{ categoryDetail.status === 'active' ? '启用' : '禁用' }}
              </el-tag>
            </div>
          </div>
        </div>

        <div class="space-y-1">
          <div class="text-sm text-slate-500 dark:text-slate-400">类目路径</div>
          <div class="font-medium">{{ categoryDetail.category_path }}</div>
        </div>

        <div v-if="categoryDetail.icon_url" class="space-y-1">
          <div class="text-sm text-slate-500 dark:text-slate-400">图标URL</div>
          <div class="font-medium break-all">{{ categoryDetail.icon_url }}</div>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-1">
            <div class="text-sm text-slate-500 dark:text-slate-400">创建时间</div>
            <div class="font-medium">{{ categoryDetail.created_at }}</div>
          </div>
          <div class="space-y-1">
            <div class="text-sm text-slate-500 dark:text-slate-400">更新时间</div>
            <div class="font-medium">{{ categoryDetail.updated_at }}</div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end">
          <el-button @click="detailDialogVisible = false" class="h-9 rounded-xl">
            关闭
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Search, Refresh, Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useCategoryStore } from '@/pinia/modules/category'
import CategoryForm from '@/components/category/categoryForm.vue'
import type { Category } from '@/types/category'

// 状态管理
const categoryStore = useCategoryStore()

// 本地状态
const loading = ref(false)
const dialogVisible = ref(false)
const detailDialogVisible = ref(false)
const currentEditData = ref<Category | null>(null)
const categoryDetail = ref<Category | null>(null)
const showSearchResults = ref(false)
const searchResults = ref<Array<Category & { full_path: string }>>([])
const searchContainerRef = ref<HTMLElement | null>(null)
// 保存当前选中的搜索结果
const selectedSearchItem = ref<(Category & { full_path: string }) | null>(null)

// 从Store获取完整类目树
const categoryTree = computed(() => categoryStore.categoryTree)

// 查询参数
const queryParams = ref({
  keyword: ''
})

// 页面初始化
onMounted(() => {
  fetchCategoryData()

  // 点击外部关闭搜索结果下拉框
  document.addEventListener('click', handleClickOutside)
})

// 页面卸载时移除事件监听
onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})

// 点击外部关闭搜索结果
const handleClickOutside = (e: MouseEvent) => {
  if (searchContainerRef.value && !searchContainerRef.value.contains(e.target as Node)) {
    showSearchResults.value = false
  }
}

// 获取类目树数据
const fetchCategoryData = async () => {
  loading.value = true
  try {
    await categoryStore.GetCategoryTree({ include_disabled: true })
  } catch (error) {
    console.error('获取类目数据失败:', error)
    ElMessage.error('获取类目数据失败')
  } finally {
    loading.value = false
  }
}

// 搜索输入时实时显示匹配结果
const handleSearchInput = () => {
  // 输入时清空已选中的搜索结果
  selectedSearchItem.value = null

  if (!queryParams.value.keyword.trim()) {
    searchResults.value = []
    showSearchResults.value = false
    return
  }

  // 在完整类目树中进行模糊匹配
  const results: Array<Category & { full_path: string }> = []
  const keyword = queryParams.value.keyword.toLowerCase()

  const traverse = (categories: Category[], parentPath: string) => {
    for (const cat of categories) {
      const fullPath = parentPath ? `${parentPath} > ${cat.category_name}` : cat.category_name

      if (cat.category_name.toLowerCase().includes(keyword)) {
        results.push({
          ...cat,
          full_path: fullPath
        })
      }

      if (cat.children && cat.children.length > 0) {
        traverse(cat.children, fullPath)
      }
    }
  }

  traverse(categoryStore.categoryTree, '')
  searchResults.value = results
  showSearchResults.value = results.length > 0
}

// 选择搜索结果
const handleSelectSearchResult = (item: Category & { full_path: string }) => {
  // 保存选中的搜索结果
  selectedSearchItem.value = item
  // 填充搜索框
  queryParams.value.keyword = item.category_name
  // 关闭下拉框
  showSearchResults.value = false
}

// 搜索按钮点击
const handleSearch = () => {
  // 关闭搜索结果下拉框
  showSearchResults.value = false

  // 检查是否有选中的搜索结果
  if (!selectedSearchItem.value) {
    ElMessage.warning('请先从搜索结果中选择一个类目')
    return
  }

  // 调用GetCategory获取完整详情并弹出弹窗
  handleViewDetail(selectedSearchItem.value)
}

// 清除搜索
const handleClear = () => {
  queryParams.value.keyword = ''
  searchResults.value = []
  showSearchResults.value = false
  selectedSearchItem.value = null
}

// 重置搜索
const handleReset = () => {
  queryParams.value = { keyword: '' }
  searchResults.value = []
  showSearchResults.value = false
  selectedSearchItem.value = null
  fetchCategoryData()
}

// 查看类目详情
const handleViewDetail = async (row: Category) => {
  loading.value = true
  try {
    const detail = await categoryStore.GetCategory(row.category_id)
    if (detail) {
      categoryDetail.value = detail
      detailDialogVisible.value = true
    } else {
      ElMessage.error('获取类目详情失败')
    }
  } catch (error) {
    console.error('获取类目详情失败:', error)
    ElMessage.error('获取类目详情失败')
  } finally {
    loading.value = false
  }
}

// 新增顶级类目
const handleAddRoot = () => {
  currentEditData.value = {
    parent_id: 0,
    category_name: '',
    sort_order: 0,
    is_visible: true,
    status: 'active'
  } as Category
  dialogVisible.value = true
}

// 新增子类目
const handleAddChild = (row: Category) => {
  currentEditData.value = {
    parent_id: row.category_id,
    category_name: '',
    sort_order: 0,
    is_visible: true,
    status: 'active'
  } as Category
  dialogVisible.value = true
}

// 编辑类目
const handleEdit = (row: Category) => {
  currentEditData.value = { ...row }
  dialogVisible.value = true
}

// 删除类目
const handleDelete = async (row: Category) => {
  try {
    // 如果有子类目禁止删除
    if (row.children && row.children.length > 0) {
      ElMessage.warning('该类目下存在子类目，无法删除！')
      return
    }

    await ElMessageBox.confirm(
        `确定要删除类目【${row.category_name}】吗？此操作不可恢复！`,
        '删除确认',
        {
          confirmButtonText: '确认删除',
          cancelButtonText: '取消',
          type: 'warning',
          draggable: true,
          customClass: 'rounded-xl'
        }
    )

    loading.value = true
    const success = await categoryStore.DeleteCategory(row.category_id)
    if (success) {
      ElMessage.success('删除成功')
      await fetchCategoryData()
    } else {
      ElMessage.error('删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  } finally {
    loading.value = false
  }
}

// 操作成功回调
const handleSuccess = () => {
  fetchCategoryData()
}
</script>
