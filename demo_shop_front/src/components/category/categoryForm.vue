<template>
  <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑类目' : '新增类目'"
      width="520px"
      @close="handleClose"
      draggable
      custom-class="rounded-xl"
      :close-on-click-modal="false"
  >
    <el-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-width="100px"
        class="pr-5"
    >
      <el-form-item label="上级类目" prop="parent_id">
        <el-tree-select
            v-model="formData.parent_id"
            :data="categoryTree"
            node-key="category_id"
            label="category_name"
            value-key="category_id"
            placeholder="不选则为顶级类目"
            clearable
            check-strictly
            :default-expand-all="true"
            class="w-full"
            :disabled="isEdit"
        />
        <p v-if="isEdit" class="mt-1 text-xs text-slate-500 dark:text-slate-400">
          注意：类目创建后不支持修改上级类目
        </p>
      </el-form-item>

      <el-form-item label="类目名称" prop="category_name">
        <el-input
            v-model="formData.category_name"
            placeholder="请输入类目名称（2-50字符）"
            maxlength="50"
            show-word-limit
        />
      </el-form-item>

      <el-form-item label="排序" prop="sort_order">
        <el-input-number
            v-model="formData.sort_order"
            :min="0"
            placeholder="数字越小越靠前"
            class="w-full"
        />
      </el-form-item>

      <el-form-item label="图标URL" prop="icon_url">
        <el-input
            v-model="formData.icon_url"
            placeholder="请输入图标URL（可选）"
        />
      </el-form-item>

      <el-form-item label="可见性" prop="is_visible">
        <el-radio-group v-model="formData.is_visible">
          <el-radio :value="true">显示</el-radio>
          <el-radio :value="false">隐藏</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item label="状态" prop="status">
        <el-radio-group v-model="formData.status">
          <el-radio value="active">启用</el-radio>
          <el-radio value="disabled">禁用</el-radio>
        </el-radio-group>
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <el-button @click="handleClose" class="h-9 rounded-xl border-slate-200 bg-white text-slate-700 hover:bg-slate-50 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-200 dark:hover:bg-slate-600">
          取消
        </el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitLoading" class="h-9 rounded-xl">
          确定
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch, computed, type PropType } from 'vue'
import type { Category, CreateCategoryData, UpdateCategoryData } from '@/api/category'
import { useCategoryStore } from '@/pinia/modules/category'
import { ElMessage, type FormInstance } from 'element-plus'

// 组件属性定义
const props = defineProps({
  // 弹窗显示状态
  visible: { type: Boolean, default: false },
  // 编辑数据（新增时为null）
  editData: { type: Object as PropType<Category | null>, default: null },
  // 类目树数据（用于上级类目选择）
  categoryTree: { type: Array as PropType<Category[]>, default: () => [] }
})

// 事件定义
const emit = defineEmits(['update:visible', 'success'])

// 状态管理
const categoryStore = useCategoryStore()

// 本地状态
const formRef = ref<FormInstance | null>(null)
const submitLoading = ref(false)
const isEdit = ref(false)

// 表单数据
const formData = ref<CreateCategoryData & { category_id?: number }>({
  parent_id: 0,
  category_name: '',
  sort_order: 0,
  icon_url: '',
  is_visible: true,
  status: 'active'
})

// 表单验证规则
const rules = {
  category_name: [
    { required: true, message: '请输入类目名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在2-50个字符', trigger: 'blur' }
  ],
  sort_order: [{ required: true, message: '请输入排序值', trigger: 'blur' }],
  is_visible: [{ required: true, message: '请选择可见性', trigger: 'change' }],
  status: [{ required: true, message: '请选择状态', trigger: 'change' }]
}

// 双向绑定弹窗显示状态
const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
})

// 监听弹窗打开/关闭
watch(
    () => props.visible,
    (val) => {
      if (val) {
        // 判断是否为编辑模式
        isEdit.value = !!props.editData?.category_id

        // 初始化表单数据
        if (props.editData) {
          // 编辑模式：深拷贝编辑数据
          formData.value = { ...props.editData }
        } else {
          // 新增模式：重置为默认值
          formData.value = {
            parent_id: 0,
            category_name: '',
            sort_order: 0,
            icon_url: '',
            is_visible: true,
            status: 'active'
          }
        }

        // 清除表单验证状态
        formRef.value?.clearValidate()
      }
    }
)

// 关闭弹窗
const handleClose = () => {
  formRef.value?.resetFields()
  emit('update:visible', false)
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return

  // 表单验证
  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitLoading.value = true
    try {
      let success: boolean | Category

      if (isEdit.value) {
        // 编辑模式：调用UpdateCategory方法
        const updateData: UpdateCategoryData = {
          category_name: formData.value.category_name,
          sort_order: formData.value.sort_order,
          icon_url: formData.value.icon_url,
          is_visible: formData.value.is_visible,
          status: formData.value.status
        }

        success = await categoryStore.UpdateCategory(formData.value.category_id!, updateData)
      } else {
        // 新增模式：调用CreateCategory方法
        const createData: CreateCategoryData = {
          parent_id: formData.value.parent_id || 0,
          category_name: formData.value.category_name,
          sort_order: formData.value.sort_order,
          icon_url: formData.value.icon_url,
          is_visible: formData.value.is_visible
        }

        success = await categoryStore.CreateCategory(createData)
      }

      if (success) {
        ElMessage.success(isEdit.value ? '编辑类目成功' : '新增类目成功')
        emit('success')
        handleClose()
      } else {
        ElMessage.error(isEdit.value ? '编辑类目失败' : '新增类目失败')
      }
    } catch (error) {
      ElMessage.error('操作失败，请稍后重试')
      console.error('表单提交失败:', error)
    } finally {
      submitLoading.value = false
    }
  })
}
</script>