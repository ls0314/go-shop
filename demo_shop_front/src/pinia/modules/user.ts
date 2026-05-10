import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { loginApi, getUserInfoApi } from '@/api/user'
import router from '@/router/index'
import { ElMessage } from 'element-plus'
import {useRouter} from "vue-router";

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
        userName: '',
    })

    const isLoggedIn = computed(() => !!userToken.value.access_token)

    function setUserInfo(val:UserInfo){
        userInfo.value = val
    }

    function setToken(val:UserToken) {
        userToken.value = val
    }

    // 获取用户信息
    async function GetUserInfo(){
        try {
            const res: ApiResponse<UserInfo> = await getUserInfoApi()
            console.log(res.data)
            if (res.data.code === 200) {
                setUserInfo(res.data.data)
                return res
            }

            ElMessage.error(res.data.message || '获取用户信息失败')
            return false
        } catch (err: any) {
            console.error('获取用户信息失败', err)
            ElMessage.error('获取用户信息失败')
            return false
        }
    }

    // 登录
    async function LoginIn(loginInfo:LoginForm) {
        try {
            const res: ApiResponse<UserToken> = await loginApi(loginInfo)

            // console.log('登录接口返回 res:', res)
            // console.log('res.code:', res.data.code)
            // console.log('res.data:', res.data)
            if (res.data.code !== 200) {
                console.log(res.data.message) //
                return false
            }
            setToken(res.data.data)
            await GetUserInfo()

            ElMessage.success('登录成功')
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
        // 清空本地状态
        userToken.value = { access_token: '', refresh_token: '' }
        userInfo.value = { user_id: '', username: '' }
        // 跳转到登录页
        router.push('/login')
        ElMessage.success('已退出登录')
    }

        return {
            userToken,
            userInfo,
            isLoggedIn,
            LoginIn,
            GetUserInfo,
            Logout,
            setToken,
            setUserInfo
        }
},
    {
        persist: {
            key: 'user',
            storage: sessionStorage,
            paths: ['userToken']
        } as any
    }
)