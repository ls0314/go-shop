<template>
  <div v-if="item.is_visible">
    <!-- 有子菜单 -->
    <div v-if="hasChildren">
      <button
          type="button"
          class="group relative flex h-11 w-full items-center gap-3 rounded-lg px-3 text-sm font-medium transition-all duration-200 hover:bg-white/5 hover:text-white"
          :class="[
            isActive
              ? 'bg-gradient-to-r from-indigo-500/20 to-transparent text-white'
              : 'text-slate-400',
            collapsed ? 'justify-center px-0' : ''
          ]"
          @click="toggleOpen"
      >
        <!-- 激活指示条 -->
        <span
            v-if="isActive"
            class="absolute left-0 top-1/2 h-5 w-[3px] -translate-y-1/2 rounded-r-full bg-indigo-400 shadow-sm shadow-indigo-400/50"
        />

        <!-- 图标容器:有图标显示 emoji,无图标显示层级圆点 -->
        <span
            class="grid h-7 w-7 shrink-0 place-items-center rounded-lg bg-white/5 text-base transition-colors group-hover:bg-white/10"
            :class="isActive ? 'bg-indigo-500/25' : ''"
        >
          <span v-if="hasIcon" class="leading-none">{{ showIcon }}</span>
          <span v-else class="h-1.5 w-1.5 rounded-full bg-current opacity-40" />
        </span>

        <span
            v-show="!collapsed"
            class="flex-1 whitespace-nowrap text-left"
        >
          {{ item.menu_name }}
        </span>

        <!-- 展开箭头：SVG chevron + 旋转动画 -->
        <svg
            v-show="!collapsed"
            class="h-3.5 w-3.5 text-slate-500 transition-transform duration-200 group-hover:text-slate-300"
            :class="open ? 'rotate-180' : ''"
            viewBox="0 0 20 20"
            fill="currentColor"
        >
          <path
              fill-rule="evenodd"
              d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z"
              clip-rule="evenodd"
          />
        </svg>
      </button>

      <!-- 子菜单：左侧引导线 -->
      <div
          v-show="open && !collapsed"
          class="relative ml-3 mt-1 space-y-1 border-l border-white/10 pl-3"
      >
        <AsideMenuItem
            v-for="child in visibleChildren"
            :key="child.menu_id"
            :item="child"
            :collapsed="collapsed"
            @close="emit('close')"
        />
      </div>
    </div>

    <!-- 没有子菜单 -->
    <button
        v-else
        type="button"
        class="group relative flex h-11 w-full items-center gap-3 rounded-lg px-3 text-sm font-medium transition-all duration-200 hover:bg-white/5 hover:text-white"
        :class="[
          isActive
            ? 'bg-gradient-to-r from-indigo-500/20 to-transparent text-white'
            : 'text-slate-400',
          collapsed ? 'justify-center px-0' : ''
        ]"
        @click="handleClick"
    >
      <!-- 激活指示条 -->
      <span
          v-if="isActive"
          class="absolute left-0 top-1/2 h-5 w-[3px] -translate-y-1/2 rounded-r-full bg-indigo-400 shadow-sm shadow-indigo-400/50"
      />

      <!-- 图标容器:有图标显示 emoji,无图标显示层级圆点 -->
      <span
          class="grid h-7 w-7 shrink-0 place-items-center rounded-lg bg-white/5 text-base transition-colors group-hover:bg-white/10"
          :class="isActive ? 'bg-indigo-500/25' : ''"
      >
        <span v-if="hasIcon" class="leading-none">{{ showIcon }}</span>
        <span v-else class="h-1.5 w-1.5 rounded-full bg-current opacity-40" />
      </span>

      <span
          v-show="!collapsed"
          class="whitespace-nowrap"
      >
        {{ item.menu_name }}
      </span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { MenuItem } from '@/types/menu'

defineOptions({
  name: 'AsideMenuItem'
})

const props = defineProps<{
  item: MenuItem
  collapsed: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const route = useRoute()
const router = useRouter()

// 一级菜单(parent_id=0)默认展开,二级及更深默认收起(刷新后只展开顶层骨架);
// 当前路由的激活菜单链由下方 watch 自动展开
const open = ref((props.item.parent_id ?? 0) === 0)

const visibleChildren = computed(() => {
  return (props.item.children || [])
      .filter(child => child.is_visible)
      .sort((a, b) => a.sort_order - b.sort_order)
})

const hasChildren = computed(() => {
  return visibleChildren.value.length > 0
})

const isActive = computed(() => {
  if (route.path === props.item.route_path) {
    return true
  }

  return route.path.startsWith(props.item.route_path + '/')
})

const showIcon = computed(() => {
  const raw = props.item.meta_info?.icon || props.item.icon
  const icon = (raw || '').toLowerCase()

  // key 统一小写,只映射菜单数据中实际出现的 icon 值
  const iconMap: Record<string, string> = {
    // 后台管理(admin)
    dashboard: '🏠', // 后台管理
    system: '⚙️',    // 系统管理
    user: '👤',       // 用户管理
    role: '🔐',       // 角色管理
    menu: '📋',       // 菜单管理
    lock: '🔒',       // 权限管理 / 数据权限
    office: '🏢',     // 部门管理
    // 平台管理(platform)
    setting: '🛠️',    // 平台管理
    list: '🗂️',       // 类目管理
    product: '📦',    // 商品管理
    inventory: '📊',  // 库存管理
    document: '🧾',   // 订单管理
    money: '💰',      // 支付管理
    ticket: '🎟️',     // 优惠券管理
  }

  return iconMap[icon] || ''
})

// 是否为有图标项(决定容器里显示 emoji 还是占位圆点)
const hasIcon = computed(() => !!showIcon.value)

watch(
    () => route.path,
    () => {
      if (isActive.value) {
        open.value = true
      }
    },
    {
      immediate: true
    }
)

const toggleOpen = async () => {
  if (props.collapsed) {
    await handleClick()
    return
  }

  open.value = !open.value
}

const handleClick = async () => {
  emit('close')

  if (route.path !== props.item.route_path) {
    await router.push(props.item.route_path)
  }
}
</script>
