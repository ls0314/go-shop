<template>
  <div class="text-center">
    <h1 class="text-4xl font-bold text-slate-800 dark:text-slate-100">
      Home Page
    </h1>
    <el-button type="primary">Element Plus Button</el-button>

    <!-- 测试按钮 -->
    <el-button type="success" @click="handleGetUserInfo" style="margin-left: 10px;">
      获取用户信息
    </el-button>

    <!-- 显示用户信息 -->
    <div v-if="userInfo" style="margin-top: 20px; padding: 15px; border: 1px solid #ddd;">
      <h3>用户信息：</h3>
      <pre>{{ JSON.stringify(userInfo, null, 2) }}</pre>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useUserStore} from '@/pinia/modules/user.ts'

const userInfo = ref<any>(null)

const userStore = useUserStore()

async function handleGetUserInfo() {
  const res = await userStore.GetUserInfo(userInfo.value)
  userInfo.value = res.data.data
  console.log(res.data.data)
}

</script>

<style scoped>

</style>