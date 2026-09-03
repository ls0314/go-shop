<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-800 dark:text-white">角色管理</h1>
    </div>

    <!-- 搜索与操作栏 -->
    <div class="flex flex-wrap items-center justify-between gap-4 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-800">
      <el-form :model="queryParams" inline class="flex flex-wrap gap-3" @submit.prevent="handleSearch">
        <el-form-item label="角色类型" class="mb-0">
          <el-select v-model="queryParams.roleType" placeholder="全部" clearable class="w-36" @change="handleSearch">
            <el-option label="系统角色" value="platform" />
            <el-option label="自定义角色" value="custom" />
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

      <el-button v-if="userStore.hasPerm('system:role:create')" type="primary" class="h-9 rounded-xl" @click="handleAdd">
        <el-icon class="mr-1"><Plus /></el-icon>新增角色
      </el-button>
    </div>

    <!-- 表格 -->
    <div class="rounded-xl border border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800">
      <el-table :data="roleList" v-loading="loading" border stripe class="w-full"
                :header-cell-style="{ background: 'transparent' }" :cell-style="{ background: 'transparent' }">
        <el-table-column prop="role_id" label="ID" width="70px" align="center" />
        <el-table-column prop="role_name" label="角色名称" min-width="140px" />
        <el-table-column label="类型" width="110px" align="center">
          <template #default="scope">
            <el-tag :type="scope.row.role_type === 'platform' ? 'warning' : 'primary'" size="small" class="rounded-lg">
              {{ scope.row.role_type === 'platform' ? '系统角色' : '自定义' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="200px" show-overflow-tooltip />
        <el-table-column label="数据范围" width="100px" align="center">
          <template #default="scope">
            {{ dataScopeText(scope.row.data_scope) }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170px" align="center" />
        <el-table-column label="操作" width="320px" align="center" fixed="right">
          <template #default="scope">
            <el-button v-if="userStore.hasPerm('system:role:update')" type="primary" link size="small" @click="handleEdit(scope.row)">编辑</el-button>
            <el-button v-if="userStore.hasPerm('system:role:assign-menu')" type="warning" link size="small" @click="handleAssignMenu(scope.row)">分配菜单</el-button>
            <el-button v-if="userStore.hasPerm('system:role:assign-perm')" type="success" link size="small" @click="handleAssignPerm(scope.row)">分配权限</el-button>
            <el-button v-if="userStore.hasPerm('system:role:delete')" type="danger" link size="small" @click="handleDelete(scope.row)">删除</el-button>
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
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑角色' : '新增角色'" width="480px" custom-class="rounded-xl"
               :close-on-click-modal="false" @close="handleDialogClose">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="90px" class="pr-5">
        <el-form-item label="角色名称" prop="role_name">
          <el-input v-model="formData.role_name" placeholder="请输入角色名称" maxlength="50" />
        </el-form-item>
        <el-form-item label="角色类型" prop="role_type">
          <el-radio-group v-model="formData.role_type" :disabled="isEdit">
            <el-radio value="platform">系统角色</el-radio>
            <el-radio value="custom">自定义</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="数据范围" prop="data_scope">
          <el-select v-model="formData.data_scope" class="w-full">
            <el-option label="全部数据" value="all" />
            <el-option label="本部门数据" value="dept" />
            <el-option label="仅本人数据" value="self" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="formData.description" type="textarea" :rows="2" placeholder="请输入角色描述" maxlength="200" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false" class="h-9 rounded-xl">取消</el-button>
        <el-button type="primary" :loading="submitLoading" class="h-9 rounded-xl" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 分配菜单弹窗 -->
    <el-dialog v-model="menuDialogVisible" title="分配菜单" width="480px" custom-class="rounded-xl" :close-on-click-modal="false">
      <p class="mb-3 text-sm text-slate-500 dark:text-slate-400">
        为角色 <b class="text-slate-700 dark:text-white">{{ currentRole?.role_name }}</b> 分配菜单
      </p>
      <el-tree
          ref="menuTreeRef"
          :data="menuTreeData"
          show-checkbox
          node-key="menu_id"
          :props="{ label: 'menu_name', children: 'children' }"
          default-expand-all
          class="max-h-80 overflow-auto rounded-xl border border-slate-200 p-3 dark:border-slate-700"
      />
      <template #footer>
        <el-button @click="menuDialogVisible = false" class="h-9 rounded-xl">取消</el-button>
        <el-button type="primary" :loading="menuSubmitLoading" class="h-9 rounded-xl" @click="handleSubmitMenu">确定</el-button>
      </template>
    </el-dialog>

    <!-- 分配权限弹窗 -->
    <el-dialog v-model="permDialogVisible" title="分配权限" width="520px" custom-class="rounded-xl" :close-on-click-modal="false">
      <p class="mb-3 text-sm text-slate-500 dark:text-slate-400">
        为角色 <b class="text-slate-700 dark:text-white">{{ currentRole?.role_name }}</b> 分配权限
      </p>
      <el-select v-model="permFilter" placeholder="按模块筛选" clearable class="mb-3 w-full" @change="filterPerms">
        <el-option v-for="m in permModules" :key="m" :label="m" :value="m" />
      </el-select>
      <el-checkbox-group v-model="selectedPermIds" class="max-h-80 overflow-auto rounded-xl border border-slate-200 p-3 dark:border-slate-700">
        <el-checkbox v-for="perm in filteredPerms" :key="perm.permission_id" :value="perm.permission_id" class="!mr-0 !flex !w-full !py-1">
          <span class="text-sm">{{ perm.permission_name }}</span>
          <span class="ml-2 text-xs text-slate-400">{{ perm.permission_code }}</span>
        </el-checkbox>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="permDialogVisible = false" class="h-9 rounded-xl">取消</el-button>
        <el-button type="primary" :loading="permSubmitLoading" class="h-9 rounded-xl" @click="handleSubmitPerm">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Search, Refresh, Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, type FormInstance, type ElTree } from 'element-plus'
import { useUserStore } from '@/pinia/modules/user'
import {
    GetRoleListApi, CreateRoleApi, UpdateRoleApi, DeleteRoleApi,
    GetMenuListApi, AssignRoleMenusApi, GetRoleMenuIdsApi,
    GetPermissionListApi, AssignRolePermsApi, GetRolePermIdsApi, fetchAll,
} from '@/api/rbac'
import type { SysRole, SysMenu, SysPermission } from '@/types/rbac'

const userStore = useUserStore()

const dataScopeText = (s: string) => ({ all: '全部', dept: '本部门', self: '仅本人' }[s] || s || '-')

const loading = ref(false)
const roleList = ref<SysRole[]>([])
const total = ref(0)

const queryParams = ref({ page: 1, pageSize: 10, roleType: '' })

const fetchData = async () => {
    loading.value = true
    try {
        const res = await GetRoleListApi({
            page: queryParams.value.page,
            pageSize: queryParams.value.pageSize,
            roleType: queryParams.value.roleType || undefined,
        })
        const payload = res.data.data
        roleList.value = payload?.list || []
        total.value = payload?.total || 0
    } catch (error) {
        console.error('获取角色列表失败:', error)
        ElMessage.error('获取角色列表失败')
    } finally {
        loading.value = false
    }
}

const handleSearch = () => { queryParams.value.page = 1; fetchData() }
const handleReset = () => { queryParams.value = { page: 1, pageSize: 10, roleType: '' }; fetchData() }

// ==================== 新增/编辑 ====================
const dialogVisible = ref(false)
const submitLoading = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance | null>(null)
const currentId = ref<number | null>(null)

const formData = ref({
    role_name: '',
    role_type: 'custom',
    data_scope: 'all',
    description: '',
})

const formRules = {
    role_name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
    role_type: [{ required: true, message: '请选择角色类型', trigger: 'change' }],
    data_scope: [{ required: true, message: '请选择数据范围', trigger: 'change' }],
}

const handleAdd = () => {
    isEdit.value = false
    currentId.value = null
    formData.value = { role_name: '', role_type: 'custom', data_scope: 'all', description: '' }
    dialogVisible.value = true
}

const handleEdit = (row: SysRole) => {
    isEdit.value = true
    currentId.value = row.role_id
    formData.value = { role_name: row.role_name, role_type: row.role_type, data_scope: row.data_scope, description: row.description }
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
                role_name: formData.value.role_name,
                role_type: formData.value.role_type,
                data_scope: formData.value.data_scope,
                description: formData.value.description || undefined,
            }
            if (isEdit.value) {
                await UpdateRoleApi(currentId.value!, data)
                ElMessage.success('编辑成功')
            } else {
                await CreateRoleApi(data)
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
const handleDelete = async (row: SysRole) => {
    try {
        await ElMessageBox.confirm(`确定要删除角色【${row.role_name}】吗？`, '删除确认', { type: 'warning', customClass: 'rounded-xl' })
        await DeleteRoleApi(row.role_id)
        ElMessage.success('删除成功')
        fetchData()
    } catch { /* 取消 */ }
}

// ==================== 分配菜单 ====================
const menuDialogVisible = ref(false)
const menuSubmitLoading = ref(false)
const currentRole = ref<SysRole | null>(null)
const menuTreeData = ref<SysMenu[]>([])
const menuTreeRef = ref<InstanceType<typeof ElTree> | null>(null)

// 菜单扁平列表 → 树
const buildMenuTree = (list: SysMenu[]): SysMenu[] => {
    const map = new Map<number, SysMenu>()
    list.forEach(item => map.set(item.menu_id, { ...item, children: [] }))
    const roots: SysMenu[] = []
    map.forEach(item => {
        if (item.parent_id && item.parent_id !== 0 && map.has(item.parent_id)) {
            map.get(item.parent_id)!.children!.push(item)
        } else {
            roots.push(item)
        }
    })
    return roots
}

const handleAssignMenu = async (row: SysRole) => {
    currentRole.value = row
    menuDialogVisible.value = true
    try {
        // 全量菜单(分页循环拉全)
        const allMenus = await fetchAll<SysMenu>(GetMenuListApi)
        menuTreeData.value = buildMenuTree(allMenus)
        // 已绑定菜单
        const relRes = await GetRoleMenuIdsApi(row.role_id)
        const boundList = relRes.data.data?.list || []
        const boundIds = boundList.map((m: any) => m.menu_id ?? m.MenuId)
        // 等树渲染后回显勾选
        setTimeout(() => {
            menuTreeRef.value?.setCheckedKeys(boundIds, false)
        }, 50)
    } catch (error) {
        console.error('加载菜单数据失败:', error)
        ElMessage.error('加载菜单数据失败')
    }
}

const handleSubmitMenu = async () => {
    if (!currentRole.value || !menuTreeRef.value) return
    const checked = menuTreeRef.value.getCheckedKeys(false) as number[]
    const halfChecked = menuTreeRef.value.getHalfCheckedKeys() as number[]
    const allIds = [...checked, ...halfChecked]  // 半选父节点也要提交,否则父子关系丢失
    menuSubmitLoading.value = true
    try {
        await AssignRoleMenusApi(currentRole.value.role_id, allIds)
        ElMessage.success('分配成功')
        menuDialogVisible.value = false
    } catch (error: any) {
        ElMessage.error(error?.response?.data?.message || '分配失败')
    } finally {
        menuSubmitLoading.value = false
    }
}

// ==================== 分配权限 ====================
const permDialogVisible = ref(false)
const permSubmitLoading = ref(false)
const allPerms = ref<SysPermission[]>([])
const filteredPerms = ref<SysPermission[]>([])
const selectedPermIds = ref<number[]>([])
const permFilter = ref('')

const permModules = computed(() => {
    const set = new Set<string>()
    allPerms.value.forEach(p => {
        const m = p.permission_code?.split(':')[0]
        if (m) set.add(m)
    })
    return Array.from(set)
})

const filterPerms = () => {
    if (!permFilter.value) filteredPerms.value = allPerms.value
    else filteredPerms.value = allPerms.value.filter(p => p.permission_code?.startsWith(permFilter.value + ':'))
}

const handleAssignPerm = async (row: SysRole) => {
    currentRole.value = row
    permDialogVisible.value = true
    permFilter.value = ''
    try {
        // 全量权限(分页循环拉全)
        const allPermsData = await fetchAll<SysPermission>(GetPermissionListApi, {}, 100)
        allPerms.value = allPermsData
        filteredPerms.value = allPerms.value
        // 已绑定权限
        const relRes = await GetRolePermIdsApi(row.role_id)
        const list = relRes.data.data?.list || []
        selectedPermIds.value = list.map((p: any) => p.permission_id ?? p.PermissionId ?? p.PermId)
    } catch (error) {
        console.error('加载权限数据失败:', error)
        ElMessage.error('加载权限数据失败')
    }
}

const handleSubmitPerm = async () => {
    if (!currentRole.value) return
    permSubmitLoading.value = true
    try {
        await AssignRolePermsApi(currentRole.value.role_id, selectedPermIds.value)
        ElMessage.success('分配成功')
        permDialogVisible.value = false
    } catch (error: any) {
        ElMessage.error(error?.response?.data?.message || '分配失败')
    } finally {
        permSubmitLoading.value = false
    }
}

onMounted(() => { fetchData() })
</script>
