<template>
  <div v-if="item.is_visible">
    <!-- 有子菜单 -->
    <div v-if="hasChildren">
      <button
          type="button"
          class="flex h-11 w-full items-center gap-3 rounded-xl px-3 text-sm transition hover:bg-white/10 hover:text-white"
          :class="isActive ? 'bg-white text-slate-900' : 'text-slate-400'"
          @click="toggleOpen"
      >
        <span class="w-6 shrink-0 text-center text-base">
          {{ showIcon }}
        </span>

        <span
            v-show="!collapsed"
            class="flex-1 whitespace-nowrap text-left"
        >
          {{ item.menu_name }}
        </span>

        <span
            v-show="!collapsed"
            class="text-xs"
        >
          {{ open ? '⌃' : '⌄' }}
        </span>
      </button>

      <div
          v-show="open && !collapsed"
          class="mt-2 space-y-1 pl-4"
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
        class="flex h-11 w-full items-center gap-3 rounded-xl px-3 text-sm transition hover:bg-white/10 hover:text-white"
        :class="isActive ? 'bg-white text-slate-900' : 'text-slate-400'"
        @click="handleClick"
    >
      <span class="w-6 shrink-0 text-center text-base">
        {{ showIcon }}
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

const open = ref(true)

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
  const icon = props.item.meta_info?.icon || props.item.icon

  const iconMap: Record<string, string> = {
    dashboard: '🏠',
    system: '⚙️',
    user: '👤',
    role: '🔐',
    menu: '📋'
  }

  return iconMap[icon || ''] || '📄'
})

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