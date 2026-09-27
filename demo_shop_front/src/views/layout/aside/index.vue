<template>
  <aside
      class="fixed left-0 top-0 z-50 flex h-screen flex-col border-r border-white/5 bg-slate-950 text-slate-300 transition-all duration-200 md:static md:translate-x-0"
      :class="asideClass"
  >
    <!-- 品牌区 -->
    <div class="flex h-16 shrink-0 items-center gap-3 border-b border-white/10 px-5">
      <div
          class="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-gradient-to-br from-indigo-500 via-indigo-400 to-emerald-400 text-sm font-bold text-white shadow-lg shadow-indigo-500/25"
      >
        V
      </div>

      <span
          v-show="!collapsed"
          class="whitespace-nowrap text-lg font-bold tracking-wide text-white"
      >
        Ds_demo
      </span>
    </div>

    <!-- 菜单区：可滚动 + 细滚动条 -->
    <nav class="scrollbar-thin flex-1 space-y-1 overflow-y-auto p-3">
      <AsideMenuItem
          v-for="item in visibleMenus"
          :key="item.menu_id"
          :item="item"
          :collapsed="collapsed"
          @close="emit('close')"
      />
    </nav>

    <!-- 底部装饰 -->
    <div
        v-show="!collapsed"
        class="shrink-0 border-t border-white/10 px-5 py-3 text-[11px] text-slate-600"
    >
      Demo Shop · Admin
    </div>
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

<style scoped>
/* 细滚动条：菜单过长时保持美观 */
.scrollbar-thin::-webkit-scrollbar {
  width: 4px;
}
.scrollbar-thin::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.12);
  border-radius: 4px;
}
.scrollbar-thin::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.2);
}
.scrollbar-thin::-webkit-scrollbar-track {
  background: transparent;
}
</style>
