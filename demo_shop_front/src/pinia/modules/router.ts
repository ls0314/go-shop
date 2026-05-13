import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { RouteRecordRaw } from 'vue-router'
import router from '@/router'
import { asyncMenu } from '@/api/menu'

interface MenuItem {
    menu_id: number
    parent_id: number
    menu_name: string
    menu_type: 'M' | 'C' | 'F'
    icon?: string
    route_path: string
    component?: string
    is_visible: boolean
    is_cache: boolean
    sort_order: number
    meta_info?: {
        icon?: string
        noCache?: boolean
        title?: string
    }
    created_at?: string
    children?: MenuItem[] | null
}

const Layout = () => import('@/views/layout/index.vue')
const ParentView = () => import('@/components/ParentView.vue')

// 当前文件在 src/pinia/modules/router.ts
// ../../views/**/*.vue 对应 src/views/**/*.vue
const modules = import.meta.glob('../../views/**/*.vue')

function loadView(component?: string) {
    if (!component) {
        return ParentView
    }

    if (component === 'Layout') {
        return Layout
    }

    if (component === 'ParentView') {
        return ParentView
    }

    const path = `../../views/${component}.vue`
    const view = modules[path]

    if (!view) {
        console.warn(`动态路由组件不存在: ${path}`)
        return ParentView
    }

    return view
}

function menuToRoute(menu: MenuItem): RouteRecordRaw | null {
    // 按钮权限不生成路由
    if (menu.menu_type === 'F') {
        return null
    }

    const route: RouteRecordRaw = {
        path: menu.route_path,
        name: `Menu_${menu.menu_id}`,
        component: loadView(menu.component),
        meta: {
            title: menu.meta_info?.title || menu.menu_name,
            icon: menu.meta_info?.icon || menu.icon,
            noCache: menu.meta_info?.noCache ?? !menu.is_cache,
            hidden: !menu.is_visible,
            menuId: menu.menu_id
        }
    }

    const children = menu.children || []

    if (children.length > 0) {
        route.children = children
            .map(menuToRoute)
            .filter(Boolean) as RouteRecordRaw[]
    }

    return route
}

function generateRoutes(menus: MenuItem[]) {
    return menus
        .map(menuToRoute)
        .filter(Boolean) as RouteRecordRaw[]
}

export const useRouterStore = defineStore('router', () => {
    const menuList = ref<MenuItem[]>([])
    const asyncRouter = ref<RouteRecordRaw[]>([])
    const isInitRouter = ref(false)

    const removeRouteFns = ref<Array<() => void>>([])

    async function SetAsyncRouter(userId?: string) {
        if (isInitRouter.value) {
            return
        }

        const res = await asyncMenu(userId)

        if (res.data.code !== 200) {
            throw new Error(res.data.message || '获取菜单失败')
        }

        menuList.value = res.data.data || []

        asyncRouter.value = generateRoutes(menuList.value)

        asyncRouter.value.forEach(route => {
            if (route.name && !router.hasRoute(route.name)) {
                const removeRoute = router.addRoute(route)
                removeRouteFns.value.push(removeRoute)
            }
        })

        isInitRouter.value = true
    }

    function ResetAsyncRouter() {
        removeRouteFns.value.forEach(remove => remove())

        removeRouteFns.value = []
        menuList.value = []
        asyncRouter.value = []
        isInitRouter.value = false
    }

    return {
        menuList,
        asyncRouter,
        isInitRouter,
        SetAsyncRouter,
        ResetAsyncRouter
    }
})