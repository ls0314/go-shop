<template>
  <!-- 未登录：显示登录/注册按钮 -->
  <div v-if="!isLogin" class="flex items-center gap-2">
    <el-button class="h-8 rounded-lg text-sm" @click="$router.push('/login')">登录</el-button>
    <el-button type="primary" class="h-8 rounded-lg text-sm" @click="$router.push('/register')">注册</el-button>
  </div>

  <!-- 已登录：显示头像下拉菜单 -->
  <div v-else class="relative">
    <button
        ref="avatarButtonRef"
        type="button"
        class="flex h-10 w-10 items-center justify-center rounded-full bg-indigo-500 text-sm font-bold text-white shadow-sm transition hover:bg-indigo-600"
        @click="toggleMenu"
    >
      {{ avatarText }}
    </button>

    <Transition name="dropdown">
      <div
          v-if="visible"
          ref="menuRef"
          class="absolute right-0 top-12 z-50 w-40 overflow-hidden rounded-xl border border-slate-200 bg-white shadow-lg dark:border-slate-700 dark:bg-slate-800"
      >
        <button
            type="button"
            class="block w-full px-4 py-3 text-left text-sm text-slate-700 transition hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-700"
            @click="handleAddress"
        >
          收货地址
        </button>


        <button
            type="button"
            class="block w-full px-4 py-3 text-left text-sm text-slate-700 transition hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-700"
            @click="handleOrder"
          >
            我的订单
          </button>
        <button
            type="button"
            class="block w-full px-4 py-3 text-left text-sm text-slate-700 transition hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-700"
            @click="handleCoupon"
        >
          我的优惠券
        </button>
        <button
            type="button"
            class="block w-full px-4 py-3 text-left text-sm text-slate-700 transition hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-700"
            @click="handleProfile"
        >
          个人信息
        </button>

        <button
            v-if="hasAdmin"
            type="button"
            class="block w-full border-b border-slate-100 px-4 py-3 text-left text-sm text-indigo-600 transition hover:bg-indigo-50 dark:border-slate-700 dark:text-indigo-400 dark:hover:bg-slate-700"
            @click="handleSwitchToAdmin"
        >
          管理端
        </button>

        <button
            type="button"
            class="block w-full px-4 py-3 text-left text-sm text-red-500 transition hover:bg-red-50 dark:hover:bg-slate-700"
            @click="handleLogout"
        >
          登出
        </button>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/pinia/modules/user'
import { useRouterStore } from '@/pinia/modules/router'

const router = useRouter()
const userStore = useUserStore()
const routerStore = useRouterStore()

const visible = ref(false)
const menuRef = ref<HTMLElement | null>(null)
const avatarButtonRef = ref<HTMLElement | null>(null)

const isLogin = computed(() => !!userStore.userToken.access_token)

const hasAdmin = computed(() => routerStore.hasAdmin)

const avatarText = computed(() => {
  const username = userStore.userInfo?.username || 'U'
  return username.slice(0, 1).toUpperCase()
})

// 登录后自动检测管理端权限
watch(isLogin, async (val) => {
  if (val) {
    await routerStore.CheckAdminAccess(userStore.userInfo?.user_id)
  }
})

const toggleMenu = () => {
  visible.value = !visible.value
}

const closeMenu = () => {
  visible.value = false
}

const handleOrder = () => {
  closeMenu()
  router.push('/shop/order/list')
}

const handleCoupon = () => {
  closeMenu()
  // tab=mine 直达"我的卡券"页签
  router.push('/shop/coupon?tab=mine')
}

const handleAddress = () => {
  closeMenu()
  router.push('/shop/address')
}

const handleProfile = async () => {
  closeMenu()
  ElMessage.info('个人信息功能开发中')
}

const handleSwitchToAdmin = () => {
  closeMenu()
  router.push('/home')
}

const handleLogout = () => {
  closeMenu()
  userStore.Logout()
}

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as Node

  if (
      menuRef.value?.contains(target) ||
      avatarButtonRef.value?.contains(target)
  ) {
    return
  }

  closeMenu()
}

onMounted(async () => {
  document.addEventListener('click', handleClickOutside)
  // 页面刷新后 watch 不会触发，需主动检查管理端权限
  if (isLogin.value) {
    await routerStore.CheckAdminAccess(userStore.userInfo?.user_id)
  }
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.dropdown-enter-active,
.dropdown-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
</style>
