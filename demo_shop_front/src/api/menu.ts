import service from '@/utils/request'

export const asyncMenu = (data) => {
    return service({
        url: '/api/v1/menu/tree',
        method: 'post',
        data: data
    })
}