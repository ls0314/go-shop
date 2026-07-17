<template>
  <div class="relative">
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
            @click="handleProfile"
        >
          个人信息
        </button>

        <button
            type="button"
            class="block w-full border-b border-slate-100 px-4 py-3 text-left text-sm text-indigo-600 transition hover:bg-indigo-50 dark:border-slate-700 dark:text-indigo-400 dark:hover:bg-slate-700"
            @click="handleSwitchToShop"
        >
          返回商城
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
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/pinia/modules/user'

const router = useRouter()
const userStore = useUserStore()

const visible = ref(false)
const menuRef = ref<HTMLElement | null>(null)
const avatarButtonRef = ref<HTMLElement | null>(null)

const avatarText = computed(() => {
  const username = userStore.userInfo?.username || 'U'
  return username.slice(0, 1).toUpperCase()
})

const toggleMenu = () => {
  visible.value = !visible.value
}

const closeMenu = () => {
  visible.value = false
}

const handleProfile = async () => {
  closeMenu()
  ElMessage.info('个人信息功能开发中')
}

const handleSwitchToShop = () => {
  closeMenu()
  router.push('/shop/home')
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

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
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