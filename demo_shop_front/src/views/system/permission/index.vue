<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-800 dark:text-white">权限管理</h1>
    </div>

    <!-- 搜索与操作栏 -->
    <div class="flex flex-wrap items-center justify-between gap-4 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-800">
      <el-form :model="queryParams" inline class="flex flex-wrap gap-3" @submit.prevent="handleSearch">
        <el-form-item label="类型" class="mb-0">
          <el-select v-model="queryParams.permType" placeholder="全部" clearable class="w-32" @change="handleSearch">
            <el-option label="API" value="api" />
            <el-option label="菜单" value="menu" />
            <el-option label="按钮" value="button" />
          </el-select>
        </el-form-item>
        <el-form-item class="mb-0">
          <el-button type="primary" :loading="loading" class="h-9 rounded-xl" @click="handleSearch">
            <el-icon class="mr-1"><Search /></el-icon>搜索
          </el-button>
          <el-button class="h-9 rounded-xl" @click="handleReset">
            <el-icon class="mr-1"><Refresh /></el-icon>重置
          </el-button>
        </el-form-item>
      </el-form>

      <el-button v-if="userStore.hasPerm('system:perm:create')" type="primary" class="h-9 rounded-xl" @click="handleAdd">
        <el-icon class="mr-1"><Plus /></el-icon>新增权限
      </el-button>
    </div>

    <!-- 表格 -->
    <div class="rounded-xl border border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800">
      <el-table :data="permList" v-loading="loading" border stripe class="w-full"
                :header-cell-style="{ background: 'transparent' }" :cell-style="{ background: 'transparent' }">
        <el-table-column prop="permission_id" label="ID" width="70px" align="center" />
        <el-table-column prop="permission_code" label="权限编码" min-width="180px" />
        <el-table-column prop="permission_name" label="权限名称" min-width="140px" />
        <el-table-column label="类型" width="90px" align="center">
          <template #default="scope">
            <el-tag :type="tagType(scope.row.permission_type)" size="small" class="rounded-lg">
              {{ scope.row.permission_type }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="request_method" label="方法" width="90px" align="center" />
        <el-table-column prop="api_path" label="API 路径" min-width="220px" show-overflow-tooltip />
        <el-table-column label="系统内置" width="90px" align="center">
          <template #default="scope">
            <el-tag v-if="scope.row.is_system" type="info" size="small" class="rounded-lg">是</el-tag>
            <span v-else class="text-slate-400">-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150px" align="center" fixed="right">
          <template #default="scope">
            <el-button v-if="userStore.hasPerm('system:perm:update')" type="primary" link size="small" @click="handleEdit(scope.row)">编辑</el-button>
            <el-button v-if="userStore.hasPerm('system:perm:delete')" type="danger" link size="small" @click="handleDelete(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="flex justify-end p-4">
        <el-pagination v-model:current-page="queryParams.page" v-model:page-size="queryParams.pageSize"
                       :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next" background
                       @change="fetchData" />
      </div>
    </div>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑权限' : '新增权限'" width="520px" custom-class="rounded-xl"
               :close-on-click-modal="false" @close="handleDialogClose">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="90px" class="pr-5">
        <el-form-item label="权限编码" prop="permission_code">
          <el-input v-model="formData.permission_code" placeholder="如 system:user:view" :disabled="isEdit" maxlength="100" />
        </el-form-item>
        <el-form-item label="权限名称" prop="permission_name">
          <el-input v-model="formData.permission_name" placeholder="请输入权限名称" maxlength="100" />
        </el-form-item>
        <el-form-item label="类型" prop="permission_type">
          <el-radio-group v-model="formData.permission_type">
            <el-radio value="api">API</el-radio>
            <el-radio value="menu">菜单</el-radio>
            <el-radio value="button">按钮</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="请求方法" prop="request_method">
          <el-select v-model="formData.request_method" clearable placeholder="请选择" class="w-full">
            <el-option v-for="m in ['GET', 'POST', 'PUT', 'DELETE']" :key="m" :label="m" :value="m" />
          </el-select>
        </el-form-item>
        <el-form-item label="API 路径" prop="api_path">
          <el-input v-model="formData.api_path" placeholder="如 /api/v1/user/:id" maxlength="500" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="formData.description" type="textarea" :rows="2" maxlength="200" />
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
import { ref, onMounted } from 'vue'
import { Search, Refresh, Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { useUserStore } from '@/pinia/modules/user'
import {
    GetPermissionListApi, CreatePermissionApi, UpdatePermissionApi, DeletePermissionApi,
} from '@/api/rbac'
import type { SysPermission } from '@/types/rbac'

const userStore = useUserStore()

const loading = ref(false)
const permList = ref<SysPermission[]>([])
const total = ref(0)

const queryParams = ref({ page: 1, pageSize: 10, permType: '' })

const fetchData = async () => {
    loading.value = true
    try {
        const res = await GetPermissionListApi({
            page: queryParams.value.page,
            pageSize: queryParams.value.pageSize,
            permType: queryParams.value.permType || undefined,
        })
        const payload = res.data.data
        permList.value = payload?.list || []
        total.value = payload?.total || 0
    } catch (error) {
        console.error('获取权限列表失败:', error)
        ElMessage.error('获取权限列表失败')
    } finally {
        loading.value = false
    }
}

const handleSearch = () => { queryParams.value.page = 1; fetchData() }
const handleReset = () => { queryParams.value = { page: 1, pageSize: 10, permType: '' }; fetchData() }

const tagType = (t: string) => ({ api: 'primary', menu: 'warning', button: 'success' }[t] || 'info') as any

// ==================== 新增/编辑 ====================
const dialogVisible = ref(false)
const submitLoading = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance | null>(null)
const currentId = ref<number | null>(null)

const formData = ref({
    permission_code: '',
    permission_name: '',
    permission_type: 'api',
    request_method: '',
    api_path: '',
    description: '',
})

const formRules = {
    permission_code: [{ required: true, message: '请输入权限编码', trigger: 'blur' }],
    permission_name: [{ required: true, message: '请输入权限名称', trigger: 'blur' }],
    api_path: [{ required: true, message: '请输入 API 路径', trigger: 'blur' }],
}

const handleAdd = () => {
    isEdit.value = false
    currentId.value = null
    formData.value = { permission_code: '', permission_name: '', permission_type: 'api', request_method: '', api_path: '', description: '' }
    dialogVisible.value = true
}

const handleEdit = (row: SysPermission) => {
    isEdit.value = true
    currentId.value = row.permission_id
    formData.value = {
        permission_code: row.permission_code,
        permission_name: row.permission_name,
        permission_type: row.permission_type,
        request_method: row.request_method,
        api_path: row.api_path,
        description: row.description,
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
                permission_code: formData.value.permission_code,
                permission_name: formData.value.permission_name,
                permission_type: formData.value.permission_type,
                request_method: formData.value.request_method || undefined,
                api_path: formData.value.api_path,
                description: formData.value.description || undefined,
            }
            if (isEdit.value) {
                await UpdatePermissionApi(currentId.value!, data)
                ElMessage.success('编辑成功')
            } else {
                await CreatePermissionApi(data)
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
const handleDelete = async (row: SysPermission) => {
    try {
        await ElMessageBox.confirm(`确定要删除权限【${row.permission_code}】吗？`, '删除确认', { type: 'warning', customClass: 'rounded-xl' })
        await DeletePermissionApi(row.permission_id)
        ElMessage.success('删除成功')
        fetchData()
    } catch { /* 取消 */ }
}

onMounted(() => { fetchData() })
</script>
