import axios from 'axios'
import type {
    AxiosError,
    InternalAxiosRequestConfig,
    AxiosResponse
} from 'axios'
import { getUserStore } from '../store'
import router from "../router";

const request = axios.create({
    baseURL: "http://localhost:9001",
    timeout: 5000,
})

let isRefreshing = false
let requestsQueue: ((token: string) => void)[] = []

// 请求拦截
request.interceptors.request.use((config: InternalAxiosRequestConfig) => {
    const excludePaths = ['/api/v1/user/login','/api/v1/user/refresh', '/api/v1/user/register']
    if (config.url && excludePaths.includes(config.url)) {
        return config
    }

    const userStore = getUserStore()


    if (userStore.access_token) {
        config.headers.Authorization = `Bearer ${userStore.access_token}`
    } else {
        console.log('没有 Token！')
    }

    return config
})

// 响应拦截
request.interceptors.response.use(
    (response: AxiosResponse) => response,
    async (error: AxiosError) => {
        const userStore = getUserStore()
        const originalRequest: any = error.config

        if (error.response?.status === 401) {
            if (originalRequest._retry) {
                userStore.clearTokens()
                router.push('/login')
                return Promise.reject(error)
            }

            originalRequest._retry = true

            if (isRefreshing) {
                return new Promise((resolve) => {
                    requestsQueue.push((token: string) => {
                        originalRequest.headers.Authorization = `Bearer ${token}`
                        resolve(request(originalRequest))
                    })
                })
            }

            isRefreshing = true

            try {
                const res = await request.post('/api/v1/user/refresh', {
                    refresh_token: userStore.refresh_token
                })

                const newAccessToken = res.data.data.access_token
                userStore.setAccessToken(newAccessToken)

                requestsQueue.forEach(cb => cb(newAccessToken))
                requestsQueue = []

                originalRequest.headers.Authorization = `Bearer ${newAccessToken}`
                return request(originalRequest)

            } catch (err) {
                userStore.clearTokens()
                router.push('/login')
                return Promise.reject(err)
            } finally {
                isRefreshing = false
            }
        }

        return Promise.reject(error)
    }
)

export default request