import request from './request.ts'
import authRequest from '../utils/request.ts'


export const registerApi = (data) => {
    return request.post('/api/v1/user/register', data)
}

export const loginApi = (from) => {
    return request.post('/api/v1/user/login', from)
}

export const getUserInfoApi = () => {
    return authRequest.get('/api/v1/user/info')
}