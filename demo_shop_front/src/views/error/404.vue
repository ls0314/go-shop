<template>
  <div class="flex h-screen flex-col items-center justify-center bg-slate-50 dark:bg-slate-900">
    <h1 class="text-8xl font-bold text-slate-300 dark:text-slate-600">404</h1>
    <p class="mt-4 text-lg text-slate-500 dark:text-slate-400">页面不存在</p>
    <el-button type="primary" class="mt-6 h-9 rounded-xl" @click="$router.replace('/home')">
      返回首页
    </el-button>
  </div>
</template>
