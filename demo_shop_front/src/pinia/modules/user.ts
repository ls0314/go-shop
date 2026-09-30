import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { loginApi, getUserInfoApi, getPermsApi } from '@/api/user'
import router from '@/router/index'
import { ElMessage } from 'element-plus'
import {useRouter} from "vue-router";
import {useRouterStore} from "@/pinia/modules/router";

interface UserToken {
    access_token: string
    refresh_token: string
}

interface UserInfo {
    user_id: string
    username: string
}

interface LoginForm {
    username: string
    password: string
}




export const useUserStore = defineStore('user', () => {
    const userToken = ref<UserToken>({
        access_token: "",
        refresh_token: ""
    })
    const userInfo = ref<UserInfo>({
        user_id: '',
        username: '',
    })
    // 当前用户权限码集合(按钮级权限控制用)
    const permCodes = ref<string[]>([])

    const isLoggedIn = computed(() => !!userToken.value.access_token)

    // setUserInfo 设置用户信息
    function setUserInfo(val:UserInfo){
        userInfo.value = val
    }
    // setToken设置Token信息
    function setToken(val:UserToken) {
        userToken.value = val
    }

    // hasPerm 判断用户是否持有某权限码(支持多个任一)
    function hasPerm(...codes: string[]): boolean {
        if (codes.length === 0) return true
        return codes.some(code => permCodes.value.includes(code))
    }

    // GetPerms 拉取当前用户权限码(登录后调用;按钮级权限控制数据源)
    async function GetPerms() {
        try {
            const res: any = await getPermsApi()
            if (res.data.code === 200) {
                permCodes.value = res.data.data?.perms || []
                return true
            }
            return false
        } catch (err: any) {
            console.error('获取用户权限失败', err)
            return false
        }
    }

    // GetUserInfo 获取用户信息
    async function GetUserInfo(){
        try {
            // 调用API从后端获取用户信息
            const res: ApiResponse<UserInfo> = await getUserInfoApi()
            console.log(res.data)
            //  保存后端传回的用户信息
            if (res.data.code === 200) {
                setUserInfo(res.data.data)
                //  返回用户信息
                return true
            }

            ElMessage.error(res.data.message || '获取用户信息失败')
            return false
        } catch (err: any) {
            console.error('获取用户信息失败', err)
            ElMessage.error('获取用户信息失败')
            return false
        }
    }

    // LoginIn 登录
    async function LoginIn(loginInfo:LoginForm) {
        try {
            //  调用API执行登录操作并获取后端传回的Token信息
            const res: ApiResponse<UserToken> = await loginApi(loginInfo)
            if (res.data.code !== 200) {
                console.log(res.data.message) //
                return false
            }
            //  成功调用后保存Token信息
            setToken(res.data.data)

            //  获取并设置用户信息
            await GetUserInfo()
            //  拉取用户权限码(按钮级权限控制)
            await GetPerms()

            // 初始化路由(菜单树按当前登录用户返回)
            const routerStore = useRouterStore()
            await routerStore.SetAsyncRouter()

            ElMessage.success('登录成功')
            // 登录成功后跳转到首页
            router.push('/home')

            return true
        } catch (error: any) {
            console.error('登录请求失败:', error)
            if (error.response && error.response.data) {
                ElMessage.error(error.response.data.Message || '登录失败')
            } else {
                ElMessage.error('网络错误或服务器未响应')
            }
            return false
        }
    }

    function Logout() {
        const routerStore = useRouterStore()
        // 清空本地状态
        userToken.value = { access_token: '', refresh_token: '' }
        userInfo.value = { user_id: '', username: '' }
        permCodes.value = []
        // 清空动态路由
        routerStore.ResetAsyncRouter()
        // 跳转到登录页
        router.push('/login')
        ElMessage.success('已退出登录')
    }

        return {
            userToken,
            userInfo,
            permCodes,
            isLoggedIn,
            LoginIn,
            GetUserInfo,
            GetPerms,
            hasPerm,
            Logout,
            setToken,
            setUserInfo
        }
},
    {
        persist: {
            key: 'user',
            storage: sessionStorage,
            paths: ['userToken', 'userInfo']
        } as any
    }
)