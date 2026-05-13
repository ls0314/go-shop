import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { RouteRecordRaw } from 'vue-router'
import router from '@/router'
import { asyncMenu } from '@/api/menu'

// 后端返回菜单格式
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

// layout组件
const Layout = () => import('@/views/layout/index.vue')
// ParentView组件
const ParentView = () => import('@/components/ParentView.vue')

// 动态导入 views 目录下的所有 .vue 页面
const modules = import.meta.glob('../../views/**/*.vue')
// 根据后端返回的 component 字符串，找到对应的前端组件
function loadView(component?: string) {
    // 如果后端没有返回 component，默认使用 ParentView
    if (!component) {
        return ParentView
    }
    // 特殊组件：Layout
    if (component === 'Layout') {
        return Layout
    }
    // 特殊组件：ParentView
    if (component === 'ParentView') {
        return ParentView
    }
    // 普通业务页面组件
    const path = `../../views/${component}.vue`
    const view = modules[path]

    if (!view) {
        console.warn(`动态路由组件不存在: ${path}`)
        return ParentView
    }

    return view
}

// 把单个菜单节点转换成 Vue Router 路由对象
function menuToRoute(menu: MenuItem): RouteRecordRaw | null {
    // 按钮权限不生成路由
    if (menu.menu_type === 'F') {
        return null
    }
    // 创建 Vue Router 路由对象
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
    // 递归处理子菜单
    const children = menu.children || []

    // 递归把子菜单转换成子路由
    if (children.length > 0) {
        route.children = children
            .map(menuToRoute)
            .filter(Boolean) as RouteRecordRaw[]
    }
    return route
}

// 把后端返回的整个菜单树转换成动态路由数组
function generateRoutes(menus: MenuItem[]) {
    return menus
        .map(menuToRoute)
        .filter(Boolean) as RouteRecordRaw[]
}

export const useRouterStore = defineStore('router', () => {
    // 后端返回的菜单树
    const menuList = ref<MenuItem[]>([])
    // 由菜单树转换出来的 Vue Router 动态路由
    const asyncRouter = ref<RouteRecordRaw[]>([])
    // 动态路由是否已经初始化,防止重复请求菜单、重复 addRoute
    const isInitRouter = ref(false)
    // 保存这些删除函数，退出登录时可以把动态路由移除
    const removeRouteFns = ref<Array<() => void>>([])

    // 初始化动态路由
    async function SetAsyncRouter(userId?: string) {
        // 如果已经初始化过动态路由，就不重复执行
        if (isInitRouter.value) {
            return
        }

        // 请求后端菜单树
        const res = await asyncMenu(userId)

        // 判断后端响应状态
        if (res.data.code !== 200) {
            throw new Error(res.data.message || '获取菜单失败')
        }

        // 保存请求得到的菜单树
        menuList.value = res.data.data || []

        // 把菜单树转换成 Vue Router 路由数组
        asyncRouter.value = generateRoutes(menuList.value)

        // 把动态路由逐个添加到 Vue Router 中
        asyncRouter.value.forEach(route => {
            // 路由必须有 name，且不能重复添加。
            if (route.name && !router.hasRoute(route.name)) {
                // 动态添加路由
                const removeRoute = router.addRoute(route)
                // 保存删除函数，退出登录时使用
                removeRouteFns.value.push(removeRoute)
            }
        })
        // 标记动态路由已经初始化完成。
        isInitRouter.value = true
    }
    // 重置动态路由
    function ResetAsyncRouter() {
        // 移除所有动态添加的路由
        removeRouteFns.value.forEach(remove => remove())
        // 清空删除函数列表
        removeRouteFns.value = []
        // 清空菜单树
        menuList.value = []
        // 清空动态路由数组
        asyncRouter.value = []
        // 标记动态路由未初始化
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