import axios from 'axios'

const request = axios.create({
    baseURL: 'http://localhost:9001',
    timeout: 5000,
    headers: {
        'Content-Type': 'application/json'
    }
})

request.interceptors.request.use(
    config => {
        return config
    },
    error => Promise.reject(error)
)

request.interceptors.response.use(
    response => response.data,
    error => {
        console.error('请求错误:', error)
        return Promise.reject(error)
    }
)

export default request