<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-800 dark:text-white">部门管理</h1>
    </div>

    <!-- 操作栏 -->
    <div class="flex flex-wrap items-center justify-between gap-4 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-800">
      <p class="text-sm text-slate-500 dark:text-slate-400">部门层级用于数据权限范围控制</p>
      <el-button v-if="userStore.hasPerm('system:dept:create')" type="primary" class="h-9 rounded-xl" @click="handleAddRoot">
        <el-icon class="mr-1"><Plus /></el-icon>新增顶级部门
      </el-button>
    </div>

    <!-- 树形表格 -->
    <div class="rounded-xl border border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800">
      <el-table :data="deptTree" v-loading="loading" row-key="dept_id" border default-expand-all
                :tree-props="{ children: 'children' }" class="w-full"
                :header-cell-style="{ background: 'transparent' }" :cell-style="{ background: 'transparent' }">
        <el-table-column prop="dept_name" label="部门名称" min-width="200px" />
        <el-table-column label="类型" width="110px" align="center">
          <template #default="scope">
            <el-tag :type="scope.row.dept_type === 'platform' ? 'warning' : 'primary'" size="small" class="rounded-lg">
              {{ scope.row.dept_type === 'platform' ? '平台部门' : '普通部门' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90px" align="center">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'active' ? 'success' : 'danger'" size="small" class="rounded-lg">
              {{ scope.row.status === 'active' ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sort_order" label="排序" width="80px" align="center" />
        <el-table-column label="操作" width="230px" align="center" fixed="right">
          <template #default="scope">
            <el-button v-if="userStore.hasPerm('system:dept:create')" type="primary" link size="small" @click="handleAddChild(scope.row)">新增子部门</el-button>
            <el-button v-if="userStore.hasPerm('system:dept:update')" type="warning" link size="small" @click="handleEdit(scope.row)">编辑</el-button>
            <el-button v-if="userStore.hasPerm('system:dept:delete')" type="danger" link size="small" @click="handleDelete(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="formTitle" width="500px" custom-class="rounded-xl"
               :close-on-click-modal="false" @close="handleDialogClose">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="100px" class="pr-5">
        <el-form-item label="上级部门" prop="parent_id">
          <el-tree-select
              v-model="formData.parent_id"
              :data="parentOptions"
              node-key="dept_id"
              :props="{ label: 'dept_name', children: 'children' }"
              placeholder="不选则为顶级部门"
              clearable check-strictly class="w-full"
          />
        </el-form-item>
        <el-form-item label="部门名称" prop="dept_name">
          <el-input v-model="formData.dept_name" placeholder="请输入部门名称" maxlength="100" />
        </el-form-item>
        <el-form-item label="部门类型" prop="dept_type">
          <el-radio-group v-model="formData.dept_type">
            <el-radio value="platform">平台部门</el-radio>
            <el-radio value="normal">普通部门</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio value="active">启用</el-radio>
            <el-radio value="disabled">停用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="排序" prop="sort_order">
          <el-input-number v-model="formData.sort_order" :min="0" class="w-full" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false" class="h-9 rounded-xl">取消</el-button>
        <el-button type="primary" :loading="submitLoading" class="h-9 rounded-xl" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { useUserStore } from '@/pinia/modules/user'
import {
    GetDeptListApi, CreateDeptApi, UpdateDeptApi, DeleteDeptApi, fetchAll,
} from '@/api/rbac'
import type { SysDept } from '@/types/rbac'

const userStore = useUserStore()

const loading = ref(false)
const deptTree = ref<SysDept[]>([])

const fetchData = async () => {
    loading.value = true
    try {
        // 全量拉取(循环翻页),再组树
        const list = await fetchAll<SysDept>(GetDeptListApi)
        deptTree.value = buildDeptTree(list)
    } catch (error) {
        console.error('获取部门列表失败:', error)
        ElMessage.error('获取部门列表失败')
    } finally {
        loading.value = false
    }
}

const buildDeptTree = (list: SysDept[]): SysDept[] => {
    const map = new Map<number, SysDept>()
    list.forEach(item => map.set(item.dept_id, { ...item, children: [] }))
    const roots: SysDept[] = []
    map.forEach(item => {
        if (item.parent_id && item.parent_id !== 0 && map.has(item.parent_id)) {
            map.get(item.parent_id)!.children!.push(item)
        } else {
            roots.push(item)
        }
    })
    return roots
}

// ==================== 新增/编辑 ====================
const dialogVisible = ref(false)
const submitLoading = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance | null>(null)
const currentId = ref<number | null>(null)

const formData = ref({
    parent_id: 0,
    dept_name: '',
    dept_type: 'normal',
    status: 'active',
    sort_order: 0,
})

const formRules = {
    dept_name: [{ required: true, message: '请输入部门名称', trigger: 'blur' }],
}

const formTitle = computed(() => {
    if (isEdit.value) return '编辑部门'
    return formData.value.parent_id ? '新增子部门' : '新增顶级部门'
})

// 可选父级(排除自身及后代)
const parentOptions = computed(() => {
    const excludeIds = new Set<number>()
    if (isEdit.value && currentId.value) {
        const collect = (nodes: SysDept[]) => {
            nodes.forEach(n => { excludeIds.add(n.dept_id); if (n.children?.length) collect(n.children) })
        }
        const findAndCollect = (nodes: SysDept[]): boolean => {
            for (const n of nodes) {
                if (n.dept_id === currentId.value) { collect([n]); return true }
                if (n.children?.length && findAndCollect(n.children)) return true
            }
            return false
        }
        findAndCollect(deptTree.value)
    }
    const filter = (nodes: SysDept[]): SysDept[] => {
        return nodes
            .filter(n => !excludeIds.has(n.dept_id))
            .map(n => ({ ...n, children: n.children?.length ? filter(n.children) : undefined }))
    }
    return filter(deptTree.value)
})

const handleAddRoot = () => {
    isEdit.value = false
    currentId.value = null
    formData.value = { parent_id: 0, dept_name: '', dept_type: 'normal', status: 'active', sort_order: 0 }
    dialogVisible.value = true
}

const handleAddChild = (row: SysDept) => {
    isEdit.value = false
    currentId.value = null
    formData.value = { parent_id: row.dept_id, dept_name: '', dept_type: 'normal', status: 'active', sort_order: 0 }
    dialogVisible.value = true
}

const handleEdit = (row: SysDept) => {
    isEdit.value = true
    currentId.value = row.dept_id
    formData.value = {
        parent_id: row.parent_id,
        dept_name: row.dept_name,
        dept_type: row.dept_type || 'normal',
        status: row.status || 'active',
        sort_order: Number(row.sort_order) || 0,
    }
    dialogVisible.value = true
}

const handleDialogClose = () => { formRef.value?.resetFields() }

const handleSubmit = async () => {
    if (!formRef.value) return
    await formRef.value.validate(async (valid) => {
        if (!valid) return
        submitLoading.value = true
        try {
            const data = {
                parent_id: formData.value.parent_id || 0,
                dept_name: formData.value.dept_name,
                dept_type: formData.value.dept_type,
                status: formData.value.status,
                sort_order: formData.value.sort_order,
            }
            if (isEdit.value) {
                await UpdateDeptApi(currentId.value!, data)
                ElMessage.success('编辑成功')
            } else {
                await CreateDeptApi(data)
                ElMessage.success('新增成功')
            }
            dialogVisible.value = false
            fetchData()
        } catch (error: any) {
            ElMessage.error(error?.response?.data?.message || '操作失败')
        } finally {
            submitLoading.value = false
        }
    })
}

// ==================== 删除 ====================
const handleDelete = async (row: SysDept) => {
    try {
        await ElMessageBox.confirm(`确定要删除部门【${row.dept_name}】吗？其子部门将一并处理。`, '删除确认', { type: 'warning', customClass: 'rounded-xl' })
        await DeleteDeptApi(row.dept_id)
        ElMessage.success('删除成功')
        fetchData()
    } catch { /* 取消 */ }
}

onMounted(() => { fetchData() })
</script>
