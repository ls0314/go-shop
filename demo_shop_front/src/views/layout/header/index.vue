<template>
  <header class="flex h-16 shrink-0 items-center justify-between border-b border-slate-200 bg-white px-4 dark:border-slate-800 dark:bg-slate-900 md:px-6">
    <div class="flex items-center gap-4">
      <button
          type="button"
          class="grid h-9 w-9 place-items-center rounded-xl border border-slate-200 bg-white text-slate-600 transition hover:bg-slate-100 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-300 dark:hover:bg-slate-700 md:hidden"
          @click="emit('open-mobile')"
      >
        ☰
      </button>

      <button
          type="button"
          class="hidden h-9 w-9 place-items-center rounded-xl border border-slate-200 bg-white text-slate-600 transition hover:bg-slate-100 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-300 dark:hover:bg-slate-700 md:grid"
          @click="emit('toggle')"
      >
        {{ collapsed ? '»' : '«' }}
      </button>

      <RouterLink
          to="/home"
          class="text-xl font-bold text-indigo-500"
      >
        Ds_demo
      </RouterLink>

      <nav class="hidden items-center gap-2 md:flex">
        <RouterLink
            v-for="item in menus"
            :key="item.path"
            :to="item.path"
            class="rounded-lg px-3 py-2 text-sm font-medium text-indigo-500 transition hover:bg-indigo-50 dark:hover:bg-slate-800"
            :class="isActive(item.path) ? 'bg-indigo-50 dark:bg-slate-800' : ''"
        >
          {{ item.title }}
        </RouterLink>
      </nav>
    </div>

    <div class="flex items-center gap-2">
      <UserAvatarMenu />
    </div>
  </header>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import UserAvatarMenu from "./UserAvatarMenu.vue";

defineProps<{
  collapsed: boolean
}>()

const emit = defineEmits<{
  toggle: []
  'open-mobile': []
}>()

const route = useRoute()
const isDark = ref(false)

const menus = [
  {
    title: '首页',
    path: '/home'
  },
  {
    title: '关于',
    path: '/about'
  }
]

const isActive = (path: string) => {
  return route.path === path
}

const toggleDark = () => {
  document.documentElement.classList.toggle('dark')
  isDark.value = document.documentElement.classList.contains('dark')
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

onMounted(() => {
  const theme = localStorage.getItem('theme')

  if (theme === 'dark') {
    document.documentElement.classList.add('dark')
  }

  if (theme === 'light') {
    document.documentElement.classList.remove('dark')
  }

  isDark.value = document.documentElement.classList.contains('dark')
})
</script>