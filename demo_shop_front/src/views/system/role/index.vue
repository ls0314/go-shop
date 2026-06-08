<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="rounded-2xl bg-white p-6 shadow-sm dark:bg-slate-800">
      <h1 class="text-2xl font-bold text-slate-800 dark:text-white">
        用户管理
      </h1>
      <p class="mt-2 text-sm text-slate-500 dark:text-slate-400">
        这是动态路由加载的用户管理页面，用于测试菜单跳转是否正常。
      </p>
    </div>

    <!-- 统计卡片 -->
    <div class="grid gap-4 md:grid-cols-3">
      <div class="rounded-2xl bg-white p-5 shadow-sm dark:bg-slate-800">
        <p class="text-sm text-slate-500 dark:text-slate-400">用户总数</p>
        <p class="mt-2 text-3xl font-bold text-slate-800 dark:text-white">
          128
        </p>
      </div>

      <div class="rounded-2xl bg-white p-5 shadow-sm dark:bg-slate-800">
        <p class="text-sm text-slate-500 dark:text-slate-400">启用用户</p>
        <p class="mt-2 text-3xl font-bold text-emerald-500">
          96
        </p>
      </div>

      <div class="rounded-2xl bg-white p-5 shadow-sm dark:bg-slate-800">
        <p class="text-sm text-slate-500 dark:text-slate-400">禁用用户</p>
        <p class="mt-2 text-3xl font-bold text-rose-500">
          32
        </p>
      </div>
    </div>

    <!-- 操作栏 -->
    <div class="flex flex-col gap-3 rounded-2xl bg-white p-5 shadow-sm dark:bg-slate-800 md:flex-row md:items-center md:justify-between">
      <div>
        <h2 class="text-lg font-semibold text-slate-800 dark:text-white">
          用户列表
        </h2>
        <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">
          当前页面由后端菜单 component: system/user/index 动态加载。
        </p>
      </div>

      <button
          type="button"
          class="rounded-xl bg-emerald-500 px-4 py-2 text-sm font-medium text-white transition hover:bg-emerald-600"
      >
        新增用户
      </button>
    </div>

    <!-- 表格 -->
    <div class="overflow-hidden rounded-2xl bg-white shadow-sm dark:bg-slate-800">
      <table class="w-full border-collapse text-left text-sm">
        <thead class="bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300">
        <tr>
          <th class="px-5 py-3">ID</th>
          <th class="px-5 py-3">用户名</th>
          <th class="px-5 py-3">角色</th>
          <th class="px-5 py-3">状态</th>
          <th class="px-5 py-3">创建时间</th>
          <th class="px-5 py-3">操作</th>
        </tr>
        </thead>

        <tbody>
        <tr
            v-for="user in users"
            :key="user.id"
            class="border-t border-slate-100 dark:border-slate-700"
        >
          <td class="px-5 py-4">
            {{ user.id }}
          </td>

          <td class="px-5 py-4 font-medium text-slate-800 dark:text-white">
            {{ user.username }}
          </td>

          <td class="px-5 py-4">
            {{ user.role }}
          </td>

          <td class="px-5 py-4">
              <span
                  class="rounded-full px-3 py-1 text-xs font-medium"
                  :class="user.status === '启用'
                  ? 'bg-emerald-100 text-emerald-600'
                  : 'bg-rose-100 text-rose-600'"
              >
                {{ user.status }}
              </span>
          </td>

          <td class="px-5 py-4">
            {{ user.createdAt }}
          </td>

          <td class="px-5 py-4">
            <button
                type="button"
                class="text-emerald-500 hover:text-emerald-600"
            >
              编辑
            </button>
            <button
                type="button"
                class="ml-3 text-rose-500 hover:text-rose-600"
            >
              删除
            </button>
          </td>
        </tr>
        </tbody>
      </table>
    </div>

    <!-- 测试提示 -->
    <div class="rounded-2xl border border-dashed border-emerald-400 bg-emerald-50 p-5 text-sm text-emerald-700 dark:border-emerald-600 dark:bg-emerald-950/40 dark:text-emerald-300">
      如果你能看到这个页面，说明动态菜单和动态路由已经成功加载。
    </div>
  </div>
</template>

<script setup lang="ts">
interface UserItem {
  id: number
  username: string
  role: string
  status: '启用' | '禁用'
  createdAt: string
}

const users: UserItem[] = [
  {
    id: 1,
    username: 'admin',
    role: '超级管理员',
    status: '启用',
    createdAt: '2026-05-13'
  },
  {
    id: 2,
    username: 'test_user',
    role: '普通用户',
    status: '启用',
    createdAt: '2026-05-13'
  },
  {
    id: 3,
    username: 'guest',
    role: '访客',
    status: '禁用',
    createdAt: '2026-05-13'
  }
]
</script>