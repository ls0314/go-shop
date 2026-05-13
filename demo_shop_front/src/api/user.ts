import service from '@/utils/request'

export const loginApi = (data) => {
    return service({
        url: '/api/v1/user/login',
        method: 'post',
        data: data
    })
}

export const getUserInfoApi = () => {
    return service({
        url: '/api/v1/user/info',
        method: 'get'
    })
}

export const registerApi = (data) => {
    return service({
        url: '/api/v1/user/register',
        method: 'post',
        data: data
    })
}
