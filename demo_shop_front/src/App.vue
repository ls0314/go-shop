<script setup lang="ts">
// App component
</script>

<template>
  <div class="app">
    <el-menu :router="true" mode="horizontal">
      <el-menu-item index="/">Home</el-menu-item>
      <el-menu-item index="/about">About</el-menu-item>
    </el-menu>
    <router-view></router-view>
  </div>
</template>

<style scoped>
.app {
  min-height: 100vh;
}
</style>
