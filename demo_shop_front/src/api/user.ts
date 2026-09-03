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

// 获取当前登录用户的权限码数组(按钮级权限控制)
export const getPermsApi = () => {
    return service({
        url: '/api/v1/user/perms',
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
