<template>
  <div class="flex h-screen w-screen overflow-hidden bg-gray-50 text-slate-700 dark:bg-slate-900 dark:text-slate-200">

    <div
        v-if="mobileOpen"
        class="fixed inset-0 z-40 bg-black/40 md:hidden"
        @click="mobileOpen = false"
    />

    <section class="flex h-screen min-w-0 flex-1 flex-col">
      <LayoutHeader
          :collapsed="collapsed"
          @toggle="collapsed = !collapsed"
          @open-mobile="mobileOpen = true"
      />

      <main class="min-h-0 flex-1 overflow-auto bg-gray-50 p-4 dark:bg-slate-900 md:p-6">
        <RouterView v-slot="{ Component, route }">
          <Transition name="fade" mode="out-in">
            <component
                :is="Component"
                :key="route.fullPath"
            />
          </Transition>
        </RouterView>
      </main>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import LayoutHeader from './header/index.vue'

const collapsed = ref(false)
const mobileOpen = ref(false)
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.18s ease, transform 0.18s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(6px);
}
</style>