<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center gap-3">
      <el-button @click="$router.back()" class="h-9 rounded-xl" text>
        <el-icon><ArrowLeft /></el-icon>
      </el-button>
      <h1 class="text-2xl font-bold text-slate-800 dark:text-white">
        {{ isEdit ? '编辑商品 #' + spuId : '新增商品' }}
      </h1>
    </div>

    <el-form :model="form" label-width="100px" class="max-w-4xl space-y-6" ref="formRef">
      <!-- ==================== 基本信息 ==================== -->
      <div class="rounded-xl border border-slate-200 bg-white p-6 dark:border-slate-700 dark:bg-slate-800">
        <h2 class="mb-4 text-lg font-semibold text-slate-800 dark:text-white">基本信息</h2>
        <div class="grid grid-cols-2 gap-4">
          <el-form-item label="商品名称" required>
            <el-input v-model="form.spu_name" placeholder="1-200 字符" maxlength="200" />
          </el-form-item>
          <el-form-item label="品牌">
            <el-input v-model="form.brand" placeholder="品牌名" />
          </el-form-item>
          <el-form-item label="所属类目" required>
            <el-cascader
              v-model="form.category_id"
              :options="categoryOptions"
              :props="{ checkStrictly: true, label: 'category_name', value: 'category_id', emitPath: false }"
              placeholder="选择叶子类目"
              clearable
              class="w-full"
            />
          </el-form-item>
          <el-form-item label="优先级">
            <el-input-number v-model="form.priority" :min="0" :max="9999" placeholder="越大越靠前" />
          </el-form-item>
        </div>
        <el-form-item label="商品描述">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="商品描述（支持 HTML）" />
        </el-form-item>
        <el-form-item label="主图">
          <el-input v-model="form.main_image" placeholder="主图 URL 地址" />
        </el-form-item>
      </div>

      <!-- ==================== 规格模板 ==================== -->
      <div class="rounded-xl border border-slate-200 bg-white p-6 dark:border-slate-700 dark:bg-slate-800">
        <div class="mb-4 flex items-center justify-between">
          <h2 class="text-lg font-semibold text-slate-800 dark:text-white">规格模板</h2>
          <el-button size="small" @click="addSpec">
            <el-icon class="mr-1"><Plus /></el-icon>添加规格
          </el-button>
        </div>

        <div v-if="specTemplate.length === 0" class="py-8 text-center text-sm text-slate-400">
          暂无规格，单规格商品可不填
        </div>

        <div v-for="(spec, si) in specTemplate" :key="si" class="mb-3 flex items-start gap-3 rounded-lg border border-slate-100 p-3 dark:border-slate-600">
          <div class="flex-1 space-y-2">
            <el-input v-model="spec.name" placeholder="规格名（如：颜色）" class="w-48" />
            <div class="flex flex-wrap items-center gap-2">
              <el-tag
                v-for="(v, vi) in spec.values"
                :key="vi"
                closable
                @close="removeSpecValue(si, vi)"
                class="rounded-lg"
              >{{ v }}</el-tag>
              <el-input
                v-if="spec.inputVisible"
                v-model="spec.inputValue"
                ref="specInputRef"
                size="small"
                class="w-24"
                @keyup.enter="confirmSpecValue(si)"
                @blur="confirmSpecValue(si)"
              />
              <el-button v-else size="small" @click="showSpecInput(si)">+ 添加值</el-button>
            </div>
          </div>
          <el-button type="danger" size="small" text @click="removeSpec(si)">删除</el-button>
        </div>
      </div>

      <!-- ==================== SKU 表格 ==================== -->
      <div class="rounded-xl border border-slate-200 bg-white p-6 dark:border-slate-700 dark:bg-slate-800">
        <div class="mb-4 flex items-center justify-between">
          <h2 class="text-lg font-semibold text-slate-800 dark:text-white">
            SKU 列表
            <span class="ml-2 text-sm font-normal text-slate-400">
              {{ specTemplate.length > 0 ? '（笛卡尔积 ' + cartesianCount + ' 个）' : '（共 ' + form.sku_list.length + ' 个）' }}
            </span>
          </h2>
          <el-button size="small" type="primary" @click="generateSkuTable" :disabled="specTemplate.length === 0">生成 SKU</el-button>
        </div>

        <el-table :data="form.sku_list" border size="small" max-height="500">
          <el-table-column label="SKU 名称" min-width="160">
            <template #default="{ row, $index }">
              <el-input v-model="row.sku_name" :placeholder="autoSkuName($index)" size="small" />
            </template>
          </el-table-column>
          <el-table-column label="规格值" min-width="180">
            <template #default="{ row }">
              <span class="text-xs text-slate-600">{{ specDisplay(row.spec_values) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="售价" width="130">
            <template #default="{ row }">
              <el-input-number v-model="row.price" :min="0.01" :precision="2" size="small" controls-position="right" class="w-full" />
            </template>
          </el-table-column>
          <el-table-column label="成本价" width="130">
            <template #default="{ row }">
              <el-input-number v-model="row.cost_price" :min="0" :precision="2" size="small" controls-position="right" class="w-full" />
            </template>
          </el-table-column>
          <el-table-column label="库存" width="110">
            <template #default="{ row }">
              <el-input-number v-model="row.stock" :min="0" size="small" controls-position="right" class="w-full" />
            </template>
          </el-table-column>
          <el-table-column label="编码" width="140">
            <template #default="{ row }">
              <el-input v-model="row.sku_code" size="small" placeholder="SKU 编码" />
            </template>
          </el-table-column>
          <el-table-column label="状态" width="90" align="center">
            <template #default="{ row }">
              <el-switch v-model="row.sku_status" active-value="active" inactive-value="inactive" size="small" />
            </template>
          </el-table-column>
        </el-table>
        <div v-if="form.sku_list.length === 0" class="py-8 text-center text-sm text-slate-400">
          请先配置规格模板，然后点击"生成 SKU"
        </div>
      </div>

      <!-- ==================== 图片列表 ==================== -->
      <div class="rounded-xl border border-slate-200 bg-white p-6 dark:border-slate-700 dark:bg-slate-800">
        <div class="mb-4 flex items-center justify-between">
          <h2 class="text-lg font-semibold text-slate-800 dark:text-white">商品图片</h2>
          <div class="flex gap-2">
            <el-button size="small" @click="triggerUpload">
              <el-icon class="mr-1"><Upload /></el-icon>上传图片
            </el-button>
            <el-button size="small" @click="addImageUrl">
              <el-icon class="mr-1"><Link /></el-icon>添加 URL
            </el-button>
          </div>
        </div>

        <!-- 隐藏的文件选择器 -->
        <input
          ref="fileInputRef"
          type="file"
          accept="image/*"
          multiple
          class="hidden"
          @change="handleFileChange"
        />

        <!-- 上传进度提示 -->
        <div v-if="uploading" class="mb-3 flex items-center gap-2 text-sm text-blue-600">
          <el-icon class="animate-spin"><Loading /></el-icon>
          正在上传 {{ uploadIndex + 1 }} / {{ uploadTotal }} ...
        </div>

        <div class="space-y-2">
          <div v-for="(img, ii) in form.image_list" :key="ii" class="flex items-center gap-3 rounded-lg border border-slate-100 p-2 dark:border-slate-600">
            <el-image v-if="img.image_url" :src="img.image_url" fit="cover" class="h-16 w-16 rounded-lg" />
            <div v-else class="flex h-16 w-16 items-center justify-center rounded-lg bg-slate-100 text-slate-400 text-xs dark:bg-slate-700">无图</div>
            <div class="flex-1 space-y-1">
              <el-input v-model="img.image_url" size="small" placeholder="图片 URL 或上传后自动填充" />
              <div class="flex items-center gap-3">
                <span class="text-xs text-slate-400">排序</span>
                <el-input-number v-model="img.sort_order" :min="0" size="small" controls-position="right" class="w-28" />
                <el-checkbox v-model="img.is_main" size="small" @change="(v: boolean) => onMainChange(ii, v)">主图</el-checkbox>
              </div>
            </div>
            <el-button type="danger" size="small" text @click="removeImage(ii)">删除</el-button>
          </div>
        </div>
        <div v-if="form.image_list.length === 0" class="py-8 text-center text-sm text-slate-400">
          点击"上传图片"选择本地文件，或"添加 URL"手动输入
        </div>
      </div>

      <!-- ==================== 提交 ==================== -->
      <div class="flex justify-end gap-3">
        <el-button @click="$router.back()" class="h-10 rounded-xl px-6">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting" class="h-10 rounded-xl px-6">
          {{ isEdit ? '保存修改' : '创建商品' }}
        </el-button>
      </div>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Plus, Upload, Link, Loading } from '@element-plus/icons-vue'
import { useProductStore } from '@/pinia/modules/product'
import { useCategoryStore } from '@/pinia/modules/category'
import { uploadFile } from '@/api/upload'
import type { CreateSkuItem, SpecTemplateItem, ProductImage } from '@/types/product'

const route = useRoute()
const router = useRouter()
const productStore = useProductStore()
const categoryStore = useCategoryStore()

const isEdit = computed(() => !!route.params.id)
const spuId = computed(() => Number(route.params.id))
const submitting = ref(false)
const categoryOptions = ref<any[]>([])

// --- 表单 ---
interface SkuFormItem extends CreateSkuItem {
  sku_id?: number
}
const form = reactive({
  spu_name: '',
  category_id: null as number | null,
  brand: '',
  description: '',
  main_image: '',
  spu_status:"published",
  priority: 0,
  sku_list: [] as SkuFormItem[],
  image_list: [] as ProductImage[],
})

// --- 规格模板 ---
interface SpecItemExt extends SpecTemplateItem {
  inputVisible: boolean
  inputValue: string
}
const specTemplate = ref<SpecItemExt[]>([])

const addSpec = () => {
  specTemplate.value.push({ name: '', values: [], inputVisible: false, inputValue: '' })
}
const removeSpec = (i: number) => {
  specTemplate.value.splice(i, 1)
}
const showSpecInput = (i: number) => {
  specTemplate.value[i].inputVisible = true
  nextTick(() => {
    // focus handled by el-input
  })
}
const confirmSpecValue = (i: number) => {
  const spec = specTemplate.value[i]
  const v = spec.inputValue.trim()
  if (v && !spec.values.includes(v)) {
    spec.values.push(v)
  }
  spec.inputVisible = false
  spec.inputValue = ''
}
const removeSpecValue = (si: number, vi: number) => {
  specTemplate.value[si].values.splice(vi, 1)
}

// 笛卡尔积数量
const cartesianCount = computed(() => {
  if (specTemplate.value.length === 0) return 0
  return specTemplate.value.reduce((acc, s) => acc * Math.max(s.values.length, 1), 1)
})

// 自动拼接 SKU 名称
const autoSkuName = (i: number) => {
  const sku = form.sku_list[i]
  if (!sku || !sku.spec_values) return ''
  return Object.values(sku.spec_values).join(' / ')
}

// 规格值展示
const specDisplay = (sv: Record<string, string> | undefined) => {
  if (!sv) return '-'
  return Object.entries(sv).map(([k, v]) => `${k}:${v}`).join('，')
}

// 生成 SKU 表
const generateSkuTable = () => {
  // 验证规格模板
  for (const s of specTemplate.value) {
    if (!s.name.trim() || s.values.length === 0) {
      ElMessage.warning('请完善规格模板（名称和值不能为空）')
      return
    }
  }

  // 保留旧 SKU 中仍匹配的（编辑模式）
  const oldMap = new Map<string, SkuFormItem>()
  for (const sku of form.sku_list) {
    if (sku.spec_values) {
      oldMap.set(JSON.stringify(sku.spec_values), { ...sku })
    }
  }

  // 笛卡尔积
  const dims = specTemplate.value.map(s => ({ name: s.name, values: s.values }))
  const combinations = cartesianProduct(dims)
  form.sku_list = combinations.map(combo => {
    const key = JSON.stringify(combo)
    const old = oldMap.get(key)
    if (old) return old
    return {
      sku_name: '',
      spec_values: { ...combo } as any,
      price: 0,
      cost_price: 0,
      stock: 0,
      sku_code: '',
      sku_image: '',
      sku_status: 'active',
    } as SkuFormItem
  })
}

const cartesianProduct = (dims: { name: string; values: string[] }[]): Record<string, string>[] => {
  if (dims.length === 0) return [{}]
  const [first, ...rest] = dims
  const restProduct = cartesianProduct(rest)
  const result: Record<string, string>[] = []
  for (const val of first.values) {
    for (const r of restProduct) {
      result.push({ [first.name]: val, ...r })
    }
  }
  return result
}

// --- 图片 ---
const fileInputRef = ref<HTMLInputElement | null>(null)
const uploading = ref(false)
const uploadIndex = ref(0)
const uploadTotal = ref(0)

/** 触发文件选择 */
const triggerUpload = () => {
  fileInputRef.value?.click()
}

/** 选择文件后逐个上传 */
const handleFileChange = async (e: Event) => {
  const input = e.target as HTMLInputElement
  const files = input.files
  if (!files || files.length === 0) return

  uploading.value = true
  uploadTotal.value = files.length
  uploadIndex.value = 0

  for (let i = 0; i < files.length; i++) {
    uploadIndex.value = i
    try {
      const url = await uploadFile(files[i])
      form.image_list.push({
        image_url: url,
        sort_order: form.image_list.length,
        is_main: form.image_list.length === 0,
      })
    } catch {
      ElMessage.error(`文件 ${files[i].name} 上传失败`)
    }
  }

  uploading.value = false
  input.value = '' // 清空 input 以支持重复选择同一文件
}

/** 手动添加 URL */
const addImageUrl = () => {
  form.image_list.push({
    image_url: '',
    sort_order: form.image_list.length,
    is_main: form.image_list.length === 0,
  })
}

/** 删除图片 */
const removeImage = (i: number) => {
  form.image_list.splice(i, 1)
}

/** 主图互斥：只能设一张 */
const onMainChange = (idx: number, checked: boolean) => {
  if (checked) {
    form.image_list.forEach((img, i) => {
      img.is_main = i === idx
    })
    form.main_image = form.image_list[idx].image_url  // 同步到 SPU 主图
  }
}

// --- 提交 ---
const handleSubmit = async () => {
  if (!form.spu_name.trim()) { ElMessage.warning('请输入商品名称'); return }
  if (!form.category_id) { ElMessage.warning('请选择所属类目'); return }
  if (form.sku_list.length === 0) { ElMessage.warning('请生成 SKU 列表'); return }

  // 校验 SKU
  for (const sku of form.sku_list) {
    if (!sku.price || sku.price <= 0) { ElMessage.warning('每个 SKU 的售价必须大于 0'); return }
  }

  submitting.value = true
  try {
    const payload: any = {
      spu_name: form.spu_name,
      category_id: form.category_id,
      brand: form.brand,
      description: form.description,
      spu_status:"published",
      main_image: form.main_image,
      spec_template: specTemplate.value.map(s => ({ name: s.name, values: s.values })),
      priority: form.priority,
      sku_list: form.sku_list.map((s, i) => ({
        sku_id: (s as any).sku_id || 0,
        // 名称为空时用规格值拼接兜底(如 "红色 / XL"),与输入框的
        // placeholder 提示一致(placeholder 显示的就是 autoSkuName)。
        //
        // 为什么必须在这里补而不是让服务端补:BFF 的 .api 把 sku_name
        // 声明成必填,传 undefined 会被 go-zero 以
        // `field "sku_list[0].sku_name" is not set` 挡成 400。
        // 而 DB 该列可空、服务端也不校验 —— 从源头补上是唯一能保证
        // "名称永不为空"的做法(名称留空在商品详情里就是一片空白)。
        //
        // 规格值也为空时(用户没配规格模板)autoSkuName 返回 ''。
        // 最后用"商品名 + 序号"兜底,而**不是**直接用商品名 ——
        // 那样多个 SKU 会同名(全靠 autoSkuName 时名字天然不同,
        // 因为每个 SKU 的规格组合不一样),而同名 SKU 在列表里
        // 完全分不出谁是谁。
        sku_name: s.sku_name?.trim() || autoSkuName(i) || `${form.spu_name} #${i + 1}`,
        spec_values: s.spec_values,
        price: s.price,
        cost_price: s.cost_price || 0,
        stock: s.stock || 0,
        sku_code: s.sku_code || undefined,
        sku_image: s.sku_image || undefined,
        sku_status: s.sku_status || 'active',
      })),
      image_list: form.image_list.map(img => ({
        image_id: (img as any).image_id || 0,
        image_url: img.image_url,
        sort_order: img.sort_order || 0,
        is_main: img.is_main || false,
      })),
    }

    if (isEdit.value) {
      // 全量更新
      const res = await productStore.UpdateProductFull(spuId.value, payload)
      if (res) {
        ElMessage.success('更新成功')
        router.back()
      } else {
        ElMessage.error('更新失败')
      }
    } else {
      const res = await productStore.CreateProduct(payload)
      if (res) {
        ElMessage.success(`创建成功，SPU ID: ${res.spu_id}`)
        router.back()
      } else {
        ElMessage.error('创建失败')
      }
    }
  } finally {
    submitting.value = false
  }
}

// --- 编辑模式：加载已有数据 ---
onMounted(async () => {
  // 加载类目树用于级联选择
  try {
    await categoryStore.GetCategoryTree({ include_disabled: false })
    categoryOptions.value = categoryStore.categoryTree
  } catch { /* ignore */ }

  if (isEdit.value) {
    const detail = await productStore.GetProduct(spuId.value)
    if (detail) {
      form.spu_name = detail.spu_name
      form.category_id = detail.category_id
      form.brand = detail.brand || ''
      form.description = detail.description || ''
      form.main_image = detail.main_image || ''
      form.priority = detail.priority

      // 规格模板
      if (detail.spec_template && Array.isArray(detail.spec_template)) {
        specTemplate.value = (detail.spec_template as SpecTemplateItem[]).map(s => ({
          ...s,
          inputVisible: false,
          inputValue: '',
        }))
      }

      // SKU 列表
      form.sku_list = (detail.sku_list || []).map((s: any) => ({
        sku_id: s.sku_id,
        sku_name: s.sku_name || '',
        spec_values: s.spec_value,
        price: s.price,
        cost_price: s.cost_price,
        stock: s.stock,
        sku_code: s.sku_code || '',
        sku_image: s.sku_image || '',
        sku_status: s.sku_status || 'active',
      }))

      // 图片列表
      form.image_list = (detail.image_list || []).map((img: any) => ({
        image_id: img.image_id,
        image_url: img.image_url,
        sort_order: img.sort_order,
        is_main: img.is_main,
      }))
    }
  }
})
</script>
