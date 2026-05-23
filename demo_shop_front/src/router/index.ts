import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/pinia/modules/user'
import { useRouterStore } from '@/pinia/modules/router'
import Layout from '@/views/layout/index.vue'

const routes = [
    {
        path: '/',
        component: Layout,
        redirect: '/home',
        children: [
            {
                path: 'home',
                name: 'Home',
                component: () => import('@/views/Home.vue'),
                meta: {
                    requiresAuth: true,
                },
            },
            {
                path: 'about',
                name: 'About',
                component: () => import('@/views/About.vue'),
                meta: {
                    requiresAuth: true,
                },
            },
        ],
    },
    {
        path: '/register',
        name: 'Register',
        component: () => import('@/views/register.vue'),
        meta: {
            hideNav: true,
        },
    },
    {
        path: '/login',
        name: 'Login',
        component: () => import('@/views/login.vue'),
        meta: {
            hideNav: true,
        },
    },
]

const router = createRouter({
    history: createWebHistory(),
    routes,
})

router.beforeEach(async (to, from) => {
    const userStore = useUserStore()
    const routerStore = useRouterStore()
    const isLogin = !!userStore.userToken.access_token

    // 白名单：登录、注册页 直接放行，不拦截
    if (to.path === '/login' || to.path === '/register') {
        // 已登录 → 去首页
        if (isLogin) {
            return '/home'
        }
        // 未登录 → 允许访问登录页
        return true
    }

    // 未登录 + 访问非白名单 → 跳登录
    if (!isLogin) {
        console.log('请登录')
        return '/login'
    }

    // 已登录，处理动态路由
    if (!routerStore.isInitRouter) {
        try {
            await routerStore.SetAsyncRouter({
                user_id: userStore.userInfo.user_id, // 你原来写的 userInfo.value 是错的！
            })
            // 重新导航，让动态路由生效
            return { ...to, replace: true }
        } catch {
            return '/login'
        }
    }

    console.log('路由跳转 →', to.path, '登录状态:', isLogin)
})

export default router