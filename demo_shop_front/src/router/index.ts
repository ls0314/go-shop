import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/pinia/modules/user'
import { useRouterStore } from '@/pinia/modules/router'
import { ElMessage } from 'element-plus'
import Layout from '@/views/layout/index.vue'

const routes = [
    // ==================== 管理端路由（需登录） ====================
    {
        path: '/',
        component: Layout,
        redirect: '/home',
        children: [
            {
                path: 'home',
                name: 'Home',
                component: () => import('@/views/Home.vue'),
                meta: { requiresAuth: true },
            },
            {
                path: 'about',
                name: 'About',
                component: () => import('@/views/About.vue'),
                meta: { requiresAuth: true },
            },
            {
                path: 'platform/product/create/:id?',
                name: 'ProductCreate',
                component: () => import('@/views/platform/product/create/index.vue'),
                meta: { requiresAuth: true },
            },
        ],
    },
    // ==================== 用户端路由（无需登录） ====================
    {
        path: '/shop',
        component: () => import('@/views/userLayout/index.vue'),
        redirect: '/shop/home',
        children: [
            {
                path: 'home',
                name: 'ShopHome',
                component: () => import('@/views/shop/home/index.vue'),
            },
            {
                path: 'product/list',
                name: 'ShopProductList',
                component: () => import('@/views/shop/product/list/index.vue'),
            },
            {
                path: 'product/:id',
                name: 'ShopProductDetail',
                component: () => import('@/views/shop/product/detail/index.vue'),
            },
            {
                path: 'cart',
                name: 'ShopCart',
                component: () => import('@/views/shop/cart/index.vue'),
            },
        ],
    },
    // ==================== 公共路由 ====================
    {
        path: '/register',
        name: 'Register',
        component: () => import('@/views/register.vue'),
        meta: { hideNav: true },
    },
    {
        path: '/login',
        name: 'Login',
        component: () => import('@/views/login.vue'),
        meta: { hideNav: true },
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

// 无需登录即可访问的路径前缀
const PUBLIC_PREFIXES = ['/shop', '/login', '/register']

router.beforeEach(async (to, from) => {
    const userStore = useUserStore()
    const routerStore = useRouterStore()
    const isLogin = !!userStore.userToken.access_token

    // 404 直接放行
    if (to.name === 'NotFound') return true

    // 是否公开路径（商城、登录、注册）
    const isPublic = PUBLIC_PREFIXES.some(p => to.path.startsWith(p))

    // 登录/注册页：已登录跳商城首页
    if (to.path === '/login' || to.path === '/register') {
        if (isLogin) return '/shop/home'
        return true
    }

    // 用户端商城路由：无需登录，直接放行
    if (to.path.startsWith('/shop')) {
        return true
    }

    // ===== 以下为管理端路由，需要登录 =====

    // 未登录 → 跳商城首页
    if (!isLogin) {
        return '/shop/home'
    }

    // 已登录，加载动态菜单路由
    if (!routerStore.isInitRouter) {
        try {
            await routerStore.SetAsyncRouter({
                user_id: userStore.userInfo.user_id,
            })
            // 无菜单用户（非管理员） → 跳商城首页
            if (!routerStore.hasAdmin) {
                return '/shop/home'
            }
            return { ...to, replace: true }
        } catch (err: any) {
            if (err?.code === 'ERR_NETWORK' || err?.message?.includes('Network Error')) {
                ElMessage.error('网络异常，获取菜单失败，请检查网络后刷新重试')
                return false
            }
            userStore.setToken({ access_token: '', refresh_token: '' })
            routerStore.ResetAsyncRouter()
            return '/login'
        }
    }

    // 已登录但无管理端权限 → 跳商城首页
    if (!routerStore.hasAdmin) {
        return '/shop/home'
    }

    // 动态路由已加载但匹配不到 → 404
    if (to.matched.length === 0) {
        return '/404'
    }
})

export default router