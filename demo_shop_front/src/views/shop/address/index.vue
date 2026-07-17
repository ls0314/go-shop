<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-800 dark:text-white">收货地址</h1>
      <el-button type="primary" class="h-9 rounded-xl" style="background: #ff6700; border-color: #ff6700" @click="openAdd">
        <el-icon class="mr-1"><Plus /></el-icon>
        新增地址
      </el-button>
    </div>

    <!-- 加载 -->
    <div v-if="loading" class="flex justify-center py-20">
      <el-icon class="is-loading text-3xl text-slate-300"><Loading /></el-icon>
    </div>

    <!-- 空状态 -->
    <div
      v-else-if="addressList.length === 0"
      class="flex flex-col items-center justify-center rounded-xl bg-white py-24 dark:bg-slate-800"
    >
      <span class="text-5xl">📍</span>
      <p class="mt-4 text-slate-400">暂无收货地址</p>
      <el-button class="mt-4 h-9 rounded-xl" style="background: #ff6700; border-color: #ff6700; color: #fff" @click="openAdd">
        添加新地址
      </el-button>
    </div>

    <!-- 地址卡片列表 -->
    <div v-else class="space-y-3">
      <div
        v-for="item in addressList"
        :key="item.address_id"
        class="rounded-xl border bg-white p-5 transition-shadow hover:shadow-sm dark:border-slate-700 dark:bg-slate-800"
        :class="item.is_default ? 'border-orange-300 dark:border-orange-600' : 'border-slate-200'"
      >
        <div class="flex items-start justify-between">
          <!-- 左侧地址信息 -->
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <span class="text-base font-semibold text-slate-800 dark:text-white">{{ item.receiver_name }}</span>
              <span class="text-sm text-slate-500">{{ item.receiver_phone }}</span>
              <el-tag v-if="item.is_default" size="small" class="rounded-md" type="warning">默认</el-tag>
              <el-tag v-if="item.address_tag" size="small" class="rounded-md" type="info">{{ addressTagLabel(item.address_tag) }}</el-tag>
            </div>
            <p class="mt-2 text-sm text-slate-600 dark:text-slate-300">
              {{ item.province }}{{ item.city }}{{ item.district }} {{ item.detail_address }}
            </p>
            <p v-if="item.postal_code" class="mt-1 text-xs text-slate-400">邮编：{{ item.postal_code }}</p>
          </div>

          <!-- 右侧操作 -->
          <div class="ml-4 flex flex-shrink-0 items-center gap-1">
            <el-button link size="small" class="text-slate-500" @click="openEdit(item)">编辑</el-button>
            <el-button v-if="!item.is_default" link size="small" class="text-indigo-500" @click="handleSetDefault(item.address_id)">设为默认</el-button>
            <el-button link size="small" class="text-slate-400" @click="handleDelete(item.address_id)">删除</el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- ==================== 新增/编辑弹窗 ==================== -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑地址' : '新增地址'"
      width="520px"
      draggable
      custom-class="rounded-xl"
      :close-on-click-modal="false"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="收货人" prop="receiver_name">
              <el-input v-model="form.receiver_name" placeholder="请输入姓名" maxlength="50" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="手机号" prop="receiver_phone">
              <el-input v-model="form.receiver_phone" placeholder="11位手机号" maxlength="11" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item label="省份" prop="province">
              <el-input v-model="form.province" placeholder="省" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="城市" prop="city">
              <el-input v-model="form.city" placeholder="市" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="区/县" prop="district">
              <el-input v-model="form.district" placeholder="区" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="详细地址" prop="detail_address">
          <el-input v-model="form.detail_address" placeholder="街道、门牌号等" maxlength="200" />
        </el-form-item>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="邮编">
              <el-input v-model="form.postal_code" placeholder="选填" maxlength="6" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="标签">
              <el-select v-model="form.address_tag" clearable class="w-full">
                <el-option label="家" value="家" />
                <el-option label="公司" value="公司" />
                <el-option label="学校" value="学校" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item v-if="!isEdit || !editingAddress?.is_default" label=" ">
          <el-checkbox v-model="form.is_default">设为默认地址</el-checkbox>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false" class="h-9 rounded-xl">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting" class="h-9 rounded-xl" style="background: #ff6700; border-color: #ff6700">
          {{ isEdit ? '保存' : '添加' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Plus, Loading } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { useAddressStore } from '@/pinia/modules/address'
import type { UserAddress } from '@/types/address'

const addressStore = useAddressStore()
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const editingAddress = ref<UserAddress | null>(null)
const formRef = ref<FormInstance>()

const addressList = computed(() => addressStore.addressList)

const form = ref({
  receiver_name: '',
  receiver_phone: '',
  province: '',
  city: '',
  district: '',
  detail_address: '',
  postal_code: '',
  is_default: false,
  address_tag: '',
})

const rules: FormRules = {
  receiver_name: [
    { required: true, message: '请输入收货人姓名', trigger: 'blur' },
    { min: 1, max: 50, message: '1-50个字符', trigger: 'blur' },
  ],
  receiver_phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '手机号格式不正确', trigger: 'blur' },
  ],
  province: [{ required: true, message: '请输入省份', trigger: 'blur' }],
  city: [{ required: true, message: '请输入城市', trigger: 'blur' }],
  district: [{ required: true, message: '请输入区/县', trigger: 'blur' }],
  detail_address: [
    { required: true, message: '请输入详细地址', trigger: 'blur' },
    { min: 1, max: 200, message: '1-200个字符', trigger: 'blur' },
  ],
}

function addressTagLabel(tag: string): string {
  const map: Record<string, string> = { '家': '🏠 家', '公司': '🏢 公司', '学校': '🏫 学校' }
  return map[tag] || tag
}

function openAdd() {
  isEdit.value = false
  editingAddress.value = null
  form.value = { receiver_name: '', receiver_phone: '', province: '', city: '', district: '', detail_address: '', postal_code: '', is_default: false, address_tag: '' }
  formRef.value?.clearValidate()
  dialogVisible.value = true
}

function openEdit(item: UserAddress) {
  isEdit.value = true
  editingAddress.value = item
  form.value = {
    receiver_name: item.receiver_name,
    receiver_phone: item.receiver_phone,
    province: item.province,
    city: item.city,
    district: item.district,
    detail_address: item.detail_address,
    postal_code: item.postal_code || '',
    is_default: item.is_default,
    address_tag: item.address_tag || '',
  }
  formRef.value?.clearValidate()
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      let ok: boolean
      if (isEdit.value && editingAddress.value) {
        ok = await addressStore.UpdateAddress(editingAddress.value.address_id, form.value)
      } else {
        ok = await addressStore.CreateAddress(form.value)
      }
      if (ok) {
        ElMessage.success(isEdit.value ? '地址已更新' : '地址已添加')
        dialogVisible.value = false
      } else {
        ElMessage.error('操作失败')
      }
    } finally {
      submitting.value = false
    }
  })
}

async function handleSetDefault(id: number) {
  const ok = await addressStore.SetDefaultAddress(id)
  if (ok) ElMessage.success('已设为默认地址')
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确定要删除该收货地址吗？', '删除确认', { type: 'warning' })
    const ok = await addressStore.DeleteAddress(id)
    if (ok) ElMessage.success('已删除')
  } catch { /* 取消 */ }
}

onMounted(async () => {
  loading.value = true
  try { await addressStore.GetAddressList() }
  finally { loading.value = false }
})
</script>
