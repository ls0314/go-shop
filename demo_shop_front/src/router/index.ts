import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/pinia/modules/user'
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

router.beforeEach((to) => {
    const userStore = useUserStore()

    const isLogin = !!userStore.userToken.access_token

    console.log('路由跳转 →', to.path, '登录状态:', isLogin)

    if (to.meta.requiresAuth && !isLogin) {
        console.log('请登录')
        return '/login'
    }

    if ((to.path === '/login' || to.path === '/register') && isLogin) {
        console.log('用户已登录')
        return '/home'
    }
})

export default router