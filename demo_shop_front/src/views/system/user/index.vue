<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-800 dark:text-white">用户管理</h1>
    </div>

    <!-- 搜索与操作栏 -->
    <div class="flex flex-wrap items-center justify-between gap-4 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-800">
      <el-form :model="queryParams" inline class="flex flex-wrap gap-3" @submit.prevent="handleSearch">
        <el-form-item label="状态" class="mb-0">
          <el-select v-model="queryParams.status" placeholder="全部" clearable class="w-32" @change="handleSearch">
            <el-option label="启用" value="active" />
            <el-option label="禁用" value="disabled" />
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

      <el-button v-if="userStore.hasPerm('system:user:create')" type="primary" class="h-9 rounded-xl" @click="handleAdd">
        <el-icon class="mr-1"><Plus /></el-icon>新增用户
      </el-button>
    </div>

    <!-- 表格 -->
    <div class="rounded-xl border border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800">
      <el-table :data="userList" v-loading="loading" border stripe class="w-full"
                :header-cell-style="{ background: 'transparent' }" :cell-style="{ background: 'transparent' }">
        <el-table-column prop="user_id" label="ID" width="70px" align="center" />
        <el-table-column prop="username" label="用户名" min-width="120px" />
        <el-table-column prop="email" label="邮箱" min-width="180px" show-overflow-tooltip />
        <el-table-column prop="phone" label="手机号" min-width="130px" />
        <el-table-column label="状态" width="90px" align="center">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'active' ? 'success' : 'danger'" size="small" class="rounded-lg">
              {{ scope.row.status === 'active' ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170px" align="center" />
        <el-table-column label="操作" width="230px" align="center" fixed="right">
          <template #default="scope">
            <el-button v-if="userStore.hasPerm('system:user:update')" type="primary" link size="small" @click="handleEdit(scope.row)">编辑</el-button>
            <el-button v-if="userStore.hasPerm('system:user:assign-role')" type="warning" link size="small" @click="handleAssignRole(scope.row)">分配角色</el-button>
            <el-button v-if="userStore.hasPerm('system:user:delete')" type="danger" link size="small" @click="handleDelete(scope.row)">删除</el-button>
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
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑用户' : '新增用户'" width="480px" custom-class="rounded-xl"
               :close-on-click-modal="false" @close="handleDialogClose">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="90px" class="pr-5">
        <el-form-item v-if="!isEdit" label="用户名" prop="username">
          <el-input v-model="formData.username" placeholder="请输入用户名" maxlength="50" />
        </el-form-item>
        <el-form-item v-if="!isEdit" label="密码" prop="password">
          <el-input v-model="formData.password" type="password" show-password placeholder="请输入密码" maxlength="50" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="formData.email" placeholder="请输入邮箱" maxlength="100" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="formData.phone" placeholder="请输入手机号" maxlength="20" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio value="active">启用</el-radio>
            <el-radio value="disabled">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false" class="h-9 rounded-xl">取消</el-button>
        <el-button type="primary" :loading="submitLoading" class="h-9 rounded-xl" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 分配角色弹窗 -->
    <el-dialog v-model="roleDialogVisible" title="分配角色" width="420px" custom-class="rounded-xl"
               :close-on-click-modal="false">
      <p class="mb-3 text-sm text-slate-500 dark:text-slate-400">
        为用户 <b class="text-slate-700 dark:text-white">{{ currentUser?.username }}</b> 分配角色
      </p>
      <el-select v-model="selectedRoleIds" multiple placeholder="请选择角色" class="w-full" :loading="roleLoading">
        <el-option v-for="role in roleOptions" :key="role.role_id" :label="role.role_name" :value="role.role_id" />
      </el-select>
      <template #footer>
        <el-button @click="roleDialogVisible = false" class="h-9 rounded-xl">取消</el-button>
        <el-button type="primary" :loading="roleSubmitLoading" class="h-9 rounded-xl" @click="handleSubmitRoles">确定</el-button>
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
    GetUserListApi, CreateUserApi, UpdateUserApi, DeleteUserApi,
    GetRoleListApi, AssignUserRolesApi, GetUserRoleIdsApi, fetchAll,
} from '@/api/rbac'
import type { SysUser, SysRole } from '@/types/rbac'

const userStore = useUserStore()

const loading = ref(false)
const userList = ref<SysUser[]>([])
const total = ref(0)

const queryParams = ref({
    page: 1,
    pageSize: 10,
    status: '',
})

const fetchData = async () => {
    loading.value = true
    try {
        const res = await GetUserListApi({
            page: queryParams.value.page,
            pageSize: queryParams.value.pageSize,
            status: queryParams.value.status || undefined,
        })
        const payload = res.data.data
        userList.value = payload?.list || []
        total.value = payload?.total || 0
    } catch (error) {
        console.error('获取用户列表失败:', error)
        ElMessage.error('获取用户列表失败')
    } finally {
        loading.value = false
    }
}

const handleSearch = () => { queryParams.value.page = 1; fetchData() }
const handleReset = () => {
    queryParams.value = { page: 1, pageSize: 10, status: '' }
    fetchData()
}

// ==================== 新增/编辑 ====================
const dialogVisible = ref(false)
const submitLoading = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance | null>(null)
const currentId = ref<number | null>(null)

const formData = ref({
    username: '',
    password: '',
    email: '',
    phone: '',
    status: 'active',
})

const formRules = {
    username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
    password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
    status: [{ required: true, message: '请选择状态', trigger: 'change' }],
}

const handleAdd = () => {
    isEdit.value = false
    currentId.value = null
    formData.value = { username: '', password: '', email: '', phone: '', status: 'active' }
    dialogVisible.value = true
}

const handleEdit = (row: SysUser) => {
    isEdit.value = true
    currentId.value = row.user_id
    formData.value = { username: row.username, password: '', email: row.email, phone: row.phone, status: row.status }
    dialogVisible.value = true
}

const handleDialogClose = () => { formRef.value?.resetFields() }

const handleSubmit = async () => {
    if (!formRef.value) return
    await formRef.value.validate(async (valid) => {
        if (!valid) return
        submitLoading.value = true
        try {
            if (isEdit.value) {
                const data: any = {
                    email: formData.value.email || undefined,
                    phone: formData.value.phone || undefined,
                    status: formData.value.status,
                }
                await UpdateUserApi(currentId.value!, data)
                ElMessage.success('编辑成功')
            } else {
                await CreateUserApi({
                    username: formData.value.username,
                    password: formData.value.password,
                    email: formData.value.email || undefined,
                    phone: formData.value.phone || undefined,
                    status: formData.value.status,
                })
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
const handleDelete = async (row: SysUser) => {
    try {
        await ElMessageBox.confirm(`确定要删除用户【${row.username}】吗？`, '删除确认', { type: 'warning', customClass: 'rounded-xl' })
        await DeleteUserApi(row.user_id)
        ElMessage.success('删除成功')
        fetchData()
    } catch { /* 取消 */ }
}

// ==================== 分配角色 ====================
const roleDialogVisible = ref(false)
const roleLoading = ref(false)
const roleSubmitLoading = ref(false)
const currentUser = ref<SysUser | null>(null)
const selectedRoleIds = ref<number[]>([])
const roleOptions = ref<SysRole[]>([])

const handleAssignRole = async (row: SysUser) => {
    currentUser.value = row
    roleDialogVisible.value = true
    roleLoading.value = true
    try {
        // 拉全量角色做选项
        const allRoles = await fetchAll<SysRole>(GetRoleListApi)
        roleOptions.value = allRoles
        // 拉当前用户已绑角色ID
        const relRes = await GetUserRoleIdsApi(row.user_id)
        const list = relRes.data.data?.List || relRes.data.data?.list || []
        selectedRoleIds.value = list.map((r: any) => r.role_id ?? r.RoleId)
    } catch (error) {
        console.error('加载角色数据失败:', error)
        ElMessage.error('加载角色数据失败')
    } finally {
        roleLoading.value = false
    }
}

const handleSubmitRoles = async () => {
    if (!currentUser.value) return
    roleSubmitLoading.value = true
    try {
        await AssignUserRolesApi(currentUser.value.user_id, selectedRoleIds.value)
        ElMessage.success('分配成功')
        roleDialogVisible.value = false
    } catch (error: any) {
        ElMessage.error(error?.response?.data?.message || '分配失败')
    } finally {
        roleSubmitLoading.value = false
    }
}

onMounted(() => { fetchData() })
</script>
