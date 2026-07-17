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

      <RouterLink
          to="/shop/home"
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

    <div class="flex items-center gap-3">
      <!-- 购物车图标 -->
      <RouterLink
        to="/shop/cart"
        class="relative flex h-9 w-9 items-center justify-center rounded-xl text-slate-600 transition hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800"
      >
        <el-icon size="20"><ShoppingCart /></el-icon>
        <span
          v-if="cartCount > 0"
          class="absolute -right-1 -top-1 flex h-5 min-w-[20px] items-center justify-center rounded-full bg-rose-500 px-1 text-xs font-bold text-white"
        >
          {{ cartCount > 99 ? '99+' : cartCount }}
        </span>
      </RouterLink>

      <UserAvatarMenu />
    </div>
  </header>
</template>

<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { ShoppingCart } from '@element-plus/icons-vue'
import { useCartStore } from '@/pinia/modules/cart'
import { useUserStore } from '@/pinia/modules/user'
import UserAvatarMenu from "./UserAvatarMenu.vue";

defineProps<{
  collapsed: boolean
}>()

const emit = defineEmits<{
  toggle: []
  'open-mobile': []
}>()

const route = useRoute()
const cartStore = useCartStore()
const userStore = useUserStore()
const isDark = ref(false)
const cartCount = computed(() => cartStore.cartCount)

const menus = [
  {
    title: '首页',
    path: '/shop/home'
  },
  {
    title: '全部商品',
    path: '/shop/product/list'
  },
]

const isActive = (path: string) => {
  return route.path === path
}

const toggleDark = () => {
  document.documentElement.classList.toggle('dark')
  isDark.value = document.documentElement.classList.contains('dark')
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

onMounted(async () => {
  // 仅已登录时获取购物车数量，避免未登录 401 触发 axios 拦截器跳转登录页
  if (userStore.userToken.access_token) {
    await cartStore.GetCartCount()
  }

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