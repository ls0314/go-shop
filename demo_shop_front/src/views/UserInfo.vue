<template>
  <div>
    <h2>用户信息</h2>
    <p>用户ID：{{ user.user_id }}</p>
    <p>用户名：{{ user.username }}</p>
    <button @click="getInfo">获取用户信息</button>
  </div>
</template>

<script setup lang="ts">
import { reactive } from "vue";
import { getUserInfoApi } from "../api/user";

const user = reactive({
  user_id: "",
  username: "",
});

const getInfo = async () => {
  try {
    const res = await getUserInfoApi();
    user.user_id = res.user_id;
    user.username = res.username;
  } catch (e) {
    alert("获取失败，请先登录");
  }
};
</script>