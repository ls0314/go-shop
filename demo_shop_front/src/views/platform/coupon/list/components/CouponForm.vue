<template>
  <el-dialog
    v-model="dialogVisible"
    title="新增优惠券"
    width="560px"
    draggable
    custom-class="rounded-xl"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-form ref="formRef" :model="formData" :rules="rules" label-width="110px" class="pr-5">
      <el-form-item label="券名称" prop="coupon_name">
        <el-input v-model="formData.coupon_name" placeholder="请输入优惠券名称" maxlength="50" show-word-limit />
      </el-form-item>

      <el-form-item label="券类型" prop="coupon_type">
        <el-radio-group v-model="formData.coupon_type">
          <el-radio value="full_reduction">满减券</el-radio>
          <el-radio value="direct_discount">直减券（折扣）</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item label="使用门槛" prop="threshold_amount">
        <el-input-number
          v-model="formData.threshold_amount"
          :min="0"
          :precision="2"
          placeholder="0 表示无门槛"
          class="w-full"
        />
        <p class="mt-1 text-xs text-slate-400">满多少可用，0 表示无门槛</p>
      </el-form-item>

      <el-form-item label="优惠力度" prop="discount_amount">
        <el-input-number
          v-if="formData.coupon_type === 'full_reduction'"
          v-model="formData.discount_amount"
          :min="0.01"
          :precision="2"
          class="w-full"
        />
        <el-input-number
          v-else
          v-model="formData.discount_amount"
          :min="0.01"
          :max="0.99"
          :precision="2"
          :step="0.05"
          class="w-full"
        />
        <p class="mt-1 text-xs text-slate-400">
          {{ formData.coupon_type === 'full_reduction' ? '立减金额（元）' : '折扣率（0~1，如 0.9 表示 9 折）' }}
        </p>
      </el-form-item>

      <el-form-item label="发放总量" prop="total_count">
        <el-input-number v-model="formData.total_count" :min="1" :max="1000000" class="w-full" />
      </el-form-item>

      <el-form-item label="每人限领" prop="per_user_limit">
        <el-input-number v-model="formData.per_user_limit" :min="1" :max="formData.total_count" class="w-full" />
        <p class="mt-1 text-xs text-slate-400">不能超过发放总量</p>
      </el-form-item>

      <el-form-item label="有效期模式" prop="validityMode">
        <el-radio-group v-model="validityMode">
          <el-radio value="relative">领取后 N 天</el-radio>
          <el-radio value="fixed">固定日期</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item v-if="validityMode === 'relative'" label="有效天数" prop="usable_days">
        <el-input-number v-model="formData.usable_days" :min="1" :max="365" class="w-full" />
        <p class="mt-1 text-xs text-slate-400">用户领取后 N 天内有效</p>
      </el-form-item>

      <el-form-item v-else label="有效期" prop="dateRange">
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          value-format="YYYY-MM-DDTHH:mm:ssZ"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          class="w-full"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <el-button @click="handleClose" class="h-9 rounded-xl">取消</el-button>
        <el-button type="primary" :loading="submitLoading" class="h-9 rounded-xl" @click="handleSubmit">
          确定
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { ElMessage, type FormInstance } from 'element-plus'
import { CreateCouponApi } from '@/api/coupon'
import type { CreateCouponReq } from '@/types/coupon'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits(['update:visible', 'success'])

const formRef = ref<FormInstance | null>(null)
const submitLoading = ref(false)
const validityMode = ref<'relative' | 'fixed'>('relative')
const dateRange = ref<[string, string] | null>(null)

const formData = ref<CreateCouponReq>({
  coupon_name: '',
  coupon_type: 'full_reduction',
  threshold_amount: 0,
  discount_amount: 0,
  total_count: 100,
  per_user_limit: 1,
  usable_days: 30,
  start_time: '',
  end_time: '',
})

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
})

const rules = {
  coupon_name: [
    { required: true, message: '请输入优惠券名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2-50 个字符', trigger: 'blur' },
  ],
  coupon_type: [{ required: true, message: '请选择券类型', trigger: 'change' }],
  threshold_amount: [{ required: true, message: '请输入使用门槛', trigger: 'blur' }],
  discount_amount: [{ required: true, message: '请输入优惠力度', trigger: 'blur' }],
  total_count: [{ required: true, message: '请输入发放总量', trigger: 'blur' }],
  per_user_limit: [{ required: true, message: '请输入每人限领数量', trigger: 'blur' }],
  usable_days: [{ required: true, message: '请输入有效天数', trigger: 'blur' }],
  dateRange: [{ required: true, message: '请选择有效期', trigger: 'change' }],
}

watch(
  () => props.visible,
  (val) => {
    if (val) {
      validityMode.value = 'relative'
      dateRange.value = null
      formData.value = {
        coupon_name: '',
        coupon_type: 'full_reduction',
        threshold_amount: 0,
        discount_amount: 0,
        total_count: 100,
        per_user_limit: 1,
        usable_days: 30,
        start_time: '',
        end_time: '',
      }
      formRef.value?.clearValidate()
    }
  }
)

const handleClose = () => {
  formRef.value?.resetFields()
  emit('update:visible', false)
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return

    // 有效期模式二选一:相对天数 or 固定日期(对应后端 11008 校验)
    if (validityMode.value === 'relative') {
      // 不发送 start_time/end_time:后端 time.Time 无法解析空字符串,
      // undefined 会被 axios 序列化跳过,后端取零值(与 usable_days>0 分支匹配)
      formData.value.start_time = undefined
      formData.value.end_time = undefined
    } else {
      if (!dateRange.value || dateRange.value.length !== 2) {
        ElMessage.warning('请选择有效期')
        return
      }
      formData.value.usable_days = 0
      formData.value.start_time = dateRange.value[0]
      formData.value.end_time = dateRange.value[1]
    }

    submitLoading.value = true
    try {
      await CreateCouponApi(formData.value)
      ElMessage.success('新增优惠券成功')
      emit('success')
      handleClose()
    } catch (error: any) {
      console.error('创建优惠券失败:', error)
      ElMessage.error(error?.response?.data?.message || '新增优惠券失败')
    } finally {
      submitLoading.value = false
    }
  })
}
</script>
