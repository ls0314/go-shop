import service from '@/utils/request'

export const asyncMenu = (data) => {
    return service({
        url: '/menu/tree',
        method: 'post',
        data: data
    })
}