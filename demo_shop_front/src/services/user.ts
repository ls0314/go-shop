import request from './request.ts'

// 注册接口
export const registerApi = (data) => {
    return request.post('/api/v1/user/register', data)
}
export const loginApi = (from) => {
    return request.post('/api/v1/user/login', from)
}
