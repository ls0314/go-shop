<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-800 dark:text-white">数据权限</h1>
    </div>

    <!-- 搜索与操作栏 -->
    <div class="flex flex-wrap items-center justify-between gap-4 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-800">
      <el-form :model="queryParams" inline class="flex flex-wrap gap-3" @submit.prevent="handleSearch">
        <el-form-item label="资源类型" class="mb-0">
          <el-select v-model="queryParams.resourceType" placeholder="全部" clearable class="w-32" @change="handleSearch">
            <el-option v-for="t in resourceTypes" :key="t.value" :label="t.label" :value="t.value" />
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

      <el-button v-if="userStore.hasPerm('system:scope:create')" type="primary" class="h-9 rounded-xl" @click="handleAdd">
        <el-icon class="mr-1"><Plus /></el-icon>新增规则
      </el-button>
    </div>

    <!-- 表格 -->
    <div class="rounded-xl border border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800">
      <el-table :data="scopeList" v-loading="loading" border stripe class="w-full"
                :header-cell-style="{ background: 'transparent' }" :cell-style="{ background: 'transparent' }">
        <el-table-column prop="scope_id" label="ID" width="70px" align="center" />
        <el-table-column label="角色" min-width="130px">
          <template #default="scope">
            {{ roleNameOf(scope.row.role_id) }}
          </template>
        </el-table-column>
        <el-table-column label="资源类型" width="120px" align="center">
          <template #default="scope">
            <el-tag size="small" class="rounded-lg">{{ scope.row.resource_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="field_name" label="字段" width="110px" align="center" />
        <el-table-column label="条件" width="110px" align="center">
          <template #default="scope">
            <el-tag :type="scope.row.condition_type === 'eq' ? 'primary' : 'warning'" size="small" class="rounded-lg">
              {{ scope.row.condition_type }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="condition_value" label="条件值" min-width="140px" show-overflow-tooltip />
        <el-table-column prop="description" label="描述" min-width="180px" show-overflow-tooltip />
        <el-table-column label="操作" width="150px" align="center" fixed="right">
          <template #default="scope">
            <el-button v-if="userStore.hasPerm('system:scope:update')" type="primary" link size="small" @click="handleEdit(scope.row)">编辑</el-button>
            <el-button v-if="userStore.hasPerm('system:scope:delete')" type="danger" link size="small" @click="handleDelete(scope.row)">删除</el-button>
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
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑规则' : '新增规则'" width="520px" custom-class="rounded-xl"
               :close-on-click-modal="false" @close="handleDialogClose">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="100px" class="pr-5">
        <el-form-item label="角色" prop="role_id">
          <el-select v-model="formData.role_id" placeholder="请选择角色" filterable class="w-full" :disabled="isEdit">
            <el-option v-for="role in roleOptions" :key="role.role_id" :label="role.role_name" :value="role.role_id" />
          </el-select>
        </el-form-item>
        <el-form-item label="资源类型" prop="resource_type">
          <el-select v-model="formData.resource_type" class="w-full">
            <el-option v-for="t in resourceTypes" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="字段名" prop="field_name">
          <el-input v-model="formData.field_name" placeholder="如 dept_id / user_id" maxlength="50" />
        </el-form-item>
        <el-form-item label="条件" prop="condition_type">
          <el-select v-model="formData.condition_type" class="w-full">
            <el-option label="等于(=)" value="eq" />
            <el-option label="属于(in)" value="in" />
            <el-option label="包含(like)" value="like" />
          </el-select>
        </el-form-item>
        <el-form-item label="条件值" prop="condition_value">
          <el-input v-model="formData.condition_value" placeholder="多个值用逗号分隔" maxlength="200" />
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
    GetScopeListApi, CreateScopeApi, UpdateScopeApi, DeleteScopeApi, GetRoleListApi, fetchAll,
} from '@/api/rbac'
import type { SysScope, SysRole } from '@/types/rbac'

const userStore = useUserStore()

const resourceTypes = [
    { label: '部门数据', value: 'dept' },
    { label: '商品数据', value: 'product' },
    { label: '订单数据', value: 'order' },
]

const loading = ref(false)
const scopeList = ref<SysScope[]>([])
const total = ref(0)
const roleOptions = ref<SysRole[]>([])
const roleMap = ref<Record<number, string>>({})

const queryParams = ref({ page: 1, pageSize: 10, resourceType: '' })

const fetchData = async () => {
    loading.value = true
    try {
        const res = await GetScopeListApi({
            page: queryParams.value.page,
            pageSize: queryParams.value.pageSize,
            resourceType: queryParams.value.resourceType || undefined,
        })
        const payload = res.data.data
        scopeList.value = payload?.list || []
        total.value = payload?.total || 0
    } catch (error) {
        console.error('获取数据权限失败:', error)
        ElMessage.error('获取数据权限失败')
    } finally {
        loading.value = false
    }
}

const loadRoles = async () => {
    try {
        const allRoles = await fetchAll<SysRole>(GetRoleListApi)
        roleOptions.value = allRoles
        roleMap.value = {}
        roleOptions.value.forEach(r => { roleMap.value[r.role_id] = r.role_name })
    } catch (error) {
        console.error('加载角色失败:', error)
    }
}

const roleNameOf = (roleId: number) => roleMap.value[roleId] || `#${roleId}`

const handleSearch = () => { queryParams.value.page = 1; fetchData() }
const handleReset = () => { queryParams.value = { page: 1, pageSize: 10, resourceType: '' }; fetchData() }

// ==================== 新增/编辑 ====================
const dialogVisible = ref(false)
const submitLoading = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance | null>(null)
const currentId = ref<number | null>(null)

const formData = ref({
    role_id: 0,
    resource_type: 'dept',
    field_name: '',
    condition_type: 'eq',
    condition_value: '',
    description: '',
})

const formRules = {
    role_id: [{ required: true, message: '请选择角色', trigger: 'change' }],
    resource_type: [{ required: true, message: '请选择资源类型', trigger: 'change' }],
    field_name: [{ required: true, message: '请输入字段名', trigger: 'blur' }],
    condition_type: [{ required: true, message: '请选择条件', trigger: 'change' }],
    condition_value: [{ required: true, message: '请输入条件值', trigger: 'blur' }],
}

const handleAdd = () => {
    isEdit.value = false
    currentId.value = null
    formData.value = { role_id: 0, resource_type: 'dept', field_name: '', condition_type: 'eq', condition_value: '', description: '' }
    dialogVisible.value = true
}

const handleEdit = (row: SysScope) => {
    isEdit.value = true
    currentId.value = row.scope_id
    formData.value = {
        role_id: row.role_id,
        resource_type: row.resource_type,
        field_name: row.field_name,
        condition_type: row.condition_type,
        condition_value: row.condition_value,
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
                role_id: formData.value.role_id,
                resource_type: formData.value.resource_type,
                field_name: formData.value.field_name,
                condition_type: formData.value.condition_type,
                condition_value: formData.value.condition_value,
                description: formData.value.description || undefined,
            }
            if (isEdit.value) {
                await UpdateScopeApi(currentId.value!, data)
                ElMessage.success('编辑成功')
            } else {
                await CreateScopeApi(data)
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
const handleDelete = async (row: SysScope) => {
    try {
        await ElMessageBox.confirm('确定要删除该数据权限规则吗？', '删除确认', { type: 'warning', customClass: 'rounded-xl' })
        await DeleteScopeApi(row.scope_id)
        ElMessage.success('删除成功')
        fetchData()
    } catch { /* 取消 */ }
}

onMounted(() => { fetchData(); loadRoles() })
</script>
