<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-800 dark:text-white">菜单管理</h1>
    </div>

    <!-- 操作栏 -->
    <div class="flex flex-wrap items-center justify-between gap-4 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-800">
      <p class="text-sm text-slate-500 dark:text-slate-400">菜单层级决定管理端左侧导航结构(M=目录 / C=页面 / F=按钮)</p>
      <el-button v-if="userStore.hasPerm('system:menu:create')" type="primary" class="h-9 rounded-xl" @click="handleAddRoot">
        <el-icon class="mr-1"><Plus /></el-icon>新增顶级菜单
      </el-button>
    </div>

    <!-- 树形表格 -->
    <div class="rounded-xl border border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800">
      <el-table :data="menuTree" v-loading="loading" row-key="menu_id" border default-expand-all
                :tree-props="{ children: 'children' }" class="w-full"
                :header-cell-style="{ background: 'transparent' }" :cell-style="{ background: 'transparent' }">
        <el-table-column prop="menu_name" label="菜单名称" min-width="200px" />
        <el-table-column label="类型" width="90px" align="center">
          <template #default="scope">
            <el-tag :type="menuTypeTag(scope.row.menu_type)" size="small" class="rounded-lg">
              {{ menuTypeText(scope.row.menu_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="route_path" label="路由路径" min-width="180px" show-overflow-tooltip />
        <el-table-column prop="component" label="组件" min-width="200px" show-overflow-tooltip />
        <el-table-column prop="icon" label="图标" width="80px" align="center" />
        <el-table-column label="可见" width="70px" align="center">
          <template #default="scope">
            <el-tag v-if="scope.row.is_visible" type="success" size="small" class="rounded-lg">是</el-tag>
            <span v-else class="text-slate-400">否</span>
          </template>
        </el-table-column>
        <el-table-column prop="sort_order" label="排序" width="80px" align="center" />
        <el-table-column label="操作" width="230px" align="center" fixed="right">
          <template #default="scope">
            <template v-if="scope.row.menu_type !== 'F'">
              <el-button v-if="userStore.hasPerm('system:menu:create')" type="primary" link size="small" @click="handleAddChild(scope.row)">新增子级</el-button>
            </template>
            <el-button v-if="userStore.hasPerm('system:menu:update')" type="warning" link size="small" @click="handleEdit(scope.row)">编辑</el-button>
            <el-button v-if="userStore.hasPerm('system:menu:delete')" type="danger" link size="small" @click="handleDelete(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="formTitle" width="540px" custom-class="rounded-xl"
               :close-on-click-modal="false" @close="handleDialogClose">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="100px" class="pr-5">
        <el-form-item label="上级菜单" prop="parent_id">
          <el-tree-select
              v-model="formData.parent_id"
              :data="parentOptions"
              node-key="menu_id"
              :props="{ label: 'menu_name', children: 'children' }"
              placeholder="不选则为顶级菜单"
              clearable check-strictly class="w-full"
          />
        </el-form-item>
        <el-form-item label="菜单名称" prop="menu_name">
          <el-input v-model="formData.menu_name" placeholder="请输入菜单名称" maxlength="50" />
        </el-form-item>
        <el-form-item label="菜单类型" prop="menu_type">
          <el-radio-group v-model="formData.menu_type">
            <el-radio value="M">目录</el-radio>
            <el-radio value="C">页面</el-radio>
            <el-radio value="F">按钮</el-radio>
          </el-radio-group>
        </el-form-item>
        <template v-if="formData.menu_type !== 'F'">
          <el-form-item label="路由路径" prop="route_path">
            <el-input v-model="formData.route_path" placeholder="如 /admin/system/permission" maxlength="200" />
          </el-form-item>
          <el-form-item label="组件" prop="component">
            <el-input v-model="formData.component" placeholder="如 system/permission/index" maxlength="200" />
          </el-form-item>
          <el-form-item label="图标">
            <el-input v-model="formData.icon" placeholder="图标 key 或 emoji" maxlength="100" />
          </el-form-item>
        </template>
        <el-form-item label="可见" prop="is_visible">
          <el-radio-group v-model="formData.is_visible">
            <el-radio :value="true">显示</el-radio>
            <el-radio :value="false">隐藏</el-radio>
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
    GetMenuListApi, CreateMenuApi, UpdateMenuApi, DeleteMenuApi, fetchAll,
} from '@/api/rbac'
import type { SysMenu } from '@/types/rbac'

const userStore = useUserStore()

const loading = ref(false)
const menuTree = ref<SysMenu[]>([])

const fetchData = async () => {
    loading.value = true
    try {
        // 全量拉取(循环翻页),再组树
        const list = await fetchAll<SysMenu>(GetMenuListApi)
        menuTree.value = buildMenuTree(list)
    } catch (error) {
        console.error('获取菜单列表失败:', error)
        ElMessage.error('获取菜单列表失败')
    } finally {
        loading.value = false
    }
}

// 扁平列表 → 树
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

const menuTypeTag = (t: string) => ({ M: 'warning', C: 'primary', F: 'success' }[t] || 'info') as any

const menuTypeText = (t: string) => ({ M: '目录', C: '页面', F: '按钮' }[t] || t)

// ==================== 新增/编辑 ====================
const dialogVisible = ref(false)
const submitLoading = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance | null>(null)
const currentId = ref<number | null>(null)

const formData = ref({
    parent_id: 0,
    menu_name: '',
    menu_type: 'M',
    route_path: '',
    component: '',
    icon: '',
    is_visible: true,
    sort_order: 0,
})

const formRules = {
    menu_name: [{ required: true, message: '请输入菜单名称', trigger: 'blur' }],
    menu_type: [{ required: true, message: '请选择菜单类型', trigger: 'change' }],
    route_path: [{ required: true, message: '请输入路由路径', trigger: 'blur' }],
    component: [{ required: true, message: '请输入组件路径', trigger: 'blur' }],
}

const formTitle = computed(() => {
    if (isEdit.value) return '编辑菜单'
    return formData.value.parent_id ? '新增子菜单' : '新增顶级菜单'
})

// 可选父级(顶级 + 非按钮节点;编辑时排除自身及后代)
const parentOptions = computed(() => {
    const excludeIds = new Set<number>()
    if (isEdit.value && currentId.value) {
        const collect = (nodes: SysMenu[]) => {
            nodes.forEach(n => {
                excludeIds.add(n.menu_id)
                if (n.children?.length) collect(n.children)
            })
        }
        // 找当前节点整棵子树
        const findAndCollect = (nodes: SysMenu[]): boolean => {
            for (const n of nodes) {
                if (n.menu_id === currentId.value) { collect([n]); return true }
                if (n.children?.length && findAndCollect(n.children)) return true
            }
            return false
        }
        findAndCollect(menuTree.value)
    }
    const filter = (nodes: SysMenu[]): SysMenu[] => {
        return nodes
            .filter(n => !excludeIds.has(n.menu_id) && n.menu_type !== 'F')
            .map(n => ({ ...n, children: n.children?.length ? filter(n.children) : undefined }))
    }
    return filter(menuTree.value)
})

const handleAddRoot = () => {
    isEdit.value = false
    currentId.value = null
    formData.value = { parent_id: 0, menu_name: '', menu_type: 'M', route_path: '', component: '', icon: '', is_visible: true, sort_order: 0 }
    dialogVisible.value = true
}

const handleAddChild = (row: SysMenu) => {
    isEdit.value = false
    currentId.value = null
    formData.value = { parent_id: row.menu_id, menu_name: '', menu_type: 'C', route_path: '', component: '', icon: '', is_visible: true, sort_order: 0 }
    dialogVisible.value = true
}

const handleEdit = (row: SysMenu) => {
    isEdit.value = true
    currentId.value = row.menu_id
    formData.value = {
        parent_id: row.parent_id,
        menu_name: row.menu_name,
        menu_type: row.menu_type,
        route_path: row.route_path || '',
        component: row.component || '',
        icon: row.icon || '',
        is_visible: row.is_visible,
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
            const data: any = {
                parent_id: formData.value.parent_id || 0,
                menu_name: formData.value.menu_name,
                menu_type: formData.value.menu_type,
                is_visible: formData.value.is_visible,
                sort_order: formData.value.sort_order,
            }
            if (formData.value.menu_type !== 'F') {
                data.route_path = formData.value.route_path
                data.component = formData.value.component
                data.icon = formData.value.icon || undefined
            }
            if (isEdit.value) {
                await UpdateMenuApi(currentId.value!, data)
                ElMessage.success('编辑成功')
            } else {
                await CreateMenuApi(data)
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
const handleDelete = async (row: SysMenu) => {
    try {
        await ElMessageBox.confirm(`确定要删除菜单【${row.menu_name}】吗？其子菜单将一并处理。`, '删除确认', { type: 'warning', customClass: 'rounded-xl' })
        await DeleteMenuApi(row.menu_id)
        ElMessage.success('删除成功')
        fetchData()
    } catch { /* 取消 */ }
}

onMounted(() => { fetchData() })
</script>
