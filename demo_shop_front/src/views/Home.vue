<template>
  <div class="home">
    <h1>Home Page</h1>
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
import { getUserInfoApi } from '../services/user'
import { ElMessage } from 'element-plus'

const userInfo = ref<any>(null)

const handleGetUserInfo = async () => {
  try {
    const res = await getUserInfoApi()
    console.log('获取用户信息成功：', res)
    userInfo.value = res.data.data
    console.log('获取用户信息成功：', res.data.data.user_id)
    console.log('获取用户信息成功：', res.data.data.username)
    ElMessage.success('获取用户信息成功')
  } catch (err) {
    console.log('获取用户信息失败：', err)
    ElMessage.error('获取用户信息失败')
  }
}
</script>

<style scoped>
.home {
  padding: 20px;
}
</style>