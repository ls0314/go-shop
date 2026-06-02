import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/pinia/modules/user'
import { useRouterStore } from '@/pinia/modules/router'
import { ElMessage } from 'element-plus'
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
    {
        path: '/404',
        name: 'NotFound',
        component: () => import('@/views/error/404.vue'),
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

    // 404 页面直接放行
    if (to.name === 'NotFound') {
        return true
    }

    // 白名单：登录、注册页 直接放行
    if (to.path === '/login' || to.path === '/register') {
        if (isLogin) {
            return '/home'
        }
        return true
    }

    // 未登录 → 跳登录
    if (!isLogin) {
        return '/login'
    }

    // 已登录，处理动态路由
    if (!routerStore.isInitRouter) {
        try {
            await routerStore.SetAsyncRouter({
                user_id: userStore.userInfo.user_id,
            })
            return { ...to, replace: true }
        } catch (err: any) {
            // 区分网络错误和权限错误
            if (err?.code === 'ERR_NETWORK' || err?.message?.includes('Network Error')) {
                ElMessage.error('网络异常，获取菜单失败，请检查网络后刷新重试')
                return false
            }
            // 无权限或其他错误 → 清除登录态跳登录
            userStore.setToken({ access_token: '', refresh_token: '' })
            routerStore.ResetAsyncRouter()
            return '/login'
        }
    }

    // 动态路由已加载，但匹配不到任何路由 → 跳 404
    if (to.matched.length === 0) {
        return '/404'
    }
})

export default router