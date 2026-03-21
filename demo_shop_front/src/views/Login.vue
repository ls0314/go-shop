<template>
  <div class="register-wrapper">
    <!-- 顶部导航栏 -->
    <div class="top-nav">
      <div class="logo-area">
<!--        <div class="mi-logo">mi</div>-->
        <span class="logo-text">demo_shop</span>
      </div>
      <div class="nav-links">
        <a href="#">用户协议</a>
        <a href="#">隐私政策</a>
        <a href="#">帮助中心</a>
        <span class="lang-switch">中文 (简体) ▼</span>
      </div>
    </div>

    <div class="main-content">
      <!-- 左侧装饰 -->
      <div class="left-banner">
        <div class="banner-decoration">
          <div class="circle c1"></div>
          <div class="circle c2"></div>
          <div class="astronaut-placeholder">
            🚀
          </div>
        </div>
      </div>

      <!-- 右侧登录表单 -->
      <div class="right-form-container">
        <div class="form-card">
          <div class="form-tabs">
            <a href="/Login" class="tab-item active">登录</a>
            <a href="/Register" class="tab-item">注册</a>
          </div>

          <form @submit.prevent="handleLogin" class="xiaomi-form">

            <!-- 用户名 -->
            <div class="input-group">
              <input
                  type="text"
                  v-model="username"
                  required
                  placeholder="请输入用户名"
              />
            </div>

            <!-- 密码 -->
            <div class="input-group">
              <input
                  type="password"
                  v-model="password"
                  required
                  placeholder="请输入密码"
              />
            </div>

            <button type="submit" class="submit-btn">
              登录
            </button>

            <div class="form-footer">
              <span>还没有账号？</span>
              <a href="/Register">立即注册</a>
            </div>

          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { loginApi } from '../services/user'

const username = ref('')
const password = ref('')

const handleLogin = async () => {
  if (!username.value || !password.value) {
    alert('请输入用户名和密码')
    return
  }

    const res = await loginApi({
      username: username.value,
      password: password.value
    }).then(res => {
      // console.log('后端返回:', res)

      const {access_token, refresh_token} = res.data

      localStorage.setItem('access_token', access_token)
      localStorage.setItem('refresh_token', refresh_token)
      // console.log(access_token)

      alert('登录成功')

      window.location.href = '/home'
    }).catch(err => {
      console.log('请求失败', err)
      if (err.response && err.response.data) {
        alert(err.response.data)
      } else {
        alert('网络错误或服务器未响应')
      }
    })
}
</script>

<style scoped>
/* 全局重置 */
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

.register-wrapper {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  font-family: "MiSans", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  background-color: #fff;
  color: #333;
}

/* 顶部导航 */
.top-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 40px;
  width: 100%;
}

.logo-area {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 24px;
  font-weight: 500;
  color: #333;
}

.mi-logo {
  width: 40px;
  height: 40px;
  background-color: #ff6700;
  color: white;
  border-radius: 12px; /* 小米新 Logo 圆角 */
  display: flex;
  align-items: center;
  justify-content: flex-start;
  font-weight: bold;
  font-size: 20px;
}

.nav-links {
  display: flex;
  gap: 20px;
  font-size: 14px;
  color: #666;
}

.nav-links a {
  text-decoration: none;
  color: #666;
  transition: color 0.3s;
}

.nav-links a:hover {
  color: #ff6700;
}

.lang-switch {
  cursor: pointer;
  color: #666;
}

/* 主体内容布局 */
.main-content {
  flex: 1;
  display: flex;
  width: 100%;
}

/* 左侧插画区域 */
.left-banner {
  flex: 1;
  background: linear-gradient(135deg, #6a11cb 0%, #2575fc 100%);
  position: relative;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 400px;
  z-index: 1;
}

/* 使用 CSS 模拟梦幻背景 */
.banner-decoration {
  position: relative;
  width: 100%;
  height: 100%;
}

.circle {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(5px);
}

.c1 {
  width: 300px;
  height: 300px;
  top: 10%;
  left: 20%;
  background: radial-gradient(circle, rgba(255,255,255,0.2) 0%, rgba(255,255,255,0) 70%);
}

.c2 {
  width: 150px;
  height: 150px;
  bottom: 20%;
  right: 20%;
  background: rgba(255, 255, 255, 0.15);
}

.astronaut-placeholder {
  position: absolute;
  bottom: 15%;
  left: 15%;
  font-size: 100px;
  filter: drop-shadow(0 10px 20px rgba(0,0,0,0.2));
  animation: float 6s ease-in-out infinite;
}

@keyframes float {
  0% { transform: translateY(0px); }
  50% { transform: translateY(-20px); }
  100% { transform: translateY(0px); }
}

/* 右侧表单容器 */
.right-form-container {
  width: 500px; /* 固定宽度或百分比 */
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  padding: 40px;
}

.form-card {
  width: 100%;
  max-width: 400px;
}

/* 选项卡 */
.form-tabs {
  display: flex;
  gap: 30px;
  margin-bottom: 30px;
  font-size: 20px;
  color: #999;
  font-weight: 500;
}

.tab-item {
  text-decoration: none; /* 1. 去除下划线 */
  color: #999;           /* 2. 强制字体颜色为灰色/黑色 (覆盖默认蓝色) */
  cursor: pointer;
  position: relative;
  padding-bottom: 5px;
}

.tab-item.active {
  color: #333;           /* 激活时改为深黑色 */
  font-weight: bold;
}

.tab-item.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  height: 3px;
  background-color: #ff6700;
  border-radius: 2px;
}

/* 表单样式 */
.xiaomi-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.input-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.input-label {
  font-size: 14px;
  color: #666;
  font-weight: 500;
}

.input-group input {
  width: 100%;
  padding: 12px 15px;
  background-color: #f5f5f5;
  border: 1px solid transparent;
  border-radius: 4px;
  font-size: 16px;
  color: #333;
  transition: all 0.3s;
  outline: none;
}

.input-group input:focus {
  background-color: #fff;
  border-color: #ff6700;
  box-shadow: 0 0 0 2px rgba(255, 103, 0, 0.1);
}

.input-tip {
  font-size: 12px;
  color: #999;
  margin-top: -4px;
}

/* 手机号特殊布局 */
.phone-input-wrapper {
  display: flex;
  background-color: #f5f5f5;
  border-radius: 4px;
  padding: 2px;
  transition: all 0.3s;
}

.phone-input-wrapper:focus-within {
  background-color: #fff;
  border: 1px solid #ff6700;
  box-shadow: 0 0 0 2px rgba(255, 103, 0, 0.1);
}

.country-code {
  padding: 10px 15px;
  color: #333;
  font-size: 16px;
  border-right: 1px solid #e0e0e0;
  display: flex;
  align-items: center;
}

.full-width-input {
  border: none !important;
  background: transparent !important;
  box-shadow: none !important;
  flex: 1;
}

/* 协议 */
.agreement-group {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  font-size: 12px;
  color: #666;
  margin-top: 10px;
}

.agreement-group input {
  margin-top: 3px;
  accent-color: #ff6700;
}

.agreement-group a {
  color: #666;
  text-decoration: underline;
}

/* 提交按钮 */
.submit-btn {
  width: 100%;
  padding: 14px;
  background-color: #ffca99; /* 未激活或默认状态浅色 */
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  margin-top: 10px;
  transition: background-color 0.3s;
}

/* 简单模拟：如果有内容则变深橙色，实际项目中可用 :valid 或 JS 控制 */
.submit-btn:hover {
  background-color: #ff6700;
}

/* 底部链接 */
.form-footer {
  text-align: center;
  margin-top: 20px;
  font-size: 14px;
  color: #666;
}

.form-footer a {
  color: #ff6700;
  text-decoration: none;
  margin-left: 5px;
  font-weight: 500;
}

.form-footer a:hover {
  text-decoration: underline;
}

/* 响应式调整 */
@media (max-width: 900px) {
  .left-banner {
    display: none; /* 小屏幕隐藏左侧插画 */
  }

  .right-form-container {
    width: 100%;
    padding: 20px;
  }

  .top-nav {
    padding: 15px 20px;
    background: rgba(255,255,255,0.9);
  }
}
</style>