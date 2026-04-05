import { createRouter, createWebHistory } from 'vue-router'
import {useUserStore} from "../store/user.ts";

const routes = [
  {
    path: '/home',
    name: 'Home',
    component: () => import('../views/Home.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/about',
    name: 'About',
    component: () => import('../views/About.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/Register.vue'),
    meta: { hideNav: true,
        requiresAuth: false}
  },
  {
      path: '/login',
      name: 'Login',
      component: () => import('../views/Login.vue'),
      meta: { hideNav: true,
          requiresAuth: false}
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from) => {
    const userStore = useUserStore()

    const isLogin = !!userStore.access_token
    console.log('路由跳转 →', to.path, '登录状态:', isLogin)


    if (to.meta.requiresAuth && !isLogin) {
            console.log("请登录")
            alert("请登录")
            return '/login'
    }

    if ((to.path === '/login' || to.path === '/register') && isLogin) {
        console.log("用户已登录")
        alert("用户已登录")
        return '/home'
    }
    return
})

export default router


