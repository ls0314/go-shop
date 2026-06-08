<template>
  <aside
      class="fixed left-0 top-0 z-50 h-screen bg-slate-950 text-slate-300 transition-all duration-200 md:static md:translate-x-0"
      :class="asideClass"
  >
    <div class="flex h-16 items-center gap-3 border-b border-white/10 px-5">
      <div class="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-emerald-500 text-sm font-bold text-white">
        V
      </div>

      <span
          v-show="!collapsed"
          class="whitespace-nowrap text-lg font-bold text-white"
      >
        Ds_demo
      </span>
    </div>

    <nav class="space-y-2 p-3">
      <AsideMenuItem
          v-for="item in visibleMenus"
          :key="item.menu_id"
          :item="item"
          :collapsed="collapsed"
          @close="emit('close')"
      />
    </nav>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouterStore } from '@/pinia/modules/router'
import AsideMenuItem from './menu_item.vue'

const props = defineProps<{
  collapsed: boolean
  mobileOpen: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const routerStore = useRouterStore()

const visibleMenus = computed(() => {
  return routerStore.menuList
      .filter(item => item.is_visible)
      .sort((a, b) => a.sort_order - b.sort_order)
})

const asideClass = computed(() => {
  const desktopWidth = props.collapsed ? 'md:w-20' : 'md:w-60'
  const mobileState = props.mobileOpen ? 'translate-x-0 w-60' : '-translate-x-full w-60'

  return [desktopWidth, mobileState]
})
</script>