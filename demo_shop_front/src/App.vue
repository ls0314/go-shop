<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
// App component
</script>

<template>
  <div class="app">
    <el-menu
        v-if="!route.meta.hideNav"
        :router="true"
        mode="horizontal"
    >
      <el-menu-item index="/">Home</el-menu-item>
      <el-menu-item index="/about">About</el-menu-item>
      <el-menu-item index="/register">Register</el-menu-item>
      <el-menu-item index="/login">Login</el-menu-item>
    </el-menu>
    <router-view></router-view>
  </div>
</template>

<style scoped>
.app {
  min-height: 100vh;
}
</style>
