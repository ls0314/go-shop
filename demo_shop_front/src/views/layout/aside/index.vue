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
      <button
          v-for="item in menus"
          :key="item.path"
          type="button"
          class="flex h-11 w-full items-center gap-3 rounded-xl px-3 text-sm transition hover:bg-white/10 hover:text-white"
          :class="isActive(item.path) ? 'bg-white text-slate-900' : 'text-slate-400'"
          @click="handleMenuClick(item.path)"
      >
        <span class="w-6 shrink-0 text-center text-base">
          {{ item.icon }}
        </span>

        <span
            v-show="!collapsed"
            class="whitespace-nowrap"
        >
          {{ item.title }}
        </span>
      </button>
    </nav>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

interface MenuItem {
  title: string
  path: string
  icon: string
}

const props = defineProps<{
  collapsed: boolean
  mobileOpen: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const route = useRoute()
const router = useRouter()

const menus: MenuItem[] = [
  {
    title: '占位1',
    path: '/home',
    icon: '🏠'
  },
  {
    title: '占位2',
    path: '/home',
    icon: '📘'
  },
  {
    title: '占位3',
    path: '/home',
    icon: '📝'
  },
  {
    title: '占位4',
    path: '/home',
    icon: '🔐'
  }
]

const asideClass = computed(() => {
  const desktopWidth = props.collapsed ? 'md:w-20' : 'md:w-60'
  const mobileState = props.mobileOpen ? 'translate-x-0 w-60' : '-translate-x-full w-60'

  return [desktopWidth, mobileState]
})

const isActive = (path: string) => {
  return route.path === path
}

const handleMenuClick = async (path: string) => {
  emit('close')

  if (route.path !== path) {
    await router.push(path)
  }
}
</script>